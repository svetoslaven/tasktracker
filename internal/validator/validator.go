package validator

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Check(cond bool, field, msg string) {
	if !cond {
		v.AddError(field, msg)
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
