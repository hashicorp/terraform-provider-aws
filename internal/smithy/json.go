// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package smithy

import (
	"strings"

	smithydocument "github.com/aws/smithy-go/document"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

// DocumentFromJSONString converts a JSON string to a [Smithy document](https://smithy.io/2.0/spec/simple-types.html#document).
func DocumentFromJSONString[T any](s string, f func(any) T) (T, error) {
	var v any

	err := tfjson.DecodeFromString(s, &v)
	if err != nil {
		var zero T
		return zero, err
	}

	return f(v), nil
}

// DocumentToJSONString converts a [Smithy document](https://smithy.io/2.0/spec/simple-types.html#document) to a JSON string.
func DocumentToJSONString(document smithydocument.Marshaler) (string, error) {
	bytes, err := document.MarshalSmithyDocument()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(bytes)), nil
}

// JSONStringer interface is used to marshal and unmarshal JSON interface objects.
type JSONStringer interface {
	smithydocument.Marshaler
	smithydocument.Unmarshaler
}
