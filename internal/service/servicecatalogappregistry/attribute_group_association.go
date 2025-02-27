// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_servicecatalogappregistry_attribute_group_association", name="Attribute Group Association")
func newResourceAttributeGroupAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceAttributeGroupAssociation{}, nil
}

const (
	ResNameAttributeGroupAssociation = "Attribute Group Association"

	attributeGroupAssociationIDParts = 2
)

type resourceAttributeGroupAssociation struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
}

func (r *resourceAttributeGroupAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrApplicationID: schema.StringAttribute{
				Description: "ID of the application.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"attribute_group_id": schema.StringAttribute{
				Description: "ID of the attribute group to associate with the application.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceAttributeGroupAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ServiceCatalogAppRegistryClient(ctx)

	var plan resourceAttributeGroupAssociationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &servicecatalogappregistry.AssociateAttributeGroupInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.AssociateAttributeGroup(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionCreating, ResNameAttributeGroupAssociation, plan.AttributeGroup.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceAttributeGroupAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ServiceCatalogAppRegistryClient(ctx)

	var state resourceAttributeGroupAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := findAttributeGroupAssociationByTwoPartKey(ctx, conn, state.Application.ValueString(), state.AttributeGroup.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionReading, ResNameAttributeGroupAssociation, state.AttributeGroup.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAttributeGroupAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ServiceCatalogAppRegistryClient(ctx)

	var state resourceAttributeGroupAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &servicecatalogappregistry.DisassociateAttributeGroupInput{
		Application:    state.Application.ValueStringPointer(),
		AttributeGroup: state.AttributeGroup.ValueStringPointer(),
	}

	_, err := conn.DisassociateAttributeGroup(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionDeleting, ResNameAttributeGroupAssociation, state.AttributeGroup.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceAttributeGroupAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(req.ID, attributeGroupAssociationIDParts, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: application_id,attribute_group_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrApplicationID), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("attribute_group_id"), parts[1])...)
}

func findAttributeGroupAssociationByTwoPartKey(ctx context.Context, conn *servicecatalogappregistry.Client, applicationId string, attributeGroupId string) (string, error) {
	in := &servicecatalogappregistry.ListAssociatedAttributeGroupsInput{
		Application: aws.String(applicationId),
	}

	paginator := servicecatalogappregistry.NewListAssociatedAttributeGroupsPaginator(conn, in)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return "", err
		}

		for _, item := range page.AttributeGroups {
			if item == attributeGroupId {
				return item, nil
			}
		}
	}

	return "", &retry.NotFoundError{
		LastRequest: in,
	}
}

type resourceAttributeGroupAssociationData struct {
	Application    types.String `tfsdk:"application_id"`
	AttributeGroup types.String `tfsdk:"attribute_group_id"`
}
