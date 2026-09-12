// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package agentregistry

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol/types"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwactions "github.com/hashicorp/terraform-provider-aws/internal/framework/actions"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_agentregistry_submit_registry_record_for_approval, name="Submit Registry Record For Approval")
func newSubmitRegistryRecordForApprovalAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &submitRegistryRecordForApprovalAction{}, nil
}

var (
	_ action.Action = (*submitRegistryRecordForApprovalAction)(nil)
)

type submitRegistryRecordForApprovalAction struct {
	framework.ActionWithModel[submitRegistryRecordForApprovalActionModel]
}

type submitRegistryRecordForApprovalActionModel struct {
	framework.WithRegionModel
	RecordID   types.String `tfsdk:"record_id"`
	RegistryID types.String `tfsdk:"registry_id"`
	Timeout    types.Int64  `tfsdk:"timeout"`
}

func (a *submitRegistryRecordForApprovalAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Submits a DRAFT registry record for approval, moving it into the registry's approval workflow. Depending on the registry's approval configuration, the record is either auto-approved or set to PENDING_APPROVAL for a curator to approve or reject. Records already in the approval workflow are left unchanged.",
		Attributes: map[string]schema.Attribute{
			"record_id": schema.StringAttribute{
				Description: "Identifier of the registry record to submit (ID or ARN).",
				Required:    true,
			},
			"registry_id": schema.StringAttribute{
				Description: "Identifier of the registry containing the record (ID or ARN).",
				Required:    true,
			},
			names.AttrTimeout: schema.Int64Attribute{
				Description: "Timeout in seconds to wait for a record that is still being created or updated to settle before submitting it (default: 600).",
				Optional:    true,
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

	conn := a.Meta().AgentRegistryClient(ctx)

	registryID := fwflex.StringValueFromFramework(ctx, config.RegistryID)
	recordID := fwflex.StringValueFromFramework(ctx, config.RecordID)
	timeout := fwactions.TimeoutOr(config.Timeout, 10*time.Minute)

	tflog.Info(ctx, "Starting Agent Registry submit registry record for approval action", map[string]any{
		"registry_id":     registryID,
		"record_id":       recordID,
		names.AttrTimeout: timeout.String(),
	})

	cb := fwactions.NewSendProgressFunc(resp)
	cb(ctx, "Checking status of Agent Registry Registry Record %s...", recordID)

	// A record that is still being created or updated must settle into DRAFT
	// before it can be submitted. This is the common case when the action is
	// triggered by the record resource's after_update lifecycle event.
	record, err := waitRegistryRecordSettled(ctx, conn, registryID, recordID, timeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Reading Agent Registry Registry Record",
			fmt.Sprintf("Could not read Agent Registry Registry Record (%s): %s", recordID, err),
		)
		return
	}

	switch status := record.Status; status {
	case awstypes.RegistryRecordStatusDraft:
		// Proceed to submission.

	case awstypes.RegistryRecordStatusPendingApproval, awstypes.RegistryRecordStatusApproved:
		cb(ctx, "Agent Registry Registry Record %s is already %s, nothing to submit", recordID, status)
		return

	case awstypes.RegistryRecordStatusRejected:
		resp.Diagnostics.AddError(
			"Agent Registry Registry Record Rejected",
			fmt.Sprintf("Agent Registry Registry Record (%s) is REJECTED and cannot be resubmitted as-is. Update the record's content to return it to DRAFT, then submit it again. Status reason: %s", recordID, aws.ToString(record.StatusReason)),
		)
		return

	case awstypes.RegistryRecordStatusDeprecated:
		resp.Diagnostics.AddError(
			"Agent Registry Registry Record Deprecated",
			fmt.Sprintf("Agent Registry Registry Record (%s) is DEPRECATED, which is a terminal state. It cannot be submitted or modified; delete it or create a new record instead.", recordID),
		)
		return

	default:
		resp.Diagnostics.AddError(
			"Unexpected Agent Registry Registry Record Status",
			fmt.Sprintf("Agent Registry Registry Record (%s) is in status %s and cannot be submitted for approval. Status reason: %s", recordID, status, aws.ToString(record.StatusReason)),
		)
		return
	}

	cb(ctx, "Submitting Agent Registry Registry Record %s for approval...", recordID)

	input := agentregistrycontrol.SubmitRegistryRecordForApprovalInput{
		RegistryId: aws.String(registryID),
		RecordId:   aws.String(recordID),
	}
	out, err := conn.SubmitRegistryRecordForApproval(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Submitting Agent Registry Registry Record For Approval",
			fmt.Sprintf("Could not submit Agent Registry Registry Record (%s) for approval: %s", recordID, err),
		)
		return
	}

	cb(ctx, "Agent Registry Registry Record %s submitted for approval, resulting status: %s", recordID, out.Status)

	tflog.Info(ctx, "Agent Registry submit registry record for approval action completed", map[string]any{
		"registry_id":    registryID,
		"record_id":      recordID,
		names.AttrStatus: string(out.Status),
	})
}

// waitRegistryRecordSettled waits for a record that is being created or updated
// to reach a stable status, returning the record in that status.
func waitRegistryRecordSettled(ctx context.Context, conn *agentregistrycontrol.Client, registryID, recordID string, timeout time.Duration) (*agentregistrycontrol.GetRegistryRecordOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RegistryRecordStatusCreating, awstypes.RegistryRecordStatusUpdating),
		Target: enum.Slice(
			awstypes.RegistryRecordStatusDraft,
			awstypes.RegistryRecordStatusPendingApproval,
			awstypes.RegistryRecordStatusApproved,
			awstypes.RegistryRecordStatusRejected,
			awstypes.RegistryRecordStatusDeprecated,
			awstypes.RegistryRecordStatusCreateFailed,
			awstypes.RegistryRecordStatusUpdateFailed,
		),
		Refresh: statusRegistryRecord(conn, registryID, recordID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*agentregistrycontrol.GetRegistryRecordOutput); ok {
		return out, err
	}

	return nil, err
}
