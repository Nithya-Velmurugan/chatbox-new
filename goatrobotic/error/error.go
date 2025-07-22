package errcom

import "fmt"

type CustomError struct {
	Code string
	Err  error
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Err.Error())
}

func NewCustomError(code string, err error) error {
	return &CustomError{
		Code: code,
		Err:  err,
	}
}
