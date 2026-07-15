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

type whenNotEquals[T ~string] struct {
	value T
}

func (w whenNotEquals[T]) Eval(_ context.Context, v attr.Value) bool {
	return !v.Equal(types.StringValue(string(w.value)))
}

func (w whenNotEquals[T]) String() string {
	return "not equals " + strconv.Quote(string(w.value))
}

// ConflictsWithWhenNotEquals checks that each path.Expression has a null
// configuration value when the stringy attribute being validated does not have
// the specified value.
//
// Relative path.Expressions are resolved using the attribute being
// validated.
func ConflictsWithWhenNotEquals[T ~string](value T, expressions ...path.Expression) validator.String {
	return internal.ConflictsWithWhenValidator(whenNotEquals[T]{value: value}, expressions...)
}
