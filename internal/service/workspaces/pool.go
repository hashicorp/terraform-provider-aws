// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_workspaces_pool", name="Pool")
// @Tags(identifierAttribute="pool_id")
// @IdentityAttribute("pool_id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/workspaces/types;awstypes;awstypes.WorkspacesPool")
// @Testing(importStateIdAttribute="pool_id")
// @Testing(importIgnore="capacity")
// @Testing(serialize=true)
func newResourcePool(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePool{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNamePool = "Pool"
)

type resourcePool struct {
	framework.ResourceWithModel[resourcePoolModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *resourcePool) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_settings": framework.ResourceOptionalComputedListOfObjectsAttribute[applicationSettingsModel](ctx, 1, nil, listplanmodifier.UseStateForUnknown()),
			"bundle_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^wsb-[0-9a-z]{8,63}$`),
						"Bundle ID must be in the format 'wsb-xxxxxxxx' where 'xxxxxxxx' is a 8-63 character long string of lowercase letters and numbers",
					),
				},
			},
			"capacity_status": framework.ResourceComputedListOfObjectsAttribute[capacityStatusModel](ctx, listplanmodifier.UseStateForUnknown()),
			"created_at": schema.StringAttribute{
				Computed:   true,
				CustomType: timetypes.RFC3339Type{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[a-zA-Z0-9_./() -]+$`),
						"Description must contain only alphanumeric characters, underscores, periods, forward slashes, parentheses, spaces, and hyphens",
					),
				},
			},
			"directory_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^(d-[0-9a-f]{8,63}$)|(wsd-[0-9a-z]{8,63}$)`),
						"Directory ID must be in the format 'wsd-xxxxxxxx' where 'xxxxxxxx' is a 8-63 character long string of lowercase letters and numbers",
					),
				},
			},
			"pool_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pool_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pool_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`),
						"Name must start with an alphanumeric character and can only contain alphanumeric characters, underscores, periods, and hyphens",
					),
				},
			},
			"running_mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PoolsRunningMode](),
				Required:   true,
			},
			"s3_bucket_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:     tftags.TagsAttribute(),
			names.AttrTagsAll:  tftags.TagsAttributeComputedOnly(),
			"timeout_settings": framework.ResourceOptionalComputedListOfObjectsAttribute[timeoutSettingsModel](ctx, 1, nil, listplanmodifier.UseStateForUnknown()),
		},
		Blocks: map[string]schema.Block{
			"capacity": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[capacityModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"desired_user_sessions": schema.Int64Attribute{
							Required: true,
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourcePool) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().WorkSpacesClient(ctx)

	var plan resourcePoolModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input workspaces.CreateWorkspacesPoolInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateWorkspacesPool(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.PoolName.String())
		return
	}
	if out == nil || out.WorkspacesPool == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.PoolName.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.WorkspacesPool, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitPoolCreated(ctx, conn, plan.PoolId.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.PoolName.String())
		return
	}

	pool, err := findPoolByID(ctx, conn, plan.PoolId.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.PoolName.String())
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, pool, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourcePool) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().WorkSpacesClient(ctx)

	var state resourcePoolModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findPoolByID(ctx, conn, state.PoolId.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.PoolId.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourcePool) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().WorkSpacesClient(ctx)

	var plan, state resourcePoolModel
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

	if diff.HasChanges() {
		var input workspaces.UpdateWorkspacesPoolInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateWorkspacesPool(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.PoolId.String())
			return
		}
		if out == nil || out.WorkspacesPool == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.PoolId.String())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.WorkspacesPool, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitPoolUpdated(ctx, conn, plan.PoolId.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.PoolId.String())
		return
	}

	pool, err := findPoolByID(ctx, conn, plan.PoolId.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.PoolId.String())
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, pool, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourcePool) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().WorkSpacesClient(ctx)

	var state resourcePoolModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := workspaces.TerminateWorkspacesPoolInput{
		PoolId: state.PoolId.ValueStringPointer(),
	}

	_, err := conn.TerminateWorkspacesPool(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.PoolId.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitPoolDeleted(ctx, conn, state.PoolId.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.PoolId.String())
		return
	}
}

const (
	statusCreating = awstypes.WorkspacesPoolStateCreating
	statusDeleting = awstypes.WorkspacesPoolStateDeleting
	statusRunning  = awstypes.WorkspacesPoolStateRunning
	statusStarting = awstypes.WorkspacesPoolStateStarting
	statusStopped  = awstypes.WorkspacesPoolStateStopped
	statusStopping = awstypes.WorkspacesPoolStateStopping
	statusUpdating = awstypes.WorkspacesPoolStateUpdating
)

