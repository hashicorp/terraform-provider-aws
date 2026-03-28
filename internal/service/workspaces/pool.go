// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_workspaces_pool", name="Pool")
// @Tags(identifierAttribute="id")
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
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourcePool) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
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
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
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
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("AUTO_STOP", "ALWAYS_ON"),
				},
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"application_settings": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[applicationSettingsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(0),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrS3BucketName: schema.StringAttribute{
							Computed: true,
						},
						"settings_group": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 64),
								stringvalidator.RegexMatches(
									regexache.MustCompile(`^[a-zA-Z0-9_.-]+$`),
									"Settings group must contain only alphanumeric characters, underscores, periods, and hyphens",
								),
							},
						},
						names.AttrStatus: schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(awstypes.ApplicationSettingsStatusEnumEnabled),
									string(awstypes.ApplicationSettingsStatusEnumDisabled),
								),
							},
						},
					},
				},
			},
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
			"timeout_settings": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[timeoutSettingsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(0),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"disconnect_timeout_in_seconds": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							Validators: []validator.Int64{
								int64validator.Between(60, 36000),
							},
						},
						"idle_disconnect_timeout_in_seconds": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							Validators: []validator.Int64{
								int64validator.Between(0, 36000),
							},
						},
						"max_user_duration_in_seconds": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							Validators: []validator.Int64{
								int64validator.Between(600, 432000),
							},
						},
					},
				},
			},
		},
	}

	if s.Blocks == nil {
		s.Blocks = make(map[string]schema.Block)
	}
	s.Blocks[names.AttrTimeouts] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
		Update: true,
		Delete: true,
	})

	resp.Schema = s
}

func (r *resourcePool) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().WorkSpacesClient(ctx)

	var plan resourcePoolModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input workspaces.CreateWorkspacesPoolInput
	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Pool")))
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateWorkspacesPool(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil || out.WorkspacesPool == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.WorkspacesPool, &plan, flex.WithFieldNamePrefix("Pool")))
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitPoolCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourcePool) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().WorkSpacesClient(ctx)

	var state resourcePoolModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findPoolByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("Pool")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourcePool) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().WorkSpacesClient(ctx)

	var plan, state resourcePoolModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input workspaces.UpdateWorkspacesPoolInput
		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Pool")))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateWorkspacesPool(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
		if out == nil || out.WorkspacesPool == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ID.String())
			return
		}

		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.WorkspacesPool, &plan, flex.WithFieldNamePrefix("Pool")))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitPoolUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourcePool) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().WorkSpacesClient(ctx)

	var state resourcePoolModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := workspaces.TerminateWorkspacesPoolInput{
		PoolId: state.ID.ValueStringPointer(),
	}

	_, err := conn.TerminateWorkspacesPool(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitPoolDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
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
	return func() (any, string, error) {
		out, err := findPoolByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.State), nil
	}
}

func findPoolByID(ctx context.Context, conn *workspaces.Client, id string) (*awstypes.WorkspacesPool, error) {
	input := &workspaces.DescribeWorkspacesPoolsInput{
		PoolIds: []string{id},
	}

	out, err := conn.DescribeWorkspacesPools(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.WorkspacesPools == nil || len(out.WorkspacesPools) == 0 {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(input))
	}

	return &out.WorkspacesPools[0], nil
}

type resourcePoolModel struct {
	framework.WithRegionModel
	ApplicationSettings fwtypes.ListNestedObjectValueOf[applicationSettingsModel] `tfsdk:"application_settings"`
	ARN                 types.String                                              `tfsdk:"arn"`
	BundleId            types.String                                              `tfsdk:"bundle_id"`
	Capacity            fwtypes.ListNestedObjectValueOf[capacityModel]            `tfsdk:"capacity"`
	Description         types.String                                              `tfsdk:"description"`
	DirectoryId         types.String                                              `tfsdk:"directory_id"`
	ID                  types.String                                              `tfsdk:"id"`
	Name                types.String                                              `tfsdk:"name"`
	RunningMode         types.String                                              `tfsdk:"running_mode"`
	State               types.String                                              `tfsdk:"state"`
	Tags                tftags.Map                                                `tfsdk:"tags"`
	TagsAll             tftags.Map                                                `tfsdk:"tags_all"`
	TimeoutSettings     fwtypes.ListNestedObjectValueOf[timeoutSettingsModel]     `tfsdk:"timeout_settings"`
	Timeouts            timeouts.Value                                            `tfsdk:"timeouts"`
}

type applicationSettingsModel struct {
	S3BucketName  types.String `tfsdk:"s3_bucket_name"`
	SettingsGroup types.String `tfsdk:"settings_group"`
	Status        types.String `tfsdk:"status"`
}

type capacityModel struct {
	DesiredUserSessions types.Int64 `tfsdk:"desired_user_sessions"`
}

type timeoutSettingsModel struct {
	DisconnectTimeoutInSeconds     types.Int64 `tfsdk:"disconnect_timeout_in_seconds"`
	IdleDisconnectTimeoutInSeconds types.Int64 `tfsdk:"idle_disconnect_timeout_in_seconds"`
	MaxUserDurationInSeconds       types.Int64 `tfsdk:"max_user_duration_in_seconds"`
}

func sweepPools(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := &workspaces.DescribeWorkspacesPoolsInput{}
	conn := client.WorkSpacesClient(ctx)
	var sweepResources []sweep.Sweepable

	output, err := conn.DescribeWorkspacesPools(ctx, input)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output != nil && len(output.WorkspacesPools) > 0 {
		for _, v := range output.WorkspacesPools {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourcePool, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.PoolId))),
			)
		}
	}

	return sweepResources, nil
}
