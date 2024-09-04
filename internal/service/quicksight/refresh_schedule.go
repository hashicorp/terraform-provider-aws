// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_refresh_schedule", name="Refresh Schedule")
func newRefreshScheduleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &refreshScheduleResource{}, nil
}

const (
	resNameRefreshSchedule = "Refresh Schedule"

	dayOfMonthRegex          = "^(?:LAST_DAY_OF_MONTH|1[0-9]|2[0-8]|[12]|[3-9])$"
	timeOfTheDayLayout       = "15:04"
	timeOfTheDayFormat       = "HH:MM"
	startAfterDateTimeLayout = "2006-01-02T15:04:05"
	startAfterDateTimeFormat = "YYYY-MM-DDTHH:MM:SS"
)

type refreshScheduleResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *refreshScheduleResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_quicksight_refresh_schedule"
}

func (r *refreshScheduleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrAWSAccountID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data_set_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"schedule_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrSchedule: schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"refresh_type": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.IngestionType](),
							},
						},
						"start_after_date_time": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Validators: []validator.String{
								startAfterDateTimeValidator(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"schedule_frequency": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrInterval: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											enum.FrameworkValidate[awstypes.RefreshInterval](),
										},
									},
									"time_of_the_day": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											timeOfTheDayValidator(),
										},
									},
									"timezone": schema.StringAttribute{
										Optional: true,
										Computed: true,
									},
								},
								Blocks: map[string]schema.Block{
									"refresh_on_day": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"day_of_month": schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														stringvalidator.RegexMatches(regexache.MustCompile(dayOfMonthRegex), "day of month must match regex: "+dayOfMonthRegex),
														stringvalidator.ConflictsWith(
															path.MatchRelative().AtParent().AtName("day_of_week"),
														),
													},
												},
												"day_of_week": schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														enum.FrameworkValidate[awstypes.DayOfWeek](),
														stringvalidator.ConflictsWith(
															path.MatchRelative().AtParent().AtName("day_of_month"),
														),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

type resourceRefreshScheduleData struct {
	ARN          types.String `tfsdk:"arn"`
	AWSAccountID types.String `tfsdk:"aws_account_id"`
	DataSetID    types.String `tfsdk:"data_set_id"`
	ID           types.String `tfsdk:"id"`
	ScheduleID   types.String `tfsdk:"schedule_id"`
	Schedule     types.List   `tfsdk:"schedule"`
}

type scheduleData struct {
	RefreshType        types.String `tfsdk:"refresh_type"`
	ScheduleFrequency  types.List   `tfsdk:"schedule_frequency"`
	StartAfterDateTime types.String `tfsdk:"start_after_date_time"`
}

type refreshFrequencyData struct {
	Interval     types.String `tfsdk:"interval"`
	RefreshOnDay types.List   `tfsdk:"refresh_on_day"`
	TimeOfTheDay types.String `tfsdk:"time_of_the_day"`
	Timezone     types.String `tfsdk:"timezone"`
}

type refreshOnDayData struct {
	DayOfMonth types.String `tfsdk:"day_of_month"`
	DayOfWeek  types.String `tfsdk:"day_of_week"`
}

var (
	refreshOnDayAttrTypes = map[string]attr.Type{
		"day_of_month": types.StringType,
		"day_of_week":  types.StringType,
	}
	refreshFrequencyAttrTypes = map[string]attr.Type{
		names.AttrInterval: types.StringType,
		"refresh_on_day": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: refreshOnDayAttrTypes,
			},
		},
		"time_of_the_day": types.StringType,
		"timezone":        types.StringType,
	}
	scheduleAttrTypes = map[string]attr.Type{
		"refresh_type": types.StringType,
		"schedule_frequency": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: refreshFrequencyAttrTypes,
			},
		},
		"start_after_date_time": types.StringType,
	}
)

