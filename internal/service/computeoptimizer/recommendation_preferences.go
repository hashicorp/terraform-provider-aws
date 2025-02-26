// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package computeoptimizer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/computeoptimizer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/computeoptimizer/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_computeoptimizer_recommendation_preferences", name="Recommendation Preferences")
func newRecommendationPreferencesResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &recommendationPreferencesResource{}

	return r, nil
}

type recommendationPreferencesResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *recommendationPreferencesResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{

		Attributes: map[string]schema.Attribute{
			"enhanced_infrastructure_metrics": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EnhancedInfrastructureMetrics](),
				Optional:   true,
			},
			names.AttrID: framework.IDAttribute(),
			"inferred_workload_types": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.InferredWorkloadTypesPreference](),
				Optional:   true,
			},
			"look_back_period": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.LookBackPeriodPreference](),
				Optional:   true,
				Computed:   true,
			},
			names.AttrResourceType: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(enum.Slice(awstypes.ResourceTypeAutoScalingGroup, awstypes.ResourceTypeEc2Instance, awstypes.ResourceTypeRdsDbInstance)...),
				},
			},
			"savings_estimation_mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.SavingsEstimationMode](),
				Optional:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"external_metrics_preference": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[externalMetricsPreferenceModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrSource: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ExternalMetricsSource](),
							Required:   true,
						},
					},
				},
			},
			"preferred_resource": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[preferredResourceModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						objectvalidator.AtLeastOneOf(
							path.MatchRelative().AtParent().AtName("exclude_list"),
							path.MatchRelative().AtParent().AtName("include_list"),
						),
					},
					Attributes: map[string]schema.Attribute{
						"exclude_list": schema.SetAttribute{
							CustomType: fwtypes.SetOfStringType,
							Optional:   true,
							Validators: []validator.Set{
								setvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("include_list")),
								setvalidator.SizeAtMost(1000),
							},
						},
						"include_list": schema.SetAttribute{
							CustomType: fwtypes.SetOfStringType,
							Optional:   true,
							Validators: []validator.Set{
								setvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("exclude_list")),
								setvalidator.SizeAtMost(1000),
							},
						},
						names.AttrName: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.PreferredResourceName](),
							Required:   true,
						},
					},
				},
			},
			names.AttrScope: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[scopeModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ScopeName](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						names.AttrValue: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			"utilization_preference": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[utilizationPreferenceModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrMetricName: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.CustomizableMetricName](),
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"metric_parameters": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[customizableMetricParametersModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"headroom": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.CustomizableMetricHeadroom](),
										Required:   true,
									},
									"threshold": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.CustomizableMetricThreshold](),
										Optional:   true,
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

func (r *recommendationPreferencesResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data recommendationPreferencesResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ComputeOptimizerClient(ctx)

	input := &computeoptimizer.PutRecommendationPreferencesInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutRecommendationPreferences(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating Compute Optimizer Recommendation Preferences", err.Error())

		return
	}

	// Set values for unknowns.
	id, err := data.setID(ctx)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("flattening resource ID Compute Optimizer Recommendation Preferences (%s)", data.ID.ValueString()), err.Error())
		return
	}
	data.ID = types.StringValue(id)

	// Read the resource to get other Computed attribute values.
	scope := fwdiag.Must(data.Scope.ToPtr(ctx))
	output, err := findRecommendationPreferencesByThreePartKey(ctx, conn, data.ResourceType.ValueString(), scope.Name.ValueString(), scope.Value.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Compute Optimizer Recommendation Preferences (%s)", data.ID.ValueString()), err.Error())

		return
	}

	data.LookBackPeriod = fwtypes.StringEnumValue(output.LookBackPeriod)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *recommendationPreferencesResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data recommendationPreferencesResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(ctx); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().ComputeOptimizerClient(ctx)

	scope := fwdiag.Must(data.Scope.ToPtr(ctx))
	output, err := findRecommendationPreferencesByThreePartKey(ctx, conn, data.ResourceType.ValueString(), scope.Name.ValueString(), scope.Value.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Compute Optimizer Recommendation Preferences (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *recommendationPreferencesResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new recommendationPreferencesResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ComputeOptimizerClient(ctx)

	input := &computeoptimizer.PutRecommendationPreferencesInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutRecommendationPreferences(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Compute Optimizer Recommendation Preferences (%s)", new.ID.String()), err.Error())

		return
	}

	// Read the resource to get Computed attribute values.
	scope := fwdiag.Must(new.Scope.ToPtr(ctx))
	output, err := findRecommendationPreferencesByThreePartKey(ctx, conn, new.ResourceType.ValueString(), scope.Name.ValueString(), scope.Value.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Compute Optimizer Recommendation Preferences (%s)", new.ID.ValueString()), err.Error())

		return
	}

	new.LookBackPeriod = fwtypes.StringEnumValue(output.LookBackPeriod)

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *recommendationPreferencesResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data recommendationPreferencesResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ComputeOptimizerClient(ctx)

	input := &computeoptimizer.DeleteRecommendationPreferencesInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.RecommendationPreferenceNames = enum.EnumValues[awstypes.RecommendationPreferenceName]()

	_, err := conn.DeleteRecommendationPreferences(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Compute Optimizer Recommendation Preferences (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *recommendationPreferencesResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("enhanced_infrastructure_metrics"),
			path.MatchRoot("external_metrics_preference"),
			path.MatchRoot("inferred_workload_types"),
			path.MatchRoot("look_back_period"),
			path.MatchRoot("preferred_resource"),
			path.MatchRoot("savings_estimation_mode"),
			path.MatchRoot("utilization_preference"),
		),
	}
}

func findRecommendationPreferencesByThreePartKey(ctx context.Context, conn *computeoptimizer.Client, resourceType, scopeName, scopeValue string) (*awstypes.RecommendationPreferencesDetail, error) {
	input := &computeoptimizer.GetRecommendationPreferencesInput{
		ResourceType: awstypes.ResourceType(resourceType),
	}
	if scopeName != "" && scopeValue != "" {
		input.Scope = &awstypes.Scope{
			Name:  awstypes.ScopeName(scopeName),
			Value: aws.String(scopeValue),
		}
	}

	return findRecommendationPreferences(ctx, conn, input)
}

func findRecommendationPreferences(ctx context.Context, conn *computeoptimizer.Client, input *computeoptimizer.GetRecommendationPreferencesInput) (*awstypes.RecommendationPreferencesDetail, error) {
	output, err := findRecommendationPreferenceses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findRecommendationPreferenceses(ctx context.Context, conn *computeoptimizer.Client, input *computeoptimizer.GetRecommendationPreferencesInput) ([]awstypes.RecommendationPreferencesDetail, error) {
	var output []awstypes.RecommendationPreferencesDetail

	pages := computeoptimizer.NewGetRecommendationPreferencesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.RecommendationPreferencesDetails...)
	}

	return output, nil
}

