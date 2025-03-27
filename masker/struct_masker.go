// Package masker provides functionality to recursively mask struct fields based on tags.
package masker

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

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
//	masked := MaskStruct(original).(MyStruct)
//	// masked => MyStruct{Name: "***** ***", Age: 30, Phone: "123456####"}
func MaskStruct(v interface{}) interface{} {
	return maskValue(reflect.ValueOf(v)).Interface()
}

// maskValue creates a masked copy of the reflect.Value, handling both structs and pointers.
func maskValue(v reflect.Value) reflect.Value {
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
			newStruct.Field(i).Set(maskField(field, maskTag, maskCharTag))
		} else if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct) {
			if field.Kind() == reflect.Ptr {
				if !field.IsNil() {
					newField := reflect.New(field.Type().Elem())
					newField.Elem().Set(maskValue(field.Elem()))
					newStruct.Field(i).Set(newField)
				}
			} else {
				newStruct.Field(i).Set(maskValue(field))
			}
		} else {
			newStruct.Field(i).Set(field)
		}
	}

	return newStruct
}

// maskField creates a masked copy of a single field based on its type and tag instructions.
func maskField(field reflect.Value, maskTag, maskCharTag string) reflect.Value {
	if maskCharTag == "" {
		maskCharTag = "*"
	}

	switch field.Kind() {
	case reflect.String:
		tagParts := strings.Split(maskTag, ",")
		method := tagParts[0]
		switch method {
		case "all":
			return reflect.ValueOf(MaskStringAll(field.String(), maskCharTag))
		case "regex":
			if len(tagParts) > 1 {
				return reflect.ValueOf(MaskStringRegex(field.String(), tagParts[1], maskCharTag))
			}
		case "first":
			if len(tagParts) > 1 {
				if n, err := strconv.Atoi(tagParts[1]); err == nil {
					return reflect.ValueOf(MaskStringFirst(field.String(), n, maskCharTag))
				}
			} else {
				return reflect.ValueOf(MaskStringFirst(field.String(), 1, maskCharTag))
			}
		case "last":
			if len(tagParts) > 1 {
				if n, err := strconv.Atoi(tagParts[1]); err == nil {
					return reflect.ValueOf(MaskStringLast(field.String(), n, maskCharTag))
				}
			} else {
				return reflect.ValueOf(MaskStringLast(field.String(), 1, maskCharTag))
			}
		case "corners":
			if len(tagParts) > 1 {
				cornerParts := strings.Split(tagParts[1], "-")
				if len(cornerParts) == 2 {
					if first, err1 := strconv.Atoi(cornerParts[0]); err1 == nil {
						if last, err2 := strconv.Atoi(cornerParts[1]); err2 == nil {
							return reflect.ValueOf(MaskStringCorners(field.String(), first, last, maskCharTag))
						}
					}
				}
			} else {
				return reflect.ValueOf(MaskStringCorners(field.String(), 1, 1, maskCharTag))
			}
		case "between":
			if len(tagParts) > 1 {
				cornerParts := strings.Split(tagParts[1], "-")
				if len(cornerParts) == 2 {
					if first, err1 := strconv.Atoi(cornerParts[0]); err1 == nil {
						if last, err2 := strconv.Atoi(cornerParts[1]); err2 == nil {
							return reflect.ValueOf(MaskAllExceptCorners(field.String(), first, last, maskCharTag))
						}
					}
				}
			} else {
				return reflect.ValueOf(MaskAllExceptCorners(field.String(), 1, 1, maskCharTag))
			}
		}

		return field

	default:
		return field
	}
}

// MaskStringAll masks all characters in the string.
func MaskStringAll(s, maskChar string) string {
	return strings.Repeat(maskChar, len(s))
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

// MaskStringFirst masks the first n characters in the string.
func MaskStringFirst(s string, n int, maskChar string) string {
	if len(s) <= n {
		return strings.Repeat(maskChar, len(s))
	}
	return strings.Repeat(maskChar, n) + s[n:]
}

// MaskStringLast masks the last n characters in the string.
func MaskStringLast(s string, n int, maskChar string) string {
	if len(s) <= n {
		return strings.Repeat(maskChar, len(s))
	}
	return s[:len(s)-n] + strings.Repeat(maskChar, n)
}

// MaskStringCorners masks the first n and last m characters in the string.
func MaskStringCorners(s string, n, m int, maskChar string) string {
	if len(s) <= n+m {
		return strings.Repeat(maskChar, len(s))
	}
	return strings.Repeat(maskChar, n) + s[n:len(s)-m] + strings.Repeat(maskChar, m)
}

// MaskAllExceptCorners  masks all except the first n and last m characters in the string.
func MaskAllExceptCorners(s string, n, m int, maskChar string) string {
	if len(s) <= n+m {
		return strings.Repeat(maskChar, len(s))
	}

	maskedPart := strings.Repeat(maskChar, len(s)-n-m)
	return s[:n] + maskedPart + s[len(s)-m:]
}
