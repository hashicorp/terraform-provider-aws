// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package json

import (
	"encoding/json"

	smithydocument "github.com/aws/smithy-go/document"
)

func SmithyDocumentFromString[T smithydocument.Marshaler](s string, f func(any) T) (T, error) {
	var v map[string]interface{}

	err := json.Unmarshal([]byte(s), &v)
	if err != nil {
		var zero T
		return zero, err
	}

	return f(v), nil
}

// SmithyDocumentToString converts a [Smithy document](https://smithy.io/2.0/spec/simple-types.html#document) to a JSON string.
func SmithyDocumentToString(document smithydocument.Unmarshaler) (string, error) {
	var v map[string]interface{}

	err := document.UnmarshalSmithyDocument(&v)
	if err != nil {
		return "", err
	}

	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// JSONStringer interface is used to marshal and unmarshal JSON interface objects.
type JSONStringer interface {
	smithydocument.Marshaler
	smithydocument.Unmarshaler
}
