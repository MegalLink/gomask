// Package masker provides functionality to recursively mask struct fields based on tags.
package masker

import (
	"testing"
)

func BenchmarkMaskStruct(b *testing.B) {
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

	masker := NewMasker()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = masker.MaskStruct(example)
	}
}

func BenchmarkMaskStructWithCustomMasker(b *testing.B) {
	in := &EspecialStruct{
		CardNumber: "1234567890123456",
	}

	masker := NewMasker()
	masker.RegisterMasker("card_number", &MaskCard{})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = masker.MaskStruct(in)
	}
}

func BenchmarkMaskAll(b *testing.B) {
	input := "This is a test string for benchmarking"
	masker := &MaskAll{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = masker.Mask(input, "*", []string{"all"})
	}
}

func BenchmarkMaskRegex(b *testing.B) {
	input := "john.doe@example.com"
	masker := &MaskRegex{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = masker.Mask(input, "X", []string{"regex", "^[^@]+"})
	}
}

func BenchmarkMaskFirst(b *testing.B) {
	input := "0998695861"
	masker := &MaskFirst{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = masker.Mask(input, "*", []string{"first", "1"})
	}
}

func BenchmarkMaskLast(b *testing.B) {
	input := "2999999"
	masker := &MaskLast{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = masker.Mask(input, "*", []string{"last", "3"})
	}
}

func BenchmarkMaskCorners(b *testing.B) {
	input := "0455555554459999"
	masker := &MaskCorners{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = masker.Mask(input, "*", []string{"corners", "5-4"})
	}
}

func BenchmarkMaskBetween(b *testing.B) {
	input := "Wolfenstein"
	masker := &MaskBetween{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = masker.Mask(input, "*", []string{"between", "2-3"})
	}
}

func BenchmarkMaskerManager_ComplexStruct(b *testing.B) {
	// Create a complex nested structure to benchmark performance with deep nesting
	type Level3 struct {
		Field1 string `mask:"all"`
		Field2 string `mask:"regex,\\w+"`
	}

	type Level2 struct {
		Field1  string `mask:"first,2"`
		Field2  string `mask:"last,3"`
		Nested  Level3
		Pointer *Level3 `mask:"corners"`
	}

	type Level1 struct {
		Field1  string `mask:"between,1-1"`
		Field2  string `mask:"corners,2-2"`
		Nested  Level2
		Pointer *Level2
	}

	level3Ptr := &Level3{
		Field1: "Secret Data",
		Field2: "More Secret Data",
	}

	level2Ptr := &Level2{
		Field1:  "Confidential",
		Field2:  "Private",
		Nested:  Level3{Field1: "Internal", Field2: "Classified"},
		Pointer: level3Ptr,
	}

	complex := &Level1{
		Field1:  "TopSecret",
		Field2:  "Restricted",
		Nested:  Level2{Field1: "Hidden", Field2: "Protected", Nested: Level3{Field1: "Sensitive", Field2: "Personal"}},
		Pointer: level2Ptr,
	}

	masker := NewMasker()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = masker.MaskStruct(complex)
	}
}

// BenchmarkParallelMasking tests the performance of masking in parallel
func BenchmarkParallelMasking(b *testing.B) {
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

	masker := NewMasker()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = masker.MaskStruct(example)
		}
	})
}

// BenchmarkMaskStringFunctions benchmarks the individual string masking functions
func BenchmarkMaskStringFunctions(b *testing.B) {
	b.Run("MaskStringAll", func(b *testing.B) {
		input := "This is a test string for benchmarking"
		for i := 0; i < b.N; i++ {
			_ = MaskStringAll(input, "*")
		}
	})

	b.Run("MaskStringRegex", func(b *testing.B) {
		input := "john.doe@example.com"
		for i := 0; i < b.N; i++ {
			_ = MaskStringRegex(input, "^[^@]+", "X")
		}
	})

	b.Run("MaskStringFirst", func(b *testing.B) {
		input := "0998695861"
		for i := 0; i < b.N; i++ {
			_ = MaskStringFirst(input, 1, "*")
		}
	})

	b.Run("MaskStringLast", func(b *testing.B) {
		input := "2999999"
		for i := 0; i < b.N; i++ {
			_ = MaskStringLast(input, 3, "*")
		}
	})

	b.Run("MaskStringCorners", func(b *testing.B) {
		input := "0455555554459999"
		for i := 0; i < b.N; i++ {
			_ = MaskStringCorners(input, 5, 4, "*")
		}
	})

	b.Run("MaskAllExceptCorners", func(b *testing.B) {
		input := "Wolfenstein"
		for i := 0; i < b.N; i++ {
			_ = MaskAllExceptCorners(input, 2, 3, "*")
		}
	})
}

// BenchmarkMaskerCreation benchmarks the creation of a new masker
func BenchmarkMaskerCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewMasker()
	}
}

// BenchmarkMaskerRegistration benchmarks the registration of a custom masker
func BenchmarkMaskerRegistration(b *testing.B) {
	masker := NewMasker()
	customMasker := &MaskCard{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		masker.RegisterMasker("custom"+string(rune(i%100)), customMasker)
	}
}

// BenchmarkMaskerGetMasker benchmarks retrieving a masker by name
func BenchmarkMaskerGetMasker(b *testing.B) {
	masker := NewMasker()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = masker.GetMasker("all")
	}
}
