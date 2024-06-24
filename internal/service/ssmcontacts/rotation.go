// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssmcontacts/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameRotation = "Rotation"
)

// @FrameworkResource(name="Rotation")
// @Tags(identifierAttribute="arn")
func newResourceRotation(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceRotation{}

	return r, nil
}

type resourceRotation struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *resourceRotation) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_ssmcontacts_rotation"
}

func (r *resourceRotation) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"contact_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Required:    true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrStartTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Optional:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"time_zone_id": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"recurrence": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[recurrenceData](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"number_of_on_calls": schema.Int64Attribute{
							Required: true,
						},
						"recurrence_multiplier": schema.Int64Attribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"daily_settings": handOffTimeSchema(ctx, nil),
						"monthly_settings": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[monthlySettingsData](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"day_of_month": schema.Int64Attribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"hand_off_time": handOffTimeSchema(ctx, aws.Int(1)),
								},
							},
						},
						"shift_coverages": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[shiftCoveragesData](ctx),
							PlanModifiers: []planmodifier.List{
								ShiftCoveragesPlanModifier(),
								listplanmodifier.UseStateForUnknown(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"map_block_key": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.DayOfWeek](),
										Required:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"coverage_times": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[coverageTimesData](ctx),
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"end":   handOffTimeSchema(ctx, aws.Int(1)),
												"start": handOffTimeSchema(ctx, aws.Int(1)),
											},
										},
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
										},
									},
								},
							},
						},
						"weekly_settings": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[weeklySettingsData](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"day_of_week": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.DayOfWeek](),
										Required:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"hand_off_time": handOffTimeSchema(ctx, aws.Int(1)),
								},
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

	response.Schema = s
}

func handOffTimeSchema(ctx context.Context, size *int) schema.ListNestedBlock {
	listSchema := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[handOffTime](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"hour_of_day": schema.Int64Attribute{
					Required: true,
				},
				"minute_of_hour": schema.Int64Attribute{
					Required: true,
				},
			},
		},
	}

	if size != nil {
		listSchema.Validators = []validator.List{
			listvalidator.SizeAtMost(*size),
		}
	}

	return listSchema
}