func (r *refreshScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var plan resourceRefreshScheduleData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.AWSAccountID.IsUnknown() || plan.AWSAccountID.IsNull() {
		plan.AWSAccountID = types.StringValue(r.Meta().AccountID)
	}
	awsAccountID, dataSetID, scheduleID := flex.StringValueFromFramework(ctx, plan.AWSAccountID), flex.StringValueFromFramework(ctx, plan.DataSetID), flex.StringValueFromFramework(ctx, plan.ScheduleID)

	scheduleInput, d := expandSchedule(ctx, scheduleID, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := quicksight.CreateRefreshScheduleInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSetId:    aws.String(dataSetID),
		Schedule:     scheduleInput,
	}

	out, err := conn.CreateRefreshSchedule(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, resNameRefreshSchedule, plan.ScheduleID.String(), nil),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, resNameRefreshSchedule, plan.ScheduleID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringValueToFramework(ctx, refreshScheduleCreateResourceID(awsAccountID, dataSetID, scheduleID))

	_, outFind, err := findRefreshScheduleByThreePartKey(ctx, conn, awsAccountID, dataSetID, scheduleID)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionReading, resNameRefreshSchedule, plan.ID.String(), nil),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(plan.refreshFromRead(ctx, out.Arn, outFind)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *refreshScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceRefreshScheduleData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountID, dataSetID, scheduleID, err := refreshScheduleParseResourceID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionReading, resNameRefreshSchedule, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	arn, outFind, err := findRefreshScheduleByThreePartKey(ctx, conn, awsAccountID, dataSetID, scheduleID)
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionReading, resNameRefreshSchedule, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	state.AWSAccountID = flex.StringValueToFramework(ctx, awsAccountID)
	state.DataSetID = flex.StringValueToFramework(ctx, dataSetID)
	state.ScheduleID = flex.StringValueToFramework(ctx, scheduleID)
	resp.Diagnostics.Append(state.refreshFromRead(ctx, arn, outFind)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *refreshScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var config, plan, state resourceRefreshScheduleData
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountID, dataSetID, scheduleID, err := refreshScheduleParseResourceID(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, resNameRefreshSchedule, plan.ID.String(), nil),
			err.Error(),
		)
		return
	}

	if !plan.Schedule.Equal(state.Schedule) {
		scheduleInput, d := expandSchedule(ctx, scheduleID, plan)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		in := quicksight.UpdateRefreshScheduleInput{
			AwsAccountId: aws.String(awsAccountID),
			DataSetId:    aws.String(dataSetID),
			Schedule:     scheduleInput,
		}

		// NOTE: Do not set StartAfterDateTime if not defined in config anymore or the value is unchanged

		var configTfList, planTfList, stateTfList []scheduleData
		resp.Diagnostics.Append(config.Schedule.ElementsAs(ctx, &configTfList, false)...)
		resp.Diagnostics.Append(plan.Schedule.ElementsAs(ctx, &planTfList, false)...)
		resp.Diagnostics.Append(state.Schedule.ElementsAs(ctx, &stateTfList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		configSchedule := configTfList[0]
		planSchedule := planTfList[0]
		stateSchedule := stateTfList[0]

		if configSchedule.StartAfterDateTime.IsNull() ||
			planSchedule.StartAfterDateTime.Equal(stateSchedule.StartAfterDateTime) {
			in.Schedule.StartAfterDateTime = nil
		}
		out, err := conn.UpdateRefreshSchedule(ctx, &in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, resNameRefreshSchedule, plan.ID.String(), nil),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, resNameRefreshSchedule, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		_, outFind, err := findRefreshScheduleByThreePartKey(ctx, conn, awsAccountID, dataSetID, scheduleID)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionReading, resNameRefreshSchedule, plan.ID.String(), nil),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(plan.refreshFromRead(ctx, out.Arn, outFind)...)
		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	}
}

func (r *refreshScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceRefreshScheduleData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountID, dataSetID, scheduleID, err := refreshScheduleParseResourceID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, resNameRefreshSchedule, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	_, err = conn.DeleteRefreshSchedule(ctx, &quicksight.DeleteRefreshScheduleInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSetId:    aws.String(dataSetID),
		ScheduleId:   aws.String(scheduleID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, resNameRefreshSchedule, state.ID.String(), nil),
			err.Error(),
		)
	}
}

