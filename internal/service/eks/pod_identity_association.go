// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Pod Identity Association")
// @Tags(identifierAttribute="association_arn")
func newPodIdentityAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &podIdentityAssociationResource{}

	return r, nil
}

type podIdentityAssociationResourceModel struct {
	AssociationARN types.String `tfsdk:"association_arn"`
	AssociationID  types.String `tfsdk:"association_id"`
	ClusterName    types.String `tfsdk:"cluster_name"`
	ID             types.String `tfsdk:"id"`
	Namespace      types.String `tfsdk:"namespace"`
	RoleARN        fwtypes.ARN  `tfsdk:"role_arn"`
	ServiceAccount types.String `tfsdk:"service_account"`
	Tags           types.Map    `tfsdk:"tags"`
	TagsAll        types.Map    `tfsdk:"tags_all"`
}

func (model *podIdentityAssociationResourceModel) setID() {
	model.ID = model.AssociationID
}

const (
	ResNamePodIdentityAssociation = "Pod Identity Association"
)

type podIdentityAssociationResource struct {
	framework.ResourceWithConfigure
}

func (r *podIdentityAssociationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_eks_pod_identity_association"
}

func (r *podIdentityAssociationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"association_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrAssociationID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrClusterName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrNamespace: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"service_account": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *podIdentityAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan podIdentityAssociationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EKSClient(ctx)

	input := &eks.CreatePodIdentityAssociationInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.ClientRequestToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidParameterException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreatePodIdentityAssociation(ctx, input)
	}, "Role provided in the request does not exist")

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EKS, create.ErrActionCreating, ResNamePodIdentityAssociation, plan.AssociationID.String(), err),
			err.Error(),
		)
		return
	}

	// Set values for unknowns.
	output := outputRaw.(*eks.CreatePodIdentityAssociationOutput)
	plan.AssociationARN = fwflex.StringToFramework(ctx, output.Association.AssociationArn)
	plan.AssociationID = fwflex.StringToFramework(ctx, output.Association.AssociationId)
	plan.setID()

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *podIdentityAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EKSClient(ctx)

	var data podIdentityAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pia, err := findPodIdentityAssociationByTwoPartKey(ctx, conn, data.AssociationID.ValueString(), data.ClusterName.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EKS, create.ErrActionSetting, ResNamePodIdentityAssociation, data.AssociationID.String(), err),
			err.Error(),
		)
		return
	}

	// Set attributes for import.
	resp.Diagnostics.Append(fwflex.Flatten(ctx, pia, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, pia.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *podIdentityAssociationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new podIdentityAssociationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EKSClient(ctx)

	if !new.RoleARN.Equal(old.RoleARN) {
		input := &eks.UpdatePodIdentityAssociationInput{}
		resp.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		input.ClientRequestToken = aws.String(sdkid.UniqueId())

		_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidParameterException](ctx, propagationTimeout, func() (interface{}, error) {
			return conn.UpdatePodIdentityAssociation(ctx, input)
		}, "Role provided in the request does not exist")

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.EKS, create.ErrActionUpdating, ResNamePodIdentityAssociation, new.AssociationID.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *podIdentityAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state podIdentityAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EKSClient(ctx)

	input := &eks.DeletePodIdentityAssociationInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, state, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeletePodIdentityAssociation(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EKS, create.ErrActionDeleting, ResNamePodIdentityAssociation, state.AssociationID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *podIdentityAssociationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	const (
		partCount = 2
	)
	parts, err := flex.ExpandResourceId(req.ID, partCount, false)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("importing Pod Identity Association (%s)", req.ID), err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrAssociationID), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrClusterName), parts[0])...)
}

func (r *podIdentityAssociationResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func findPodIdentityAssociationByTwoPartKey(ctx context.Context, conn *eks.Client, associationID, clusterName string) (*awstypes.PodIdentityAssociation, error) {
	input := &eks.DescribePodIdentityAssociationInput{
		AssociationId: aws.String(associationID),
		ClusterName:   aws.String(clusterName),
	}

	output, err := conn.DescribePodIdentityAssociation(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Association == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Association, nil
}
