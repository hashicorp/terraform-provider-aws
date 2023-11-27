// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Pod Identity Association")
// @Tags(identifierAttribute="association_id")
func newResourcePodIdentityAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePodIdentityAssociation{}

	return r, nil
}

type resourcePodIdentityAssociationData struct {
	AssociationArn types.String `tfsdk:"association_arn"`
	AssociationId  types.String `tfsdk:"association_id"`
	ClusterName    types.String `tfsdk:"cluster_name"`
	CreatedAt      types.String `tfsdk:"created_at"`
	Namespace      types.String `tfsdk:"namespace"`
	ModifiedAt     types.String `tfsdk:"modified_at"`
	RoleArn        fwtypes.ARN  `tfsdk:"role_arn"`
	ServiceAccount types.String `tfsdk:"service_account"`
	Tags           types.Map    `tfsdk:"tags"`
	TagsAll        types.Map    `tfsdk:"tags_all"`
}

const (
	ResNamePodIdentityAssociation = "Pod Identity Association"
)

type resourcePodIdentityAssociation struct {
	framework.ResourceWithConfigure
}

func (r *resourcePodIdentityAssociation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_eks_pod_identity_association"
}

func (r *resourcePodIdentityAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"association_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"association_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"namespace": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_arn": schema.StringAttribute{
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

func (r *resourcePodIdentityAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EKSClient(ctx)

	var plan resourcePodIdentityAssociationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &eks.CreatePodIdentityAssociationInput{
		ClusterName:        fwflex.StringFromFramework(ctx, plan.ClusterName),
		ClientRequestToken: aws.String(sdkid.UniqueId()),
		Namespace:          fwflex.StringFromFramework(ctx, plan.Namespace),
		RoleArn:            fwflex.StringFromFramework(ctx, plan.RoleArn),
		ServiceAccount:     fwflex.StringFromFramework(ctx, plan.ServiceAccount),
		Tags:               getTagsIn(ctx),
	}

	out, err := conn.CreatePodIdentityAssociation(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EKS, create.ErrActionCreating, ResNamePodIdentityAssociation, plan.AssociationId.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Association == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EKS, create.ErrActionCreating, ResNamePodIdentityAssociation, plan.AssociationId.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.AssociationArn = fwflex.StringToFramework(ctx, out.Association.AssociationArn)
	plan.AssociationId = fwflex.StringToFramework(ctx, out.Association.AssociationId)
	plan.CreatedAt = fwflex.StringToFramework(ctx, aws.String(out.Association.CreatedAt.Format(time.RFC3339)))
	plan.ModifiedAt = fwflex.StringToFramework(ctx, aws.String(out.Association.ModifiedAt.Format(time.RFC3339)))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourcePodIdentityAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EKSClient(ctx)

	var data resourcePodIdentityAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findPodIdentityAssociationByTwoPartKey(ctx, conn, data.AssociationId.ValueString(), data.ClusterName.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EKS, create.ErrActionSetting, ResNamePodIdentityAssociation, data.AssociationId.String(), err),
			err.Error(),
		)
		return
	}

	// Set attributes for import.
	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags, err := listTags(ctx, conn, data.AssociationArn.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("listing tags for Pod Identity Association (%s)", data.AssociationId.ValueString()), err.Error())

		return
	}

	setTagsOut(ctx, Tags(tags))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourcePodIdentityAssociation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourcePodIdentityAssociationData

	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EKSClient(ctx)

	if !plan.RoleArn.Equal(state.RoleArn) {
		in := &eks.UpdatePodIdentityAssociationInput{
			AssociationId: aws.String(plan.AssociationId.ValueString()),
			ClusterName:   aws.String(plan.ClusterName.ValueString()),
		}

		if !plan.RoleArn.IsNull() {
			in.RoleArn = aws.String(plan.RoleArn.ValueString())
		}

		out, err := conn.UpdatePodIdentityAssociation(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.EKS, create.ErrActionUpdating, ResNamePodIdentityAssociation, plan.AssociationId.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Association == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.EKS, create.ErrActionUpdating, ResNamePodIdentityAssociation, plan.AssociationId.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.ModifiedAt = fwflex.StringToFramework(ctx, aws.String(out.Association.ModifiedAt.Format(time.RFC3339)))
	}

	if oldTagsAll, newTagsAll := state.TagsAll, plan.TagsAll; !newTagsAll.Equal(oldTagsAll) {
		if err := updateTags(ctx, conn, plan.AssociationArn.ValueString(), oldTagsAll, newTagsAll); err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("updating tags for Pod Identity Association (%s)", plan.AssociationId.ValueString()), err.Error())

			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourcePodIdentityAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourcePodIdentityAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EKSClient(ctx)

	in := &eks.DeletePodIdentityAssociationInput{
		AssociationId: fwflex.StringFromFramework(ctx, state.AssociationId),
		ClusterName:   fwflex.StringFromFramework(ctx, state.ClusterName),
	}

	_, err := conn.DeletePodIdentityAssociation(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EKS, create.ErrActionDeleting, ResNamePodIdentityAssociation, state.AssociationId.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourcePodIdentityAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, idSeparator)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		err := fmt.Errorf("unexpected format for ID (%[1]s), expected association-id%[2]scluster-name", req.ID, idSeparator)
		resp.Diagnostics.AddError(fmt.Sprintf("importing Pod Identity Association (%s)", req.ID), err.Error())
		return
	}

	state := resourcePodIdentityAssociationData{
		AssociationId: types.StringValue(parts[0]),
		ClusterName:   types.StringValue(parts[1]),
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
func (r *resourcePodIdentityAssociation) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func findPodIdentityAssociationByTwoPartKey(ctx context.Context, conn *eks.Client, AssociationId, ClusterName string) (*awstypes.PodIdentityAssociation, error) {
	in := &eks.DescribePodIdentityAssociationInput{
		AssociationId: aws.String(AssociationId),
		ClusterName:   aws.String(ClusterName),
	}

	out, err := conn.DescribePodIdentityAssociation(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Association == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Association, nil
}
