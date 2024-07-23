// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func ValidateDataSourceConfigRequest(in *tfplugin5.ValidateDataSourceConfig_Request) *tfprotov5.ValidateDataSourceConfigRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov5.ValidateDataSourceConfigRequest{
		Config:   DynamicValue(in.Config),
		TypeName: in.TypeName,
	}

	return resp
}

func ReadDataSourceRequest(in *tfplugin5.ReadDataSource_Request) *tfprotov5.ReadDataSourceRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov5.ReadDataSourceRequest{
		Config:             DynamicValue(in.Config),
		ProviderMeta:       DynamicValue(in.ProviderMeta),
		TypeName:           in.TypeName,
		ClientCapabilities: ReadDataSourceClientCapabilities(in.ClientCapabilities),
	}

	return resp
}
