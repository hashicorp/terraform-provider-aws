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
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networksecuritymanager_template", name="Template")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(preCheck="testAccPreCheck")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/networksecuritymanager;networksecuritymanager;networksecuritymanager.GetTemplateOutput")
// @Testing(identityRegionOverrideTest=false)
func newTemplateResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &templateResource{}

	return r, nil
}

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
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9 _.:/=+\-@]*$`), "must contain only alphanumeric characters, spaces, and _.:/=+-@"),
				},
			},
			"firewall_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TemplateFirewallType](),
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
			// The list is ordered: an INSPECTION rule's position among the
			// template's rules is its priority in the web ACL.
			"rule_arns": schema.ListAttribute{
				CustomType:  fwtypes.ListOfARNType,
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 50),
					// The API returns an internal error for a duplicate rule.
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(fwvalidators.ARN()),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EntityStatus](),
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"template_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
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

func (r *templateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var plan templateResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	var input networksecuritymanager.CreateTemplateInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Template")))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.AssociatedRuleList = expandRuleReferences(ctx, plan.RuleARNs)
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateTemplate(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenTemplate(ctx, (*networksecuritymanager.GetTemplateOutput)(output), &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *templateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var state templateResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := ruleBaseARN(state.ARN.ValueString())
	output, err := findTemplateByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenTemplate(ctx, output, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *templateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var plan, state templateResourceModel
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
	current, err := findTemplateByARN(ctx, conn, arn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	if !plan.RuleARNs.Equal(state.RuleARNs) || !plan.Description.Equal(state.Description) || !plan.IsPublished.Equal(state.IsPublished) {
		var input networksecuritymanager.UpdateTemplateInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Template")))
		if resp.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		// A pending draft can only be updated (or published) through its
		// DRAFT-qualified ARN; the base ARN is rejected while a draft exists.
		input.TemplateIdentifier = aws.String(ruleUpdateIdentifier(arn, current.Status))
		input.UpdateToken = current.UpdateToken
		input.AssociatedRuleList = expandRuleReferences(ctx, plan.RuleARNs)
		if plan.Description.IsNull() {
			// An omitted description leaves the current one in place; an
			// empty string clears it.
			input.TemplateDescription = aws.String("")
		}

		output, err := conn.UpdateTemplate(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		current = (*networksecuritymanager.GetTemplateOutput)(output)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenTemplate(ctx, current, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *templateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var state templateResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := ruleBaseARN(state.ARN.ValueString())

	// A published template with a pending draft cannot be deleted until the
	// draft is; a template that was never published exists only as a draft,
	// and deleting its base ARN succeeds without removing anything.
	for _, identifier := range []string{arn + ruleDraftQualifier, arn} {
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

// findTemplateByARN returns the version of the template that an update would
// act on: the pending draft when there is one, otherwise the published template.
func findTemplateByARN(ctx context.Context, conn *networksecuritymanager.Client, arn string) (*networksecuritymanager.GetTemplateOutput, error) {
	output, err := findTemplateByIdentifier(ctx, conn, arn+ruleDraftQualifier)
	if retry.NotFound(err) {
		return findTemplateByIdentifier(ctx, conn, arn)
	}
	if err != nil {
		return nil, err
	}

	return output, nil
}

func findTemplateByIdentifier(ctx context.Context, conn *networksecuritymanager.Client, identifier string) (*networksecuritymanager.GetTemplateOutput, error) {
	input := networksecuritymanager.GetTemplateInput{
		TemplateIdentifier: aws.String(identifier),
	}

	output, err := conn.GetTemplate(ctx, &input)
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

func expandRuleReferences(ctx context.Context, tfList fwtypes.ListOfARN) []awstypes.RuleReference { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	arns := fwflex.ExpandFrameworkStringValueList(ctx, tfList)
	apiObjects := make([]awstypes.RuleReference, 0, len(arns))

	for _, arn := range arns {
		apiObjects = append(apiObjects, awstypes.RuleReference{
			RuleIdentifier: aws.String(arn),
		})
	}

	return apiObjects
}

func flattenAssociatedRules(ctx context.Context, apiObjects []awstypes.AssociatedRule) fwtypes.ListOfARN { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	elems := make([]attr.Value, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		elems = append(elems, fwtypes.ARNValue(aws.ToString(apiObject.RuleArn)))
	}

	return fwtypes.NewListValueOfMust[fwtypes.ARN](ctx, elems)
}

func flattenTemplate(ctx context.Context, output *networksecuritymanager.GetTemplateOutput, data *templateResourceModel) diag.Diagnostics { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	// The API returns an empty description when none was set.
	description := data.Description

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, output, data, fwflex.WithFieldNamePrefix("Template")))
	if diags.HasError() {
		return diags
	}

	// A draft is returned with the DRAFT-qualified ARN.
	data.ARN = types.StringValue(ruleBaseARN(aws.ToString(output.TemplateArn)))
	data.HasPublishedVersion = types.BoolValue(aws.ToBool(output.HasPublishedVersion))
	data.IsPublished = types.BoolValue(output.Status != awstypes.EntityStatusDraft)
	data.RuleARNs = flattenAssociatedRules(ctx, output.AssociatedRuleList)
	if aws.ToString(output.TemplateDescription) == "" && (description.IsNull() || description.IsUnknown()) {
		data.Description = types.StringNull()
	}

	return diags
}

type templateResourceModel struct {
	framework.WithRegionModel
	ARN                 types.String                                      `tfsdk:"arn"`
	Description         types.String                                      `tfsdk:"description"`
	FirewallType        fwtypes.StringEnum[awstypes.TemplateFirewallType] `tfsdk:"firewall_type"`
	HasPublishedVersion types.Bool                                        `tfsdk:"has_published_version"`
	IsPublished         types.Bool                                        `tfsdk:"is_published"`
	Name                types.String                                      `tfsdk:"name"`
	RuleARNs            fwtypes.ListOfARN                                 `tfsdk:"rule_arns" autoflex:"-"`
	Status              fwtypes.StringEnum[awstypes.EntityStatus]         `tfsdk:"status"`
	Tags                tftags.Map                                        `tfsdk:"tags"`
	TagsAll             tftags.Map                                        `tfsdk:"tags_all"`
	TemplateID          types.String                                      `tfsdk:"template_id"`
	UpdatedAt           timetypes.RFC3339                                 `tfsdk:"updated_at"`
	Version             types.String                                      `tfsdk:"version"`
}
