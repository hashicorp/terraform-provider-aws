package datafy

import (
	"errors"
	"strings"
)

var (
	NotFoundError = errors.New("resource not found")
)

func NotFound(err error) bool {
	return errors.Is(err, NotFoundError)
}

func ApiNotFound(err error) bool {
	return strings.Contains(err.Error(), "404")
}
