# GoMask

[![Go Reference](https://pkg.go.dev/badge/github.com/MegalLink/gomask.svg)](https://pkg.go.dev/github.com/MegalLink/gomask)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

GoMask is a powerful and flexible Go library designed to help you protect sensitive data by masking struct fields based on simple tag annotations. It's perfect for logging, displaying, or transmitting data while maintaining security and privacy.

## 📦 Installation

Add GoMask to your Go module:

```bash
go get github.com/MegalLink/gomask
```

## ✨ Features

- **Simple Tag-Based Masking**: Annotate struct fields with `mask` tags to define masking behavior
- **Nested Structures**: Seamlessly handles nested structs and pointers
- **Multiple Built-in Strategies**:
  - `all`: Completely masks the entire string
  - `regex`: Masks based on regular expression patterns
  - `first`: Masks the first n characters
  - `last`: Masks the last n characters
  - `corners`: Masks the beginning and end of strings
  - `between`: Masks the middle portion of strings
- **Customizable**: Define your own masking strategies
- **Thread-Safe**: Safe for concurrent use
- **Lightweight**: No external dependencies

## 🚀 Quick Start

### Basic Usage

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/MegalLink/gomask/masker"
)

type User struct {
    Name     string `mask:"regex,\\b[A-Za-z]+\\b"` // Mask words, leave spaces
    Email    string `mask:"regex,^[^@]+" maskTag:"X"` // Mask local part of email
    Password string `mask:"all"`                      // Mask entire string
    Phone    string `mask:"last,4"`                   // Mask last 4 digits
}

func main() {
    user := User{
        Name:     "John Doe",
        Email:    "john.doe@example.com",
        Password: "secret123",
        Phone:    "1234567890",
    }
    
    maskedUser := masker.NewMasker().MaskStruct(user).(User)
    
    maskedJSON, _ := json.MarshalIndent(maskedUser, "", "  ")
    fmt.Println(string(maskedJSON))
}
```

**Output:**
```json
{
  "Name": "**** ***",
  "Email": "XXXXXXXX@example.com",
  "Password": "**********",
  "Phone": "123456****"
}
```

## 📚 Usage Guide

### Working with Nested Structures

GoMask handles nested structs and pointers seamlessly:

```go
type Address struct {
    Street  string `mask:"first,5" json:"street"`
    City    string `mask:"all" json:"city"`
    Country string `mask:"last" json:"country"`
}

type User struct {
    Name    string  `mask:"regex,\\b[A-Za-z]+\\b"`
    Email   string  `mask:"regex,^[^@]+" maskTag:"X"`
    Address Address
}

// Usage:
user := User{
    Name:  "John Doe",
    Email: "john.doe@example.com",
    Address: Address{
        Street:  "Floresta",
        City:    "New York",
        Country: "Ecuador",
    },
}

maskedUser := masker.NewMasker().MaskStruct(user).(User)
```

## 🔧 Custom Masking Strategies

Easily extend GoMask with your own masking logic:

```go
// Define a custom masker
type MaskCard struct{}


func (m *MaskCard) Mask(value string, maskChar string, tags []string) reflect.Value {
    // Show first 4 and last 4 digits by default
    if len(value) < 8 {
        return reflect.ValueOf(strings.Repeat(maskChar, len(value)-1) + value[len(value)-1:])
    }
    
    firstVisible := 4
    lastVisible := 4
    
    // Parse custom visibility from tags if provided
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

// Register and use your custom masker
type PaymentInfo struct {
    CardNumber  string `mask:"card_number"`
    CardNumber2 string `mask:"card_number,2-3"` // Custom format: first 2 and last 3 digits visible
}

masker := masker.NewMasker()
masker.RegisterMasker("card_number", &MaskCard{})

payment := PaymentInfo{
    CardNumber:  "1234567890123456",
    CardNumber2: "1234567890123456",
}

maskedPayment := masker.MaskStruct(payment).(PaymentInfo)
// CardNumber: "1234********3456"
// CardNumber2: "12***********456"
```

## 📋 Available Masking Methods

### `all`
Masks all characters in the string.

```go
Password string `mask:"all"` // "secret123" → "*********"
```

### `regex`
Masks substrings matching the provided regular expression pattern.

```go
Email string `mask:"regex,^[^@]+" maskTag:"X"` // "john.doe@example.com" → "XXXXXXXX@example.com"
```

### `first`
Masks the first n characters of the string.

```go
// Mask first character (default)
Phone string `mask:"first"` // "0998695861" → "*998695861"

// Mask first 5 characters
Street string `mask:"first,5"` // "Floresta" → "*****esta"
```

### `last`
Masks the last n characters of the string.

```go
// Mask last character (default)
Country string `mask:"last"` // "Ecuador" → "Ecuado*"

// Mask last 3 characters
Phone string `mask:"last,3"` // "2999999" → "2999***"
```

### `corners`
Masks the first n and last m characters of the string.

```go
// Format: corners,n-m
CreditCard string `mask:"corners,5-4"` // "0455555554459999" → "*****5555445****"
```

### `between`
Masks all characters except the first n and last m characters.

```go
// Default: keep first and last character
DogName string `mask:"between"` // "Firulais" → "F******s"

// Custom format: keep first 2 and last 3 characters
DogLastName string `mask:"between,2-3"` // "Wolfenstein" → "Wo******ein"
```

## 🎨 Customization

### Custom Mask Character

Change the default mask character (`*`) using the `maskTag` tag:

```go
CVV string `mask:"all" maskTag:"+"` // "333" → "+++"
```

## 🔒 Thread Safety

GoMask is safe for concurrent use, utilizing read-write locks to ensure thread safety during masker registration and retrieval.

## 📄 License

This project is open source and available under the [MIT License](LICENSE).
