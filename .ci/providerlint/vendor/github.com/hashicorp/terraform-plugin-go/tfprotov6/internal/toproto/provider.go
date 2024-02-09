// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
)

func GetMetadata_Response(in *tfprotov6.GetMetadataResponse) *tfplugin6.GetMetadata_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.GetMetadata_Response{
		DataSources:        make([]*tfplugin6.GetMetadata_DataSourceMetadata, 0, len(in.DataSources)),
		Diagnostics:        Diagnostics(in.Diagnostics),
		Functions:          make([]*tfplugin6.GetMetadata_FunctionMetadata, 0, len(in.Functions)),
		Resources:          make([]*tfplugin6.GetMetadata_ResourceMetadata, 0, len(in.Resources)),
		ServerCapabilities: ServerCapabilities(in.ServerCapabilities),
	}

	for _, datasource := range in.DataSources {
		resp.DataSources = append(resp.DataSources, GetMetadata_DataSourceMetadata(&datasource))
	}

	for _, function := range in.Functions {
		resp.Functions = append(resp.Functions, GetMetadata_FunctionMetadata(&function))
	}

	for _, resource := range in.Resources {
		resp.Resources = append(resp.Resources, GetMetadata_ResourceMetadata(&resource))
	}

	return resp
}

func GetProviderSchema_Response(in *tfprotov6.GetProviderSchemaResponse) *tfplugin6.GetProviderSchema_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.GetProviderSchema_Response{
		DataSourceSchemas:  make(map[string]*tfplugin6.Schema, len(in.DataSourceSchemas)),
		Diagnostics:        Diagnostics(in.Diagnostics),
		Functions:          make(map[string]*tfplugin6.Function, len(in.Functions)),
		Provider:           Schema(in.Provider),
		ProviderMeta:       Schema(in.ProviderMeta),
		ResourceSchemas:    make(map[string]*tfplugin6.Schema, len(in.ResourceSchemas)),
		ServerCapabilities: ServerCapabilities(in.ServerCapabilities),
	}

	for name, schema := range in.ResourceSchemas {
		resp.ResourceSchemas[name] = Schema(schema)
	}

	for name, schema := range in.DataSourceSchemas {
		resp.DataSourceSchemas[name] = Schema(schema)
	}

	for name, function := range in.Functions {
		resp.Functions[name] = Function(function)
	}

	return resp
}

func ValidateProviderConfig_Response(in *tfprotov6.ValidateProviderConfigResponse) *tfplugin6.ValidateProviderConfig_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.ValidateProviderConfig_Response{
		Diagnostics: Diagnostics(in.Diagnostics),
	}

	return resp
}

func ConfigureProvider_Response(in *tfprotov6.ConfigureProviderResponse) *tfplugin6.ConfigureProvider_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.ConfigureProvider_Response{
		Diagnostics: Diagnostics(in.Diagnostics),
	}

	return resp
}

func StopProvider_Response(in *tfprotov6.StopProviderResponse) *tfplugin6.StopProvider_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.StopProvider_Response{
		Error: in.Error,
	}

	return resp
}
