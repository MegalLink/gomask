// Package masker provides functionality to recursively mask struct fields based on tags.
package masker

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type MaskerManager struct {
	maskerRegistry     map[string]Masker
	maskerRegistryLock sync.RWMutex
}

// Masker defines the interface for all masking strategies
type Masker interface {
	Mask(value string, maskChar string, tags []string) reflect.Value
}

// RegisterMasker registers a new masking strategy with the given name
func (m *MaskerManager) RegisterMasker(name string, masker Masker) {
	m.maskerRegistryLock.Lock()
	defer m.maskerRegistryLock.Unlock()
	m.maskerRegistry[name] = masker
}

// GetMasker retrieves a masking strategy by name
func (m *MaskerManager) GetMasker(name string) (Masker, error) {
	m.maskerRegistryLock.RLock()
	defer m.maskerRegistryLock.RUnlock()

	masker, exists := m.maskerRegistry[name]
	if !exists {
		return nil, fmt.Errorf("masker %s not registered", name)
	}
	return masker, nil
}

// MaskStruct recursively creates a masked copy of the struct tagged with "mask".
// It traverses the struct fields and applies masking based on the tags specified.
// It allows child struct directly or pointers also.
// Supported masking methods:
//   - all: Masks all characters in a string.
//   - regex: Masks characters based on a regular expression pattern.
//   - first: Masks the first n characters in a string.
//   - last: Masks the last n characters in a string.
//   - corners: Masks the first n and last m characters in a string separated by "-" example Phone string `mask:"corners,4-5"`.
//   - between: Masks all except the first n and last m characters in a string separated by "-" example Phone string `mask:"between,4-5"`.
//
// Supported configurations:
//   - mask: Specifies the masking method and options. Format: "mask:<method>,<options>".
//   - maskTag: Specifies the character used for masking. Default is "*".
//
// Example usage:
//
//	type MyStruct struct {
//	    Name string `mask:"regex,\\b[A-Za-z]+\\b"` // Masks all characters except spaces
//	    Age int
//	    Phone string `mask:"last,4" maskTag:"#"`  // Masks last 4 digits with #
//	}
//	original := MyStruct{Name: "John Doe", Age: 30, Phone: "1234567890"}
//	masker := NewMasker()
//	masked := masker.MaskStruct(original).(MyStruct)
//	// masked => MyStruct{Name: "***** ***", Age: 30, Phone: "123456####"}

func NewMasker() *MaskerManager {
	maskerManager := &MaskerManager{
		maskerRegistry:     make(map[string]Masker),
		maskerRegistryLock: sync.RWMutex{},
	}

	maskerManager.RegisterMasker("all", &MaskAll{})
	maskerManager.RegisterMasker("regex", &MaskRegex{})
	maskerManager.RegisterMasker("first", &MaskFirst{})
	maskerManager.RegisterMasker("last", &MaskLast{})
	maskerManager.RegisterMasker("corners", &MaskCorners{})
	maskerManager.RegisterMasker("between", &MaskBetween{})

	return maskerManager
}

// MaskStruct is a convenience function that uses the default masker
func (m *MaskerManager) MaskStruct(v interface{}) interface{} {
	return m.maskValue(reflect.ValueOf(v)).Interface()
}

// maskValue creates a masked copy of the reflect.Value, handling both structs and pointers.
func (m *MaskerManager) maskValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Create a new instance of the struct
	newStruct := reflect.New(v.Type()).Elem()

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		maskTag := fieldType.Tag.Get("mask")
		maskCharTag := fieldType.Tag.Get("maskTag")

		if maskTag != "" {
			newStruct.Field(i).Set(m.maskField(field, maskTag, maskCharTag))
		} else if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct) {
			if field.Kind() == reflect.Ptr {
				if !field.IsNil() {
					newField := reflect.New(field.Type().Elem())
					newField.Elem().Set(m.maskValue(field.Elem()))
					newStruct.Field(i).Set(newField)
				}
			} else {
				newStruct.Field(i).Set(m.maskValue(field))
			}
		} else {
			newStruct.Field(i).Set(field)
		}
	}

	return newStruct
}

// maskField method for MaskerManager
func (m *MaskerManager) maskField(field reflect.Value, maskTag, maskCharTag string) reflect.Value {
	if maskCharTag == "" {
		maskCharTag = "*"
	}

	switch field.Kind() {
	case reflect.String:
		tagParts := strings.Split(maskTag, ",")
		method := tagParts[0]

		masker, err := m.GetMasker(method)
		if err == nil {
			return masker.Mask(field.String(), maskCharTag, tagParts)
		}
		// If masker not found, return original field
		return field

	default:
		return field
	}
}

