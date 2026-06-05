// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bcmdashboards

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bcmdashboards"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bcmdashboards/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bcmdashboards_scheduled_report",name="Scheduled Report")
// @Tags(identifierAttribute="arn")
// @ArnIdentity(identityDuplicateAttributes="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bcmdashboards;bcmdashboards.GetScheduledReportOutput")
// @Testing(hasNoPreExistingResource=true)
// @Testing(skipEmptyTags=true, skipNullTags=true)
func newScheduledReportResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &scheduledReportResource{}, nil
}

const (
	ResNameScheduledReport = "Scheduled Report"

	// iamPropagationTimeout bounds Create retries while a newly-created
	// execution role becomes assumable by the service (IAM eventual consistency).
	iamPropagationTimeout = 2 * time.Minute
)

type scheduledReportResource struct {
	framework.ResourceWithModel[scheduledReportResourceModel]
	framework.WithImportByIdentity
}

func (r *scheduledReportResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"dashboard_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrARN)),
			"last_execution_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"scheduled_report_execution_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"widget_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Optional:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"schedule_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[scheduleConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"schedule_expression": schema.StringAttribute{
							Optional: true,
						},
						"schedule_expression_time_zone": schema.StringAttribute{
							Optional: true,
						},
						// The schedule period is represented as flat Optional+Computed
						// attributes (rather than a nested block) because the service
						// assigns default start/end times when they are not configured,
						// and protocol v5 does not support computed nested blocks.
						"schedule_period_end_time": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Optional:   true,
							Computed:   true,
						},
						"schedule_period_start_time": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Optional:   true,
							Computed:   true,
						},
						names.AttrState: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ScheduleState](),
							Optional:   true,
							Computed:   true,
						},
					},
				},
			},
			"widget_date_range_override": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dateTimeRangeModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"end_time":   dateTimeValueSchema(ctx),
						"start_time": dateTimeValueSchema(ctx),
					},
				},
			},
		},
	}
}

func (r *scheduledReportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BCMDashboardsClient(ctx)

	var plan scheduledReportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var scheduledReport awstypes.ScheduledReportInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &scheduledReport)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := bcmdashboards.CreateScheduledReportInput{
		ScheduledReport: &scheduledReport,
		ResourceTags:    getTagsIn(ctx),
	}

	// A newly-created execution role (and its policy) may not have propagated
	// when the service validates it (IAM eventual consistency), surfacing as a
	// ValidationException for either the trust relationship or the permissions.
	out, err := tfresource.RetryWhen(ctx, iamPropagationTimeout,
		func(ctx context.Context) (*bcmdashboards.CreateScheduledReportOutput, error) {
			return conn.CreateScheduledReport(ctx, &input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "assume execution role") ||
				errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Missing permissions for dashboard") {
				return true, err
			}
			return false, err
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDashboards, create.ErrActionCreating, ResNameScheduledReport, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.Arn)
	plan.ID = plan.ARN

	report, err := findScheduledReportByARN(ctx, conn, plan.ARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDashboards, create.ErrActionReading, ResNameScheduledReport, plan.ARN.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenScheduledReport(ctx, report, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *scheduledReportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BCMDashboardsClient(ctx)

	var state scheduledReportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findScheduledReportByARN(ctx, conn, state.ARN.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDashboards, create.ErrActionReading, ResNameScheduledReport, state.ARN.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenScheduledReport(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *scheduledReportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BCMDashboardsClient(ctx)

	var plan, state scheduledReportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) ||
		!plan.DashboardARN.Equal(state.DashboardARN) ||
		!plan.ScheduledReportExecutionRoleARN.Equal(state.ScheduledReportExecutionRoleARN) ||
		!plan.ScheduleConfig.Equal(state.ScheduleConfig) ||
		!plan.WidgetDateRangeOverride.Equal(state.WidgetDateRangeOverride) ||
		!plan.WidgetIDs.Equal(state.WidgetIDs) {
		var input bcmdashboards.UpdateScheduledReportInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		input.Arn = state.ARN.ValueStringPointer()
		input.ClearWidgetDateRangeOverride = plan.WidgetDateRangeOverride.IsNull()
		input.ClearWidgetIds = plan.WidgetIDs.IsNull()

		_, err := conn.UpdateScheduledReport(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BCMDashboards, create.ErrActionUpdating, ResNameScheduledReport, state.ARN.ValueString(), err),
				err.Error(),
			)
			return
		}
	}

	report, err := findScheduledReportByARN(ctx, conn, state.ARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDashboards, create.ErrActionReading, ResNameScheduledReport, state.ARN.ValueString(), err),
			err.Error(),
		)
		return
	}

	plan.ARN = state.ARN
	plan.ID = state.ID
	resp.Diagnostics.Append(flattenScheduledReport(ctx, report, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *scheduledReportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BCMDashboardsClient(ctx)

	var state scheduledReportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := bcmdashboards.DeleteScheduledReportInput{
		Arn: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteScheduledReport(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDashboards, create.ErrActionDeleting, ResNameScheduledReport, state.ARN.ValueString(), err),
			err.Error(),
		)
		return
	}
}

