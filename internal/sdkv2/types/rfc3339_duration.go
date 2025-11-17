// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/types/duration"
)

func ValidateRFC3339Duration(i any, path cty.Path) diag.Diagnostics {
	v, ok := i.(string)
	if !ok {
		return diag.Diagnostics{errs.NewIncorrectValueTypeAttributeError(path, "string")}
	}

	_, err := duration.Parse(v)
	if err != nil {
		return diag.Diagnostics{errs.NewInvalidValueAttributeErrorf(path, "Cannot be parsed as an RFC 3339 duration: %s", err)}
	}

	return nil
}
