// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
)

func ServerCapabilities(in *tfplugin6.ServerCapabilities) *tfprotov6.ServerCapabilities {
	if in == nil {
		return nil
	}

	return &tfprotov6.ServerCapabilities{
		GetProviderSchemaOptional: in.GetProviderSchemaOptional,
		PlanDestroy:               in.PlanDestroy,
	}
}
