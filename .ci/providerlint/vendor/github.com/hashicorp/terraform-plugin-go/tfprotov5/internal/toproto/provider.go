// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func GetMetadata_Response(in *tfprotov5.GetMetadataResponse) *tfplugin5.GetMetadata_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.GetMetadata_Response{
		DataSources:        make([]*tfplugin5.GetMetadata_DataSourceMetadata, 0, len(in.DataSources)),
		Diagnostics:        Diagnostics(in.Diagnostics),
		Functions:          make([]*tfplugin5.GetMetadata_FunctionMetadata, 0, len(in.Functions)),
		Resources:          make([]*tfplugin5.GetMetadata_ResourceMetadata, 0, len(in.Resources)),
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

func GetProviderSchema_Response(in *tfprotov5.GetProviderSchemaResponse) *tfplugin5.GetProviderSchema_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.GetProviderSchema_Response{
		DataSourceSchemas:  make(map[string]*tfplugin5.Schema, len(in.DataSourceSchemas)),
		Diagnostics:        Diagnostics(in.Diagnostics),
		Functions:          make(map[string]*tfplugin5.Function, len(in.Functions)),
		Provider:           Schema(in.Provider),
		ProviderMeta:       Schema(in.ProviderMeta),
		ResourceSchemas:    make(map[string]*tfplugin5.Schema, len(in.ResourceSchemas)),
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

func PrepareProviderConfig_Response(in *tfprotov5.PrepareProviderConfigResponse) *tfplugin5.PrepareProviderConfig_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.PrepareProviderConfig_Response{
		Diagnostics:    Diagnostics(in.Diagnostics),
		PreparedConfig: DynamicValue(in.PreparedConfig),
	}

	return resp
}

func Configure_Response(in *tfprotov5.ConfigureProviderResponse) *tfplugin5.Configure_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.Configure_Response{
		Diagnostics: Diagnostics(in.Diagnostics),
	}

	return resp
}

func Stop_Response(in *tfprotov5.StopProviderResponse) *tfplugin5.Stop_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.Stop_Response{
		Error: in.Error,
	}

	return resp
}
