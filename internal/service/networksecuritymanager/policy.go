// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networksecuritymanager_policy", name="Policy")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/networksecuritymanager;networksecuritymanager.GetPolicyOutput")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheck")
// @Testing(identityRegionOverrideTest=false)
func newPolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &policyResource{}, nil
}

type policyResource struct {
	framework.ResourceWithModel[policyResourceModel]
	framework.WithImportByIdentity
}

var namePattern = regexache.MustCompile(`^[0-9A-Za-z][0-9A-Za-z _.:/=+\-@]*$`)

func (r *policyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(256),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z _.:/=+\-@]*$`), ""),
				},
			},
			"firewall_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PolicyFirewallType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(namePattern, ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrPriority: schema.Int32Attribute{
				Required: true,
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"policy_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[policyConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"remediation_enabled": schema.BoolAttribute{
							Required: true,
						},
						"resources_clean_up": schema.BoolAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"waf_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[wafConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"conflict_resolution": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.WAFConflictResolutionOptions](),
										Required:   true,
									},
									"existing_customer_web_acl_resolution": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.ExistingCustomerWebACLResolution](),
										Required:   true,
									},
								},
							},
						},
					},
				},
			},
			"template_or_rule": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[templateOrRuleModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(100),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"rule_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							Validators: []validator.String{
								stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("template_arn")),
							},
						},
						"template_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
					},
				},
			},
		},
	}
}

func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan policyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input networksecuritymanager.CreatePolicyInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Policy")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.IsPublished = aws.Bool(true)
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreatePolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan, fwflex.WithFieldNamePrefix("Policy")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state policyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	out, err := findPolicyByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state, fwflex.WithFieldNamePrefix("Policy")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state policyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		// Updates are guarded by the token returned from the most recent read or write.
		current, err := findPolicyByARN(ctx, conn, arn)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		var input networksecuritymanager.UpdatePolicyInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Policy")))
		if resp.Diagnostics.HasError() {
			return
		}

		input.ClientToken = aws.String(create.UniqueId(ctx))
		// Always publish so that a Terraform-managed policy never carries a dangling draft.
		input.IsPublished = aws.Bool(true)
		input.PolicyIdentifier = aws.String(arn)
		input.UpdateToken = current.UpdateToken

		_, err = conn.UpdatePolicy(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state policyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	input := networksecuritymanager.DeletePolicyInput{
		PolicyIdentifier: aws.String(arn),
	}
	_, err := conn.DeletePolicy(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}
}

func findPolicyByARN(ctx context.Context, conn *networksecuritymanager.Client, arn string) (*networksecuritymanager.GetPolicyOutput, error) {
	input := networksecuritymanager.GetPolicyInput{
		PolicyIdentifier: aws.String(arn),
	}

	out, err := conn.GetPolicy(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type policyResourceModel struct {
	framework.WithRegionModel
	ARN                           types.String                                              `tfsdk:"arn"`
	AssociatedTemplateAndRuleList fwtypes.ListNestedObjectValueOf[templateOrRuleModel]      `tfsdk:"template_or_rule"`
	Description                   types.String                                              `tfsdk:"description"`
	FirewallType                  fwtypes.StringEnum[awstypes.PolicyFirewallType]           `tfsdk:"firewall_type"`
	Name                          types.String                                              `tfsdk:"name"`
	PolicyConfiguration           fwtypes.ListNestedObjectValueOf[policyConfigurationModel] `tfsdk:"policy_configuration"`
	Priority                      types.Int32                                               `tfsdk:"priority"`
	Tags                          tftags.Map                                                `tfsdk:"tags"`
	TagsAll                       tftags.Map                                                `tfsdk:"tags_all"`
}

type policyConfigurationModel struct {
	RemediationEnabled types.Bool                                      `tfsdk:"remediation_enabled"`
	ResourcesCleanUp   types.Bool                                      `tfsdk:"resources_clean_up"`
	WafConfig          fwtypes.ListNestedObjectValueOf[wafConfigModel] `tfsdk:"waf_config"`
}

type wafConfigModel struct {
	ConflictResolution               fwtypes.StringEnum[awstypes.WAFConflictResolutionOptions]     `tfsdk:"conflict_resolution"`
	ExistingCustomerWebACLResolution fwtypes.StringEnum[awstypes.ExistingCustomerWebACLResolution] `tfsdk:"existing_customer_web_acl_resolution"`
}

// templateOrRuleModel maps to the TemplateOrRuleReference (input) and
// AssociatedTemplateOrRule (output) unions.
type templateOrRuleModel struct {
	RuleARN     fwtypes.ARN `tfsdk:"rule_arn"`
	TemplateARN fwtypes.ARN `tfsdk:"template_arn"`
}

var (
	_ fwflex.Expander  = templateOrRuleModel{}
	_ fwflex.Flattener = &templateOrRuleModel{}
)

func (m templateOrRuleModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.RuleARN.IsNull():
		return &awstypes.TemplateOrRuleReferenceMemberRuleIdentifier{
			Value: m.RuleARN.ValueString(),
		}, diags
	case !m.TemplateARN.IsNull():
		return &awstypes.TemplateOrRuleReferenceMemberTemplateIdentifier{
			Value: m.TemplateARN.ValueString(),
		}, diags
	}

	return nil, diags
}

func (m *templateOrRuleModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.AssociatedTemplateOrRuleMemberRuleArn:
		m.RuleARN = fwtypes.ARNValue(t.Value)
	case awstypes.AssociatedTemplateOrRuleMemberTemplateArn:
		m.TemplateARN = fwtypes.ARNValue(t.Value)
	}

	return diags
}
