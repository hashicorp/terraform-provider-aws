// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_backup_logically_air_gapped_vault", name="Logically Air Gapped Vault")
// @Tags(identifierAttribute="arn")
func newResourceLogicallyAirGappedVault(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceLogicallyAirGappedVault{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameLogicallyAirGappedVault = "Logically Air Gapped Vault"
)

type resourceLogicallyAirGappedVault struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceLogicallyAirGappedVault) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_backup_logically_air_gapped_vault"
}

func (r *resourceLogicallyAirGappedVault) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"max_retention_days": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"min_retention_days": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceLogicallyAirGappedVault) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BackupClient(ctx)

	var plan resourceLogicallyAirGappedVaultData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &backup.CreateLogicallyAirGappedBackupVaultInput{
		BackupVaultName:  plan.BackupVaultName.ValueStringPointer(),
		MaxRetentionDays: plan.MaxRetentionDays.ValueInt64Pointer(),
		MinRetentionDays: plan.MinRetentionDays.ValueInt64Pointer(),
		BackupVaultTags:  getTagsIn(ctx),
	}

	out, err := conn.CreateLogicallyAirGappedBackupVault(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionCreating, ResNameLogicallyAirGappedVault, plan.BackupVaultName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionCreating, ResNameLogicallyAirGappedVault, plan.BackupVaultName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringToFramework(ctx, out.BackupVaultName)
	plan.ARN = flex.StringToFramework(ctx, out.BackupVaultArn)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitLogicallyAirGappedVaultCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionWaitingForCreation, ResNameLogicallyAirGappedVault, plan.BackupVaultName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceLogicallyAirGappedVault) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BackupClient(ctx)

	var state resourceLogicallyAirGappedVaultData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findVaultByName(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionSetting, ResNameLogicallyAirGappedVault, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.BackupVaultArn)
	state.BackupVaultName = flex.StringToFramework(ctx, out.BackupVaultName)
	state.MaxRetentionDays = flex.Int64ToFramework(ctx, out.MaxRetentionDays)
	state.MinRetentionDays = flex.Int64ToFramework(ctx, out.MinRetentionDays)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceLogicallyAirGappedVault) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceLogicallyAirGappedVaultData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Tags only.

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceLogicallyAirGappedVault) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BackupClient(ctx)

	var state resourceLogicallyAirGappedVaultData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &backup.DeleteBackupVaultInput{
		BackupVaultName: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteBackupVault(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.MessageContains(err, "AccessDeniedException", "Insufficient privileges to perform this action") {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Backup, create.ErrActionDeleting, ResNameLogicallyAirGappedVault, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceLogicallyAirGappedVault) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func (r *resourceLogicallyAirGappedVault) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type resourceLogicallyAirGappedVaultData struct {
	ARN              types.String   `tfsdk:"arn"`
	ID               types.String   `tfsdk:"id"`
	MaxRetentionDays types.Int64    `tfsdk:"max_retention_days"`
	MinRetentionDays types.Int64    `tfsdk:"min_retention_days"`
	BackupVaultName  types.String   `tfsdk:"name"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
	Tags             types.Map      `tfsdk:"tags"`
	TagsAll          types.Map      `tfsdk:"tags_all"`
}

func waitLogicallyAirGappedVaultCreated(ctx context.Context, conn *backup.Client, id string, timeout time.Duration) (*backup.DescribeBackupVaultOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.VaultStateCreating),
		Target:                    enum.Slice(awstypes.VaultStateAvailable),
		Refresh:                   statusLogicallyAirGappedVault(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*backup.DescribeBackupVaultOutput); ok {
		return out, err
	}

	return nil, err
}

func statusLogicallyAirGappedVault(ctx context.Context, conn *backup.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findVaultByName(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.VaultState), nil
	}
}
