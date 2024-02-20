// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
)

func GetMetadata_DataSourceMetadata(in *tfprotov6.DataSourceMetadata) *tfplugin6.GetMetadata_DataSourceMetadata {
	if in == nil {
		return nil
	}

	return &tfplugin6.GetMetadata_DataSourceMetadata{
		TypeName: in.TypeName,
	}
}

func ValidateDataResourceConfig_Response(in *tfprotov6.ValidateDataResourceConfigResponse) *tfplugin6.ValidateDataResourceConfig_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.ValidateDataResourceConfig_Response{
		Diagnostics: Diagnostics(in.Diagnostics),
	}

	return resp
}

func ReadDataSource_Response(in *tfprotov6.ReadDataSourceResponse) *tfplugin6.ReadDataSource_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.ReadDataSource_Response{
		Diagnostics: Diagnostics(in.Diagnostics),
		State:       DynamicValue(in.State),
	}

	return resp
}
