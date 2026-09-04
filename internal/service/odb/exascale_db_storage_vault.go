// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_odb_exascale_db_storage_vault", name="Exascale DB Storage Vault")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/odb/types;awstypes;awstypes.ExascaleDbStorageVault")
// @Testing(preCheckRegion="us-east-1;eu-west-1")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator="testAccRandomExascaleDBStorageVaultDisplayName(t)")
func newExascaleDBStorageVaultResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &exascaleDBStorageVaultResource{}

	r.SetDefaultCreateTimeout(24 * time.Hour)
	r.SetDefaultUpdateTimeout(24 * time.Hour)
	r.SetDefaultDeleteTimeout(24 * time.Hour)

	return r, nil
}

const (
	ResNameExascaleDBStorageVault = "Exascale DB Storage Vault"
)

type exascaleDBStorageVaultResource struct {
	framework.ResourceWithModel[exascaleDBStorageVaultResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *exascaleDBStorageVaultResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"additional_flash_cache_in_percent": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int32{
					int32validator.AtLeast(0),
				},
				Description: "Additional flash cache percentage for the Exascale DB storage vault.",
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"autoscale_limit_in_gbs": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int32{
					int32validator.AtLeast(0),
				},
				Description: "Autoscale limit, in gigabytes (GB), for the Exascale DB storage vault.",
			},
			names.AttrAvailabilityZone: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				Description: "Availability Zone for the Exascale DB storage vault.",
			},
			"availability_zone_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				Description: "Availability Zone ID for the Exascale DB storage vault.",
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 400),
				},
				Description: "Description of the Exascale DB storage vault.",
			},
			names.AttrDisplayName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[a-zA-Z_](?:[a-zA-Z0-9_]|-[a-zA-Z0-9_])*-?$`),
						"must start with a letter or underscore, contain only letters, numbers, underscores, or hyphens, and must not contain consecutive hyphens",
					),
				},
				Description: "User-friendly name for the Exascale DB storage vault.",
			},
			"high_capacity_database_storage_total_size_in_gbs": schema.Int32Attribute{
				Required: true,
				Validators: []validator.Int32{
					int32validator.AtLeast(0),
				},
				Description: "Total size of the high-capacity database storage, in gigabytes (GB).",
			},
			names.AttrID: framework.IDAttribute(),
			"is_autoscale_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Whether autoscaling is enabled for the Exascale DB storage vault.",
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"time_zone": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				Description: "Time zone for the Exascale DB storage vault.",
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *exascaleDBStorageVaultResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ODBClient(ctx)

	var plan exascaleDBStorageVaultResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input odb.CreateExascaleDbStorageVaultInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("ExascaleDbStorageVault")))
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateExascaleDbStorageVault(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.DisplayName.ValueString())
		return
	}
	if out == nil || out.ExascaleDbStorageVaultId == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.DisplayName.ValueString())
		return
	}

	id := aws.ToString(out.ExascaleDbStorageVaultId)
	plan.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), id)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	created, err := waitExascaleDBStorageVaultCreated(ctx, conn, id, createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.DisplayName.ValueString())
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, created, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *exascaleDBStorageVaultResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ODBClient(ctx)

	var state exascaleDBStorageVaultResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findExascaleDBStorageVaultByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *exascaleDBStorageVaultResource) flatten(ctx context.Context, exascaleDBStorageVault *awstypes.ExascaleDbStorageVault, data *exascaleDBStorageVaultResourceModel) (diags diag.Diagnostics) {
	diags.Append(flex.Flatten(ctx, exascaleDBStorageVault, data, flex.WithFieldNamePrefix("ExascaleDbStorageVault"))...)
	if exascaleDBStorageVault.HighCapacityDatabaseStorage != nil {
		data.HighCapacityDatabaseStorageTotalSizeInGBs = types.Int32PointerValue(exascaleDBStorageVault.HighCapacityDatabaseStorage.TotalSizeInGBs)
	}

	return diags
}

func (r *exascaleDBStorageVaultResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ODBClient(ctx)

	var plan, state exascaleDBStorageVaultResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if !diff.HasChanges() {
		smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
		return
	}

	input := odb.UpdateExascaleDbStorageVaultInput{
		ExascaleDbStorageVaultId: plan.ID.ValueStringPointer(),
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, diff.IgnoredFieldNamesOpts()...))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.UpdateExascaleDbStorageVault(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}
	if out == nil || out.ExascaleDbStorageVaultId == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ID.String())
		return
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	updated, err := waitExascaleDBStorageVaultUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, updated, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *exascaleDBStorageVaultResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ODBClient(ctx)

	var state exascaleDBStorageVaultResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.DeleteExascaleDbStorageVaultInput{
		ExascaleDbStorageVaultId: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteExascaleDbStorageVault(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitExascaleDBStorageVaultDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

func waitExascaleDBStorageVaultCreated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*awstypes.ExascaleDbStorageVault, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ResourceStatusProvisioning),
		Target:                    enum.Slice(awstypes.ResourceStatusAvailable),
		Refresh:                   statusExascaleDBStorageVault(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ExascaleDbStorageVault); ok {
		if err != nil && aws.ToString(out.StatusReason) != "" {
			retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		}

		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitExascaleDBStorageVaultUpdated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*awstypes.ExascaleDbStorageVault, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ResourceStatusUpdating),
		Target:                    enum.Slice(awstypes.ResourceStatusAvailable),
		Refresh:                   statusExascaleDBStorageVault(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ExascaleDbStorageVault); ok {
		if err != nil && aws.ToString(out.StatusReason) != "" {
			retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		}

		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitExascaleDBStorageVaultDeleted(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*awstypes.ExascaleDbStorageVault, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceStatusTerminating),
		Target:  []string{},
		Refresh: statusExascaleDBStorageVault(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ExascaleDbStorageVault); ok {
		if err != nil && aws.ToString(out.StatusReason) != "" {
			retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		}

		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusExascaleDBStorageVault(conn *odb.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findExascaleDBStorageVaultByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findExascaleDBStorageVaultByID(ctx context.Context, conn *odb.Client, id string) (*awstypes.ExascaleDbStorageVault, error) {
	input := odb.GetExascaleDbStorageVaultInput{
		ExascaleDbStorageVaultId: aws.String(id),
	}

	out, err := conn.GetExascaleDbStorageVault(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.ExascaleDbStorageVault == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.ExascaleDbStorageVault, nil
}

type exascaleDBStorageVaultResourceModel struct {
	framework.WithRegionModel
	AdditionalFlashCacheInPercent             types.Int32    `tfsdk:"additional_flash_cache_in_percent"`
	ARN                                       types.String   `tfsdk:"arn"`
	AutoscaleLimitInGBs                       types.Int32    `tfsdk:"autoscale_limit_in_gbs"`
	AvailabilityZone                          types.String   `tfsdk:"availability_zone"`
	AvailabilityZoneID                        types.String   `tfsdk:"availability_zone_id"`
	Description                               types.String   `tfsdk:"description"`
	DisplayName                               types.String   `tfsdk:"display_name"`
	HighCapacityDatabaseStorageTotalSizeInGBs types.Int32    `tfsdk:"high_capacity_database_storage_total_size_in_gbs"`
	ID                                        types.String   `tfsdk:"id"`
	IsAutoscaleEnabled                        types.Bool     `tfsdk:"is_autoscale_enabled"`
	Tags                                      tftags.Map     `tfsdk:"tags"`
	TagsAll                                   tftags.Map     `tfsdk:"tags_all"`
	Timeouts                                  timeouts.Value `tfsdk:"timeouts"`
	TimeZone                                  types.String   `tfsdk:"time_zone"`
}
