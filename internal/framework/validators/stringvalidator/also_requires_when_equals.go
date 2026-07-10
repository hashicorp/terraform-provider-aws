// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package stringvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators/internal"
)

// AlsoRequiresWhenEquals checks that each path.Expression has a non-null
// configuration value when the stringy attribute being validated has the known
// specified value.
//
// Relative path.Expressions are resolved using the attribute being
// validated.
func AlsoRequiresWhenEquals[T ~string](value T, expressions ...path.Expression) validator.String {
	return internal.AlsoRequiresWhenValidator{
		When: func(_ context.Context, v attr.Value) bool {
			return v.Equal(types.StringValue(string(value)))
		},
		PathExpressions: expressions,
	}
}
