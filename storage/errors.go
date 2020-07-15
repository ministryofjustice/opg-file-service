package storage

import "fmt"

type NotFoundError struct {
	Ref string
}

type ErrFieldValidation struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrValidation struct {
	Errors []ErrFieldValidation `json:"errors"`
}

func (err NotFoundError) Error() string {
	return "Could not find entry with reference: " + err.Ref
}

func (e ErrFieldValidation) Error() string {
	return fmt.Sprintf("Field %s failed validation: %s", e.Field, e.Message)
}

func (e ErrValidation) Error() string {
	errString := "Validation errors: "
	for _, err := range e.Errors {
		errString += err.Error() + "; "
	}
	return errString
}