func waitPoolCreated(ctx context.Context, conn *workspaces.Client, id string, timeout time.Duration) (*awstypes.WorkspacesPool, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(statusCreating, statusUpdating),
		Target:                    enum.Slice(statusStopped, statusRunning, statusStarting),
		Refresh:                   statusPool(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.WorkspacesPool); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitPoolUpdated(ctx context.Context, conn *workspaces.Client, id string, timeout time.Duration) (*awstypes.WorkspacesPool, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(statusUpdating, statusCreating),
		Target:                    enum.Slice(statusStopped, statusRunning, statusStarting),
		Refresh:                   statusPool(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.WorkspacesPool); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitPoolDeleted(ctx context.Context, conn *workspaces.Client, id string, timeout time.Duration) (*awstypes.WorkspacesPool, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(statusDeleting, statusStopping),
		Target:  []string{},
		Refresh: statusPool(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.WorkspacesPool); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusPool(ctx context.Context, conn *workspaces.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findPoolByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.State), nil
	}
}

func (r *resourcePool) flatten(ctx context.Context, apiObject *awstypes.WorkspacesPool, data *resourcePoolModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(flex.Flatten(ctx, apiObject, data)...)
	if diags.HasError() {
		return diags
	}
	data.S3BucketName = flex.StringToFramework(ctx, apiObject.ApplicationSettings.S3BucketName)
	return diags
}

func findPoolByID(ctx context.Context, conn *workspaces.Client, id string) (*awstypes.WorkspacesPool, error) {
	input := &workspaces.DescribeWorkspacesPoolsInput{
		PoolIds: []string{id},
	}

	out, err := conn.DescribeWorkspacesPools(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.WorkspacesPools == nil || len(out.WorkspacesPools) == 0 {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return &out.WorkspacesPools[0], nil
}

type resourcePoolModel struct {
	framework.WithRegionModel
	ApplicationSettings fwtypes.ListNestedObjectValueOf[applicationSettingsModel] `tfsdk:"application_settings"`
	BundleId            types.String                                              `tfsdk:"bundle_id"`
	Capacity            fwtypes.ListNestedObjectValueOf[capacityModel]            `tfsdk:"capacity"`
	CapacityStatus      fwtypes.ListNestedObjectValueOf[capacityStatusModel]      `tfsdk:"capacity_status"`
	CreatedAt           timetypes.RFC3339                                         `tfsdk:"created_at"`
	Description         types.String                                              `tfsdk:"description"`
	DirectoryId         types.String                                              `tfsdk:"directory_id"`
	PoolArn             types.String                                              `tfsdk:"pool_arn"`
	PoolId              types.String                                              `tfsdk:"pool_id"`
	PoolName            types.String                                              `tfsdk:"pool_name"`
	RunningMode         fwtypes.StringEnum[awstypes.PoolsRunningMode]             `tfsdk:"running_mode"`
	S3BucketName        types.String                                              `tfsdk:"s3_bucket_name"`
	State               types.String                                              `tfsdk:"state"`
	Tags                tftags.Map                                                `tfsdk:"tags"`
	TagsAll             tftags.Map                                                `tfsdk:"tags_all"`
	TimeoutSettings     fwtypes.ListNestedObjectValueOf[timeoutSettingsModel]     `tfsdk:"timeout_settings"`
	Timeouts            timeouts.Value                                            `tfsdk:"timeouts"`
}

type applicationSettingsModel struct {
	SettingsGroup types.String `tfsdk:"settings_group"`
	Status        types.String `tfsdk:"status"`
}

type capacityModel struct {
	DesiredUserSessions types.Int64 `tfsdk:"desired_user_sessions"`
}

type capacityStatusModel struct {
	ActiveUserSessions    types.Int64 `tfsdk:"active_user_sessions"`
	ActualUserSessions    types.Int64 `tfsdk:"actual_user_sessions"`
	AvailableUserSessions types.Int64 `tfsdk:"available_user_sessions"`
	DesiredUserSessions   types.Int64 `tfsdk:"desired_user_sessions"`
}

type timeoutSettingsModel struct {
	DisconnectTimeoutInSeconds     types.Int64 `tfsdk:"disconnect_timeout_in_seconds"`
	IdleDisconnectTimeoutInSeconds types.Int64 `tfsdk:"idle_disconnect_timeout_in_seconds"`
	MaxUserDurationInSeconds       types.Int64 `tfsdk:"max_user_duration_in_seconds"`
}
