// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package stringvalidator

import (
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators/internal"
)

func WarnExactlyOneOf(expressions ...path.Expression) validator.String {
	return internal.ExactlyOneOfValidator{
		PathExpressions: expressions,
	}
}
