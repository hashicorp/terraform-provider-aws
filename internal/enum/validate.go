// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package enum

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func Validate[T valueser[T]]() schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.StringInSlice(Values[T](), false))
}

func FrameworkValidate[T valueser[T]]() validator.String {
	return stringvalidator.OneOf(Values[T]()...)
}
