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
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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

// @FrameworkResource("aws_networksecuritymanager_deployment", name="Deployment")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(preCheck="testAccPreCheck")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/networksecuritymanager;networksecuritymanager;networksecuritymanager.GetDeploymentOutput")
// @Testing(identityRegionOverrideTest=false)
func newDeploymentResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &deploymentResource{}

	return r, nil
}

const (
	// A deployment applies at most this many policies, at most one per
	// firewall type.
	deploymentMaxPolicies = 2
)

type deploymentResource struct {
	framework.ResourceWithModel[deploymentResourceModel]
	framework.WithImportByIdentity
}

func (r *deploymentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"coverage":    framework.ResourceComputedListOfObjectsAttribute[deploymentCoverageEntryModel](ctx),
			"deployment_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(256),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9 _.:/=+\-@]*$`), "must contain only alphanumeric characters, spaces, and _.:/=+-@"),
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
			"policy_arns": schema.SetAttribute{
				CustomType:  fwtypes.SetOfARNType,
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, deploymentMaxPolicies),
					setvalidator.ValueStringsAre(fwvalidators.ARN()),
				},
			},
			"scope_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
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
			"warnings": framework.ResourceComputedListOfObjectsAttribute[deploymentWarningEntryModel](ctx),
		},
		Blocks: map[string]schema.Block{
			"deployment_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[deploymentConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"enable_cross_account_visibility": schema.BoolAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

func (r *deploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var plan deploymentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	var input networksecuritymanager.CreateDeploymentInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Deployment")))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	policies, d := expandPolicyReferences(ctx, plan.PolicyARNs)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}
	input.AssociatedPolicyList = policies
	input.AssociatedScopeList = expandScopeReferences(plan.ScopeARN)
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateDeployment(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenDeployment(ctx, (*networksecuritymanager.GetDeploymentOutput)(output), &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *deploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var state deploymentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := ruleBaseARN(state.ARN.ValueString())
	output, err := findDeploymentByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenDeployment(ctx, output, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *deploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var plan, state deploymentResourceModel
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
	current, err := findDeploymentByARN(ctx, conn, arn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	if !plan.DeploymentConfiguration.Equal(state.DeploymentConfiguration) ||
		!plan.Description.Equal(state.Description) ||
		!plan.IsPublished.Equal(state.IsPublished) ||
		!plan.PolicyARNs.Equal(state.PolicyARNs) ||
		!plan.ScopeARN.Equal(state.ScopeARN) {
		var input networksecuritymanager.UpdateDeploymentInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Deployment")))
		if resp.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		// A pending draft can only be updated (or published) through its
		// DRAFT-qualified ARN; the base ARN is rejected while a draft exists.
		input.DeploymentIdentifier = aws.String(ruleUpdateIdentifier(arn, current.Status))
		input.UpdateToken = current.UpdateToken
		policies, d := expandPolicyReferences(ctx, plan.PolicyARNs)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}
		input.AssociatedPolicyList = policies
		input.AssociatedScopeList = expandScopeReferences(plan.ScopeARN)
		if plan.Description.IsNull() {
			// An omitted description leaves the current one in place; an
			// empty string clears it.
			input.DeploymentDescription = aws.String("")
		}

		output, err := conn.UpdateDeployment(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		current = (*networksecuritymanager.GetDeploymentOutput)(output)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenDeployment(ctx, current, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *deploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var state deploymentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := ruleBaseARN(state.ARN.ValueString())

	// A published deployment with a pending draft cannot be deleted until the
	// draft is; a deployment that was never published exists only as a draft,
	// and deleting its base ARN succeeds without removing anything.
	for _, identifier := range []string{arn + ruleDraftQualifier, arn} {
		input := networksecuritymanager.DeleteDeploymentInput{
			DeploymentIdentifier: aws.String(identifier),
		}

		_, err := conn.DeleteDeployment(ctx, &input)
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			continue
		}
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}
	}
}

// findDeploymentByARN returns the version of the deployment that an update
// would act on: the pending draft when there is one, otherwise the published
// deployment.
func findDeploymentByARN(ctx context.Context, conn *networksecuritymanager.Client, arn string) (*networksecuritymanager.GetDeploymentOutput, error) {
	output, err := findDeploymentByIdentifier(ctx, conn, arn+ruleDraftQualifier)
	if retry.NotFound(err) {
		return findDeploymentByIdentifier(ctx, conn, arn)
	}
	if err != nil {
		return nil, err
	}

	return output, nil
}

func findDeploymentByIdentifier(ctx context.Context, conn *networksecuritymanager.Client, identifier string) (*networksecuritymanager.GetDeploymentOutput, error) {
	input := networksecuritymanager.GetDeploymentInput{
		DeploymentIdentifier: aws.String(identifier),
	}

	output, err := conn.GetDeployment(ctx, &input)
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

func expandPolicyReferences(ctx context.Context, tfSet fwtypes.SetOfARN) ([]awstypes.PolicyReference, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	var diags diag.Diagnostics

	if tfSet.IsNull() || tfSet.IsUnknown() {
		return nil, diags
	}

	var arns []string
	smerr.AddEnrich(ctx, &diags, tfSet.ElementsAs(ctx, &arns, false))
	if diags.HasError() {
		return nil, diags
	}

	apiObjects := make([]awstypes.PolicyReference, 0, len(arns))
	for _, arn := range arns {
		apiObjects = append(apiObjects, awstypes.PolicyReference{
			PolicyIdentifier: aws.String(arn),
		})
	}

	return apiObjects, diags
}

// expandScopeReferences wraps the single scope of a deployment in the list the
// API takes: a deployment has exactly one scope.
func expandScopeReferences(scopeARN fwtypes.ARN) []awstypes.ScopeReference { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	if scopeARN.IsNull() || scopeARN.IsUnknown() {
		return nil
	}

	return []awstypes.ScopeReference{{
		ScopeIdentifier: scopeARN.ValueStringPointer(),
	}}
}

func flattenDeployment(ctx context.Context, output *networksecuritymanager.GetDeploymentOutput, data *deploymentResourceModel) diag.Diagnostics { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	// The API returns an empty description when none was set.
	description := data.Description

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, output, data, fwflex.WithFieldNamePrefix("Deployment")))
	if diags.HasError() {
		return diags
	}

	// A draft is returned with the DRAFT-qualified ARN.
	data.ARN = types.StringValue(ruleBaseARN(aws.ToString(output.DeploymentArn)))
	data.HasPublishedVersion = types.BoolValue(aws.ToBool(output.HasPublishedVersion))
	data.IsPublished = types.BoolValue(output.Status != awstypes.EntityStatusDraft)
	if aws.ToString(output.DeploymentDescription) == "" && (description.IsNull() || description.IsUnknown()) {
		data.Description = types.StringNull()
	}

	policyARNs := make([]attr.Value, 0, len(output.AssociatedPolicyList))
	for _, v := range output.AssociatedPolicyList {
		policyARNs = append(policyARNs, fwtypes.ARNValue(aws.ToString(v.PolicyArn)))
	}
	data.PolicyARNs = fwtypes.NewSetValueOfMust[fwtypes.ARN](ctx, policyARNs)

	data.ScopeARN = fwtypes.ARNNull()
	if len(output.AssociatedScopeList) > 0 {
		data.ScopeARN = fwtypes.ARNValue(aws.ToString(output.AssociatedScopeList[0].ScopeArn))
	}

	return diags
}

type deploymentResourceModel struct {
	framework.WithRegionModel
	ARN                     types.String                                                  `tfsdk:"arn"`
	Coverage                fwtypes.ListNestedObjectValueOf[deploymentCoverageEntryModel] `tfsdk:"coverage"`
	DeploymentConfiguration fwtypes.ListNestedObjectValueOf[deploymentConfigurationModel] `tfsdk:"deployment_configuration"`
	DeploymentID            types.String                                                  `tfsdk:"deployment_id"`
	Description             types.String                                                  `tfsdk:"description"`
	HasPublishedVersion     types.Bool                                                    `tfsdk:"has_published_version"`
	IsPublished             types.Bool                                                    `tfsdk:"is_published"`
	Name                    types.String                                                  `tfsdk:"name"`
	PolicyARNs              fwtypes.SetOfARN                                              `tfsdk:"policy_arns" autoflex:"-"`
	ScopeARN                fwtypes.ARN                                                   `tfsdk:"scope_arn" autoflex:"-"`
	Status                  fwtypes.StringEnum[awstypes.EntityStatus]                     `tfsdk:"status"`
	Tags                    tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                    `tfsdk:"tags_all"`
	UpdatedAt               timetypes.RFC3339                                             `tfsdk:"updated_at"`
	Version                 types.String                                                  `tfsdk:"version"`
	Warnings                fwtypes.ListNestedObjectValueOf[deploymentWarningEntryModel]  `tfsdk:"warnings"`
}

type deploymentConfigurationModel struct {
	EnableCrossAccountVisibility types.Bool `tfsdk:"enable_cross_account_visibility"`
}

type deploymentCoverageEntryModel struct {
	FirewallType         fwtypes.StringEnum[awstypes.PolicyFirewallType]      `tfsdk:"firewall_type"`
	InScopeResourceTypes fwtypes.ListOfStringEnum[awstypes.ScopeResourceType] `tfsdk:"in_scope_resource_types"`
	PolicyARNs           fwtypes.ListOfString                                 `tfsdk:"policy_arns"`
}

type deploymentWarningEntryModel struct {
	Code      types.String `tfsdk:"code"`
	Message   types.String `tfsdk:"message"`
	PolicyARN types.String `tfsdk:"policy_arn"`
}
