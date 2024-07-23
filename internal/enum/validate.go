// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package enum

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func Validate[T Valueser[T]]() schema.SchemaValidateDiagFunc {
	return validate[T](false)
}

func ValidateIgnoreCase[T Valueser[T]]() schema.SchemaValidateDiagFunc {
	return validate[T](true)
}

func validate[T Valueser[T]](ignoreCase bool) schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.StringInSlice(Values[T](), ignoreCase))
}

// TODO Move to internal/framework/validators or replace with custom types.
func FrameworkValidate[T Valueser[T]]() validator.String {
	return stringvalidator.OneOf(Values[T]()...)
}
