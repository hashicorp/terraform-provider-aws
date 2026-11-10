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
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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

// @FrameworkResource("aws_networksecuritymanager_deployment", name="Deployment")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheckDeployment")
func newDeploymentResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &deploymentResource{}, nil
}

// deploymentDraftQualifier is the ARN qualifier of a deployment's pending draft. A deployment created
// unpublished, or a draft saved on top of a published deployment, can only be read,
// updated, published or deleted through the DRAFT-qualified ARN.
const deploymentDraftQualifier = ":DRAFT"

// deploymentMaxPolicies is the maximum number of policies that can be associated with a deployment.
const deploymentMaxPolicies = 2

type deploymentResource struct {
	framework.ResourceWithModel[deploymentResourceModel]
	framework.WithImportByIdentity
}

func (r *deploymentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"policy_arns": schema.SetAttribute{
				CustomType: fwtypes.SetOfARNType,
				Required:   true,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, deploymentMaxPolicies),
				},
			},
			"scope_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrConfiguration: schema.ListNestedBlock{
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
	var plan deploymentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input networksecuritymanager.CreateDeploymentInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Deployment")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.AssociatedPolicyList = plan.policyReferences(ctx)
	input.AssociatedScopeList = plan.scopeReferences()
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateDeployment(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, (*networksecuritymanager.GetDeploymentOutput)(out), &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *deploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deploymentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	out, err := findDeploymentByARN(ctx, conn, arn)
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

func (r *deploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state deploymentResourceModel
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
		current, err := findDeploymentByARN(ctx, conn, arn)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		var input networksecuritymanager.UpdateDeploymentInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, append(diff.IgnoredFieldNamesOpts(), fwflex.WithFieldNamePrefix("Deployment"))...))
		if resp.Diagnostics.HasError() {
			return
		}

		if !plan.PolicyARNs.Equal(state.PolicyARNs) {
			input.AssociatedPolicyList = plan.policyReferences(ctx)
		}
		if !plan.ScopeARN.Equal(state.ScopeARN) {
			input.AssociatedScopeList = plan.scopeReferences()
		}

		// An omitted description leaves the current one in place, so clear it explicitly.
		if plan.Description.IsNull() && !state.Description.IsNull() {
			input.DeploymentDescription = aws.String("")
		}

		input.ClientToken = aws.String(create.UniqueId(ctx))
		input.DeploymentIdentifier = aws.String(deploymentIdentifier(arn, current.Status))
		// IsPublished is required on every update, not only when it changes.
		input.IsPublished = fwflex.BoolFromFramework(ctx, plan.IsPublished)
		input.UpdateToken = current.UpdateToken

		out, err := conn.UpdateDeployment(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, (*networksecuritymanager.GetDeploymentOutput)(out), &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *deploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deploymentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)

	// A published deployment with a pending draft cannot be deleted until the draft is,
	// and a deployment that was never published exists only as a draft.
	for _, identifier := range []string{arn + deploymentDraftQualifier, arn} {
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

// findDeploymentByARN returns the version of the deployment that an update acts on:
// the pending draft when there is one, otherwise the published deployment.
func findDeploymentByARN(ctx context.Context, conn *networksecuritymanager.Client, arn string) (*networksecuritymanager.GetDeploymentOutput, error) {
	out, err := findDeploymentByIdentifier(ctx, conn, arn+deploymentDraftQualifier)
	if retry.NotFound(err) {
		return findDeploymentByIdentifier(ctx, conn, arn)
	}
	if err != nil {
		return nil, err
	}

	return out, nil
}

func findDeploymentByIdentifier(ctx context.Context, conn *networksecuritymanager.Client, identifier string) (*networksecuritymanager.GetDeploymentOutput, error) {
	input := networksecuritymanager.GetDeploymentInput{
		DeploymentIdentifier: aws.String(identifier),
	}

	out, err := conn.GetDeployment(ctx, &input)
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

// deploymentBaseARN strips the DRAFT qualifier, so that the deployment is always identified
// by the ARN it keeps across publication.
func deploymentBaseARN(arn string) string {
	return strings.TrimSuffix(arn, deploymentDraftQualifier)
}

func deploymentIdentifier(arn string, status awstypes.EntityStatus) string {
	if status == awstypes.EntityStatusDraft {
		return arn + deploymentDraftQualifier
	}

	return arn
}

func (r *deploymentResource) flatten(ctx context.Context, out *networksecuritymanager.GetDeploymentOutput, data *deploymentResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	description := data.Description
	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, out, data, fwflex.WithFieldNamePrefix("Deployment")))
	if diags.HasError() {
		return diags
	}
	data.Description = normalizeDescription(description, data.Description)
	data.ARN = fwflex.StringValueToFramework(ctx, deploymentBaseARN(aws.ToString(out.DeploymentArn)))
	data.IsPublished = types.BoolValue(out.Status != awstypes.EntityStatusDraft)

	policyARNs := tfslices.ApplyToAll(out.AssociatedPolicyList, func(v awstypes.AssociatedPolicy) string {
		return aws.ToString(v.PolicyArn)
	})
	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, policyARNs, &data.PolicyARNs))

	// A deployment has exactly one scope.
	if len(out.AssociatedScopeList) > 0 {
		data.ScopeARN = fwtypes.ARNValue(aws.ToString(out.AssociatedScopeList[0].ScopeArn))
	}

	return diags
}

type deploymentResourceModel struct {
	framework.WithRegionModel
	ARN           types.String                                                  `tfsdk:"arn"`
	Configuration fwtypes.ListNestedObjectValueOf[deploymentConfigurationModel] `tfsdk:"configuration"`
	Description   types.String                                                  `tfsdk:"description"`
	IsPublished   types.Bool                                                    `tfsdk:"is_published"`
	Name          types.String                                                  `tfsdk:"name"`
	PolicyARNs    fwtypes.SetOfARN                                              `tfsdk:"policy_arns" autoflex:"-"`
	ScopeARN      fwtypes.ARN                                                   `tfsdk:"scope_arn" autoflex:"-"`
	Tags          tftags.Map                                                    `tfsdk:"tags"`
	TagsAll       tftags.Map                                                    `tfsdk:"tags_all"`
}

type deploymentConfigurationModel struct {
	EnableCrossAccountVisibility types.Bool `tfsdk:"enable_cross_account_visibility"`
}

// policyReferences returns the policies associated with the deployment. The service
// takes policy identifiers and returns policy ARNs, so the set is not mapped by AutoFlex.
func (m deploymentResourceModel) policyReferences(ctx context.Context) []awstypes.PolicyReference {
	return tfslices.ApplyToAll(fwflex.ExpandFrameworkStringValueSet(ctx, m.PolicyARNs), func(arn string) awstypes.PolicyReference {
		return awstypes.PolicyReference{
			PolicyIdentifier: aws.String(arn),
		}
	})
}

// scopeReferences returns the deployment's scope. The service takes a list holding exactly one scope.
func (m deploymentResourceModel) scopeReferences() []awstypes.ScopeReference {
	return []awstypes.ScopeReference{
		{
			ScopeIdentifier: m.ScopeARN.ValueStringPointer(),
		},
	}
}
