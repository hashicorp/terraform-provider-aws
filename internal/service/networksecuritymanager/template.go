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
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networksecuritymanager_template", name="Template")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheckTemplate")
func newTemplateResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &templateResource{}, nil
}

// templateDraftQualifier is the ARN qualifier of a template's pending draft. A template created
// unpublished, or a draft saved on top of a published template, can only be read,
// updated, published or deleted through the DRAFT-qualified ARN.
const templateDraftQualifier = ":DRAFT"

// templateMaxRules is the maximum number of rules that can be associated with a template.
const templateMaxRules = 50

type templateResource struct {
	framework.ResourceWithModel[templateResourceModel]
	framework.WithImportByIdentity
}

func (r *templateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				CustomType: fwtypes.StringEnumType[awstypes.TemplateFirewallType](),
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
			"rule_arns": schema.ListAttribute{
				CustomType: fwtypes.ListOfARNType,
				Required:   true,
				Validators: []validator.List{
					listvalidator.SizeBetween(1, templateMaxRules),
					listvalidator.UniqueValues(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *templateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan templateResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input networksecuritymanager.CreateTemplateInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Template")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.AssociatedRuleList = plan.ruleReferences(ctx)
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateTemplate(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, (*networksecuritymanager.GetTemplateOutput)(out), &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *templateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state templateResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	out, err := findTemplateByARN(ctx, conn, arn)
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

func (r *templateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state templateResourceModel
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
		current, err := findTemplateByARN(ctx, conn, arn)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		var input networksecuritymanager.UpdateTemplateInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, append(diff.IgnoredFieldNamesOpts(), fwflex.WithFieldNamePrefix("Template"))...))
		if resp.Diagnostics.HasError() {
			return
		}

		if !plan.RuleARNs.Equal(state.RuleARNs) {
			input.AssociatedRuleList = plan.ruleReferences(ctx)
		}

		// An omitted description leaves the current one in place, so clear it explicitly.
		if plan.Description.IsNull() && !state.Description.IsNull() {
			input.TemplateDescription = aws.String("")
		}

		input.ClientToken = aws.String(create.UniqueId(ctx))
		// IsPublished is required on every update, not only when it changes.
		input.IsPublished = fwflex.BoolFromFramework(ctx, plan.IsPublished)
		input.TemplateIdentifier = aws.String(templateIdentifier(arn, current.Status))
		input.UpdateToken = current.UpdateToken

		out, err := conn.UpdateTemplate(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, (*networksecuritymanager.GetTemplateOutput)(out), &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *templateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state templateResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)

	// A published template with a pending draft cannot be deleted until the draft is,
	// and a template that was never published exists only as a draft.
	for _, identifier := range []string{arn + templateDraftQualifier, arn} {
		input := networksecuritymanager.DeleteTemplateInput{
			TemplateIdentifier: aws.String(identifier),
		}
		_, err := conn.DeleteTemplate(ctx, &input)
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			continue
		}
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}
	}
}

// findTemplateByARN returns the version of the template that an update acts on:
// the pending draft when there is one, otherwise the published template.
func findTemplateByARN(ctx context.Context, conn *networksecuritymanager.Client, arn string) (*networksecuritymanager.GetTemplateOutput, error) {
	out, err := findTemplateByIdentifier(ctx, conn, arn+templateDraftQualifier)
	if retry.NotFound(err) {
		return findTemplateByIdentifier(ctx, conn, arn)
	}
	if err != nil {
		return nil, err
	}

	return out, nil
}

func findTemplateByIdentifier(ctx context.Context, conn *networksecuritymanager.Client, identifier string) (*networksecuritymanager.GetTemplateOutput, error) {
	input := networksecuritymanager.GetTemplateInput{
		TemplateIdentifier: aws.String(identifier),
	}

	out, err := conn.GetTemplate(ctx, &input)
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

// templateBaseARN strips the DRAFT qualifier, so that the template is always identified
// by the ARN it keeps across publication.
func templateBaseARN(arn string) string {
	return strings.TrimSuffix(arn, templateDraftQualifier)
}

func templateIdentifier(arn string, status awstypes.EntityStatus) string {
	if status == awstypes.EntityStatusDraft {
		return arn + templateDraftQualifier
	}

	return arn
}

func (r *templateResource) flatten(ctx context.Context, out *networksecuritymanager.GetTemplateOutput, data *templateResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	description := data.Description
	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, out, data, fwflex.WithFieldNamePrefix("Template")))
	if diags.HasError() {
		return diags
	}
	data.Description = normalizeDescription(description, data.Description)
	data.ARN = fwflex.StringValueToFramework(ctx, templateBaseARN(aws.ToString(out.TemplateArn)))
	data.IsPublished = types.BoolValue(out.Status != awstypes.EntityStatusDraft)

	ruleARNs := tfslices.ApplyToAll(out.AssociatedRuleList, func(v awstypes.AssociatedRule) string {
		return aws.ToString(v.RuleArn)
	})
	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, ruleARNs, &data.RuleARNs))

	return diags
}

type templateResourceModel struct {
	framework.WithRegionModel
	ARN          types.String                                      `tfsdk:"arn"`
	Description  types.String                                      `tfsdk:"description"`
	FirewallType fwtypes.StringEnum[awstypes.TemplateFirewallType] `tfsdk:"firewall_type"`
	IsPublished  types.Bool                                        `tfsdk:"is_published"`
	Name         types.String                                      `tfsdk:"name"`
	RuleARNs     fwtypes.ListOfARN                                 `tfsdk:"rule_arns" autoflex:"-"`
	Tags         tftags.Map                                        `tfsdk:"tags"`
	TagsAll      tftags.Map                                        `tfsdk:"tags_all"`
}

// ruleReferences returns the rules associated with the template, in order. The service
// takes rule identifiers and returns rule ARNs, so the list is not mapped by AutoFlex.
func (m templateResourceModel) ruleReferences(ctx context.Context) []awstypes.RuleReference {
	return tfslices.ApplyToAll(fwflex.ExpandFrameworkStringValueList(ctx, m.RuleARNs), func(arn string) awstypes.RuleReference {
		return awstypes.RuleReference{
			RuleIdentifier: aws.String(arn),
		}
	})
}
