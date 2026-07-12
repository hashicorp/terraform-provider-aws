// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package boolvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators/internal"
)

type whenTrue struct{}

func (whenTrue) Eval(_ context.Context, v attr.Value) bool {
	return v.Equal(types.BoolValue(true))
}

func (whenTrue) String() string {
	return "is true"
}

// AlsoRequiresWhenTrue checks that each path.Expression has a non-null
// configuration value when the bool attribute being validated has a known
// value of true.
//
// Unlike boolvalidator.AlsoRequires in
// github.com/hashicorp/terraform-plugin-framework-validators, which fires
// whenever the attribute is non-null regardless of value, this validator
// fires only when the attribute is explicitly true. Use it for asymmetric
// constraints such as "when `foo_enabled` is true, `foo_arn` must be set."
//
// Relative path.Expressions are resolved using the attribute being
// validated.
func AlsoRequiresWhenTrue(expressions ...path.Expression) validator.Bool {
	return internal.AlsoRequiresWhenValidator{
		When:            whenTrue{},
		PathExpressions: expressions,
	}
}
