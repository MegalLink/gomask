# GoMask

GoMask is a Go library that provides functionality to recursively mask struct fields based on tags. It allows you to easily mask sensitive data in your structs before logging, displaying, or transmitting them.

## Installation

To install the library, use the standard Go module installation:

```bash
go get github.com/MegalLink/gomask
```

## Features

- Mask struct fields using simple tag annotations
- Support for nested structs and pointers
- Multiple masking strategies:
  - `all`: Mask all characters in a string
  - `regex`: Mask characters based on a regular expression pattern
  - `first`: Mask the first n characters in a string
  - `last`: Mask the last n characters in a string
  - `corners`: Mask the first n and last m characters in a string
  - `between`: Mask all except the first n and last m characters in a string
- Custom masking character via `maskTag` tag
- Extensible architecture for custom masking strategies
- Thread-safe implementation

## Usage

### Basic Usage

```go
package main

import (
    "encoding/json"
    "fmt"
    
    "github.com/MegalLink/gomask/masker"
)

type User struct {
    Name     string `mask:"regex,\\b[A-Za-z]+\\b"` // Mask all characters not including spaces
    Email    string `mask:"regex,^[^@]+" maskTag:"X"` // Mask everything before @ with X
    Password string `mask:"all"` // Mask all characters
    Phone    string `mask:"last,4"` // Mask last 4 digits
}

func main() {
    user := User{
        Name:     "John Doe",
        Email:    "john.doe@example.com",
        Password: "secret123",
        Phone:    "1234567890",
    }
    
    // Create a new masker and mask the struct
    masker := masker.NewMasker()
    maskedUser := masker.MaskStruct(user).(User)
    
    // Print the masked user
    maskedJSON, _ := json.MarshalIndent(maskedUser, "", "  ")
    fmt.Println(string(maskedJSON))
}
```

Output:
```json
{
  "Name": "**** ***",
  "Email": "XXXXXXXX@example.com",
  "Password": "**********",
  "Phone": "123456****"
}
```

### Nested Structs

GoMask supports nested structs and pointers to structs:

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

### Custom Masking Strategies

You can create your own masking strategies by implementing the `Masker` interface:

```go
type MaskCard struct{}

func (m *MaskCard) Mask(value string, maskChar string, tags []string) reflect.Value {
    // Show first 4 and last 4 digits of a card number
    if len(value) < 8 {
        return reflect.ValueOf(strings.Repeat(maskChar, len(value)-1) + value[len(value)-1:])
    }
    
    firstVisible := 4
    lastVisible := 4
    
    // Parse options if provided
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
    CardNumber string `mask:"card_number"`
}

masker := masker.NewMasker()
masker.RegisterMasker("card_number", &MaskCard{})

payment := PaymentInfo{
    CardNumber: "1234567890123456",
}

maskedPayment := masker.MaskStruct(payment).(PaymentInfo)
// CardNumber will be "1234********3456"
```

## Available Masking Methods

### all

Masks all characters in a string.

```go
Password string `mask:"all"` // "secret123" -> "*********"
```

### regex

Masks characters based on a regular expression pattern.

```go
Email string `mask:"regex,^[^@]+" maskTag:"X"` // "john.doe@example.com" -> "XXXXXXXX@example.com"
```

### first

Masks the first n characters in a string.

```go
// Default: mask first character
Phone string `mask:"first"` // "0998695861" -> "*998695861"

// Mask first 5 characters
Street string `mask:"first,5"` // "Floresta" -> "*****esta"
```

### last

Masks the last n characters in a string.

```go
// Default: mask last character
Country string `mask:"last"` // "Ecuador" -> "Ecuado*"

// Mask last 3 characters
Phone string `mask:"last,3"` // "2999999" -> "2999***"
```

### corners

Masks the first n and last m characters in a string.

```go
// Format: corners,n-m
CreditCard string `mask:"corners,5-4"` // "0455555554459999" -> "*****5555445****"
```

### between

Masks all except the first n and last m characters in a string.

```go
// Default: keep first and last character
DogName string `mask:"between"` // "Firulais" -> "F******s"

// Format: between,n-m (keep first n and last m characters)
DogLastName string `mask:"between,2-3"` // "Wolfenstein" -> "Wo******ein"
```

## Custom Mask Character

You can specify a custom mask character using the `maskTag` tag:

```go
CVV string `mask:"all" maskTag:"+"` // "333" -> "+++"
```

## Thread Safety

GoMask uses read-write locks to ensure thread safety when registering and retrieving maskers.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