type MaskAll struct{}

func (m *MaskAll) Mask(value string, maskChar string, tags []string) reflect.Value {
	return reflect.ValueOf(MaskStringAll(value, maskChar))
}

// MaskStringAll masks all characters in the string.
func MaskStringAll(s, maskChar string) string {
	return strings.Repeat(maskChar, len(s))
}

type MaskRegex struct{}

func (m *MaskRegex) Mask(value string, maskChar string, tags []string) reflect.Value {
	if len(tags) > 1 {
		return reflect.ValueOf(MaskStringRegex(value, tags[1], maskChar))
	}

	return reflect.ValueOf(value)
}

// MaskStringRegex applies the regex-based masking to a string.
func MaskStringRegex(s, regex, maskChar string) string {
	re, err := regexp.Compile(regex)
	if err != nil {
		// If the regex is invalid, return the original string
		return s
	}
	return re.ReplaceAllStringFunc(s, func(m string) string {
		return strings.Repeat(maskChar, len(m))
	})
}

type MaskFirst struct{}

func (m *MaskFirst) Mask(value string, maskChar string, tags []string) reflect.Value {
	if len(tags) > 1 {
		if n, err := strconv.Atoi(tags[1]); err == nil {
			return reflect.ValueOf(MaskStringFirst(value, n, maskChar))
		}
	}

	return reflect.ValueOf(MaskStringFirst(value, 1, maskChar))
}

// MaskStringFirst masks the first n characters in the string.
func MaskStringFirst(s string, n int, maskChar string) string {
	if len(s) <= n {
		return strings.Repeat(maskChar, len(s))
	}
	return strings.Repeat(maskChar, n) + s[n:]
}

type MaskLast struct{}

func (m *MaskLast) Mask(value string, maskChar string, tags []string) reflect.Value {
	if len(tags) > 1 {
		if n, err := strconv.Atoi(tags[1]); err == nil {
			return reflect.ValueOf(MaskStringLast(value, n, maskChar))
		}
	}

	return reflect.ValueOf(MaskStringLast(value, 1, maskChar))
}

// MaskStringLast masks the last n characters in the string.
func MaskStringLast(s string, n int, maskChar string) string {
	if len(s) <= n {
		return strings.Repeat(maskChar, len(s))
	}
	return s[:len(s)-n] + strings.Repeat(maskChar, n)
}

type MaskCorners struct{}

func (m *MaskCorners) Mask(value string, maskChar string, tags []string) reflect.Value {
	if len(tags) > 1 {
		cornerParts := strings.Split(tags[1], "-")
		if len(cornerParts) == 2 {
			if first, err1 := strconv.Atoi(cornerParts[0]); err1 == nil {
				if last, err2 := strconv.Atoi(cornerParts[1]); err2 == nil {
					return reflect.ValueOf(MaskStringCorners(value, first, last, maskChar))
				}
			}
		}
	}

	return reflect.ValueOf(MaskStringCorners(value, 1, 1, maskChar))
}

// MaskStringCorners masks the first n and last m characters in the string.
func MaskStringCorners(s string, n, m int, maskChar string) string {
	if len(s) <= n+m {
		return strings.Repeat(maskChar, len(s))
	}
	return strings.Repeat(maskChar, n) + s[n:len(s)-m] + strings.Repeat(maskChar, m)
}

type MaskBetween struct{}

func (m *MaskBetween) Mask(value string, maskChar string, tags []string) reflect.Value {
	if len(tags) > 1 {
		betweenParts := strings.Split(tags[1], "-")
		if len(betweenParts) == 2 {
			if first, err1 := strconv.Atoi(betweenParts[0]); err1 == nil {
				if last, err2 := strconv.Atoi(betweenParts[1]); err2 == nil {
					return reflect.ValueOf(MaskAllExceptCorners(value, first, last, maskChar))
				}
			}
		}
	}

	return reflect.ValueOf(MaskAllExceptCorners(value, 1, 1, maskChar))
}

// MaskAllExceptCorners  masks all except the first n and last m characters in the string.
func MaskAllExceptCorners(s string, n, m int, maskChar string) string {
	if len(s) <= n+m {
		return s
	}
	return s[:n] + strings.Repeat(maskChar, len(s)-n-m) + s[len(s)-m:]
}
