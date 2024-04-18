// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func GetMetadata_DataSourceMetadata(in *tfprotov5.DataSourceMetadata) *tfplugin5.GetMetadata_DataSourceMetadata {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.GetMetadata_DataSourceMetadata{
		TypeName: in.TypeName,
	}

	return resp
}

func ValidateDataSourceConfig_Response(in *tfprotov5.ValidateDataSourceConfigResponse) *tfplugin5.ValidateDataSourceConfig_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.ValidateDataSourceConfig_Response{
		Diagnostics: Diagnostics(in.Diagnostics),
	}

	return resp
}

func ReadDataSource_Response(in *tfprotov5.ReadDataSourceResponse) *tfplugin5.ReadDataSource_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.ReadDataSource_Response{
		Diagnostics: Diagnostics(in.Diagnostics),
		State:       DynamicValue(in.State),
	}

	return resp
}
