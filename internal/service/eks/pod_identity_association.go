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
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
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
			"association_arn": framework.ARNAttributeComputedOnly(),
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
		ClusterName:        flex.StringFromFramework(ctx, plan.ClusterName),
		ClientRequestToken: aws.String(sdkid.UniqueId()),
		Namespace:          flex.StringFromFramework(ctx, plan.Namespace),
		RoleArn:            flex.StringFromFramework(ctx, plan.RoleArn),
		ServiceAccount:     flex.StringFromFramework(ctx, plan.ServiceAccount),
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

	plan.AssociationArn = flex.StringToFramework(ctx, out.Association.AssociationArn)
	plan.AssociationId = flex.StringToFramework(ctx, out.Association.AssociationId)
	plan.CreatedAt = flex.StringToFramework(ctx, aws.String(out.Association.CreatedAt.Format(time.RFC3339)))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourcePodIdentityAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EKSClient(ctx)

	var state resourcePodIdentityAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findPodIdentityAssociationByTwoPartKey(ctx, conn, state.AssociationId.ValueString(), state.ClusterName.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EKS, create.ErrActionSetting, ResNamePodIdentityAssociation, state.AssociationId.String(), err),
			err.Error(),
		)
		return
	}

	state.AssociationArn = flex.StringToFramework(ctx, out.AssociationArn)
	state.AssociationId = flex.StringToFramework(ctx, out.AssociationId)
	state.ClusterName = flex.StringToFramework(ctx, out.ClusterName)
	state.CreatedAt = flex.StringToFramework(ctx, aws.String(out.CreatedAt.Format(time.RFC3339)))
	state.ModifiedAt = flex.StringToFramework(ctx, aws.String(out.ModifiedAt.Format(time.RFC3339)))
	state.Namespace = flex.StringToFramework(ctx, out.Namespace)
	state.RoleArn = fwtypes.ARNValue(aws.ToString(out.RoleArn))
	state.ServiceAccount = flex.StringToFramework(ctx, out.ServiceAccount)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourcePodIdentityAssociation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().EKSClient(ctx)

	var plan, state resourcePodIdentityAssociationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

		// Using the output from the update function, re-set any computed attributes
		plan.AssociationArn = flex.StringToFramework(ctx, out.Association.AssociationArn)
		plan.AssociationId = flex.StringToFramework(ctx, out.Association.AssociationId)
		plan.CreatedAt = flex.StringToFramework(ctx, aws.String(out.Association.CreatedAt.Format(time.RFC3339)))
		plan.ModifiedAt = flex.StringToFramework(ctx, aws.String(out.Association.ModifiedAt.Format(time.RFC3339)))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourcePodIdentityAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EKSClient(ctx)

	var state resourcePodIdentityAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &eks.DeletePodIdentityAssociationInput{
		AssociationId: aws.String(state.AssociationId.ValueString()),
		ClusterName:   aws.String(state.ClusterName.ValueString()),
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
