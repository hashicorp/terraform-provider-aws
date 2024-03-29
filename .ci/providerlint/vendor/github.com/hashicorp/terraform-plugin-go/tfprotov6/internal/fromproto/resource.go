// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
)

func ValidateResourceConfigRequest(in *tfplugin6.ValidateResourceConfig_Request) *tfprotov6.ValidateResourceConfigRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov6.ValidateResourceConfigRequest{
		Config:   DynamicValue(in.Config),
		TypeName: in.TypeName,
	}

	return resp
}

func UpgradeResourceStateRequest(in *tfplugin6.UpgradeResourceState_Request) *tfprotov6.UpgradeResourceStateRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov6.UpgradeResourceStateRequest{
		RawState: RawState(in.RawState),
		TypeName: in.TypeName,
		Version:  in.Version,
	}

	return resp
}

func ReadResourceRequest(in *tfplugin6.ReadResource_Request) *tfprotov6.ReadResourceRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov6.ReadResourceRequest{
		CurrentState: DynamicValue(in.CurrentState),
		Private:      in.Private,
		ProviderMeta: DynamicValue(in.ProviderMeta),
		TypeName:     in.TypeName,
	}

	return resp
}

func PlanResourceChangeRequest(in *tfplugin6.PlanResourceChange_Request) *tfprotov6.PlanResourceChangeRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov6.PlanResourceChangeRequest{
		Config:           DynamicValue(in.Config),
		PriorPrivate:     in.PriorPrivate,
		PriorState:       DynamicValue(in.PriorState),
		ProposedNewState: DynamicValue(in.ProposedNewState),
		ProviderMeta:     DynamicValue(in.ProviderMeta),
		TypeName:         in.TypeName,
	}

	return resp
}

func ApplyResourceChangeRequest(in *tfplugin6.ApplyResourceChange_Request) *tfprotov6.ApplyResourceChangeRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov6.ApplyResourceChangeRequest{
		Config:         DynamicValue(in.Config),
		PlannedPrivate: in.PlannedPrivate,
		PlannedState:   DynamicValue(in.PlannedState),
		PriorState:     DynamicValue(in.PriorState),
		ProviderMeta:   DynamicValue(in.ProviderMeta),
		TypeName:       in.TypeName,
	}

	return resp
}

func ImportResourceStateRequest(in *tfplugin6.ImportResourceState_Request) *tfprotov6.ImportResourceStateRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov6.ImportResourceStateRequest{
		TypeName: in.TypeName,
		ID:       in.Id,
	}

	return resp
}

func MoveResourceStateRequest(in *tfplugin6.MoveResourceState_Request) *tfprotov6.MoveResourceStateRequest {
	if in == nil {
		return nil
	}

	resp := &tfprotov6.MoveResourceStateRequest{
		SourcePrivate:         in.SourcePrivate,
		SourceProviderAddress: in.SourceProviderAddress,
		SourceSchemaVersion:   in.SourceSchemaVersion,
		SourceState:           RawState(in.SourceState),
		SourceTypeName:        in.SourceTypeName,
		TargetTypeName:        in.TargetTypeName,
	}

	return resp
}
