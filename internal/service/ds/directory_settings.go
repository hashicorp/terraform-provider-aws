// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_directory_service_directory_settings", name="Directory Settings")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/directoryservice;directoryservice.DescribeSettingsOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
func newDirectorySettingsResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &directorySettingsResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDirectorySettings = "Directory Settings"
)

type directorySettingsResource struct {
	framework.ResourceWithModel[directorySettingsResourceModel]
	framework.WithTimeouts
	framework.WithNoOpDelete
}

func (r *directorySettingsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"directory_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"setting": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[directorySettingModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
						names.AttrValue: schema.StringAttribute{
							Required: true,
						},
						"applied_value": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"request_status": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						names.AttrType: schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func (r *directorySettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("directory_id"), req, resp)
}

func (r *directorySettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DSClient(ctx)

	var plan directorySettingsResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	directoryID := plan.DirectoryID.ValueString()

	var input directoryservice.UpdateSettingsInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateSettings(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, directoryID)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	entries, err := waitSettingsUpdated(ctx, conn, directoryID, createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, directoryID)
		return
	}

	// DescribeSettings returns SettingEntries; field name differs so we merge
	// computed attributes back into the plan manually.
	requested, d := plan.Settings.ToSlice(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Settings = flattenSettingEntries(ctx, entries, requested, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *directorySettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DSClient(ctx)

	var state directorySettingsResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	directoryID := state.DirectoryID.ValueString()

	entries, err := findDirectorySettingsByDirectoryID(ctx, conn, directoryID)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, directoryID)
		return
	}

	// Preserve the requested settings from state to match against returned entries.
	existing, d := state.Settings.ToSlice(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Settings = flattenSettingEntries(ctx, entries, existing, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *directorySettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DSClient(ctx)

	var plan, state directorySettingsResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		directoryID := plan.DirectoryID.ValueString()

		var input directoryservice.UpdateSettingsInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateSettings(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, directoryID)
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		entries, err := waitSettingsUpdated(ctx, conn, directoryID, updateTimeout)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, directoryID)
			return
		}

		// DescribeSettings returns SettingEntries; field name differs so we merge
		// computed attributes back into the plan manually.
		requested, d := plan.Settings.ToSlice(ctx)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}

		plan.Settings = flattenSettingEntries(ctx, entries, requested, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func waitSettingsUpdated(ctx context.Context, conn *directoryservice.Client, directoryID string, timeout time.Duration) ([]awstypes.SettingEntry, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			string(awstypes.DirectoryConfigurationStatusRequested),
			string(awstypes.DirectoryConfigurationStatusUpdating),
		},
		Target: []string{
			string(awstypes.DirectoryConfigurationStatusUpdated),
			string(awstypes.DirectoryConfigurationStatusDefault),
		},
		Refresh: statusDirectorySettings(ctx, conn, directoryID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if entries, ok := outputRaw.([]awstypes.SettingEntry); ok {
		return entries, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusDirectorySettings(ctx context.Context, conn *directoryservice.Client, directoryID string) retry.StateRefreshFunc {
	return func(_ context.Context) (any, string, error) {
		entries, err := findDirectorySettingsByDirectoryID(ctx, conn, directoryID)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		// Surface the worst status across all entries.
		overall := overallSettingsStatus(entries)
		return entries, overall, nil
	}
}

// overallSettingsStatus returns the most-pending status across all entries.
// Priority: Updating > Requested > Failed > Updated/Default.
func overallSettingsStatus(entries []awstypes.SettingEntry) string {
	result := string(awstypes.DirectoryConfigurationStatusUpdated)
	for _, e := range entries {
		switch e.RequestStatus {
		case awstypes.DirectoryConfigurationStatusUpdating:
			return string(awstypes.DirectoryConfigurationStatusUpdating)
		case awstypes.DirectoryConfigurationStatusRequested:
			result = string(awstypes.DirectoryConfigurationStatusRequested)
		case awstypes.DirectoryConfigurationStatusFailed:
			if result != string(awstypes.DirectoryConfigurationStatusRequested) {
				result = string(awstypes.DirectoryConfigurationStatusFailed)
			}
		}
	}
	return result
}

func findDirectorySettingsByDirectoryID(ctx context.Context, conn *directoryservice.Client, directoryID string) ([]awstypes.SettingEntry, error) {
	input := directoryservice.DescribeSettingsInput{
		DirectoryId: aws.String(directoryID),
	}

	var entries []awstypes.SettingEntry
	for {
		out, err := conn.DescribeSettings(ctx, &input)
		if err != nil {
			return nil, smarterr.NewError(err)
		}
		if out == nil {
			break
		}
		entries = append(entries, out.SettingEntries...)
		if out.NextToken == nil {
			break
		}
		input.NextToken = out.NextToken
	}

	if len(entries) == 0 {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return entries, nil
}

func flattenSettingEntries(ctx context.Context, entries []awstypes.SettingEntry, requested []*directorySettingModel, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[directorySettingModel] {
	// Build a lookup from name → SettingEntry for fast access.
	byName := make(map[string]awstypes.SettingEntry, len(entries))
	for _, e := range entries {
		if e.Name != nil {
			byName[aws.ToString(e.Name)] = e
		}
	}

	models := make([]*directorySettingModel, len(requested))
	for i, req := range requested {
		name := req.Name.ValueString()
		m := &directorySettingModel{
			Name:  req.Name,
			Value: req.Value,
		}
		if e, ok := byName[name]; ok {
			m.AppliedValue = types.StringPointerValue(e.AppliedValue)
			m.RequestStatus = types.StringValue(string(e.RequestStatus))
			m.Type = types.StringPointerValue(e.Type)
		}
		models[i] = m
	}

	result, d := fwtypes.NewListNestedObjectValueOfSlice(ctx, models, nil)
	diags.Append(d...)
	return result
}

type directorySettingsResourceModel struct {
	framework.WithRegionModel
	DirectoryID types.String                                           `tfsdk:"directory_id"`
	Settings    fwtypes.ListNestedObjectValueOf[directorySettingModel] `tfsdk:"setting"`
	Timeouts    timeouts.Value                                         `tfsdk:"timeouts"`
}

type directorySettingModel struct {
	Name          types.String `tfsdk:"name"`
	Value         types.String `tfsdk:"value"`
	AppliedValue  types.String `tfsdk:"applied_value"`
	RequestStatus types.String `tfsdk:"request_status"`
	Type          types.String `tfsdk:"type"`
}