// flattenScheduledReport copies a GetScheduledReportOutput into the resource
// model. AutoFlex handles the scalar and nested fields; the ARN-derived id is
// set explicitly.
func flattenScheduledReport(ctx context.Context, out *bcmdashboards.GetScheduledReportOutput, model *scheduledReportResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(flex.Flatten(ctx, out.ScheduledReport, model)...)
	if diags.HasError() {
		return diags
	}

	model.ID = model.ARN

	return diags
}

func findScheduledReportByARN(ctx context.Context, conn *bcmdashboards.Client, arn string) (*bcmdashboards.GetScheduledReportOutput, error) {
	input := bcmdashboards.GetScheduledReportInput{
		Arn: aws.String(arn),
	}

	out, err := conn.GetScheduledReport(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.ScheduledReport == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

type scheduledReportResourceModel struct {
	ARN                             types.String                                         `tfsdk:"arn"`
	CreatedAt                       timetypes.RFC3339                                    `tfsdk:"created_at"`
	DashboardARN                    fwtypes.ARN                                          `tfsdk:"dashboard_arn"`
	Description                     types.String                                         `tfsdk:"description"`
	ID                              types.String                                         `tfsdk:"id"`
	LastExecutionAt                 timetypes.RFC3339                                    `tfsdk:"last_execution_at"`
	Name                            types.String                                         `tfsdk:"name"`
	ScheduleConfig                  fwtypes.ListNestedObjectValueOf[scheduleConfigModel] `tfsdk:"schedule_config"`
	ScheduledReportExecutionRoleARN fwtypes.ARN                                          `tfsdk:"scheduled_report_execution_role_arn"`
	Tags                            tftags.Map                                           `tfsdk:"tags"`
	TagsAll                         tftags.Map                                           `tfsdk:"tags_all"`
	UpdatedAt                       timetypes.RFC3339                                    `tfsdk:"updated_at"`
	WidgetDateRangeOverride         fwtypes.ListNestedObjectValueOf[dateTimeRangeModel]  `tfsdk:"widget_date_range_override"`
	WidgetIDs                       fwtypes.ListOfString                                 `tfsdk:"widget_ids"`
}

type scheduleConfigModel struct {
	ScheduleExpression         types.String                               `tfsdk:"schedule_expression"`
	ScheduleExpressionTimeZone types.String                               `tfsdk:"schedule_expression_time_zone"`
	SchedulePeriodEndTime      timetypes.RFC3339                          `tfsdk:"schedule_period_end_time"`
	SchedulePeriodStartTime    timetypes.RFC3339                          `tfsdk:"schedule_period_start_time"`
	State                      fwtypes.StringEnum[awstypes.ScheduleState] `tfsdk:"state"`
}

var (
	_ flex.Expander  = scheduleConfigModel{}
	_ flex.Flattener = (*scheduleConfigModel)(nil)
)

// Expand maps the flat schedule-period attributes onto the nested SDK
// SchedulePeriod (protocol v5 cannot represent a computed nested block).
func (m scheduleConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	out := awstypes.ScheduleConfig{
		ScheduleExpression:         flex.StringFromFramework(ctx, m.ScheduleExpression),
		ScheduleExpressionTimeZone: flex.StringFromFramework(ctx, m.ScheduleExpressionTimeZone),
	}
	if !m.State.IsNull() && !m.State.IsUnknown() {
		out.State = m.State.ValueEnum()
	}

	if v := m.SchedulePeriodStartTime; !v.IsNull() && !v.IsUnknown() {
		t, d := v.ValueRFC3339Time()
		diags.Append(d...)
		if out.SchedulePeriod == nil {
			out.SchedulePeriod = &awstypes.SchedulePeriod{}
		}
		out.SchedulePeriod.StartTime = aws.Time(t)
	}
	if v := m.SchedulePeriodEndTime; !v.IsNull() && !v.IsUnknown() {
		t, d := v.ValueRFC3339Time()
		diags.Append(d...)
		if out.SchedulePeriod == nil {
			out.SchedulePeriod = &awstypes.SchedulePeriod{}
		}
		out.SchedulePeriod.EndTime = aws.Time(t)
	}

	return &out, diags
}

func (m *scheduleConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	var sc awstypes.ScheduleConfig
	switch t := v.(type) {
	case awstypes.ScheduleConfig:
		sc = t
	case *awstypes.ScheduleConfig:
		sc = *t
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("flattening schedule config: %T", v))
		return diags
	}

	m.ScheduleExpression = flex.StringToFramework(ctx, sc.ScheduleExpression)
	m.ScheduleExpressionTimeZone = flex.StringToFramework(ctx, sc.ScheduleExpressionTimeZone)
	m.State = fwtypes.StringEnumValue(sc.State)

	if sc.SchedulePeriod != nil {
		m.SchedulePeriodStartTime = timetypes.NewRFC3339TimePointerValue(sc.SchedulePeriod.StartTime)
		m.SchedulePeriodEndTime = timetypes.NewRFC3339TimePointerValue(sc.SchedulePeriod.EndTime)
	} else {
		m.SchedulePeriodStartTime = timetypes.NewRFC3339Null()
		m.SchedulePeriodEndTime = timetypes.NewRFC3339Null()
	}

	return diags
}