func (r *refreshScheduleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	scheduleFrequencyPath := path.Root(names.AttrSchedule).AtListIndex(0).AtName("schedule_frequency").AtListIndex(0)

	var scheduleFrequency refreshFrequencyData
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, scheduleFrequencyPath, &scheduleFrequency)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if scheduleFrequency.Interval.IsUnknown() {
		// Field is required, if it's unknown, the value is likely coming from a dynamic block and
		// ValidateConfig will be called again later with the actual value.
		return
	}

	refreshOnDayPath := scheduleFrequencyPath.AtName("refresh_on_day")

	var refreshOnDay []refreshOnDayData
	resp.Diagnostics.Append(scheduleFrequency.RefreshOnDay.ElementsAs(ctx, &refreshOnDay, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	switch interval := scheduleFrequency.Interval.ValueString(); interval {
	case string(awstypes.RefreshIntervalWeekly):
		if len(refreshOnDay) == 0 || refreshOnDay[0].DayOfWeek.IsNull() {
			resp.Diagnostics.Append(fwdiag.NewAttributeRequiredWhenError(
				refreshOnDayPath.AtListIndex(0).AtName("day_of_week"),
				scheduleFrequencyPath.AtName(names.AttrInterval),
				interval,
			))
		}
	case string(awstypes.RefreshIntervalMonthly):
		if len(refreshOnDay) == 0 || refreshOnDay[0].DayOfMonth.IsNull() {
			resp.Diagnostics.Append(fwdiag.NewAttributeRequiredWhenError(
				refreshOnDayPath.AtListIndex(0).AtName("day_of_month"),
				scheduleFrequencyPath.AtName(names.AttrInterval),
				interval,
			))
		}

	default:
		if len(refreshOnDay) != 0 {
			resp.Diagnostics.Append(fwdiag.NewAttributeConflictsWhenError(
				refreshOnDayPath,
				scheduleFrequencyPath.AtName(names.AttrInterval),
				interval,
			))
		}
	}
}

func findRefreshScheduleByThreePartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, dataSetID, scheduleID string) (*string, *awstypes.RefreshSchedule, error) {
	input := &quicksight.DescribeRefreshScheduleInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSetId:    aws.String(dataSetID),
		ScheduleId:   aws.String(scheduleID),
	}

	return findRefreshSchedule(ctx, conn, input)
}

func findRefreshSchedule(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeRefreshScheduleInput) (*string, *awstypes.RefreshSchedule, error) {
	output, err := conn.DescribeRefreshSchedule(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, nil, err
	}

	if output == nil || output.RefreshSchedule == nil {
		return nil, nil, tfresource.NewEmptyResultError(input)
	}

	return output.Arn, output.RefreshSchedule, nil
}

func (rd *resourceRefreshScheduleData) refreshFromRead(ctx context.Context, arn *string, out *awstypes.RefreshSchedule) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}

	rd.ARN = flex.StringToFramework(ctx, arn)

	schedule, d := flattenSchedule(ctx, out)
	diags.Append(d...)
	rd.Schedule = schedule

	return diags
}

func expandSchedule(ctx context.Context, scheduleId string, plan resourceRefreshScheduleData) (*awstypes.RefreshSchedule, diag.Diagnostics) {
	var diags diag.Diagnostics

	var tfList []scheduleData
	diags.Append(plan.Schedule.ElementsAs(ctx, &tfList, false)...)
	if diags.HasError() {
		return nil, diags
	}

	tfObj := tfList[0]
	in := &awstypes.RefreshSchedule{
		ScheduleId:  aws.String(scheduleId),
		RefreshType: awstypes.IngestionType(tfObj.RefreshType.ValueString()),
	}

	if !tfObj.StartAfterDateTime.IsUnknown() {
		start, _ := time.Parse(startAfterDateTimeLayout, tfObj.StartAfterDateTime.ValueString())
		in.StartAfterDateTime = aws.Time(start)
	}

	refreshFrequency, d := expandRefreshFrequency(ctx, tfObj)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	in.ScheduleFrequency = refreshFrequency
	return in, diags
}

func expandRefreshFrequency(ctx context.Context, plan scheduleData) (*awstypes.RefreshFrequency, diag.Diagnostics) {
	var diags diag.Diagnostics
	var tfList []refreshFrequencyData
	diags.Append(plan.ScheduleFrequency.ElementsAs(ctx, &tfList, false)...)
	if diags.HasError() || len(tfList) == 0 {
		return nil, diags
	}

	tfObj := tfList[0]
	freq := &awstypes.RefreshFrequency{
		Interval:     awstypes.RefreshInterval(tfObj.Interval.ValueString()),
		TimeOfTheDay: aws.String(tfObj.TimeOfTheDay.ValueString()),
		Timezone:     aws.String(tfObj.Timezone.ValueString()),
	}

	if !tfObj.RefreshOnDay.IsNull() {
		refreshOnDay, d := expandRefreshOnDayData(ctx, tfObj)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		freq.RefreshOnDay = refreshOnDay
	}
	return freq, diags
}

func expandRefreshOnDayData(ctx context.Context, plan refreshFrequencyData) (*awstypes.ScheduleRefreshOnEntity, diag.Diagnostics) {
	var diags diag.Diagnostics
	var tfList []refreshOnDayData
	diags.Append(plan.RefreshOnDay.ElementsAs(ctx, &tfList, false)...)
	if diags.HasError() || len(tfList) == 0 {
		return nil, diags
	}

	tfObj := tfList[0]
	entity := &awstypes.ScheduleRefreshOnEntity{}
	if !tfObj.DayOfMonth.IsNull() {
		entity.DayOfMonth = aws.String(tfObj.DayOfMonth.ValueString())
	}
	if !tfObj.DayOfWeek.IsNull() {
		entity.DayOfWeek = awstypes.DayOfWeek(tfObj.DayOfWeek.ValueString())
	}
	return entity, diags
}