func (r *resourceRotation) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().SSMContactsClient(ctx)
	var plan resourceRotationData

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	recurrenceData, diags := plan.Recurrence.ToPtr(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	shiftCoveragesData, diags := recurrenceData.ShiftCoverages.ToSlice(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	shiftCoverages := expandShiftCoverages(ctx, shiftCoveragesData, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	dailySettingsInput, dailySettingsOutput := setupSerializationObjects[handOffTime, awstypes.HandOffTime](recurrenceData.DailySettings)
	response.Diagnostics.Append(fwflex.Expand(ctx, dailySettingsInput, &dailySettingsOutput)...)
	if response.Diagnostics.HasError() {
		return
	}

	monthlySettingsInput, monthlySettingsOutput := setupSerializationObjects[monthlySettingsData, awstypes.MonthlySetting](recurrenceData.MonthlySettings)
	response.Diagnostics.Append(fwflex.Expand(ctx, monthlySettingsInput, &monthlySettingsOutput)...)
	if response.Diagnostics.HasError() {
		return
	}

	weeklySettingsInput, weeklySettingsOutput := setupSerializationObjects[weeklySettingsData, awstypes.WeeklySetting](recurrenceData.WeeklySettings)
	response.Diagnostics.Append(fwflex.Expand(ctx, weeklySettingsInput, &weeklySettingsOutput)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := &ssmcontacts.CreateRotationInput{
		IdempotencyToken: aws.String(id.UniqueId()),
		ContactIds:       fwflex.ExpandFrameworkStringValueList(ctx, plan.ContactIds),
		Name:             fwflex.StringFromFramework(ctx, plan.Name),
		Recurrence: &awstypes.RecurrenceSettings{
			NumberOfOnCalls:      fwflex.Int32FromFramework(ctx, recurrenceData.NumberOfOnCalls),
			RecurrenceMultiplier: fwflex.Int32FromFramework(ctx, recurrenceData.RecurrenceMultiplier),
			DailySettings:        dailySettingsOutput.Data,
			MonthlySettings:      monthlySettingsOutput.Data,
			ShiftCoverages:       shiftCoverages,
			WeeklySettings:       weeklySettingsOutput.Data,
		},
		StartTime:  fwflex.TimeFromFramework(ctx, plan.StartTime),
		TimeZoneId: fwflex.StringFromFramework(ctx, plan.TimeZoneID),
		Tags:       getTagsIn(ctx),
	}

	output, err := conn.CreateRotation(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMContacts, create.ErrActionCreating, ResNameRotation, plan.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	state := plan

	state.ID = fwflex.StringToFramework(ctx, output.RotationArn)
	state.ARN = fwflex.StringToFramework(ctx, output.RotationArn)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourceRotation) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().SSMContactsClient(ctx)
	var state resourceRotationData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	output, err := findRotationByID(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMContacts, create.ErrActionSetting, ResNameRotation, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	rc := &recurrenceData{}
	rc.RecurrenceMultiplier = fwflex.Int32ToFramework(ctx, output.Recurrence.RecurrenceMultiplier)
	rc.NumberOfOnCalls = fwflex.Int32ToFramework(ctx, output.Recurrence.NumberOfOnCalls)

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.Recurrence.DailySettings, &rc.DailySettings)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.Recurrence.MonthlySettings, &rc.MonthlySettings)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.Recurrence.WeeklySettings, &rc.WeeklySettings)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.ContactIds, &state.ContactIds)...)
	if response.Diagnostics.HasError() {
		return
	}

	rc.ShiftCoverages = flattenShiftCoverages(ctx, output.Recurrence.ShiftCoverages)

	state.ARN = fwflex.StringToFramework(ctx, output.RotationArn)
	state.Name = fwflex.StringToFramework(ctx, output.Name)
	state.Recurrence = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, rc)
	state.StartTime = fwflex.TimeToFramework(ctx, output.StartTime)
	state.TimeZoneID = fwflex.StringToFramework(ctx, output.TimeZoneId)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourceRotation) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().SSMContactsClient(ctx)
	var state, plan resourceRotationData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Recurrence.Equal(state.Recurrence) || !plan.ContactIds.Equal(state.ContactIds) ||
		!plan.StartTime.Equal(state.StartTime) || !plan.TimeZoneID.Equal(state.TimeZoneID) {
		recurrenceData, diags := plan.Recurrence.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		shiftCoveragesData, diags := recurrenceData.ShiftCoverages.ToSlice(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		shiftCoverages := expandShiftCoverages(ctx, shiftCoveragesData, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}

		dailySettingsInput, dailySettingsOutput := setupSerializationObjects[handOffTime, awstypes.HandOffTime](recurrenceData.DailySettings)
		response.Diagnostics.Append(fwflex.Expand(ctx, dailySettingsInput, &dailySettingsOutput)...)
		if response.Diagnostics.HasError() {
			return
		}

		monthlySettingsInput, monthlySettingsOutput := setupSerializationObjects[monthlySettingsData, awstypes.MonthlySetting](recurrenceData.MonthlySettings)
		response.Diagnostics.Append(fwflex.Expand(ctx, monthlySettingsInput, &monthlySettingsOutput)...)
		if response.Diagnostics.HasError() {
			return
		}

		weeklySettingsInput, weeklySettingsOutput := setupSerializationObjects[weeklySettingsData, awstypes.WeeklySetting](recurrenceData.WeeklySettings)
		response.Diagnostics.Append(fwflex.Expand(ctx, weeklySettingsInput, &weeklySettingsOutput)...)
		if response.Diagnostics.HasError() {
			return
		}

		input := &ssmcontacts.UpdateRotationInput{
			RotationId: fwflex.StringFromFramework(ctx, state.ID),
			Recurrence: &awstypes.RecurrenceSettings{
				NumberOfOnCalls:      fwflex.Int32FromFramework(ctx, recurrenceData.NumberOfOnCalls),
				RecurrenceMultiplier: fwflex.Int32FromFramework(ctx, recurrenceData.RecurrenceMultiplier),
				DailySettings:        dailySettingsOutput.Data,
				MonthlySettings:      monthlySettingsOutput.Data,
				ShiftCoverages:       shiftCoverages,
				WeeklySettings:       weeklySettingsOutput.Data,
			},
			ContactIds: fwflex.ExpandFrameworkStringValueList(ctx, plan.ContactIds),
			StartTime:  fwflex.TimeFromFramework(ctx, plan.StartTime),
			TimeZoneId: fwflex.StringFromFramework(ctx, plan.TimeZoneID),
		}

		_, err := conn.UpdateRotation(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SSMContacts, create.ErrActionUpdating, ResNameRotation, state.ID.ValueString(), err),
				err.Error(),
			)
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceRotation) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().SSMContactsClient(ctx)
	var state resourceRotationData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting TODO", map[string]interface{}{
		names.AttrID: state.ID.ValueString(),
	})

	_, err := conn.DeleteRotation(ctx, &ssmcontacts.DeleteRotationInput{
		RotationId: fwflex.StringFromFramework(ctx, state.ID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMContacts, create.ErrActionDeleting, ResNameRotation, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceRotation) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

type resourceRotationData struct {
	ARN        types.String                                    `tfsdk:"arn"`
	ContactIds fwtypes.ListValueOf[types.String]               `tfsdk:"contact_ids"`
	ID         types.String                                    `tfsdk:"id"`
	Recurrence fwtypes.ListNestedObjectValueOf[recurrenceData] `tfsdk:"recurrence"`
	Name       types.String                                    `tfsdk:"name"`
	StartTime  timetypes.RFC3339                               `tfsdk:"start_time"`
	Tags       types.Map                                       `tfsdk:"tags"`
	TagsAll    types.Map                                       `tfsdk:"tags_all"`
	TimeZoneID types.String                                    `tfsdk:"time_zone_id"`
}

type recurrenceData struct {
	DailySettings        fwtypes.ListNestedObjectValueOf[handOffTime]         `tfsdk:"daily_settings"`
	MonthlySettings      fwtypes.ListNestedObjectValueOf[monthlySettingsData] `tfsdk:"monthly_settings"`
	NumberOfOnCalls      types.Int64                                          `tfsdk:"number_of_on_calls"`
	RecurrenceMultiplier types.Int64                                          `tfsdk:"recurrence_multiplier"`
	ShiftCoverages       fwtypes.ListNestedObjectValueOf[shiftCoveragesData]  `tfsdk:"shift_coverages"`
	WeeklySettings       fwtypes.ListNestedObjectValueOf[weeklySettingsData]  `tfsdk:"weekly_settings"`
}

type monthlySettingsData struct {
	DayOfMonth  types.Int64                                  `tfsdk:"day_of_month"`
	HandOffTime fwtypes.ListNestedObjectValueOf[handOffTime] `tfsdk:"hand_off_time"`
}

type shiftCoveragesData struct {
	CoverageTimes fwtypes.ListNestedObjectValueOf[coverageTimesData] `tfsdk:"coverage_times"`
	MapBlockKey   fwtypes.StringEnum[awstypes.DayOfWeek]             `tfsdk:"map_block_key"`
}

type coverageTimesData struct {
	End   fwtypes.ListNestedObjectValueOf[handOffTime] `tfsdk:"end"`
	Start fwtypes.ListNestedObjectValueOf[handOffTime] `tfsdk:"start"`
}

type handOffTime struct {
	HourOfDay    types.Int64 `tfsdk:"hour_of_day"`
	MinuteOfHour types.Int64 `tfsdk:"minute_of_hour"`
}

type weeklySettingsData struct {
	DayOfWeek   fwtypes.StringEnum[awstypes.DayOfWeek]       `tfsdk:"day_of_week"`
	HandOffTime fwtypes.ListNestedObjectValueOf[handOffTime] `tfsdk:"hand_off_time"`
}

func expandShiftCoverages(ctx context.Context, object []*shiftCoveragesData, diags *diag.Diagnostics) map[string][]awstypes.CoverageTime {
	if len(object) == 0 {
		return nil
	}

	result := make(map[string][]awstypes.CoverageTime)
	for _, v := range object {
		covTimes, diagErr := v.CoverageTimes.ToSlice(ctx)
		diags.Append(diagErr...)
		if diags.HasError() {
			return nil
		}

		var cTimes []awstypes.CoverageTime
		for _, val := range covTimes {
			end, diagErr := val.End.ToPtr(ctx)
			diags.Append(diagErr...)
			if diags.HasError() {
				return nil
			}
			start, diagErr := val.Start.ToPtr(ctx)
			diags.Append(diagErr...)
			if diags.HasError() {
				return nil
			}

			cTimes = append(cTimes, awstypes.CoverageTime{
				End: &awstypes.HandOffTime{
					HourOfDay:    fwflex.Int32ValueFromFramework(ctx, end.HourOfDay),
					MinuteOfHour: fwflex.Int32ValueFromFramework(ctx, end.MinuteOfHour),
				},
				Start: &awstypes.HandOffTime{
					HourOfDay:    fwflex.Int32ValueFromFramework(ctx, start.HourOfDay),
					MinuteOfHour: fwflex.Int32ValueFromFramework(ctx, start.MinuteOfHour),
				},
			})
		}

		result[v.MapBlockKey.ValueString()] = cTimes
	}

	return result
}

func flattenShiftCoverages(ctx context.Context, object map[string][]awstypes.CoverageTime) fwtypes.ListNestedObjectValueOf[shiftCoveragesData] {
	if len(object) == 0 {
		return fwtypes.NewListNestedObjectValueOfNull[shiftCoveragesData](ctx)
	}

	var output []shiftCoveragesData
	for key, value := range object {
		sc := shiftCoveragesData{
			MapBlockKey: fwtypes.StringEnumValue[awstypes.DayOfWeek](awstypes.DayOfWeek(key)),
		}

		var coverageTimes []coverageTimesData
		for _, v := range value {
			ct := coverageTimesData{
				End: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &handOffTime{
					HourOfDay:    fwflex.Int32ValueToFramework(ctx, v.End.HourOfDay),
					MinuteOfHour: fwflex.Int32ValueToFramework(ctx, v.End.MinuteOfHour),
				}),
				Start: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &handOffTime{
					HourOfDay:    fwflex.Int32ValueToFramework(ctx, v.Start.HourOfDay),
					MinuteOfHour: fwflex.Int32ValueToFramework(ctx, v.End.MinuteOfHour),
				}),
			}
			coverageTimes = append(coverageTimes, ct)
		}
		sc.CoverageTimes = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, coverageTimes)

		output = append(output, sc)
	}

	return fwtypes.NewListNestedObjectValueOfValueSliceMust[shiftCoveragesData](ctx, output)
}

func findRotationByID(ctx context.Context, conn *ssmcontacts.Client, id string) (*ssmcontacts.GetRotationOutput, error) {
	in := &ssmcontacts.GetRotationInput{
		RotationId: aws.String(id),
	}
	out, err := conn.GetRotation(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type objectForInput[T any] struct {
	Data fwtypes.ListNestedObjectValueOf[T]
}

type objectForOutput[T any] struct {
	Data []T
}

func setupSerializationObjects[T any, V any](input fwtypes.ListNestedObjectValueOf[T]) (objectForInput[T], objectForOutput[V]) { //nolint:unparam
	in := objectForInput[T]{
		Data: input,
	}

	return in, objectForOutput[V]{}
}
