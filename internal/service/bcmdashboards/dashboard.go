// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bcmdashboards

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bcmdashboards"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bcmdashboards/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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

// @FrameworkResource("aws_bcmdashboards_dashboard",name="Dashboard")
// @Tags(identifierAttribute="arn")
// @ArnIdentity(identityDuplicateAttributes="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bcmdashboards;bcmdashboards.GetDashboardOutput")
// @Testing(hasNoPreExistingResource=true)
// @Testing(skipEmptyTags=true, skipNullTags=true)
func newDashboardResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &dashboardResource{}, nil
}

const (
	ResNameDashboard = "Dashboard"

	// dashboardWidgetsMax is the maximum number of widgets allowed per dashboard.
	dashboardWidgetsMax = 20
)

type dashboardResource struct {
	framework.ResourceWithModel[dashboardResourceModel]
	framework.WithImportByIdentity
}

func (r *dashboardResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"dashboard_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DashboardType](),
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrARN)),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"widget": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[widgetModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(dashboardWidgetsMax),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
						},
						"height": schema.Int64Attribute{
							Optional: true,
						},
						"horizontal_offset": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						names.AttrID: schema.StringAttribute{
							// Computed without UseStateForUnknown: the API regenerates
							// the widget ID when the widget's content changes.
							Computed: true,
						},
						"title": schema.StringAttribute{
							Required: true,
						},
						"width": schema.Int64Attribute{
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						"configs": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[widgetConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 2),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"display_config":   displayConfigSchema(ctx),
									"query_parameters": queryParametersSchema(ctx),
								},
							},
						},
					},
				},
			},
		},
	}
}

func queryParametersSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[queryParametersModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"cost_and_usage": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[costAndUsageQueryModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"granularity": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.Granularity](),
								Required:   true,
							},
							"metrics": schema.ListAttribute{
								CustomType: fwtypes.ListOfStringEnumType[awstypes.MetricName](),
								Required:   true,
							},
						},
						Blocks: map[string]schema.Block{
							names.AttrFilter: filterExpressionSchema(ctx),
							"group_by":       groupDefinitionSchema(ctx),
							"time_range":     dateTimeRangeSchema(ctx),
						},
					},
				},
				"reservation_coverage": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[reservationCoverageQueryModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"granularity": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.Granularity](),
								Optional:   true,
							},
							"metrics": schema.ListAttribute{
								CustomType: fwtypes.ListOfStringEnumType[awstypes.MetricName](),
								Optional:   true,
							},
						},
						Blocks: map[string]schema.Block{
							names.AttrFilter: filterExpressionSchema(ctx),
							"group_by":       groupDefinitionSchema(ctx),
							"time_range":     dateTimeRangeSchema(ctx),
						},
					},
				},
				"reservation_utilization": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[reservationUtilizationQueryModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"granularity": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.Granularity](),
								Optional:   true,
							},
						},
						Blocks: map[string]schema.Block{
							names.AttrFilter: filterExpressionSchema(ctx),
							"group_by":       groupDefinitionSchema(ctx),
							"time_range":     dateTimeRangeSchema(ctx),
						},
					},
				},
				"savings_plans_coverage": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[savingsPlansCoverageQueryModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"granularity": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.Granularity](),
								Optional:   true,
							},
							"metrics": schema.ListAttribute{
								CustomType: fwtypes.ListOfStringEnumType[awstypes.MetricName](),
								Optional:   true,
							},
						},
						Blocks: map[string]schema.Block{
							names.AttrFilter: filterExpressionSchema(ctx),
							"group_by":       groupDefinitionSchema(ctx),
							"time_range":     dateTimeRangeSchema(ctx),
						},
					},
				},
				"savings_plans_utilization": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[savingsPlansUtilizationQueryModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"granularity": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.Granularity](),
								Optional:   true,
							},
						},
						Blocks: map[string]schema.Block{
							names.AttrFilter: filterExpressionSchema(ctx),
							"time_range":     dateTimeRangeSchema(ctx),
						},
					},
				},
			},
		},
	}
}

func dateTimeRangeSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[dateTimeRangeModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"end_time":   dateTimeValueSchema(ctx),
				"start_time": dateTimeValueSchema(ctx),
			},
		},
	}
}

func dateTimeValueSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[dateTimeValueModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrType: schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.DateTimeType](),
					Required:   true,
				},
				names.AttrValue: schema.StringAttribute{
					Required: true,
				},
			},
		},
	}
}

func groupDefinitionSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[groupDefinitionModel](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrKey: schema.StringAttribute{
					Required: true,
				},
				names.AttrType: schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.GroupDefinitionType](),
					Optional:   true,
				},
			},
		},
	}
}

// filterExpressionSchema returns the top-level (depth-1) filter expression schema.
//
// The AWS Expression type is recursive (And/Or/Not contain further Expressions),
// but Terraform schemas must be finite. The expression is therefore capped at two
// levels: the top level supports the logical operators and/or/not plus leaf
// filters, and the operands of and/or/not are leaf-only expressions (dimensions,
// tags, or cost_categories).
func filterExpressionSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[filterExpressionModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"and":             filterExpressionLeafSchema(ctx),
				"cost_categories": filterValuesSchema[costCategoryValuesModel](ctx),
				"dimensions":      dimensionValuesSchema(ctx),
				"not": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[filterExpressionLeafModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: filterExpressionLeafNestedObject(ctx),
				},
				"or":   filterExpressionLeafSchema(ctx),
				"tags": filterValuesSchema[tagValuesModel](ctx),
			},
		},
	}
}

func filterExpressionLeafSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType:   fwtypes.NewListNestedObjectTypeOf[filterExpressionLeafModel](ctx),
		NestedObject: filterExpressionLeafNestedObject(ctx),
	}
}

func filterExpressionLeafNestedObject(ctx context.Context) schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Blocks: map[string]schema.Block{
			"cost_categories": filterValuesSchema[costCategoryValuesModel](ctx),
			"dimensions":      dimensionValuesSchema(ctx),
			"tags":            filterValuesSchema[tagValuesModel](ctx),
		},
	}
}

func dimensionValuesSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[dimensionValuesModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrKey: schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.Dimension](),
					Required:   true,
				},
				"match_options": schema.ListAttribute{
					CustomType: fwtypes.ListOfStringEnumType[awstypes.MatchOption](),
					Optional:   true,
				},
				names.AttrValues: schema.ListAttribute{
					CustomType: fwtypes.ListOfStringType,
					Required:   true,
				},
			},
		},
	}
}

// filterValuesSchema returns the schema for tag- or cost-category-based filter
// values, which share the same shape (key, values, match_options).
func filterValuesSchema[T any](ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[T](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrKey: schema.StringAttribute{
					Optional: true,
				},
				"match_options": schema.ListAttribute{
					CustomType: fwtypes.ListOfStringEnumType[awstypes.MatchOption](),
					Optional:   true,
				},
				names.AttrValues: schema.ListAttribute{
					CustomType: fwtypes.ListOfStringType,
					Optional:   true,
				},
			},
		},
	}
}

func displayConfigSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[displayConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"graph": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[graphDisplayConfigModel](ctx),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"metric": schema.StringAttribute{
								Required: true,
							},
							"visual_type": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.VisualType](),
								Required:   true,
							},
						},
					},
				},
				"table": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[tableDisplayConfigModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{},
				},
			},
		},
	}
}

