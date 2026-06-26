// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwactions "github.com/hashicorp/terraform-provider-aws/internal/framework/actions"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_bedrockagentcore_submit_registry_record_for_approval, name="Submit Registry Record For Approval")
func newSubmitRegistryRecordForApprovalAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &submitRegistryRecordForApprovalAction{}, nil
}

type submitRegistryRecordForApprovalAction struct {
	framework.ActionWithModel[submitRegistryRecordForApprovalActionModel]
}

func (a *submitRegistryRecordForApprovalAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Submits an AWS Bedrock AgentCore Registry Record for approval.",
		Attributes: map[string]schema.Attribute{
			"record_id": schema.StringAttribute{
				Description: "Registry record ID.",
				Required:    true,
			},
			"registry_id": schema.StringAttribute{
				Description: "Registry ID.",
				Required:    true,
			},
		},
	}
}

func (a *submitRegistryRecordForApprovalAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config submitRegistryRecordForApprovalActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := a.Meta().BedrockAgentCoreClient(ctx)

	registryID, recordID := fwflex.StringValueFromFramework(ctx, config.RegistryID), fwflex.StringValueFromFramework(ctx, config.RecordID)
	tflog.Info(ctx, "Starting Bedrock AgentCore submit registry record for approval action", map[string]any{
		"registry_id": registryID,
		"record_id":   recordID,
	})

	cb := fwactions.NewSendProgressFunc(resp)
	cb(ctx, "Submitting registry record (%s) for approval...", recordID)

	input := bedrockagentcorecontrol.SubmitRegistryRecordForApprovalInput{
		RecordId:   aws.String(recordID),
		RegistryId: aws.String(registryID),
	}
	output, err := conn.SubmitRegistryRecordForApproval(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("submitting Bedrock AgentCore Registry Record (%s) for approval", recordID), err.Error())

		return
	}

	status := output.Status
	cb(ctx, "Registry record (%s) status: %s", recordID, status)

	tflog.Info(ctx, "Bedrock AgentCore submit registry record for approval action completed successfully", map[string]any{
		"registry_id":    registryID,
		"record_id":      recordID,
		names.AttrStatus: status,
	})
}

type submitRegistryRecordForApprovalActionModel struct {
	framework.WithRegionModel
	RecordID   types.String `tfsdk:"record_id"`
	RegistryID types.String `tfsdk:"registry_id"`
}
