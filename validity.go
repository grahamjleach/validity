package validity

import (
	"reflect"
	"strings"
)

type Error struct {
	message string
}

func (e Error) Error() string {
	return e.message
}

func NewError(message string) *Error {
	return &Error{
		message: message,
	}
}

func IsError(err error) bool {
	_, ok := err.(*Error)
	return ok
}

// validatable interface
type validatable interface {
	Validate() *Error
}

// Validate traverses all values and invokes the Validate method on any values that implement the validatable interface.
// Validate will stop traversing and return an error when it receives a non-nil response from a validatable.
func Check(values ...interface{}) error {
	for _, v := range values {
		value := reflect.ValueOf(v)
		if value.Kind() == reflect.Ptr && value.IsNil() {
			continue
		}

		value = indirect(reflect.ValueOf(v))
		valueType := value.Type()
		valueKind := valueType.Kind()

		if valueKind == reflect.Slice || valueKind == reflect.Array {
			for i := 0; i < value.Len(); i++ {
				if err := Validate(value.Index(i).Interface()); err != nil {
					return err
				}
			}
		}

		if valueKind.String() == "struct" {
			for i := 0; i < value.NumField(); i++ {
				fieldName := valueType.Field(i).Name
				if isPrivateField(fieldName) {
					continue
				}

				if err := Validate(value.FieldByIndex([]int{i}).Interface()); err != nil {
					return err
				}
			}
		}

		if val, ok := value.Interface().(validatable); ok {
			if err := val.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

// indirect resolves pointer values
func indirect(value reflect.Value) reflect.Value {
	if value.Kind() != reflect.Ptr && value.Type().Name() != "" && value.CanAddr() {
		value = value.Addr()
	}
	for {
		if value.Kind() == reflect.Interface && !value.IsNil() {
			element := value.Elem()
			if element.Kind() == reflect.Ptr && !element.IsNil() {
				value = element
				continue
			}
		}
		if value.Kind() != reflect.Ptr {
			break
		}
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		value = value.Elem()
	}
	return value
}

func isPrivateField(fieldName string) bool {
	init := string(fieldName[0])
	return init == strings.ToLower(init)
}