func (r *dashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BCMDashboardsClient(ctx)

	var plan dashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input bcmdashboards.CreateDashboardInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.ResourceTags = getTagsIn(ctx)

	out, err := conn.CreateDashboard(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDashboards, create.ErrActionCreating, ResNameDashboard, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.Arn)
	plan.ID = plan.ARN

	dashboard, err := findDashboardByARN(ctx, conn, plan.ARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDashboards, create.ErrActionReading, ResNameDashboard, plan.ARN.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenDashboard(ctx, dashboard, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *dashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BCMDashboardsClient(ctx)

	var state dashboardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDashboardByARN(ctx, conn, state.ARN.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDashboards, create.ErrActionReading, ResNameDashboard, state.ARN.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenDashboard(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BCMDashboardsClient(ctx)

	var plan, state dashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) || !plan.Description.Equal(state.Description) || !plan.Widgets.Equal(state.Widgets) {
		var input bcmdashboards.UpdateDashboardInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		input.Arn = state.ARN.ValueStringPointer()

		_, err := conn.UpdateDashboard(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BCMDashboards, create.ErrActionUpdating, ResNameDashboard, state.ARN.ValueString(), err),
				err.Error(),
			)
			return
		}
	}

	dashboard, err := findDashboardByARN(ctx, conn, state.ARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDashboards, create.ErrActionReading, ResNameDashboard, state.ARN.ValueString(), err),
			err.Error(),
		)
		return
	}

	plan.ARN = state.ARN
	plan.ID = state.ID
	resp.Diagnostics.Append(flattenDashboard(ctx, dashboard, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BCMDashboardsClient(ctx)

	var state dashboardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := bcmdashboards.DeleteDashboardInput{
		Arn: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteDashboard(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDashboards, create.ErrActionDeleting, ResNameDashboard, state.ARN.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func findDashboardByARN(ctx context.Context, conn *bcmdashboards.Client, arn string) (*bcmdashboards.GetDashboardOutput, error) {
	input := bcmdashboards.GetDashboardInput{
		Arn: aws.String(arn),
	}

	out, err := conn.GetDashboard(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

type dashboardResourceModel struct {
	ARN           types.String                                 `tfsdk:"arn"`
	CreatedAt     timetypes.RFC3339                            `tfsdk:"created_at"`
	DashboardType fwtypes.StringEnum[awstypes.DashboardType]   `tfsdk:"dashboard_type"`
	Description   types.String                                 `tfsdk:"description"`
	ID            types.String                                 `tfsdk:"id"`
	Name          types.String                                 `tfsdk:"name"`
	Tags          tftags.Map                                   `tfsdk:"tags"`
	TagsAll       tftags.Map                                   `tfsdk:"tags_all"`
	UpdatedAt     timetypes.RFC3339                            `tfsdk:"updated_at"`
	Widgets       fwtypes.ListNestedObjectValueOf[widgetModel] `tfsdk:"widget"`
}

type widgetModel struct {
	Configs          fwtypes.ListNestedObjectValueOf[widgetConfigModel] `tfsdk:"configs"`
	Description      types.String                                       `tfsdk:"description"`
	Height           types.Int64                                        `tfsdk:"height"`
	HorizontalOffset types.Int64                                        `tfsdk:"horizontal_offset"`
	ID               types.String                                       `tfsdk:"id"`
	Title            types.String                                       `tfsdk:"title"`
	Width            types.Int64                                        `tfsdk:"width"`
}

type widgetConfigModel struct {
	DisplayConfig   fwtypes.ListNestedObjectValueOf[displayConfigModel]   `tfsdk:"display_config"`
	QueryParameters fwtypes.ListNestedObjectValueOf[queryParametersModel] `tfsdk:"query_parameters"`
}

type queryParametersModel struct {
	CostAndUsage            fwtypes.ListNestedObjectValueOf[costAndUsageQueryModel]            `tfsdk:"cost_and_usage"`
	ReservationCoverage     fwtypes.ListNestedObjectValueOf[reservationCoverageQueryModel]     `tfsdk:"reservation_coverage"`
	ReservationUtilization  fwtypes.ListNestedObjectValueOf[reservationUtilizationQueryModel]  `tfsdk:"reservation_utilization"`
	SavingsPlansCoverage    fwtypes.ListNestedObjectValueOf[savingsPlansCoverageQueryModel]    `tfsdk:"savings_plans_coverage"`
	SavingsPlansUtilization fwtypes.ListNestedObjectValueOf[savingsPlansUtilizationQueryModel] `tfsdk:"savings_plans_utilization"`
}

type costAndUsageQueryModel struct {
	Filter      fwtypes.ListNestedObjectValueOf[filterExpressionModel] `tfsdk:"filter"`
	Granularity fwtypes.StringEnum[awstypes.Granularity]               `tfsdk:"granularity"`
	GroupBy     fwtypes.ListNestedObjectValueOf[groupDefinitionModel]  `tfsdk:"group_by"`
	Metrics     fwtypes.ListOfStringEnum[awstypes.MetricName]          `tfsdk:"metrics"`
	TimeRange   fwtypes.ListNestedObjectValueOf[dateTimeRangeModel]    `tfsdk:"time_range"`
}

type reservationCoverageQueryModel struct {
	Filter      fwtypes.ListNestedObjectValueOf[filterExpressionModel] `tfsdk:"filter"`
	Granularity fwtypes.StringEnum[awstypes.Granularity]               `tfsdk:"granularity"`
	GroupBy     fwtypes.ListNestedObjectValueOf[groupDefinitionModel]  `tfsdk:"group_by"`
	Metrics     fwtypes.ListOfStringEnum[awstypes.MetricName]          `tfsdk:"metrics"`
	TimeRange   fwtypes.ListNestedObjectValueOf[dateTimeRangeModel]    `tfsdk:"time_range"`
}

type reservationUtilizationQueryModel struct {
	Filter      fwtypes.ListNestedObjectValueOf[filterExpressionModel] `tfsdk:"filter"`
	Granularity fwtypes.StringEnum[awstypes.Granularity]               `tfsdk:"granularity"`
	GroupBy     fwtypes.ListNestedObjectValueOf[groupDefinitionModel]  `tfsdk:"group_by"`
	TimeRange   fwtypes.ListNestedObjectValueOf[dateTimeRangeModel]    `tfsdk:"time_range"`
}

type savingsPlansCoverageQueryModel struct {
	Filter      fwtypes.ListNestedObjectValueOf[filterExpressionModel] `tfsdk:"filter"`
	Granularity fwtypes.StringEnum[awstypes.Granularity]               `tfsdk:"granularity"`
	GroupBy     fwtypes.ListNestedObjectValueOf[groupDefinitionModel]  `tfsdk:"group_by"`
	Metrics     fwtypes.ListOfStringEnum[awstypes.MetricName]          `tfsdk:"metrics"`
	TimeRange   fwtypes.ListNestedObjectValueOf[dateTimeRangeModel]    `tfsdk:"time_range"`
}

type savingsPlansUtilizationQueryModel struct {
	Filter      fwtypes.ListNestedObjectValueOf[filterExpressionModel] `tfsdk:"filter"`
	Granularity fwtypes.StringEnum[awstypes.Granularity]               `tfsdk:"granularity"`
	TimeRange   fwtypes.ListNestedObjectValueOf[dateTimeRangeModel]    `tfsdk:"time_range"`
}

type dateTimeRangeModel struct {
	EndTime   fwtypes.ListNestedObjectValueOf[dateTimeValueModel] `tfsdk:"end_time"`
	StartTime fwtypes.ListNestedObjectValueOf[dateTimeValueModel] `tfsdk:"start_time"`
}

type dateTimeValueModel struct {
	Type  fwtypes.StringEnum[awstypes.DateTimeType] `tfsdk:"type"`
	Value types.String                              `tfsdk:"value"`
}

type groupDefinitionModel struct {
	Key  types.String                                     `tfsdk:"key"`
	Type fwtypes.StringEnum[awstypes.GroupDefinitionType] `tfsdk:"type"`
}

type filterExpressionModel struct {
	And            fwtypes.ListNestedObjectValueOf[filterExpressionLeafModel] `tfsdk:"and"`
	CostCategories fwtypes.ListNestedObjectValueOf[costCategoryValuesModel]   `tfsdk:"cost_categories"`
	Dimensions     fwtypes.ListNestedObjectValueOf[dimensionValuesModel]      `tfsdk:"dimensions"`
	Not            fwtypes.ListNestedObjectValueOf[filterExpressionLeafModel] `tfsdk:"not"`
	Or             fwtypes.ListNestedObjectValueOf[filterExpressionLeafModel] `tfsdk:"or"`
	Tags           fwtypes.ListNestedObjectValueOf[tagValuesModel]            `tfsdk:"tags"`
}

type filterExpressionLeafModel struct {
	CostCategories fwtypes.ListNestedObjectValueOf[costCategoryValuesModel] `tfsdk:"cost_categories"`
	Dimensions     fwtypes.ListNestedObjectValueOf[dimensionValuesModel]    `tfsdk:"dimensions"`
	Tags           fwtypes.ListNestedObjectValueOf[tagValuesModel]          `tfsdk:"tags"`
}

type dimensionValuesModel struct {
	Key          fwtypes.StringEnum[awstypes.Dimension]         `tfsdk:"key"`
	MatchOptions fwtypes.ListOfStringEnum[awstypes.MatchOption] `tfsdk:"match_options"`
	Values       fwtypes.ListOfString                           `tfsdk:"values"`
}

type tagValuesModel struct {
	Key          types.String                                   `tfsdk:"key"`
	MatchOptions fwtypes.ListOfStringEnum[awstypes.MatchOption] `tfsdk:"match_options"`
	Values       fwtypes.ListOfString                           `tfsdk:"values"`
}

type costCategoryValuesModel struct {
	Key          types.String                                   `tfsdk:"key"`
	MatchOptions fwtypes.ListOfStringEnum[awstypes.MatchOption] `tfsdk:"match_options"`
	Values       fwtypes.ListOfString                           `tfsdk:"values"`
}

type displayConfigModel struct {
	Graph fwtypes.ListNestedObjectValueOf[graphDisplayConfigModel] `tfsdk:"graph"`
	Table fwtypes.ListNestedObjectValueOf[tableDisplayConfigModel] `tfsdk:"table"`
}

type graphDisplayConfigModel struct {
	Metric     types.String                            `tfsdk:"metric"`
	VisualType fwtypes.StringEnum[awstypes.VisualType] `tfsdk:"visual_type"`
}

type tableDisplayConfigModel struct{}
