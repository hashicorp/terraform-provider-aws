package errs

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

// Messager is a simple interface for types with ErrorMessage().
type Messager interface {
	ErrorMessage() string
}

func AsContains(err error, target any, message string) bool {
	if errors.As(err, target) {
		if v, ok := target.(Messager); ok && strings.Contains(v.ErrorMessage(), message) {
			return true
		}
	}
	return false
}

// Contains returns true if the error matches all these conditions:
//   - err as string contains needle
func Contains(err error, needle string) bool {
	if err != nil && strings.Contains(err.Error(), needle) {
		return true
	}
	return false
}

// MessageContains unwraps the error and returns true if the error matches
// all these conditions:
//   - err is of type awserr.Error, Error.Code() equals code, and Error.Message() contains message
//   - OR err if not of type awserr.Error as string contains both code and message
func MessageContains(err error, code string, message string) bool {
	var awsErr awserr.Error
	if AsContains(err, &awsErr, message) {
		return true
	}

	if Contains(err, code) && Contains(err, message) {
		return true
	}

	return false
}
