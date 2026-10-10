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
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

// @FrameworkResource("aws_networksecuritymanager_rule", name="Rule")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(preCheck="testAccPreCheck")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/networksecuritymanager;networksecuritymanager;networksecuritymanager.GetRuleOutput")
// @Testing(identityRegionOverrideTest=false)
func newRuleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &ruleResource{}

	return r, nil
}

// ruleDraftQualifier is the ARN qualifier of a rule's pending draft. A draft
// saved on top of a published rule, or a rule created unpublished, can only be
// read, updated, published or deleted through the DRAFT-qualified ARN.
const ruleDraftQualifier = ":DRAFT"

type ruleResource struct {
	framework.ResourceWithModel[ruleResourceModel]
	framework.WithImportByIdentity
}

func (r *ruleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrConfiguration: schema.StringAttribute{
				CustomType: ruleConfigurationType{},
				Required:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(256),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9 _.:/=+\-@]*$`), "must contain only alphanumeric characters, spaces, and _.:/=+-@"),
				},
			},
			"firewall_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RuleFirewallType](),
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
			"rule_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"rule_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RuleType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
	}
}

func (r *ruleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var plan ruleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	var input networksecuritymanager.CreateRuleInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Rule")))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateRule(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenRule(ctx, (*networksecuritymanager.GetRuleOutput)(output), &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *ruleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var state ruleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := ruleBaseARN(state.ARN.ValueString())
	output, err := findRuleByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenRule(ctx, output, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *ruleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var plan, state ruleResourceModel
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
	current, err := findRuleByARN(ctx, conn, arn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	// A configuration that differs only in formatting is planned as a change
	// but is the same document; it is not written back (every write creates
	// a new version).
	sameConfiguration, d := plan.Configuration.StringSemanticEquals(ctx, state.Configuration)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if !sameConfiguration || !plan.Description.Equal(state.Description) || !plan.IsPublished.Equal(state.IsPublished) {
		var input networksecuritymanager.UpdateRuleInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Rule")))
		if resp.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		// A pending draft can only be updated (or published) through its
		// DRAFT-qualified ARN; the base ARN is rejected while a draft exists.
		input.RuleIdentifier = aws.String(ruleUpdateIdentifier(arn, current.Status))
		input.UpdateToken = current.UpdateToken
		if plan.Description.IsNull() {
			// An omitted description leaves the current one in place; an
			// empty string clears it.
			input.RuleDescription = aws.String("")
		}

		output, err := conn.UpdateRule(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		current = (*networksecuritymanager.GetRuleOutput)(output)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenRule(ctx, current, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *ruleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var state ruleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := ruleBaseARN(state.ARN.ValueString())

	// A published rule with a pending draft cannot be deleted until the draft
	// is; a rule that was never published exists only as a draft, and deleting
	// its base ARN succeeds without removing anything.
	for _, identifier := range []string{arn + ruleDraftQualifier, arn} {
		input := networksecuritymanager.DeleteRuleInput{
			RuleIdentifier: aws.String(identifier),
		}

		_, err := conn.DeleteRule(ctx, &input)
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			continue
		}
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}
	}
}

// findRuleByARN returns the version of the rule that an update would act on:
// the pending draft when there is one, otherwise the published rule.
func findRuleByARN(ctx context.Context, conn *networksecuritymanager.Client, arn string) (*networksecuritymanager.GetRuleOutput, error) {
	output, err := findRuleByIdentifier(ctx, conn, arn+ruleDraftQualifier)
	if retry.NotFound(err) {
		return findRuleByIdentifier(ctx, conn, arn)
	}
	if err != nil {
		return nil, err
	}

	return output, nil
}

func findRuleByIdentifier(ctx context.Context, conn *networksecuritymanager.Client, identifier string) (*networksecuritymanager.GetRuleOutput, error) {
	input := networksecuritymanager.GetRuleInput{
		RuleIdentifier: aws.String(identifier),
	}

	output, err := conn.GetRule(ctx, &input)
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

// ruleBaseARN strips the DRAFT qualifier, so that the rule is always
// identified by the ARN it keeps across publication.
func ruleBaseARN(arn string) string {
	return strings.TrimSuffix(arn, ruleDraftQualifier)
}

func ruleUpdateIdentifier(arn string, status awstypes.EntityStatus) string {
	if status == awstypes.EntityStatusDraft {
		return arn + ruleDraftQualifier
	}

	return arn
}

func flattenRule(ctx context.Context, output *networksecuritymanager.GetRuleOutput, data *ruleResourceModel) diag.Diagnostics { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	// The API returns an empty description when none was set.
	description := data.Description

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, output, data, fwflex.WithFieldNamePrefix("Rule")))
	if diags.HasError() {
		return diags
	}

	// A draft is returned with the DRAFT-qualified ARN.
	data.ARN = types.StringValue(ruleBaseARN(aws.ToString(output.RuleArn)))
	data.HasPublishedVersion = types.BoolValue(aws.ToBool(output.HasPublishedVersion))
	data.IsPublished = types.BoolValue(output.Status != awstypes.EntityStatusDraft)
	if aws.ToString(output.RuleDescription) == "" && (description.IsNull() || description.IsUnknown()) {
		data.Description = types.StringNull()
	}

	return diags
}

type ruleResourceModel struct {
	framework.WithRegionModel
	ARN                 types.String                                  `tfsdk:"arn"`
	Configuration       ruleConfiguration                             `tfsdk:"configuration"`
	Description         types.String                                  `tfsdk:"description"`
	FirewallType        fwtypes.StringEnum[awstypes.RuleFirewallType] `tfsdk:"firewall_type"`
	HasPublishedVersion types.Bool                                    `tfsdk:"has_published_version"`
	IsPublished         types.Bool                                    `tfsdk:"is_published"`
	Name                types.String                                  `tfsdk:"name"`
	RuleID              types.String                                  `tfsdk:"rule_id"`
	RuleType            fwtypes.StringEnum[awstypes.RuleType]         `tfsdk:"rule_type"`
	Status              fwtypes.StringEnum[awstypes.EntityStatus]     `tfsdk:"status"`
	Tags                tftags.Map                                    `tfsdk:"tags"`
	TagsAll             tftags.Map                                    `tfsdk:"tags_all"`
	UpdatedAt           timetypes.RFC3339                             `tfsdk:"updated_at"`
	Version             types.String                                  `tfsdk:"version"`
}
