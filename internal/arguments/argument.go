package arguments

import (
	"fmt"
	"reflect"

	"github.com/spf13/cast"
)

type ParsableArgument interface {
	Parse() error
	Name() string
	FormattedValue() string
	IsSensitive() bool
}

type Argument[T any] struct {
	rawValue       string      // Raw input from CLI as string
	parsedValue    T           // Parsed valuess
	formattedValue string      // Formatted value or parsed value
	name           string      // Argument name
	parserFunc     interface{} // Parser function or raw value
	formatFunc     interface{} // Format function or string
	sensitive      bool        // True if the argument contains sensitive data
}

func NewArgument[T any](rawValue interface{}, name string, parserFunc interface{}, formatFunc interface{}, sensitive bool) *Argument[T] {
	rawValueStr := cast.ToString(rawValue)
	return &Argument[T]{
		rawValue:       rawValueStr,
		formattedValue: rawValueStr,
		name:           name,
		parserFunc:     parserFunc,
		formatFunc:     formatFunc,
		sensitive:      sensitive,
	}
}

func (a *Argument[T]) RawValue() string {
	return a.rawValue
}

func (a *Argument[T]) Value() T {
	return a.parsedValue
}

func (a *Argument[T]) FormattedValue() string {
	return a.formattedValue
}

func (a *Argument[T]) Name() string {
	return a.name
}

func (a *Argument[T]) IsSensitive() bool {
	return a.sensitive
}

func (a *Argument[T]) String() string {
	return fmt.Sprintf("%v", a.parsedValue)
}

func (a *Argument[T]) Parse() error {
	var pValue T
	var err error

	// Use a centralized fallback parser for type inference
	fallbackParser := func(value interface{}) (T, error) {
		switch any(pValue).(type) {
		case int:
			val, err := cast.ToIntE(value)
			return any(val).(T), err
		case float64:
			val, err := cast.ToFloat64E(value)
			return any(val).(T), err
		case bool:
			val, err := cast.ToBoolE(value)
			return any(val).(T), err
		case string:
			return any(value).(T), nil
		default:
			return pValue, fmt.Errorf("unsupported type; provide a custom parser or valid type")
		}
	}

	// Parse using the provided custom parser or fallback to default logic
	if a.parserFunc != nil {
		parserValue := reflect.ValueOf(a.parserFunc)
		if parserValue.Kind() == reflect.Func {
			expectedType := parserValue.Type().In(0) // first parameter type
			var rawValue reflect.Value

			// Convert raw value to the expected type
			switch expectedType.Kind() {
			case reflect.Int:
				parsed, err := cast.ToIntE(a.rawValue)
				if err != nil {
					return fmt.Errorf("failed to cast raw input to int: %w", err)
				}
				rawValue = reflect.ValueOf(parsed)
			case reflect.Float64:
				parsed, err := cast.ToFloat64E(a.rawValue)
				if err != nil {
					return fmt.Errorf("failed to cast raw input to float64: %w", err)
				}
				rawValue = reflect.ValueOf(parsed)
			case reflect.Bool:
				parsed, err := cast.ToBoolE(a.rawValue)
				if err != nil {
					return fmt.Errorf("failed to cast raw input to boolean: %w", err)
				}
				rawValue = reflect.ValueOf(parsed)
			case reflect.String:
				rawValue = reflect.ValueOf(a.rawValue)
			default:
				return fmt.Errorf("unsupported parser input type: %v", expectedType)
			}

			// Call the parser function with the converted value
			results := parserValue.Call([]reflect.Value{rawValue})
			if len(results) == 2 && !results[1].IsNil() {
				err = results[1].Interface().(error)
				if err != nil {
					return err
				}
			}
			parsedValue, ok := results[0].Interface().(T)
			if !ok {
				return fmt.Errorf("parser function returned unexpected type: expected %T, got %T", pValue, results[0].Interface())
			}
			pValue = parsedValue
		} else {
			return fmt.Errorf("argument parser must be a function of type func(T) (T, error)")
		}
	} else {
		pValue, err = fallbackParser(a.rawValue)
		if err != nil {
			return err
		}
	}

	// Formatted value
	var fText string
	if a.formatFunc != nil {
		formatFuncValue := reflect.ValueOf(a.formatFunc)
		if formatFuncValue.Kind() == reflect.Func {
			results := formatFuncValue.Call([]reflect.Value{reflect.ValueOf(pValue)})
			if len(results) == 1 {
				fText = results[0].Interface().(string)
			} else {
				fText = a.rawValue
			}
		} else {
			if str, ok := a.formatFunc.(string); ok {
				fText = str
			} else {
				fText = a.rawValue
			}
		}
	} else {
		fText = fmt.Sprintf("%v", pValue)
	}

	a.parsedValue = pValue
	a.formattedValue = fText
	return nil
}
