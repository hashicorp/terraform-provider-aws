// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func GetMetadata_ResourceMetadata(in *tfprotov5.ResourceMetadata) *tfplugin5.GetMetadata_ResourceMetadata {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.GetMetadata_ResourceMetadata{
		TypeName: in.TypeName,
	}

	return resp
}

func ValidateResourceTypeConfig_Response(in *tfprotov5.ValidateResourceTypeConfigResponse) *tfplugin5.ValidateResourceTypeConfig_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.ValidateResourceTypeConfig_Response{
		Diagnostics: Diagnostics(in.Diagnostics),
	}

	return resp
}

func UpgradeResourceState_Response(in *tfprotov5.UpgradeResourceStateResponse) *tfplugin5.UpgradeResourceState_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.UpgradeResourceState_Response{
		Diagnostics:   Diagnostics(in.Diagnostics),
		UpgradedState: DynamicValue(in.UpgradedState),
	}

	return resp
}

func ReadResource_Response(in *tfprotov5.ReadResourceResponse) *tfplugin5.ReadResource_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.ReadResource_Response{
		Diagnostics: Diagnostics(in.Diagnostics),
		NewState:    DynamicValue(in.NewState),
		Private:     in.Private,
	}

	return resp
}

func PlanResourceChange_Response(in *tfprotov5.PlanResourceChangeResponse) *tfplugin5.PlanResourceChange_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.PlanResourceChange_Response{
		Diagnostics:      Diagnostics(in.Diagnostics),
		LegacyTypeSystem: in.UnsafeToUseLegacyTypeSystem, //nolint:staticcheck
		PlannedPrivate:   in.PlannedPrivate,
		PlannedState:     DynamicValue(in.PlannedState),
		RequiresReplace:  AttributePaths(in.RequiresReplace),
	}

	return resp
}

func ApplyResourceChange_Response(in *tfprotov5.ApplyResourceChangeResponse) *tfplugin5.ApplyResourceChange_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.ApplyResourceChange_Response{
		Diagnostics:      Diagnostics(in.Diagnostics),
		LegacyTypeSystem: in.UnsafeToUseLegacyTypeSystem, //nolint:staticcheck
		NewState:         DynamicValue(in.NewState),
		Private:          in.Private,
	}

	return resp
}

func ImportResourceState_Response(in *tfprotov5.ImportResourceStateResponse) *tfplugin5.ImportResourceState_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.ImportResourceState_Response{
		Diagnostics:       Diagnostics(in.Diagnostics),
		ImportedResources: ImportResourceState_ImportedResources(in.ImportedResources),
	}

	return resp
}

func ImportResourceState_ImportedResource(in *tfprotov5.ImportedResource) *tfplugin5.ImportResourceState_ImportedResource {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.ImportResourceState_ImportedResource{
		Private:  in.Private,
		State:    DynamicValue(in.State),
		TypeName: in.TypeName,
	}

	return resp
}

func ImportResourceState_ImportedResources(in []*tfprotov5.ImportedResource) []*tfplugin5.ImportResourceState_ImportedResource {
	resp := make([]*tfplugin5.ImportResourceState_ImportedResource, 0, len(in))

	for _, i := range in {
		resp = append(resp, ImportResourceState_ImportedResource(i))
	}

	return resp
}

func MoveResourceState_Response(in *tfprotov5.MoveResourceStateResponse) *tfplugin5.MoveResourceState_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.MoveResourceState_Response{
		Diagnostics:   Diagnostics(in.Diagnostics),
		TargetPrivate: in.TargetPrivate,
		TargetState:   DynamicValue(in.TargetState),
	}

	return resp
}