type recommendationPreferencesResourceModel struct {
	EnhancedInfrastructureMetrics fwtypes.StringEnum[awstypes.EnhancedInfrastructureMetrics]      `tfsdk:"enhanced_infrastructure_metrics"`
	ExternalMetricsPreference     fwtypes.ListNestedObjectValueOf[externalMetricsPreferenceModel] `tfsdk:"external_metrics_preference"`
	ID                            types.String                                                    `tfsdk:"id"`
	InferredWorkloadTypes         fwtypes.StringEnum[awstypes.InferredWorkloadTypesPreference]    `tfsdk:"inferred_workload_types"`
	LookBackPeriod                fwtypes.StringEnum[awstypes.LookBackPeriodPreference]           `tfsdk:"look_back_period"`
	PreferredResources            fwtypes.ListNestedObjectValueOf[preferredResourceModel]         `tfsdk:"preferred_resource"`
	ResourceType                  types.String                                                    `tfsdk:"resource_type"`
	SavingsEstimationMode         fwtypes.StringEnum[awstypes.SavingsEstimationMode]              `tfsdk:"savings_estimation_mode"`
	Scope                         fwtypes.ListNestedObjectValueOf[scopeModel]                     `tfsdk:"scope"`
	UtilizationPreferences        fwtypes.ListNestedObjectValueOf[utilizationPreferenceModel]     `tfsdk:"utilization_preference"`
}

const (
	recommendationPreferencesResourceIDPartCount = 3
)

func (m *recommendationPreferencesResourceModel) InitFromID(ctx context.Context) error {
	parts, err := flex.ExpandResourceId(m.ID.ValueString(), recommendationPreferencesResourceIDPartCount, false)
	if err != nil {
		return err
	}

	m.ResourceType = types.StringValue(parts[0])
	scope := &scopeModel{
		Name:  fwtypes.StringEnumValue(awstypes.ScopeName(parts[1])),
		Value: types.StringValue(parts[2]),
	}
	m.Scope = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, scope)

	return nil
}

func (m *recommendationPreferencesResourceModel) setID(ctx context.Context) (string, error) {
	scope := fwdiag.Must(m.Scope.ToPtr(ctx))
	parts := []string{
		m.ResourceType.ValueString(),
		scope.Name.ValueString(),
		scope.Value.ValueString(),
	}

	return flex.FlattenResourceId(parts, recommendationPreferencesResourceIDPartCount, false)
}

type externalMetricsPreferenceModel struct {
	Source fwtypes.StringEnum[awstypes.ExternalMetricsSource] `tfsdk:"source"`
}

type preferredResourceModel struct {
	ExcludeList fwtypes.SetValueOf[types.String]                   `tfsdk:"exclude_list"`
	IncludeList fwtypes.SetValueOf[types.String]                   `tfsdk:"include_list"`
	Name        fwtypes.StringEnum[awstypes.PreferredResourceName] `tfsdk:"name"`
}

type scopeModel struct {
	Name  fwtypes.StringEnum[awstypes.ScopeName] `tfsdk:"name"`
	Value types.String                           `tfsdk:"value"`
}

type utilizationPreferenceModel struct {
	MetricName       fwtypes.StringEnum[awstypes.CustomizableMetricName]                `tfsdk:"metric_name"`
	MetricParameters fwtypes.ListNestedObjectValueOf[customizableMetricParametersModel] `tfsdk:"metric_parameters"`
}

type customizableMetricParametersModel struct {
	Headroom  fwtypes.StringEnum[awstypes.CustomizableMetricHeadroom]  `tfsdk:"headroom"`
	Threshold fwtypes.StringEnum[awstypes.CustomizableMetricThreshold] `tfsdk:"threshold"`
}
