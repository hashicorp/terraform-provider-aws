// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"
	"fmt"
	"strconv"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(preCheck="testAccPreCheck")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/networksecuritymanager;networksecuritymanager;networksecuritymanager.GetPolicyOutput")
// @Testing(identityRegionOverrideTest=false)
func newPolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &policyResource{}

	return r, nil
}

const (
	// A WAF policy associates at most this many templates; the rest of its
	// list must be rules.
	policyMaxTemplates = 2
)

type policyResource struct {
	framework.ResourceWithModel[policyResourceModel]
	framework.WithImportByIdentity
}

func (r *policyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(256),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9 _.:/=+\-@]*$`), "must contain only alphanumeric characters, spaces, and _.:/=+-@"),
				},
			},
			"firewall_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PolicyFirewallType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"has_published_version": schema.BoolAttribute{
				Computed: true,
			},
			"is_published": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9 _.:/=+\-@]*$`), "must start with an alphanumeric character and contain only alphanumeric characters, spaces, and _.:/=+-@"),
				},
			},
			"policy_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrPriority: schema.Int32Attribute{
				Required: true,
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EntityStatus](),
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrVersion: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			// The list is ordered: the position of a template or rule in the
			// list is its position in the web ACL the policy builds.
			"associated_template_and_rule": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[policyAssociationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(100),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"rule_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							Validators: []validator.String{
								stringvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("rule_arn"),
									path.MatchRelative().AtParent().AtName("template_arn"),
								),
							},
						},
						"template_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
					},
				},
			},
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
		},
	}
}

// ValidateConfig enforces the constraints between the firewall type and the
// rest of the configuration that the API applies at write time: a WAF policy
// needs a WAF configuration and at least one template or rule, of which at
// most two may be templates; a Shield Advanced policy takes neither.
func (r *policyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config policyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &config))
	if resp.Diagnostics.HasError() {
		return
	}

	associationsPath := path.Root("associated_template_and_rule")
	wafConfigPath := path.Root("policy_configuration").AtListIndex(0).AtName("waf_config")

	var associations []*policyAssociationModel
	if !config.AssociatedTemplateAndRule.IsUnknown() {
		var d diag.Diagnostics
		associations, d = config.AssociatedTemplateAndRule.ToSlice(ctx)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Constraints that hold for any firewall type.
	templates := 0
	seen := make(map[string]struct{}, len(associations))
	for i, association := range associations {
		if association == nil {
			continue
		}
		if !association.TemplateARN.IsNull() {
			templates++
		}
		for _, v := range []fwtypes.ARN{association.RuleARN, association.TemplateARN} {
			if v.IsNull() || v.IsUnknown() {
				continue
			}
			if _, ok := seen[v.ValueString()]; ok {
				resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
					associationsPath.AtListIndex(i),
					"must not reference the same template or rule more than once",
					v.ValueString(),
				))
			}
			seen[v.ValueString()] = struct{}{}
		}
	}
	if templates > policyMaxTemplates {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			associationsPath,
			fmt.Sprintf("must reference at most %d templates", policyMaxTemplates),
			strconv.Itoa(templates),
		))
	}

	if config.FirewallType.IsUnknown() || config.PolicyConfiguration.IsUnknown() {
		return
	}

	var wafConfigSet bool
	if policyConfigurations, d := config.PolicyConfiguration.ToSlice(ctx); !d.HasError() && len(policyConfigurations) > 0 && policyConfigurations[0] != nil {
		wafConfig := policyConfigurations[0].WafConfig
		if wafConfig.IsUnknown() {
			return
		}
		wafConfigSet = wafConfig.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0
	}

	switch firewallType := config.FirewallType.ValueEnum(); firewallType {
	case awstypes.PolicyFirewallTypeWaf:
		if !wafConfigSet {
			resp.Diagnostics.Append(fwdiag.NewAttributeRequiredWhenError(wafConfigPath, path.Root("firewall_type"), string(firewallType)))
		}
		if !config.AssociatedTemplateAndRule.IsUnknown() && len(associations) == 0 {
			resp.Diagnostics.Append(fwdiag.NewAttributeRequiredWhenError(associationsPath, path.Root("firewall_type"), string(firewallType)))
		}

	case awstypes.PolicyFirewallTypeShieldAdvanced:
		if wafConfigSet {
			resp.Diagnostics.Append(fwdiag.NewAttributeConflictsWhenError(wafConfigPath, path.Root("firewall_type"), string(firewallType)))
		}
		if len(associations) > 0 {
			resp.Diagnostics.Append(fwdiag.NewAttributeConflictsWhenError(associationsPath, path.Root("firewall_type"), string(firewallType)))
		}
	}
}

