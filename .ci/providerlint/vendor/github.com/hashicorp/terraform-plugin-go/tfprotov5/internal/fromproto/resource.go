// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func ValidateResourceTypeConfigRequest(in *tfplugin5.ValidateResourceTypeConfig_Request) *tfprotov5.ValidateResourceTypeConfigRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov5.ValidateResourceTypeConfigRequest{
		Config:   DynamicValue(in.Config),
		TypeName: in.TypeName,
	}

	return resp
}

func UpgradeResourceStateRequest(in *tfplugin5.UpgradeResourceState_Request) *tfprotov5.UpgradeResourceStateRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov5.UpgradeResourceStateRequest{
		RawState: RawState(in.RawState),
		TypeName: in.TypeName,
		Version:  in.Version,
	}

	return resp
}

func ReadResourceRequest(in *tfplugin5.ReadResource_Request) *tfprotov5.ReadResourceRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov5.ReadResourceRequest{
		CurrentState: DynamicValue(in.CurrentState),
		Private:      in.Private,
		ProviderMeta: DynamicValue(in.ProviderMeta),
		TypeName:     in.TypeName,
	}

	return resp
}

func PlanResourceChangeRequest(in *tfplugin5.PlanResourceChange_Request) *tfprotov5.PlanResourceChangeRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov5.PlanResourceChangeRequest{
		Config:           DynamicValue(in.Config),
		PriorPrivate:     in.PriorPrivate,
		PriorState:       DynamicValue(in.PriorState),
		ProposedNewState: DynamicValue(in.ProposedNewState),
		ProviderMeta:     DynamicValue(in.ProviderMeta),
		TypeName:         in.TypeName,
	}

	return resp
}

func ApplyResourceChangeRequest(in *tfplugin5.ApplyResourceChange_Request) *tfprotov5.ApplyResourceChangeRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov5.ApplyResourceChangeRequest{
		Config:         DynamicValue(in.Config),
		PlannedPrivate: in.PlannedPrivate,
		PlannedState:   DynamicValue(in.PlannedState),
		PriorState:     DynamicValue(in.PriorState),
		ProviderMeta:   DynamicValue(in.ProviderMeta),
		TypeName:       in.TypeName,
	}

	return resp
}

func ImportResourceStateRequest(in *tfplugin5.ImportResourceState_Request) *tfprotov5.ImportResourceStateRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov5.ImportResourceStateRequest{
		TypeName: in.TypeName,
		ID:       in.Id,
	}

	return resp
}

func MoveResourceStateRequest(in *tfplugin5.MoveResourceState_Request) *tfprotov5.MoveResourceStateRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov5.MoveResourceStateRequest{
		SourcePrivate:         in.SourcePrivate,
		SourceProviderAddress: in.SourceProviderAddress,
		SourceSchemaVersion:   in.SourceSchemaVersion,
		SourceState:           RawState(in.SourceState),
		SourceTypeName:        in.SourceTypeName,
		TargetTypeName:        in.TargetTypeName,
	}

	return resp
}