func flattenSchedule(ctx context.Context, apiObject *awstypes.RefreshSchedule) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiObject == nil {
		return types.ListNull(types.ObjectType{AttrTypes: scheduleAttrTypes}), diags
	}

	refreshFrequency, d := flattenRefreshFrequency(ctx, apiObject.ScheduleFrequency)
	diags.Append(d...)

	scheduleAttrs := map[string]attr.Value{
		"refresh_type":       flex.StringValueToFramework(ctx, apiObject.RefreshType),
		"schedule_frequency": refreshFrequency,
	}

	if apiObject.StartAfterDateTime != nil {
		scheduleAttrs["start_after_date_time"] = types.StringValue(apiObject.StartAfterDateTime.Format(startAfterDateTimeLayout))
	}

	objVal, d := types.ObjectValue(scheduleAttrTypes, scheduleAttrs)
	diags.Append(d...)
	listVal, d := types.ListValue(types.ObjectType{AttrTypes: scheduleAttrTypes}, []attr.Value{objVal})
	diags.Append(d...)
	return listVal, diags
}

func flattenRefreshFrequency(ctx context.Context, apiObject *awstypes.RefreshFrequency) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiObject == nil {
		return types.ListNull(types.ObjectType{AttrTypes: refreshFrequencyAttrTypes}), diags
	}

	refreshOnDay, d := flattenRefreshOnDay(ctx, apiObject.RefreshOnDay)
	diags.Append(d...)

	refreshFrequencyAttrs := map[string]attr.Value{
		names.AttrInterval: flex.StringValueToFramework(ctx, apiObject.Interval),
		"time_of_the_day":  flex.StringToFramework(ctx, apiObject.TimeOfTheDay),
		"timezone":         flex.StringToFramework(ctx, apiObject.Timezone),
		"refresh_on_day":   refreshOnDay,
	}

	objVal, d := types.ObjectValue(refreshFrequencyAttrTypes, refreshFrequencyAttrs)
	diags.Append(d...)
	listVal, d := types.ListValue(types.ObjectType{AttrTypes: refreshFrequencyAttrTypes}, []attr.Value{objVal})
	diags.Append(d...)
	return listVal, diags
}

func flattenRefreshOnDay(ctx context.Context, apiObject *awstypes.ScheduleRefreshOnEntity) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiObject == nil {
		return types.ListNull(types.ObjectType{AttrTypes: refreshOnDayAttrTypes}), diags
	}

	objVal, d := types.ObjectValue(refreshOnDayAttrTypes, map[string]attr.Value{
		"day_of_month": flex.StringToFramework(ctx, apiObject.DayOfMonth),
		"day_of_week":  flex.StringValueToFramework(ctx, apiObject.DayOfWeek),
	})
	diags.Append(d...)
	listVal, d := types.ListValue(types.ObjectType{AttrTypes: refreshOnDayAttrTypes}, []attr.Value{objVal})
	diags.Append(d...)
	return listVal, diags
}

const refreshScheduleResourceIDSeparator = ","

func refreshScheduleCreateResourceID(awsAccountID, dataSetID, scheduleID string) string {
	parts := []string{awsAccountID, dataSetID, scheduleID}
	id := strings.Join(parts, refreshScheduleResourceIDSeparator)

	return id
}

func refreshScheduleParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, refreshScheduleResourceIDSeparator, 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sDATA_SET_ID%[2]sSCHEDULE_ID", id, refreshScheduleResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], nil
}

func timeOfTheDayValidator() validator.String {
	return timeMatchesValidator{
		layout:  timeOfTheDayLayout,
		message: fmt.Sprintf("value must match '%s' format", timeOfTheDayFormat),
	}
}

func startAfterDateTimeValidator() validator.String {
	return timeMatchesValidator{
		layout:  startAfterDateTimeLayout,
		message: fmt.Sprintf("value must match '%s' format", startAfterDateTimeFormat),
	}
}

type timeMatchesValidator struct {
	layout  string
	message string
}

func (v timeMatchesValidator) Description(_ context.Context) string {
	return v.message
}

func (v timeMatchesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v timeMatchesValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	_, err := time.Parse(v.layout, value)
	if err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			value,
		))
	}
}
