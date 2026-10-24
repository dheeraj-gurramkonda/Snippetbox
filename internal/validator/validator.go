package validator

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

// EmailRX performs a pragmatic syntax check before an address is stored; it
// does not prove that the mailbox exists or can receive messages.
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Validator accumulates field-specific and form-wide errors for a submission.
type Validator struct {
	NonFieldErrors []string
	FieldErrors map[string]string
}

// Valid reports whether field validation has produced any errors.
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0
}

// AddNonFieldError records a message that does not belong to one input field.
func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

// AddFieldError keeps the first error for a field, preserving validation order.
func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	if _, ok := v.FieldErrors[key]; !ok {
		v.FieldErrors[key] = message
	}
}

// CheckField adds message for key when the supplied condition fails.
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// NotBlank rejects empty strings and strings made only of whitespace.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MaxChars counts Unicode code points rather than UTF-8 bytes.
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// PermittedValue checks membership in a list of allowed values.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

// MinChars enforces a lower bound measured in Unicode code points.
func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// MaxBytes enforces a byte limit, which is useful for storage and algorithms
// such as bcrypt that impose limits on the encoded password bytes.
func MaxBytes(value string, n int) bool {
	return len(value) <= n
}

// Matches reports whether the entire value matches the provided expression.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
