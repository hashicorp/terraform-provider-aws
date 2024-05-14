// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ssmcontacts/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Rotation")
func newDataSourceRotation(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &dataSourceRotation{}

	return d, nil
}

type dataSourceRotation struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceRotation) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_ssmcontacts_rotation"
}

func (d *dataSourceRotation) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"contact_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			"recurrence": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[dsRecurrenceData](ctx),
				ElementType: fwtypes.NewObjectTypeOf[dsRecurrenceData](ctx),
				Computed:    true,
			},
			names.AttrStartTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"time_zone_id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *dataSourceRotation) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	conn := d.Meta().SSMContactsClient(ctx)
	var data dataSourceRotationData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	output, err := findRotationByID(ctx, conn, data.ARN.ValueString())

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMContacts, create.ErrActionSetting, ResNameRotation, data.ARN.ValueString(), err),
			err.Error(),
		)
		return
	}

	rc := &dsRecurrenceData{}
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

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.ContactIds, &data.ContactIds)...)
	if response.Diagnostics.HasError() {
		return
	}

	rc.ShiftCoverages = flattenShiftCoveragesDataSource(ctx, output.Recurrence.ShiftCoverages)

	data.Name = fwflex.StringToFramework(ctx, output.Name)
	data.Recurrence = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, rc)
	data.StartTime = fwflex.TimeToFramework(ctx, output.StartTime)
	data.TimeZoneID = fwflex.StringToFramework(ctx, output.TimeZoneId)
	data.ID = fwflex.StringToFramework(ctx, output.RotationArn)

	tags, err := listTags(ctx, conn, data.ARN.ValueString())

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMContacts, create.ErrActionSetting, ResNameRotation, data.ARN.ValueString(), err),
			err.Error(),
		)
		return
	}

	data.Tags = fwflex.FlattenFrameworkStringValueMap(ctx, tags.Map())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceRotationData struct {
	ARN        fwtypes.ARN                                       `tfsdk:"arn"`
	ContactIds fwtypes.ListValueOf[types.String]                 `tfsdk:"contact_ids"`
	ID         types.String                                      `tfsdk:"id"`
	Recurrence fwtypes.ListNestedObjectValueOf[dsRecurrenceData] `tfsdk:"recurrence"`
	Name       types.String                                      `tfsdk:"name"`
	StartTime  timetypes.RFC3339                                 `tfsdk:"start_time"`
	Tags       types.Map                                         `tfsdk:"tags"`
	TimeZoneID types.String                                      `tfsdk:"time_zone_id"`
}

type dsRecurrenceData struct {
	DailySettings        fwtypes.ListNestedObjectValueOf[dsHandOffTime]         `tfsdk:"daily_settings"`
	MonthlySettings      fwtypes.ListNestedObjectValueOf[dsMonthlySettingsData] `tfsdk:"monthly_settings"`
	NumberOfOnCalls      types.Int64                                            `tfsdk:"number_of_on_calls"`
	RecurrenceMultiplier types.Int64                                            `tfsdk:"recurrence_multiplier"`
	ShiftCoverages       fwtypes.ListNestedObjectValueOf[dsShiftCoveragesData]  `tfsdk:"shift_coverages"`
	WeeklySettings       fwtypes.ListNestedObjectValueOf[dsWeeklySettingsData]  `tfsdk:"weekly_settings"`
}
type dsMonthlySettingsData struct {
	DayOfMonth  types.Int64                                    `tfsdk:"day_of_month"`
	HandOffTime fwtypes.ListNestedObjectValueOf[dsHandOffTime] `tfsdk:"hand_off_time"`
}

type dsShiftCoveragesData struct {
	CoverageTimes fwtypes.ListNestedObjectValueOf[dsCoverageTimesData] `tfsdk:"coverage_times"`
	MapBlockKey   fwtypes.StringEnum[awstypes.DayOfWeek]               `tfsdk:"map_block_key"`
}

type dsCoverageTimesData struct {
	End   fwtypes.ListNestedObjectValueOf[dsHandOffTime] `tfsdk:"end"`
	Start fwtypes.ListNestedObjectValueOf[dsHandOffTime] `tfsdk:"start"`
}

type dsHandOffTime struct {
	HourOfDay    types.Int64 `tfsdk:"hour_of_day"`
	MinuteOfHour types.Int64 `tfsdk:"minute_of_hour"`
}

type dsWeeklySettingsData struct {
	DayOfWeek   fwtypes.StringEnum[awstypes.DayOfWeek]         `tfsdk:"day_of_week"`
	HandOffTime fwtypes.ListNestedObjectValueOf[dsHandOffTime] `tfsdk:"hand_off_time"`
}

func flattenShiftCoveragesDataSource(ctx context.Context, object map[string][]awstypes.CoverageTime) fwtypes.ListNestedObjectValueOf[dsShiftCoveragesData] {
	if len(object) == 0 {
		return fwtypes.NewListNestedObjectValueOfNull[dsShiftCoveragesData](ctx)
	}

	var output []dsShiftCoveragesData
	for key, value := range object {
		sc := dsShiftCoveragesData{
			MapBlockKey: fwtypes.StringEnumValue[awstypes.DayOfWeek](awstypes.DayOfWeek(key)),
		}

		var coverageTimes []dsCoverageTimesData
		for _, v := range value {
			ct := dsCoverageTimesData{
				End: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dsHandOffTime{
					HourOfDay:    fwflex.Int32ValueToFramework(ctx, v.End.HourOfDay),
					MinuteOfHour: fwflex.Int32ValueToFramework(ctx, v.End.MinuteOfHour),
				}),
				Start: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dsHandOffTime{
					HourOfDay:    fwflex.Int32ValueToFramework(ctx, v.Start.HourOfDay),
					MinuteOfHour: fwflex.Int32ValueToFramework(ctx, v.End.MinuteOfHour),
				}),
			}
			coverageTimes = append(coverageTimes, ct)
		}
		sc.CoverageTimes = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, coverageTimes)

		output = append(output, sc)
	}

	return fwtypes.NewListNestedObjectValueOfValueSliceMust[dsShiftCoveragesData](ctx, output)
}
