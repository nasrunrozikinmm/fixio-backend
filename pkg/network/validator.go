package network

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Validatable is the interface that DTOs must implement for custom error messages
type Validatable[T any] interface {
	GetValue() *T
	ValidateErrors(errs validator.ValidationErrors) ([]string, error)
}

// ValidationResult holds the result of a validation
type ValidationResult struct {
	Valid    bool
	Messages []string
}

// FirstMessage returns the first validation error message
func (v *ValidationResult) FirstMessage() string {
	if len(v.Messages) > 0 {
		return v.Messages[0]
	}
	return "Validation failed"
}

// Validate validates a DTO and returns a ValidationResult
func Validate[T any](dto Validatable[T]) *ValidationResult {
	val := dto.GetValue()
	err := validate.Struct(val)
	if err == nil {
		return &ValidationResult{Valid: true}
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return &ValidationResult{
			Valid:    false,
			Messages: []string{fmt.Sprintf("Validation error: %v", err)},
		}
	}

	messages, customErr := dto.ValidateErrors(validationErrors)
	if customErr != nil || len(messages) == 0 {
		// Fallback to default messages
		messages = defaultValidationMessages(validationErrors)
	}

	return &ValidationResult{
		Valid:    false,
		Messages: messages,
	}
}

// defaultValidationMessages generates default validation error messages
func defaultValidationMessages(errs validator.ValidationErrors) []string {
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		msg := fmt.Sprintf("Field '%s' failed on '%s' validation", err.Field(), err.Tag())
		messages = append(messages, msg)
	}
	return messages
}

// GetValidator returns the shared validator instance
func GetValidator() *validator.Validate {
	return validate
}
