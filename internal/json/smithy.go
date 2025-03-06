// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package json

import (
	smithydocument "github.com/aws/smithy-go/document"
)

func SmithyDocumentFromString[T smithydocument.Marshaler](s string, f func(any) T) (T, error) {
	var v map[string]interface{}

	err := DecodeFromString(s, &v)
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

	return EncodeToString(v)
}

// JSONStringer interface is used to marshal and unmarshal JSON interface objects.
type JSONStringer interface {
	smithydocument.Marshaler
	smithydocument.Unmarshaler
}
