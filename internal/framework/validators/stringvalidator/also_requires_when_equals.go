// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package stringvalidator

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators/internal"
)

type whenEquals[T ~string] struct {
	value T
}

func (w whenEquals[T]) Eval(_ context.Context, v attr.Value) bool {
	return v.Equal(types.StringValue(string(w.value)))
}

func (w whenEquals[T]) String() string {
	return "equals " + strconv.Quote(string(w.value))
}

// AlsoRequiresWhenEquals checks that each path.Expression has a non-null
// configuration value when the stringy attribute being validated has the known
// specified value.
//
// Relative path.Expressions are resolved using the attribute being
// validated.
func AlsoRequiresWhenEquals[T ~string](value T, expressions ...path.Expression) validator.String {
	return internal.AlsoRequiresWhenValidator(whenEquals[T]{value: value}, expressions...)
}
