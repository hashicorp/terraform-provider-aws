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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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

// @FrameworkResource("aws_eks_pod_identity_association", name="Pod Identity Association")
// @Tags(identifierAttribute="association_arn")
func newPodIdentityAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &podIdentityAssociationResource{}

	return r, nil
}

type podIdentityAssociationResourceModel struct {
	framework.WithRegionModel
	AssociationARN     types.String `tfsdk:"association_arn"`
	AssociationID      types.String `tfsdk:"association_id"`
	ClusterName        types.String `tfsdk:"cluster_name"`
	DisableSessionTags types.Bool   `tfsdk:"disable_session_tags"`
	ExternalID         types.String `tfsdk:"external_id"`
	ID                 types.String `tfsdk:"id"`
	Namespace          types.String `tfsdk:"namespace"`
	RoleARN            fwtypes.ARN  `tfsdk:"role_arn"`
	ServiceAccount     types.String `tfsdk:"service_account"`
	Tags               tftags.Map   `tfsdk:"tags"`
	TagsAll            tftags.Map   `tfsdk:"tags_all"`
	TargetRoleARN      fwtypes.ARN  `tfsdk:"target_role_arn"`
}

func (model *podIdentityAssociationResourceModel) setID() {
	model.ID = model.AssociationID
}

const (
	ResNamePodIdentityAssociation = "Pod Identity Association"
)

type podIdentityAssociationResource struct {
	framework.ResourceWithModel[podIdentityAssociationResourceModel]
}

func (r *podIdentityAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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
			"disable_session_tags": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			names.AttrExternalID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
				Required:   true,
				CustomType: fwtypes.ARNType,
			},
			"service_account": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *podIdentityAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan podIdentityAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EKSClient(ctx)

	input := &eks.CreatePodIdentityAssociationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, plan, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.ClientRequestToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)
	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidParameterException](ctx, propagationTimeout, func() (any, error) {
		return conn.CreatePodIdentityAssociation(ctx, input)
	}, "Role provided in the request does not exist")

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EKS, create.ErrActionCreating, ResNamePodIdentityAssociation, plan.AssociationID.String(), err),
			err.Error(),
		)
		return
	}

	// Set values for unknowns.
	output := outputRaw.(*eks.CreatePodIdentityAssociationOutput)
	plan.AssociationARN = fwflex.StringToFramework(ctx, output.Association.AssociationArn)
	plan.AssociationID = fwflex.StringToFramework(ctx, output.Association.AssociationId)
	if !plan.ExternalID.IsNull() {
		plan.ExternalID = fwflex.StringToFramework(ctx, output.Association.ExternalId)
	}
	plan.DisableSessionTags = fwflex.BoolToFramework(ctx, output.Association.DisableSessionTags)
	plan.setID()
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (r *podIdentityAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().EKSClient(ctx)

	var data podIdentityAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	pia, err := findPodIdentityAssociationByTwoPartKey(ctx, conn, data.AssociationID.ValueString(), data.ClusterName.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EKS, create.ErrActionSetting, ResNamePodIdentityAssociation, data.AssociationID.String(), err),
			err.Error(),
		)
		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, pia, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, pia.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *podIdentityAssociationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new podIdentityAssociationResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EKSClient(ctx)

	if !new.RoleARN.Equal(old.RoleARN) ||
		!new.TargetRoleARN.Equal(old.TargetRoleARN) ||
		!new.DisableSessionTags.Equal(old.DisableSessionTags) {
		input := &eks.UpdatePodIdentityAssociationInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		input.ClientRequestToken = aws.String(sdkid.UniqueId())

		_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidParameterException](ctx, propagationTimeout, func() (any, error) {
			return conn.UpdatePodIdentityAssociation(ctx, input)
		}, "Role provided in the request does not exist")

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.EKS, create.ErrActionUpdating, ResNamePodIdentityAssociation, new.AssociationID.String(), err),
				err.Error(),
			)
			return
		}
	}
	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *podIdentityAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state podIdentityAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EKSClient(ctx)

	input := &eks.DeletePodIdentityAssociationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, state, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeletePodIdentityAssociation(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EKS, create.ErrActionDeleting, ResNamePodIdentityAssociation, state.AssociationID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *podIdentityAssociationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		partCount = 2
	)
	parts, err := flex.ExpandResourceId(request.ID, partCount, false)
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(fmt.Errorf("wrong format of import ID (%s), use: 'cluster-name,association-id'", request.ID)))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), parts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrAssociationID), parts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrClusterName), parts[0])...)
}

func findPodIdentityAssociationByTwoPartKey(ctx context.Context, conn *eks.Client, associationID, clusterName string) (*awstypes.PodIdentityAssociation, error) {
	input := eks.DescribePodIdentityAssociationInput{
		AssociationId: aws.String(associationID),
		ClusterName:   aws.String(clusterName),
	}

	output, err := conn.DescribePodIdentityAssociation(ctx, &input)

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
