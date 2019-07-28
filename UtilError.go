package goghostex

import "fmt"

type Error interface {
	error
	Code() int
}

type apiError struct {
	code    int
	message string
}

func (this *apiError) Error() string {
	return this.message
}

func (this *apiError) Code() int {
	return this.code
}

// New creates a new API error with a code and a message
func NewError(code int, message string, args ...interface{}) Error {
	if len(args) > 0 {
		return &apiError{code, fmt.Sprintf(message, args...)}
	}
	return &apiError{code, message}
}
