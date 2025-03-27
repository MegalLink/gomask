// Package masker provides functionality to recursively mask struct fields based on tags.
package masker

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ExampleStruct struct {
	Name        string `mask:"regex,\\b[A-Za-z]+\\b"` // Mask all characters not including spaces
	Age         int    `mask:"all"`                   // it should not be masked
	Address     NestedStruct
	DogName     string `mask:"between"`                  // mask all except first and last characters
	DogLastName string `mask:"between,2-3"`              // mask all except first 2 and 3 last characters
	Email       string `mask:"regex,^[^@]+" maskTag:"X"` // mask with regex and maskTag should be X instead of *
}

type NestedStruct struct {
	City      string             `mask:"all" json:"city"`         // Mask all characters
	State     string             `mask:"regex,\\w+" json:"state"` // Mask all words
	Phone     string             `mask:"last,3" json:"phone"`     // mask last 3 digits
	Cellphone string             `mask:"first" json:"cellphone"`  // mask first 1 digit
	Street    string             `mask:"first,5" json:"street"`   // mask last 5 digits
	Country   string             `mask:"last" json:"country"`
	Child     *ChildNestedStruct `json:"child"`
}

type ChildNestedStruct struct {
	CreditCard string `mask:"corners,5-4" json:"credit_card"` // Mask first 5 digits and last 4 digits
	CVV        string `mask:"all" maskTag:"+" json:"cvv"`     //Mask all and mask tag should be +
}

func TestMaskStruct(t *testing.T) {
	example := &ExampleStruct{
		Name:        "Jeferson Narvae",
		Age:         30,
		DogName:     "Firulais",
		DogLastName: "Wolfenstein",
		Address: NestedStruct{
			City:      "New York",
			State:     "NY",
			Phone:     "2999999",
			Cellphone: "0998695861",
			Street:    "Floresta",
			Country:   "Ecuador",
			Child: &ChildNestedStruct{
				CreditCard: "0455555554459999",
				CVV:        "333",
			},
		},
		Email: "john.doe@example.com",
	}

	res, err := json.Marshal(example)
	if err != nil {
		t.FailNow()
	}

	t.Log("Before masking:", string(res))
	maskedStruct := NewMasker().MaskStruct(example)
	resMasked, err := json.Marshal(maskedStruct)
	if err != nil {
		t.FailNow()
	}
	t.Log("After masking:", string(resMasked))

	assert.Equal(t,
		ExampleStruct{
			Name:        "******** ******",
			Age:         30,
			DogName:     "F******s",
			DogLastName: "Wo******ein",
			Address: NestedStruct{
				City:      "********",
				State:     "**",
				Phone:     "2999***",
				Cellphone: "*998695861",
				Street:    "*****sta",
				Country:   "Ecuado*",
				Child: &ChildNestedStruct{
					CreditCard: "*****5555445****",
					CVV:        "+++",
				},
			},
			Email: "XXXXXXXX@example.com",
		},
		maskedStruct,
	)
}

func TestMaskStruct_with_bad_fields(t *testing.T) {
	type EspecialChild struct {
	}
	type EspecialStruct struct {
		Name               string `mask:"regex,[A-Z"` // bad regex.
		LastName           string `mask:"corners"`
		Nacionality        string `mask:"corners,10-20"`     // smaller than limits
		NotRegister        string `mask:"not_registerd_tag"` // tag doesnt exist
		EmptyTag           string `mask:""`
		Example            ExampleStruct
		EspecialChild      *EspecialChild `mask:"corners"` // it ignores struct fields to mask
		PointerString      *string        `mask:"all"`
		OtherPointerString *string        `mask:"all"`
		PointerNilString   *string        `mask:"all"`
		EmptyRegex         string         `mask:"regex"`
	}
	pointerString := new(string)
	*pointerString = "test"

	example := &EspecialStruct{
		Name:               "Jhon",
		LastName:           "Doe",
		Nacionality:        "Ec",
		NotRegister:        "Test",
		EmptyTag:           "Test",
		EspecialChild:      nil,
		PointerString:      pointerString,
		PointerNilString:   nil,
		OtherPointerString: new(string),
		Example: ExampleStruct{
			Age: 30,
			Address: NestedStruct{
				Child: nil,
			},
			Email: "john.doe@example.com",
		},
	}

	res, err := json.Marshal(example)
	if err != nil {
		t.FailNow()
	}

	t.Log("Before masking:", string(res))
	maskedStruct := NewMasker().MaskStruct(example)
	resMasked, err := json.Marshal(maskedStruct)
	if err != nil {
		t.FailNow()
	}
	t.Log("After masking:", string(resMasked))

	assert.Equal(t,
		EspecialStruct{
			Name:               "Jhon",
			LastName:           "*o*",
			Nacionality:        "**",
			NotRegister:        "Test",
			EmptyTag:           "Test",
			EspecialChild:      nil,
			PointerString:      pointerString,
			PointerNilString:   nil,
			OtherPointerString: new(string),
			Example: ExampleStruct{
				Age: 30,
				Address: NestedStruct{
					Child: nil,
				},
				Email: "XXXXXXXX@example.com",
			},
		},
		maskedStruct,
	)
}

type MaskCard struct{}

func (m *MaskCard) Mask(value string, maskChar string, tags []string) reflect.Value {
	if len(value) < 8 {
		return reflect.ValueOf(strings.Repeat(maskChar, len(value)-1) + value[len(value)-1:])
	}

	firstVisible := 4
	lastVisible := 4

	if len(tags) > 1 {
		parts := strings.Split(tags[1], "-")
		if len(parts) == 2 {
			if first, err := strconv.Atoi(parts[0]); err == nil && first > 0 {
				firstVisible = first
			}
			if last, err := strconv.Atoi(parts[1]); err == nil && last > 0 {
				lastVisible = last
			}
		}
	}

	if firstVisible+lastVisible > len(value) {
		return reflect.ValueOf(value)
	}

	maskedPart := strings.Repeat(maskChar, len(value)-firstVisible-lastVisible)
	return reflect.ValueOf(value[:firstVisible] + maskedPart + value[len(value)-lastVisible:])
}

type EspecialStruct struct {
	CardNumber string `mask:"card_number"`
}

func TestMaskCustom(t *testing.T) {
	masker := NewMasker()
	masker.RegisterMasker("card_number", &MaskCard{})
	in := &EspecialStruct{
		CardNumber: "1234567890123456",
	}

	res, err := json.Marshal(in)
	if err != nil {
		t.FailNow()
	}

	t.Log("Before masking:", string(res))
	maskedStruct := masker.MaskStruct(in)
	resMasked, err := json.Marshal(maskedStruct)
	if err != nil {
		t.FailNow()
	}
	t.Log("After masking:", string(resMasked))
	assert.Equal(t,
		EspecialStruct{
			CardNumber: "1234********3456",
		},
		maskedStruct,
	)
}
