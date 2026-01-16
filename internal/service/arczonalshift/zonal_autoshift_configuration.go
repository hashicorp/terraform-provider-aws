// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package arczonalshift

import (
	"context"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arczonalshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/arczonalshift/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_arczonalshift_zonal_autoshift_configuration", name="Zonal Autoshift Configuration")
func newResourceZonalAutoshiftConfiguration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceZonalAutoshiftConfiguration{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameZonalAutoshiftConfiguration = "Zonal Autoshift Configuration"
)

type resourceZonalAutoshiftConfiguration struct {
	framework.ResourceWithModel[resourceZonalAutoshiftConfigurationModel]
	framework.WithTimeouts
}

func (r *resourceZonalAutoshiftConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrResourceARN: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"outcome_alarm_arns": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Required:    true,
				ElementType: types.StringType,
			},
			"blocking_alarm_arns": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Optional:    true,
				ElementType: types.StringType,
			},
			"blocked_dates": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("allowed_windows")),
				},
			},
			"blocked_windows": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("allowed_windows")),
				},
			},
			"allowed_windows": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("blocked_windows")),
				},
			},
			"autoshift_enabled": schema.BoolAttribute{
				Required: true,
			},
		},
	}
}

func (r *resourceZonalAutoshiftConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ARCZonalShiftClient(ctx)

	var plan resourceZonalAutoshiftConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	input := arczonalshift.CreatePracticeRunConfigurationInput{
		ResourceIdentifier: plan.ResourceARN.ValueStringPointer(),
		OutcomeAlarms:      expandControlConditions(ctx, plan.OutcomeAlarmARNs),
	}

	if !plan.BlockingAlarmARNs.IsNull() {
		input.BlockingAlarms = expandControlConditions(ctx, plan.BlockingAlarmARNs)
	}

	if !plan.BlockedDates.IsNull() {
		input.BlockedDates = expandStringList(ctx, plan.BlockedDates)
	}

	if !plan.BlockedWindows.IsNull() {
		input.BlockedWindows = expandStringList(ctx, plan.BlockedWindows)
	}

	if !plan.AllowedWindows.IsNull() {
		input.AllowedWindows = expandStringList(ctx, plan.AllowedWindows)
	}

	out, err := conn.CreatePracticeRunConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ResourceARN.String())
		return
	}

	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ResourceARN.String())
		return
	}

	statusInput := arczonalshift.UpdateZonalAutoshiftConfigurationInput{
		ResourceIdentifier:   plan.ResourceARN.ValueStringPointer(),
		ZonalAutoshiftStatus: awstypes.ZonalAutoshiftStatusDisabled,
	}

	if plan.AutoshiftEnabled.ValueBool() {
		statusInput.ZonalAutoshiftStatus = awstypes.ZonalAutoshiftStatusEnabled
	}

	out2, err := conn.UpdateZonalAutoshiftConfiguration(ctx, &statusInput)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ResourceARN.String())
		return
	}

	if out2 == nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ResourceARN.String())
		return
	}

	plan.ResourceARN = types.StringValue(aws.ToString(out2.ResourceIdentifier))
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceZonalAutoshiftConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ARCZonalShiftClient(ctx)

	var state resourceZonalAutoshiftConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findManagedResourceByIdentifier(ctx, conn, state.ResourceARN.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ARCZonalShift, create.ErrActionReading, ResNameZonalAutoshiftConfiguration, state.ResourceARN.String(), err),
			err.Error(),
		)
		return
	}

	if out == nil || out.PracticeRunConfiguration == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.AutoshiftEnabled = types.BoolValue(out.ZonalAutoshiftStatus == awstypes.ZonalAutoshiftStatusEnabled)
	state.OutcomeAlarmARNs = flattenControlConditions(ctx, out.PracticeRunConfiguration.OutcomeAlarms)
	state.BlockingAlarmARNs = flattenControlConditions(ctx, out.PracticeRunConfiguration.BlockingAlarms)
	state.BlockedDates = flattenStringList(ctx, out.PracticeRunConfiguration.BlockedDates)
	state.BlockedWindows = flattenStringList(ctx, out.PracticeRunConfiguration.BlockedWindows)
	state.AllowedWindows = flattenStringList(ctx, out.PracticeRunConfiguration.AllowedWindows)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceZonalAutoshiftConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ARCZonalShiftClient(ctx)

	var plan, state resourceZonalAutoshiftConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	resourceIdentifier := plan.ResourceARN.ValueString()

	practiceRunChanged := !plan.OutcomeAlarmARNs.Equal(state.OutcomeAlarmARNs) ||
		!plan.BlockingAlarmARNs.Equal(state.BlockingAlarmARNs) ||
		!plan.BlockedDates.Equal(state.BlockedDates) ||
		!plan.BlockedWindows.Equal(state.BlockedWindows) ||
		!plan.AllowedWindows.Equal(state.AllowedWindows)

	if practiceRunChanged {
		input := arczonalshift.UpdatePracticeRunConfigurationInput{
			ResourceIdentifier: aws.String(resourceIdentifier),
			OutcomeAlarms:      expandControlConditions(ctx, plan.OutcomeAlarmARNs),
		}

		if !plan.BlockingAlarmARNs.IsNull() {
			input.BlockingAlarms = expandControlConditions(ctx, plan.BlockingAlarmARNs)
		}

		if !plan.BlockedDates.IsNull() {
			input.BlockedDates = expandStringList(ctx, plan.BlockedDates)
		}

		if !plan.BlockedWindows.IsNull() {
			input.BlockedWindows = expandStringList(ctx, plan.BlockedWindows)
			input.AllowedWindows = []string{}
		} else if !plan.AllowedWindows.IsNull() {
			input.AllowedWindows = expandStringList(ctx, plan.AllowedWindows)
			input.BlockedWindows = []string{}
		} else {
			input.BlockedWindows = []string{}
			input.AllowedWindows = []string{}
		}

		_, err := conn.UpdatePracticeRunConfiguration(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceIdentifier)
			return
		}
	}

	if !plan.AutoshiftEnabled.Equal(state.AutoshiftEnabled) {
		status := awstypes.ZonalAutoshiftStatusDisabled
		if plan.AutoshiftEnabled.ValueBool() {
			status = awstypes.ZonalAutoshiftStatusEnabled
		}

		input := arczonalshift.UpdateZonalAutoshiftConfigurationInput{
			ResourceIdentifier:   aws.String(resourceIdentifier),
			ZonalAutoshiftStatus: status,
		}
		_, err := conn.UpdateZonalAutoshiftConfiguration(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceIdentifier)
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceZonalAutoshiftConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ARCZonalShiftClient(ctx)

	var state resourceZonalAutoshiftConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	resourceIdentifier := state.ResourceARN.ValueString()

	statusInput := arczonalshift.UpdateZonalAutoshiftConfigurationInput{
		ResourceIdentifier:   aws.String(resourceIdentifier),
		ZonalAutoshiftStatus: awstypes.ZonalAutoshiftStatusDisabled,
	}
	_, err := conn.UpdateZonalAutoshiftConfiguration(ctx, &statusInput)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceIdentifier)
		return
	}

	deleteInput := arczonalshift.DeletePracticeRunConfigurationInput{
		ResourceIdentifier: aws.String(resourceIdentifier),
	}
	_, err = conn.DeletePracticeRunConfiguration(ctx, &deleteInput)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceIdentifier)
		return
	}
}

func (r *resourceZonalAutoshiftConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrResourceARN), req, resp)
}

func expandControlConditions(ctx context.Context, list fwtypes.ListOfString) []awstypes.ControlCondition {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var arns []string
	list.ElementsAs(ctx, &arns, false)

	conditions := make([]awstypes.ControlCondition, len(arns))
	for i, arn := range arns {
		conditions[i] = awstypes.ControlCondition{
			AlarmIdentifier: aws.String(arn),
			Type:            awstypes.ControlConditionTypeCloudwatch,
		}
	}
	return conditions
}

func flattenControlConditions(ctx context.Context, conditions []awstypes.ControlCondition) fwtypes.ListOfString {
	if len(conditions) == 0 {
		return fwtypes.NewListValueOfNull[basetypes.StringValue](ctx)
	}

	elements := make([]attr.Value, len(conditions))
	for i, condition := range conditions {
		elements[i] = basetypes.NewStringValue(aws.ToString(condition.AlarmIdentifier))
	}

	return fwtypes.NewListValueOfMust[basetypes.StringValue](ctx, elements)
}

func expandStringList(ctx context.Context, list fwtypes.ListOfString) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var result []string
	list.ElementsAs(ctx, &result, false)
	return result
}

func flattenStringList(ctx context.Context, list []string) fwtypes.ListOfString {
	if len(list) == 0 {
		return fwtypes.NewListValueOfNull[basetypes.StringValue](ctx)
	}

	elements := make([]attr.Value, len(list))
	for i, s := range list {
		elements[i] = basetypes.NewStringValue(s)
	}

	return fwtypes.NewListValueOfMust[basetypes.StringValue](ctx, elements)
}

// Finder functions
func findManagedResourceByIdentifier(ctx context.Context, conn *arczonalshift.Client, resourceIdentifier string) (*arczonalshift.GetManagedResourceOutput, error) {
	input := arczonalshift.GetManagedResourceInput{
		ResourceIdentifier: aws.String(resourceIdentifier),
	}

	out, err := conn.GetManagedResource(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return out, nil
}

type resourceZonalAutoshiftConfigurationModel struct {
	framework.WithRegionModel
	ResourceARN       types.String         `tfsdk:"resource_arn"`
	OutcomeAlarmARNs  fwtypes.ListOfString `tfsdk:"outcome_alarm_arns"`
	BlockingAlarmARNs fwtypes.ListOfString `tfsdk:"blocking_alarm_arns"`
	BlockedDates      fwtypes.ListOfString `tfsdk:"blocked_dates"`
	BlockedWindows    fwtypes.ListOfString `tfsdk:"blocked_windows"`
	AllowedWindows    fwtypes.ListOfString `tfsdk:"allowed_windows"`
	AutoshiftEnabled  types.Bool           `tfsdk:"autoshift_enabled"`
}

func sweepZonalAutoshiftConfigurations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := arczonalshift.ListManagedResourcesInput{}
	conn := client.ARCZonalShiftClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := arczonalshift.NewListManagedResourcesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Items {
			if v.ZonalAutoshiftStatus == awstypes.ZonalAutoshiftStatusEnabled || v.PracticeRunStatus != "" {
				sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceZonalAutoshiftConfiguration, client,
					sweepfw.NewAttribute(names.AttrResourceARN, aws.ToString(v.Arn))),
				)
			}
		}
	}

	return sweepResources, nil
}
