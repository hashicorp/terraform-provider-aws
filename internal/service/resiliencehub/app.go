// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehub

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehub/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_resiliencehub_app", name="App")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/resiliencehub;resiliencehub.DescribeAppOutput")
// @Testing(importStateIdAttribute="arn")
// @Testing(hasNoPreExistingResource=true)
// @ArnIdentity
func newAppResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &appResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameApp = "App"
)

type appResource struct {
	framework.ResourceWithModel[appResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *appResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"assessment_schedule": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AppAssessmentScheduleType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 500),
				},
			},
			"drift_status": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 60),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_\-]*$`), "must start with an alphanumeric character and contain only alphanumeric characters, underscores, and hyphens"),
				},
			},
			"resiliency_policy_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
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

func (r *appResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan appResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubClient(ctx)

	name := plan.Name.ValueString()
	input := &resiliencehub.CreateAppInput{
		ClientToken: aws.String(create.UniqueId(ctx)),
		Tags:        getTagsIn(ctx),
	}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateApp(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionCreating, ResNameApp, name, err),
			err.Error(),
		)
		return
	}

	plan.AppArn = types.StringValue(aws.ToString(output.App.AppArn))

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	app, err := waitAppCreated(ctx, conn, plan.AppArn.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionWaitingForCreation, ResNameApp, name, err),
			err.Error(),
		)
		return
	}

	plan.DriftStatus = types.StringValue(string(app.DriftStatus))
	plan.AssessmentSchedule = fwtypes.StringEnumValue(app.AssessmentSchedule)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *appResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state appResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubClient(ctx)

	output, err := FindAppByARN(ctx, conn, state.AppArn.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionReading, ResNameApp, state.AppArn.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, output, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.AppArn = types.StringValue(aws.ToString(output.AppArn))

	setTagsOut(ctx, output.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *appResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state appResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubClient(ctx)

	if !plan.Description.Equal(state.Description) ||
		!plan.AssessmentSchedule.Equal(state.AssessmentSchedule) ||
		!plan.PolicyArn.Equal(state.PolicyArn) {
		input := &resiliencehub.UpdateAppInput{
			AppArn: plan.AppArn.ValueStringPointer(),
		}

		resp.Diagnostics.Append(flex.Expand(ctx, plan, input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateApp(ctx, input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionUpdating, ResNameApp, plan.AppArn.ValueString(), err),
				err.Error(),
			)
			return
		}
	}

	if !plan.TagsAll.Equal(state.TagsAll) {
		if err := updateTags(ctx, conn, plan.AppArn.ValueString(), state.TagsAll, plan.TagsAll); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionUpdating, ResNameApp, plan.AppArn.ValueString(), err),
				err.Error(),
			)
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	app, err := waitAppUpdated(ctx, conn, plan.AppArn.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionWaitingForUpdate, ResNameApp, plan.AppArn.ValueString(), err),
			err.Error(),
		)
		return
	}

	plan.DriftStatus = types.StringValue(string(app.DriftStatus))
	plan.AssessmentSchedule = fwtypes.StringEnumValue(app.AssessmentSchedule)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *appResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state appResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubClient(ctx)

	_, err := conn.DeleteApp(ctx, &resiliencehub.DeleteAppInput{
		AppArn:      state.AppArn.ValueStringPointer(),
		ClientToken: aws.String(create.UniqueId(ctx)),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionDeleting, ResNameApp, state.AppArn.ValueString(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	err = waitAppDeleted(ctx, conn, state.AppArn.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionWaitingForDeletion, ResNameApp, state.AppArn.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func FindAppByARN(ctx context.Context, conn *resiliencehub.Client, arn string) (*awstypes.App, error) {
	input := &resiliencehub.DescribeAppInput{
		AppArn: aws.String(arn),
	}

	output, err := conn.DescribeApp(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.App == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.App, nil
}

func waitAppCreated(ctx context.Context, conn *resiliencehub.Client, arn string, timeout time.Duration) (*awstypes.App, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.AppStatusTypeActive),
		Refresh:                   statusApp(ctx, conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.App); ok {
		return output, err
	}

	return nil, err
}

func waitAppUpdated(ctx context.Context, conn *resiliencehub.Client, arn string, timeout time.Duration) (*awstypes.App, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.AppStatusTypeActive),
		Refresh:                   statusApp(ctx, conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.App); ok {
		return output, err
	}

	return nil, err
}

func waitAppDeleted(ctx context.Context, conn *resiliencehub.Client, arn string, timeout time.Duration) error {
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(awstypes.AppStatusTypeActive, awstypes.AppStatusTypeDeleting),
		Target:  []string{},
		Refresh: statusApp(ctx, conn, arn),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func statusApp(ctx context.Context, conn *resiliencehub.Client, arn string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := FindAppByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

type appResourceModel struct {
	framework.WithRegionModel
	AppArn             types.String                                           `tfsdk:"arn"`
	AssessmentSchedule fwtypes.StringEnum[awstypes.AppAssessmentScheduleType] `tfsdk:"assessment_schedule"`
	Description        types.String                                           `tfsdk:"description"`
	DriftStatus        types.String                                           `tfsdk:"drift_status"`
	Name               types.String                                           `tfsdk:"name"`
	PolicyArn          fwtypes.ARN                                            `tfsdk:"resiliency_policy_arn"`
	Tags               tftags.Map                                             `tfsdk:"tags" autoflex:"-"`
	TagsAll            tftags.Map                                             `tfsdk:"tags_all" autoflex:"-"`
	Timeouts           timeouts.Value                                         `tfsdk:"timeouts" autoflex:"-"`
}
