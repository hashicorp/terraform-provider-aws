// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"
	"strings"

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
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
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheckPolicy")
func newPolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &policyResource{}, nil
}

// policyDraftQualifier is the ARN qualifier of a policy's pending draft. A policy created
// unpublished, or a draft saved on top of a published policy, can only be read,
// updated, published or deleted through the DRAFT-qualified ARN.
const policyDraftQualifier = ":DRAFT"

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
			"is_published": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z][0-9A-Za-z _.:/=+\-@]*$`), ""),
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
			names.AttrConfiguration: schema.ListNestedBlock{
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
						"rule_identifier": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							Validators: []validator.String{
								stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("template_identifier")),
							},
						},
						"template_identifier": schema.StringAttribute{
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
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreatePolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, (*networksecuritymanager.GetPolicyOutput)(out), &plan))
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

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state), smerr.ID, arn)
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
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, append(diff.IgnoredFieldNamesOpts(), fwflex.WithFieldNamePrefix("Policy"))...))
		if resp.Diagnostics.HasError() {
			return
		}

		// An omitted description leaves the current one in place, so clear it explicitly.
		if plan.Description.IsNull() && !state.Description.IsNull() {
			input.PolicyDescription = aws.String("")
		}

		input.ClientToken = aws.String(create.UniqueId(ctx))
		// IsPublished is required on every update, not only when it changes.
		input.IsPublished = fwflex.BoolFromFramework(ctx, plan.IsPublished)
		input.PolicyIdentifier = aws.String(policyIdentifier(arn, current.Status))
		input.UpdateToken = current.UpdateToken

		out, err := conn.UpdatePolicy(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, (*networksecuritymanager.GetPolicyOutput)(out), &plan))
		if resp.Diagnostics.HasError() {
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

	// A published policy with a pending draft cannot be deleted until the draft is,
	// and a policy that was never published exists only as a draft.
	for _, identifier := range []string{arn + policyDraftQualifier, arn} {
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

// findPolicyByARN returns the version of the policy that an update acts on:
// the pending draft when there is one, otherwise the published policy.
func findPolicyByARN(ctx context.Context, conn *networksecuritymanager.Client, arn string) (*networksecuritymanager.GetPolicyOutput, error) {
	out, err := findPolicyByIdentifier(ctx, conn, arn+policyDraftQualifier)
	if retry.NotFound(err) {
		return findPolicyByIdentifier(ctx, conn, arn)
	}
	if err != nil {
		return nil, err
	}

	return out, nil
}

func findPolicyByIdentifier(ctx context.Context, conn *networksecuritymanager.Client, identifier string) (*networksecuritymanager.GetPolicyOutput, error) {
	input := networksecuritymanager.GetPolicyInput{
		PolicyIdentifier: aws.String(identifier),
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

// policyBaseARN strips the DRAFT qualifier, so that the policy is always identified
// by the ARN it keeps across publication.
func policyBaseARN(arn string) string {
	return strings.TrimSuffix(arn, policyDraftQualifier)
}

func policyIdentifier(arn string, status awstypes.EntityStatus) string {
	if status == awstypes.EntityStatusDraft {
		return arn + policyDraftQualifier
	}

	return arn
}

func (r *policyResource) flatten(ctx context.Context, out *networksecuritymanager.GetPolicyOutput, data *policyResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	description := data.Description
	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, out, data, fwflex.WithFieldNamePrefix("Policy")))
	if diags.HasError() {
		return diags
	}
	data.Description = normalizeDescription(description, data.Description)
	data.ARN = fwflex.StringValueToFramework(ctx, policyBaseARN(aws.ToString(out.PolicyArn)))
	data.IsPublished = types.BoolValue(out.Status != awstypes.EntityStatusDraft)

	return diags
}

type policyResourceModel struct {
	framework.WithRegionModel
	ARN                           types.String                                              `tfsdk:"arn"`
	AssociatedTemplateAndRuleList fwtypes.ListNestedObjectValueOf[templateOrRuleModel]      `tfsdk:"template_or_rule"`
	Configuration                 fwtypes.ListNestedObjectValueOf[policyConfigurationModel] `tfsdk:"configuration"`
	Description                   types.String                                              `tfsdk:"description"`
	FirewallType                  fwtypes.StringEnum[awstypes.PolicyFirewallType]           `tfsdk:"firewall_type"`
	IsPublished                   types.Bool                                                `tfsdk:"is_published"`
	Name                          types.String                                              `tfsdk:"name"`
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
	RuleIdentifier     fwtypes.ARN `tfsdk:"rule_identifier"`
	TemplateIdentifier fwtypes.ARN `tfsdk:"template_identifier"`
}

var (
	_ fwflex.Expander  = templateOrRuleModel{}
	_ fwflex.Flattener = &templateOrRuleModel{}
)

func (m templateOrRuleModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.RuleIdentifier.IsNull():
		return &awstypes.TemplateOrRuleReferenceMemberRuleIdentifier{
			Value: m.RuleIdentifier.ValueString(),
		}, diags
	case !m.TemplateIdentifier.IsNull():
		return &awstypes.TemplateOrRuleReferenceMemberTemplateIdentifier{
			Value: m.TemplateIdentifier.ValueString(),
		}, diags
	}

	return nil, diags
}

func (m *templateOrRuleModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.AssociatedTemplateOrRuleMemberRuleArn:
		m.RuleIdentifier = fwtypes.ARNValue(t.Value)
	case awstypes.AssociatedTemplateOrRuleMemberTemplateArn:
		m.TemplateIdentifier = fwtypes.ARNValue(t.Value)
	}

	return diags
}
