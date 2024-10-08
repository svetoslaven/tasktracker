package validator

import (
	"fmt"
	"net/mail"
	"reflect"
)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) CheckGreaterThanOrEqualTo(value, threshold int, field string) {
	v.Check(value >= threshold, field, fmt.Sprintf("Must be greater than or equal to %d.", threshold))
}

func (v *Validator) CheckLessThanOrEqualTo(value, threshold int, field string) {
	v.Check(value <= threshold, field, fmt.Sprintf("Must be less than or equal to %d.", threshold))
}

func (v *Validator) CheckStringMaxLength(value string, maxLength int, field string) {
	v.Check(len(value) <= maxLength, field, fmt.Sprintf("Must be no more than %d bytes long.", maxLength))
}

func (v *Validator) CheckStringMinLength(value string, minLength int, field string) {
	v.Check(len(value) >= minLength, field, fmt.Sprintf("Must be at least %d bytes long.", minLength))
}

func (v *Validator) CheckValidEmail(email string, field string) {
	addr, err := mail.ParseAddress(email)
	v.Check(err == nil && email == addr.Address, field, "Must be a valid email address.")
}

func (v *Validator) CheckNonZero(data any, field string) {
	v.Check(!reflect.ValueOf(data).IsZero(), field, "Must be provided.")
}

func (v *Validator) Check(cond bool, field, msg string) {
	if _, ok := v.Errors[field]; ok {
		return
	}

	if !cond {
		v.Errors[field] = msg
	}
}

func (v *Validator) AddError(field, msg string) {
	if _, ok := v.Errors[field]; !ok {
		v.Errors[field] = msg
	}
}

func (v *Validator) HasErrors() bool {
	return len(v.Errors) > 0
}
