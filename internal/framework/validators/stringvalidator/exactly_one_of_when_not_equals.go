// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package stringvalidator

import (
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators/internal"
)

// ExactlyOneOfWhenNotEquals checks that one and only one of a set of path.Expressions
// has a non-null configuration value when the stringy attribute being validated does
// not have the specified value.
//
// The attribute or block the validator is applied to is not included in the set of path.Expressions.
//
// Relative path.Expressions are resolved using the attribute being validated.
func ExactlyOneOfWhenNotEquals[T ~string](value T, expressions ...path.Expression) validator.String {
	return internal.ExactlyOneOfWhenValidator(whenNotEquals[T]{value: value}, expressions...)
}
