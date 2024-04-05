// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
)

func GetMetadataRequest(in *tfplugin6.GetMetadata_Request) *tfprotov6.GetMetadataRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov6.GetMetadataRequest{}

	return resp
}

func GetProviderSchemaRequest(in *tfplugin6.GetProviderSchema_Request) *tfprotov6.GetProviderSchemaRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov6.GetProviderSchemaRequest{}

	return resp
}

func ValidateProviderConfigRequest(in *tfplugin6.ValidateProviderConfig_Request) *tfprotov6.ValidateProviderConfigRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov6.ValidateProviderConfigRequest{
		Config: DynamicValue(in.Config),
	}

	return resp
}

func ConfigureProviderRequest(in *tfplugin6.ConfigureProvider_Request) *tfprotov6.ConfigureProviderRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov6.ConfigureProviderRequest{
		Config:           DynamicValue(in.Config),
		TerraformVersion: in.TerraformVersion,
	}

	return resp
}

func StopProviderRequest(in *tfplugin6.StopProvider_Request) *tfprotov6.StopProviderRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov6.StopProviderRequest{}

	return resp
}
