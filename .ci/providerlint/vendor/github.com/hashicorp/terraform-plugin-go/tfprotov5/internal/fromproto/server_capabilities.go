// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func ServerCapabilities(in *tfplugin5.ServerCapabilities) *tfprotov5.ServerCapabilities {
	if in == nil {
		return nil
	}

	return &tfprotov5.ServerCapabilities{
		GetProviderSchemaOptional: in.GetProviderSchemaOptional,
		PlanDestroy:               in.PlanDestroy,
	}
}
