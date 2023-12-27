// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func ServerCapabilities(in *tfprotov5.ServerCapabilities) *tfplugin5.ServerCapabilities {
	if in == nil {
		return nil
	}

	return &tfplugin5.ServerCapabilities{
		GetProviderSchemaOptional: in.GetProviderSchemaOptional,
		PlanDestroy:               in.PlanDestroy,
	}
}
