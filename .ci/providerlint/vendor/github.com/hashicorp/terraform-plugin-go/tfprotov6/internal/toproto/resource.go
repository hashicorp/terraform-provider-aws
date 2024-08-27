// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
)

func GetMetadata_ResourceMetadata(in *tfprotov6.ResourceMetadata) *tfplugin6.GetMetadata_ResourceMetadata {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.GetMetadata_ResourceMetadata{
		TypeName: in.TypeName,
	}

	return resp
}

func ValidateResourceConfig_Response(in *tfprotov6.ValidateResourceConfigResponse) *tfplugin6.ValidateResourceConfig_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.ValidateResourceConfig_Response{
		Diagnostics: Diagnostics(in.Diagnostics),
	}

	return resp
}

func UpgradeResourceState_Response(in *tfprotov6.UpgradeResourceStateResponse) *tfplugin6.UpgradeResourceState_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.UpgradeResourceState_Response{
		Diagnostics:   Diagnostics(in.Diagnostics),
		UpgradedState: DynamicValue(in.UpgradedState),
	}

	return resp
}

func ReadResource_Response(in *tfprotov6.ReadResourceResponse) *tfplugin6.ReadResource_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.ReadResource_Response{
		Diagnostics: Diagnostics(in.Diagnostics),
		NewState:    DynamicValue(in.NewState),
		Private:     in.Private,
		Deferred:    Deferred(in.Deferred),
	}

	return resp
}

func PlanResourceChange_Response(in *tfprotov6.PlanResourceChangeResponse) *tfplugin6.PlanResourceChange_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.PlanResourceChange_Response{
		Diagnostics:      Diagnostics(in.Diagnostics),
		LegacyTypeSystem: in.UnsafeToUseLegacyTypeSystem, //nolint:staticcheck
		PlannedPrivate:   in.PlannedPrivate,
		PlannedState:     DynamicValue(in.PlannedState),
		RequiresReplace:  AttributePaths(in.RequiresReplace),
		Deferred:         Deferred(in.Deferred),
	}

	return resp
}

func ApplyResourceChange_Response(in *tfprotov6.ApplyResourceChangeResponse) *tfplugin6.ApplyResourceChange_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.ApplyResourceChange_Response{
		Diagnostics:      Diagnostics(in.Diagnostics),
		LegacyTypeSystem: in.UnsafeToUseLegacyTypeSystem, //nolint:staticcheck
		NewState:         DynamicValue(in.NewState),
		Private:          in.Private,
	}

	return resp
}

func ImportResourceState_Response(in *tfprotov6.ImportResourceStateResponse) *tfplugin6.ImportResourceState_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.ImportResourceState_Response{
		Diagnostics:       Diagnostics(in.Diagnostics),
		ImportedResources: ImportResourceState_ImportedResources(in.ImportedResources),
		Deferred:          Deferred(in.Deferred),
	}

	return resp
}

func ImportResourceState_ImportedResource(in *tfprotov6.ImportedResource) *tfplugin6.ImportResourceState_ImportedResource {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.ImportResourceState_ImportedResource{
		Private:  in.Private,
		State:    DynamicValue(in.State),
		TypeName: in.TypeName,
	}

	return resp
}

func ImportResourceState_ImportedResources(in []*tfprotov6.ImportedResource) []*tfplugin6.ImportResourceState_ImportedResource {
	resp := make([]*tfplugin6.ImportResourceState_ImportedResource, 0, len(in))

	for _, i := range in {
		resp = append(resp, ImportResourceState_ImportedResource(i))
	}

	return resp
}

func MoveResourceState_Response(in *tfprotov6.MoveResourceStateResponse) *tfplugin6.MoveResourceState_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.MoveResourceState_Response{
		Diagnostics:   Diagnostics(in.Diagnostics),
		TargetPrivate: in.TargetPrivate,
		TargetState:   DynamicValue(in.TargetState),
	}

	return resp
}
