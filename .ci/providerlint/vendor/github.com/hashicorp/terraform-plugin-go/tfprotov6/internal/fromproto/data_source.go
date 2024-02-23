// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
)

func ValidateDataResourceConfigRequest(in *tfplugin6.ValidateDataResourceConfig_Request) *tfprotov6.ValidateDataResourceConfigRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov6.ValidateDataResourceConfigRequest{
		Config:   DynamicValue(in.Config),
		TypeName: in.TypeName,
	}

	return resp
}

func ReadDataSourceRequest(in *tfplugin6.ReadDataSource_Request) *tfprotov6.ReadDataSourceRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov6.ReadDataSourceRequest{
		Config:       DynamicValue(in.Config),
		ProviderMeta: DynamicValue(in.ProviderMeta),
		TypeName:     in.TypeName,
	}

	return resp
}
