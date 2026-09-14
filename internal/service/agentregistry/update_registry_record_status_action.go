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
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwactions "github.com/hashicorp/terraform-provider-aws/internal/framework/actions"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_agentregistry_update_registry_record_status, name="Update Registry Record Status")
func newUpdateRegistryRecordStatusAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &updateRegistryRecordStatusAction{}, nil
}

var (
	_ action.Action = (*updateRegistryRecordStatusAction)(nil)
)

type updateRegistryRecordStatusAction struct {
	framework.ActionWithModel[updateRegistryRecordStatusActionModel]
}

type updateRegistryRecordStatusActionModel struct {
	framework.WithRegionModel
	RecordID     types.String                                      `tfsdk:"record_id"`
	RegistryID   types.String                                      `tfsdk:"registry_id"`
	Status       fwtypes.StringEnum[awstypes.RegistryRecordStatus] `tfsdk:"status"`
	StatusReason types.String                                      `tfsdk:"status_reason"`
	Timeout      types.Int64                                       `tfsdk:"timeout"`
}

func (a *updateRegistryRecordStatusAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Updates the status of a registry record as part of the registry's curation workflow: approve or reject a record that is pending approval, or deprecate a record so that it is no longer discoverable. Deprecation is irreversible.",
		Attributes: map[string]schema.Attribute{
			"record_id": schema.StringAttribute{
				Description: "Identifier of the registry record (ID or ARN).",
				Required:    true,
			},
			"registry_id": schema.StringAttribute{
				Description: "Identifier of the registry containing the record (ID or ARN).",
				Required:    true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.RegistryRecordStatus](),
				Description: "Target status for the record. Valid values: APPROVED, REJECTED, DEPRECATED.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(enum.Slice(
						awstypes.RegistryRecordStatusApproved,
						awstypes.RegistryRecordStatusRejected,
						awstypes.RegistryRecordStatusDeprecated,
					)...),
				},
			},
			names.AttrStatusReason: schema.StringAttribute{
				Description: "Reason for the status change, for example why the record was approved, rejected, or deprecated.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			names.AttrTimeout: schema.Int64Attribute{
				Description: "Timeout in seconds to wait for a record that is still being created or updated to settle before changing its status (default: 600).",
				Optional:    true,
			},
		},
	}
}

func (a *updateRegistryRecordStatusAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config updateRegistryRecordStatusActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := a.Meta().AgentRegistryClient(ctx)

	registryID := fwflex.StringValueFromFramework(ctx, config.RegistryID)
	recordID := fwflex.StringValueFromFramework(ctx, config.RecordID)
	targetStatus := config.Status.ValueEnum()
	timeout := fwactions.TimeoutOr(config.Timeout, 10*time.Minute)

	tflog.Info(ctx, "Starting Agent Registry update registry record status action", map[string]any{
		"registry_id":     registryID,
		"record_id":       recordID,
		names.AttrStatus:  string(targetStatus),
		names.AttrTimeout: timeout.String(),
	})

	cb := fwactions.NewSendProgressFunc(resp)
	cb(ctx, "Checking status of Agent Registry Registry Record %s...", recordID)

	record, err := waitRegistryRecordSettled(ctx, conn, registryID, recordID, timeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Reading Agent Registry Registry Record",
			fmt.Sprintf("Could not read Agent Registry Registry Record (%s): %s", recordID, err),
		)
		return
	}

	// Setting the same status again succeeds but overwrites the status reason,
	// so treat it as a no-op to preserve the original curation record.
	if record.Status == targetStatus {
		cb(ctx, "Agent Registry Registry Record %s is already %s, nothing to change", recordID, targetStatus)
		return
	}

	cb(ctx, "Updating Agent Registry Registry Record %s status from %s to %s...", recordID, record.Status, targetStatus)

	input := agentregistrycontrol.UpdateRegistryRecordStatusInput{
		RegistryId:   aws.String(registryID),
		RecordId:     aws.String(recordID),
		Status:       targetStatus,
		StatusReason: fwflex.StringFromFramework(ctx, config.StatusReason),
	}
	out, err := conn.UpdateRegistryRecordStatus(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating Agent Registry Registry Record Status",
			fmt.Sprintf("Could not update Agent Registry Registry Record (%s) status from %s to %s: %s", recordID, record.Status, targetStatus, err),
		)
		return
	}

	cb(ctx, "Agent Registry Registry Record %s status is now %s", recordID, out.Status)

	tflog.Info(ctx, "Agent Registry update registry record status action completed", map[string]any{
		"registry_id":    registryID,
		"record_id":      recordID,
		names.AttrStatus: string(out.Status),
	})
}