func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var plan policyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	var input networksecuritymanager.CreatePolicyInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Policy")))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	associations, d := expandTemplateOrRuleReferences(ctx, plan.AssociatedTemplateAndRule)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}
	input.AssociatedTemplateAndRuleList = associations
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreatePolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenPolicy(ctx, (*networksecuritymanager.GetPolicyOutput)(output), &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var state policyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := ruleBaseARN(state.ARN.ValueString())
	output, err := findPolicyByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenPolicy(ctx, output, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var plan, state policyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := ruleBaseARN(state.ARN.ValueString())

	// Every write returns a new update token and the token is required on the
	// next update, so the current one is read back rather than kept in state:
	// a change made outside Terraform would otherwise fail the update with a
	// ConflictException.
	current, err := findPolicyByARN(ctx, conn, arn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	if !plan.AssociatedTemplateAndRule.Equal(state.AssociatedTemplateAndRule) ||
		!plan.Description.Equal(state.Description) ||
		!plan.IsPublished.Equal(state.IsPublished) ||
		!plan.PolicyConfiguration.Equal(state.PolicyConfiguration) ||
		!plan.Priority.Equal(state.Priority) {
		var input networksecuritymanager.UpdatePolicyInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Policy")))
		if resp.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		// A pending draft can only be updated (or published) through its
		// DRAFT-qualified ARN; the base ARN is rejected while a draft exists.
		input.PolicyIdentifier = aws.String(ruleUpdateIdentifier(arn, current.Status))
		input.UpdateToken = current.UpdateToken
		associations, d := expandTemplateOrRuleReferences(ctx, plan.AssociatedTemplateAndRule)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}
		input.AssociatedTemplateAndRuleList = associations
		if plan.Description.IsNull() {
			// An omitted description leaves the current one in place; an
			// empty string clears it.
			input.PolicyDescription = aws.String("")
		}

		output, err := conn.UpdatePolicy(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		current = (*networksecuritymanager.GetPolicyOutput)(output)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenPolicy(ctx, current, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var state policyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := ruleBaseARN(state.ARN.ValueString())

	// A published policy with a pending draft cannot be deleted until the
	// draft is; a policy that was never published exists only as a draft,
	// and deleting its base ARN succeeds without removing anything.
	for _, identifier := range []string{arn + ruleDraftQualifier, arn} {
		input := networksecuritymanager.DeletePolicyInput{
			PolicyIdentifier: aws.String(identifier),
		}

		_, err := conn.DeletePolicy(ctx, &input)
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			continue
		}
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}
	}
}

// findPolicyByARN returns the version of the policy that an update would act
// on: the pending draft when there is one, otherwise the published policy.
func findPolicyByARN(ctx context.Context, conn *networksecuritymanager.Client, arn string) (*networksecuritymanager.GetPolicyOutput, error) {
	output, err := findPolicyByIdentifier(ctx, conn, arn+ruleDraftQualifier)
	if retry.NotFound(err) {
		return findPolicyByIdentifier(ctx, conn, arn)
	}
	if err != nil {
		return nil, err
	}

	return output, nil
}

func findPolicyByIdentifier(ctx context.Context, conn *networksecuritymanager.Client, identifier string) (*networksecuritymanager.GetPolicyOutput, error) {
	input := networksecuritymanager.GetPolicyInput{
		PolicyIdentifier: aws.String(identifier),
	}

	output, err := conn.GetPolicy(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}

// expandTemplateOrRuleReferences returns nil for an empty list: the API
// rejects an empty list for a WAF policy and any list at all for a Shield
// Advanced policy, and an omitted list on an update leaves the current one.
func expandTemplateOrRuleReferences(ctx context.Context, tfList fwtypes.ListNestedObjectValueOf[policyAssociationModel]) ([]awstypes.TemplateOrRuleReference, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	var diags diag.Diagnostics

	if tfList.IsNull() || tfList.IsUnknown() {
		return nil, diags
	}

	tfObjects, d := tfList.ToSlice(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, diags
	}

	if len(tfObjects) == 0 {
		return nil, diags
	}

	apiObjects := make([]awstypes.TemplateOrRuleReference, 0, len(tfObjects))

	for _, tfObject := range tfObjects {
		switch {
		case !tfObject.RuleARN.IsNull():
			apiObjects = append(apiObjects, &awstypes.TemplateOrRuleReferenceMemberRuleIdentifier{
				Value: tfObject.RuleARN.ValueString(),
			})
		case !tfObject.TemplateARN.IsNull():
			apiObjects = append(apiObjects, &awstypes.TemplateOrRuleReferenceMemberTemplateIdentifier{
				Value: tfObject.TemplateARN.ValueString(),
			})
		}
	}

	return apiObjects, diags
}

func flattenAssociatedTemplatesAndRules(ctx context.Context, apiObjects []awstypes.AssociatedTemplateOrRule) (fwtypes.ListNestedObjectValueOf[policyAssociationModel], diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	if len(apiObjects) == 0 {
		return fwtypes.NewListNestedObjectValueOfNull[policyAssociationModel](ctx), diags
	}

	tfObjects := make([]*policyAssociationModel, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfObject := &policyAssociationModel{
			RuleARN:     fwtypes.ARNNull(),
			TemplateARN: fwtypes.ARNNull(),
		}

		switch v := apiObject.(type) {
		case *awstypes.AssociatedTemplateOrRuleMemberRuleArn:
			tfObject.RuleARN = fwtypes.ARNValue(v.Value)
		case *awstypes.AssociatedTemplateOrRuleMemberTemplateArn:
			tfObject.TemplateARN = fwtypes.ARNValue(v.Value)
		case *awstypes.UnknownUnionMember:
			fwflex.HandleFlattenUnknownUnionMember(ctx, v.Tag, &diags)
			return fwtypes.NewListNestedObjectValueOfNull[policyAssociationModel](ctx), diags
		default:
			diags.AddError(
				"Unsupported Associated Template Or Rule",
				fmt.Sprintf("flattenAssociatedTemplatesAndRules: %T", apiObject),
			)
			return fwtypes.NewListNestedObjectValueOfNull[policyAssociationModel](ctx), diags
		}

		tfObjects = append(tfObjects, tfObject)
	}

	tfList, d := fwtypes.NewListNestedObjectValueOfSlice(ctx, tfObjects, nil)
	smerr.AddEnrich(ctx, &diags, d)

	return tfList, diags
}

func flattenPolicy(ctx context.Context, output *networksecuritymanager.GetPolicyOutput, data *policyResourceModel) diag.Diagnostics { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	// The API returns an empty description when none was set.
	description := data.Description

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, output, data, fwflex.WithFieldNamePrefix("Policy")))
	if diags.HasError() {
		return diags
	}

	// A draft is returned with the DRAFT-qualified ARN.
	data.ARN = types.StringValue(ruleBaseARN(aws.ToString(output.PolicyArn)))
	data.HasPublishedVersion = types.BoolValue(aws.ToBool(output.HasPublishedVersion))
	data.IsPublished = types.BoolValue(output.Status != awstypes.EntityStatusDraft)
	associations, d := flattenAssociatedTemplatesAndRules(ctx, output.AssociatedTemplateAndRuleList)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return diags
	}
	data.AssociatedTemplateAndRule = associations
	if aws.ToString(output.PolicyDescription) == "" && (description.IsNull() || description.IsUnknown()) {
		data.Description = types.StringNull()
	}

	return diags
}

