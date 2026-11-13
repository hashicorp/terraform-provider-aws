// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networksecuritymanager_rule", name="Rule")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheckRule")
func newRuleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &ruleResource{}, nil
}

// ruleDraftQualifier is the ARN qualifier of a rule's pending draft. A rule created
// unpublished, or a draft saved on top of a published rule, can only be read,
// updated, published or deleted through the DRAFT-qualified ARN.
const ruleDraftQualifier = ":DRAFT"

// The first Network Security Manager resource created in an account onboards the
// account, and a create issued while that is still in progress fails with
// "ConflictException: Account or org is already onboarded". Terraform creates
// independent resources in parallel, so the first apply of a configuration with
// more than one rule hits it.
const (
	onboardingConflictMessage = "already onboarded"
	onboardingTimeout         = 5 * time.Minute
)

// retryWhenOnboarding retries f while the account is being onboarded. Only the
// onboarding conflict is retried; any other ConflictException is returned.
func retryWhenOnboarding[T any](ctx context.Context, f func(context.Context) (T, error)) (T, error) {
	return tfresource.RetryWhenIsAErrorMessageContains[T, *awstypes.ConflictException](ctx, onboardingTimeout, f, onboardingConflictMessage)
}

type ruleResource struct {
	framework.ResourceWithModel[ruleResourceModel]
	framework.WithImportByIdentity
}

func (r *ruleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrConfiguration: schema.StringAttribute{
				CustomType: fwtypes.NewSmithyJSONType(ctx, document.NewLazyDocument),
				Required:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(256),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z _.:/=+\-@]*$`), ""),
				},
			},
			"firewall_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RuleFirewallType](),
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
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RuleType](),
				Required:   true,
			},
		},
	}
}

func (r *ruleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ruleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input networksecuritymanager.CreateRuleInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Rule")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	out, err := retryWhenOnboarding(ctx, func(ctx context.Context) (*networksecuritymanager.CreateRuleOutput, error) {
		return conn.CreateRule(ctx, &input)
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, (*networksecuritymanager.GetRuleOutput)(out), &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *ruleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ruleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	out, err := findRuleByARN(ctx, conn, arn)
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

func (r *ruleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ruleResourceModel
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
		current, err := findRuleByARN(ctx, conn, arn)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		var input networksecuritymanager.UpdateRuleInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, append(diff.IgnoredFieldNamesOpts(), fwflex.WithFieldNamePrefix("Rule"))...))
		if resp.Diagnostics.HasError() {
			return
		}

		// An omitted description leaves the current one in place, so clear it explicitly.
		if plan.Description.IsNull() && !state.Description.IsNull() {
			input.RuleDescription = aws.String("")
		}

		input.ClientToken = aws.String(create.UniqueId(ctx))
		// IsPublished is required on every update, not only when it changes.
		input.IsPublished = fwflex.BoolFromFramework(ctx, plan.IsPublished)
		input.RuleIdentifier = aws.String(ruleIdentifier(arn, current.Status))
		input.UpdateToken = current.UpdateToken

		out, err := conn.UpdateRule(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, (*networksecuritymanager.GetRuleOutput)(out), &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *ruleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ruleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)

	// A published rule with a pending draft cannot be deleted until the draft is,
	// and a rule that was never published exists only as a draft.
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

// findRuleByARN returns the version of the rule that an update acts on:
// the pending draft when there is one, otherwise the published rule.
func findRuleByARN(ctx context.Context, conn *networksecuritymanager.Client, arn string) (*networksecuritymanager.GetRuleOutput, error) {
	out, err := findRuleByIdentifier(ctx, conn, arn+ruleDraftQualifier)
	if retry.NotFound(err) {
		return findRuleByIdentifier(ctx, conn, arn)
	}
	if err != nil {
		return nil, err
	}

	return out, nil
}

func findRuleByIdentifier(ctx context.Context, conn *networksecuritymanager.Client, identifier string) (*networksecuritymanager.GetRuleOutput, error) {
	input := networksecuritymanager.GetRuleInput{
		RuleIdentifier: aws.String(identifier),
	}

	out, err := conn.GetRule(ctx, &input)
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

// ruleBaseARN strips the DRAFT qualifier, so that the rule is always identified
// by the ARN it keeps across publication.
func ruleBaseARN(arn string) string {
	return strings.TrimSuffix(arn, ruleDraftQualifier)
}

func ruleIdentifier(arn string, status awstypes.EntityStatus) string {
	if status == awstypes.EntityStatusDraft {
		return arn + ruleDraftQualifier
	}

	return arn
}

func (r *ruleResource) flatten(ctx context.Context, out *networksecuritymanager.GetRuleOutput, data *ruleResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	description := data.Description
	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, out, data, fwflex.WithFieldNamePrefix("Rule")))
	if diags.HasError() {
		return diags
	}
	data.Description = normalizeDescription(description, data.Description)
	data.ARN = fwflex.StringValueToFramework(ctx, ruleBaseARN(aws.ToString(out.RuleArn)))
	data.IsPublished = types.BoolValue(out.Status != awstypes.EntityStatusDraft)

	// The configuration is not flattened because the service returns its numbers as strings.
	// Set it from the response when it is missing (for example, during import or listing).
	if data.Configuration.IsNull() && out.Configuration != nil {
		configuration, err := tfsmithy.DocumentToJSONString(out.Configuration)
		if err != nil {
			smerr.AddError(ctx, &diags, err)
			return diags
		}
		data.Configuration = fwtypes.NewSmithyJSONValue(configuration, document.NewLazyDocument)
	}

	return diags
}

// normalizeDescription keeps a null description null when the service returns
// an empty one, which it does after the description has been cleared.
func normalizeDescription(prior, current types.String) types.String {
	if prior.IsNull() && current.ValueString() == "" {
		return types.StringNull()
	}
	return current
}

type ruleResourceModel struct {
	framework.WithRegionModel
	ARN           types.String                                  `tfsdk:"arn"`
	Configuration fwtypes.SmithyJSON[document.Interface]        `tfsdk:"configuration" autoflex:",noflatten"`
	Description   types.String                                  `tfsdk:"description"`
	FirewallType  fwtypes.StringEnum[awstypes.RuleFirewallType] `tfsdk:"firewall_type"`
	IsPublished   types.Bool                                    `tfsdk:"is_published"`
	Name          types.String                                  `tfsdk:"name"`
	Tags          tftags.Map                                    `tfsdk:"tags"`
	TagsAll       tftags.Map                                    `tfsdk:"tags_all"`
	Type          fwtypes.StringEnum[awstypes.RuleType]         `tfsdk:"type"`
}
