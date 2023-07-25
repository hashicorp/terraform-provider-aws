// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type ValueAsPtr interface {
	ValueAsPtr(context.Context) (any, diag.Diagnostics)
}
