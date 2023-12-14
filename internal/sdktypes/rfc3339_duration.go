// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdktypes

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/types/duration"
)

const (
	TypeRFC3339Duration = schema.TypeString
)

type RFC3339Duration string

func (d RFC3339Duration) IsNull() bool {
	return d == ""
}

func (d RFC3339Duration) Value() (duration.Duration, bool, error) {
	if d.IsNull() {
		return duration.Duration{}, true, nil
	}

	value, err := duration.Parse(string(d))
	if err != nil {
		return duration.Duration{}, false, err
	}
	return value, false, nil
}

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
