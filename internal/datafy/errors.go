package datafy

import (
	"errors"
)

var (
	NotFoundError = errors.New("resource not found")
)

func NotFound(err error) bool {
	return errors.Is(err, NotFoundError)
}
