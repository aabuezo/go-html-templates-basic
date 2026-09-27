package forms

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/asaskevich/govalidator"
)

// Form creates a custom form struct, embeds a url.Values object
type Form struct {
	url.Values // holds some values from the form
	Errors     ValidationErrors
}

// New initializes a form struct
func New(data url.Values) *Form {
	return &Form{
		data,
		ValidationErrors(map[string][]string{}),
	}
}

// Required checks for required fields
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

// Has checks if form field is in post and not empty
func (f *Form) Has(field string) bool {
	x := f.Get(field)
	if x == "" {
		f.Errors.Add(field, "This field cannot be blank")
		return false
	}
	return true
}

// Valid returns true if there are no errors, otherwise false
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// MinLength checks for string minimum length
func (f *Form) MinLength(field string, length int) bool {
	x := strings.TrimSpace(f.Get(field))
	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	}
	return true
}

// CheckEmail validates the syntax of an email address.
func (f *Form) CheckEmail(field string) bool {
	email := strings.TrimSpace(f.Get(field))

	if len(email) > 254 || !govalidator.IsEmail(email) {
		f.Errors.Add(field, "Please enter a valid email address")
		return false
	}
	return true
}

// CheckPassword checks for valid passwords
func (f *Form) CheckPassword(field string) bool {
	pass := f.Get(field) // space is a valid character in passwords
	const message = "Password must be 8-128 characters and contain uppercase, lowercase, number, and symbol"

	length := utf8.RuneCountInString(pass)
	if length < 8 || length > 128 {
		f.Errors.Add(field, message)
		return false
	}

	var upper, lower, digit, symbol bool
	for _, r := range pass {
		switch {
		case unicode.IsControl(r):
			f.Errors.Add(field, "Password contains an invalid control character")
			return false
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsLower(r):
			lower = true
		case unicode.IsDigit(r):
			digit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			symbol = true
		}
	}

	if !upper || !lower || !digit || !symbol {
		f.Errors.Add(field, message)
		return false
	}

	return true
}
