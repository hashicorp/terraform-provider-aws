// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"target_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
		},
	}
}

func (r *podIdentityAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data podIdentityAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EKSClient(ctx)

	var input eks.CreatePodIdentityAssociationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientRequestToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.InvalidParameterException](ctx, propagationTimeout, func(ctx context.Context) (any, error) {
		return conn.CreatePodIdentityAssociation(ctx, &input)
	}, "Role provided in the request does not exist")

	if err != nil {
		response.Diagnostics.AddError("creating EKS Pod Identity Association", err.Error())

		return
	}

	// Set values for unknowns.
	association := outputRaw.(*eks.CreatePodIdentityAssociationOutput).Association
	data.AssociationID = fwflex.StringToFramework(ctx, association.AssociationId)
	data.AssociationARN = fwflex.StringToFramework(ctx, association.AssociationArn)
	data.ExternalID = fwflex.StringToFramework(ctx, association.ExternalId)
	data.ID = data.AssociationID

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *podIdentityAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data podIdentityAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EKSClient(ctx)

	association, err := findPodIdentityAssociationByTwoPartKey(ctx, conn, data.AssociationID.ValueString(), data.ClusterName.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EKS Pod Identity Association (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, association, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, association.Tags)

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

	if !new.DisableSessionTags.Equal(old.DisableSessionTags) ||
		!new.RoleARN.Equal(old.RoleARN) ||
		!new.TargetRoleARN.Equal(old.TargetRoleARN) {
		var input eks.UpdatePodIdentityAssociationInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Set values for unknowns.
		input.ClientRequestToken = aws.String(sdkid.UniqueId())

		outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.InvalidParameterException](ctx, propagationTimeout, func(ctx context.Context) (any, error) {
			return conn.UpdatePodIdentityAssociation(ctx, &input)
		}, "Role provided in the request does not exist")

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating EKS Pod Identity Association (%s)", new.ID.ValueString()), err.Error())

			return
		}

		// Set values for unknowns.
		association := outputRaw.(*eks.UpdatePodIdentityAssociationOutput).Association
		new.ExternalID = fwflex.StringToFramework(ctx, association.ExternalId)
	} else {
		new.ExternalID = old.ExternalID
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *podIdentityAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data podIdentityAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EKSClient(ctx)

	var input eks.DeletePodIdentityAssociationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeletePodIdentityAssociation(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EKS Pod Identity Association (%s)", data.ID.ValueString()), err.Error())

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
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Association == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Association, nil
}
