// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func GetMetadataRequest(in *tfplugin5.GetMetadata_Request) *tfprotov5.GetMetadataRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov5.GetMetadataRequest{}

	return resp
}

func GetProviderSchemaRequest(in *tfplugin5.GetProviderSchema_Request) *tfprotov5.GetProviderSchemaRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov5.GetProviderSchemaRequest{}

	return resp
}

func PrepareProviderConfigRequest(in *tfplugin5.PrepareProviderConfig_Request) *tfprotov5.PrepareProviderConfigRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov5.PrepareProviderConfigRequest{
		Config: DynamicValue(in.Config),
	}

	return resp
}

func ConfigureProviderRequest(in *tfplugin5.Configure_Request) *tfprotov5.ConfigureProviderRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov5.ConfigureProviderRequest{
		Config:             DynamicValue(in.Config),
		TerraformVersion:   in.TerraformVersion,
		ClientCapabilities: ConfigureProviderClientCapabilities(in.ClientCapabilities),
	}

	return resp
}

func StopProviderRequest(in *tfplugin5.Stop_Request) *tfprotov5.StopProviderRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov5.StopProviderRequest{}

	return resp
}