type policyResourceModel struct {
	framework.WithRegionModel
	ARN                       types.String                                              `tfsdk:"arn"`
	AssociatedTemplateAndRule fwtypes.ListNestedObjectValueOf[policyAssociationModel]   `tfsdk:"associated_template_and_rule" autoflex:"-"`
	Description               types.String                                              `tfsdk:"description"`
	FirewallType              fwtypes.StringEnum[awstypes.PolicyFirewallType]           `tfsdk:"firewall_type"`
	HasPublishedVersion       types.Bool                                                `tfsdk:"has_published_version"`
	IsPublished               types.Bool                                                `tfsdk:"is_published"`
	Name                      types.String                                              `tfsdk:"name"`
	PolicyConfiguration       fwtypes.ListNestedObjectValueOf[policyConfigurationModel] `tfsdk:"policy_configuration"`
	PolicyID                  types.String                                              `tfsdk:"policy_id"`
	Priority                  types.Int32                                               `tfsdk:"priority"`
	Status                    fwtypes.StringEnum[awstypes.EntityStatus]                 `tfsdk:"status"`
	Tags                      tftags.Map                                                `tfsdk:"tags"`
	TagsAll                   tftags.Map                                                `tfsdk:"tags_all"`
	UpdatedAt                 timetypes.RFC3339                                         `tfsdk:"updated_at"`
	Version                   types.String                                              `tfsdk:"version"`
}

// policyAssociationModel models the TemplateOrRuleReference union: exactly one
// of the two ARNs is set.
type policyAssociationModel struct {
	RuleARN     fwtypes.ARN `tfsdk:"rule_arn"`
	TemplateARN fwtypes.ARN `tfsdk:"template_arn"`
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
