// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package stringvalidator

import (
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators/internal"
)

func AlsoRequiresWhenEquals[T ~string](value T, expressions ...path.Expression) validator.String {
	return internal.AlsoRequiresWhenEqualsValidator{
		Value:           types.StringValue(string(value)),
		PathExpressions: expressions,
	}
}
