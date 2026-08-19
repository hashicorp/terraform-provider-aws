// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package xray

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/xray"
	awstypes "github.com/aws/aws-sdk-go-v2/service/xray/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_xray_indexing_rule", name="Indexing Rule")
// @IdentityAttribute("name")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="name")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/xray/types;awstypes;awstypes.IndexingRule")
// @Testing(checkDestroyNoop=true)
// @Testing(generator=false)
// @Testing(serialize=true)
// @Testing(plannableImportAction="Update")
func newIndexingRuleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &indexingRuleResource{}

	return r, nil
}

type indexingRuleResource struct {
	framework.ResourceWithModel[indexingRuleResourceModel]
	framework.WithNoOpDelete
	framework.WithImportByIdentity
}

func (r *indexingRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrRule: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[indexingRuleValueModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"probabilistic": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[indexingProbabilisticRuleValueModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"actual_sampling_percentage": schema.Float64Attribute{
										Computed: true,
									},
									"desired_sampling_percentage": schema.Float64Attribute{
										Required: true,
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

func (r *indexingRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan indexingRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().XRayClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var in xray.UpdateIndexingRuleInput
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.UpdateIndexingRule(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("creating XRay Indexing Rule (%s)", name), err.Error())
		return
	}

	// Set values for unknowns.
	resp.Diagnostics.Append(fwflex.Flatten(ctx, out.IndexingRule, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *indexingRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state indexingRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().XRayClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, state.Name)
	out, err := findIndexingRuleByName(ctx, conn, name)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading XRay Indexing Rule (%s)", name), err.Error())
		return
	}

	// Set attributes for import.
	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *indexingRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan indexingRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().XRayClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var in xray.UpdateIndexingRuleInput
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.UpdateIndexingRule(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("updating XRay Indexing Rule (%s)", name), err.Error())
		return
	}

	// Set values for unknowns.
	resp.Diagnostics.Append(fwflex.Flatten(ctx, out.IndexingRule, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func findIndexingRuleByName(ctx context.Context, conn *xray.Client, name string) (*awstypes.IndexingRule, error) {
	var input xray.GetIndexingRulesInput
	output, err := findIndexingRule(ctx, conn, &input, func(v awstypes.IndexingRule) bool {
		return aws.ToString(v.Name) == name
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func findIndexingRule(ctx context.Context, conn *xray.Client, input *xray.GetIndexingRulesInput, filter tfslices.Predicate[awstypes.IndexingRule]) (*awstypes.IndexingRule, error) {
	output, err := findIndexingRules(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIndexingRules(ctx context.Context, conn *xray.Client, input *xray.GetIndexingRulesInput, filter tfslices.Predicate[awstypes.IndexingRule]) ([]awstypes.IndexingRule, error) {
	var output []awstypes.IndexingRule
	err := getIndexingRulesPages(ctx, conn, input, func(page *xray.GetIndexingRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.IndexingRules {
			if filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

type indexingRuleResourceModel struct {
	framework.WithRegionModel
	Name types.String                                            `tfsdk:"name"`
	Rule fwtypes.ListNestedObjectValueOf[indexingRuleValueModel] `tfsdk:"rule"`
}

type indexingRuleValueModel struct {
	Probabilistic fwtypes.ListNestedObjectValueOf[indexingProbabilisticRuleValueModel] `tfsdk:"probabilistic"`
}

var (
	_ fwflex.Expander  = indexingRuleValueModel{}
	_ fwflex.Flattener = &indexingRuleValueModel{}
)

func (m *indexingRuleValueModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.IndexingRuleValueMemberProbabilistic:
		var data indexingProbabilisticRuleValueModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}
		m.Probabilistic = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("rule flatten: %T", v))
	}
	return diags
}

func (m indexingRuleValueModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.Probabilistic.IsNull():
		data, d := m.Probabilistic.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.IndexingRuleValueUpdateMemberProbabilistic
		diags.Append(fwflex.Expand(ctx, data, &r.Value)...)
		return &r, diags
	}
	return nil, diags
}

type indexingProbabilisticRuleValueModel struct {
	ActualSamplingPercentage  types.Float64 `tfsdk:"actual_sampling_percentage"`
	DesiredSamplingPercentage types.Float64 `tfsdk:"desired_sampling_percentage"`
}
