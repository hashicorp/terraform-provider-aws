// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	flex2 "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_servicecatalogappregistry_application_attribute_group_association", name="Application Attribute Group Association")
func newResourceApplicationAttributeGroupAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceApplicationAttributeGroupAssociation{}

	return r, nil
}

const (
	ResNameApplicationAttributeGroupAssociation = "Application Attribute Group Association"
)

const applicationAttributeGroupAssociationIDParts = 2

type resourceApplicationAttributeGroupAssociation struct {
	framework.ResourceWithConfigure
}

func (r *resourceApplicationAttributeGroupAssociation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_servicecatalogappregistry_application_attribute_group_association"
}

func (r *resourceApplicationAttributeGroupAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrApplicationID: schema.StringAttribute{
				Description: "ID of the application to associate with",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"attribute_group_id": schema.StringAttribute{
				Description: "ID of the attribute group to associate with",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceApplicationAttributeGroupAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ServiceCatalogAppRegistryClient(ctx)

	var plan resourceApplicationAttributeGroupAssociationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &servicecatalogappregistry.AssociateAttributeGroupInput{}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)

	out, err := conn.AssociateAttributeGroup(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionCreating, ResNameApplicationAttributeGroupAssociation, plan.AttributeGroup.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionCreating, ResNameApplicationAttributeGroupAssociation, plan.AttributeGroup.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}
	plan.setId()

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceApplicationAttributeGroupAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ServiceCatalogAppRegistryClient(ctx)

	var state resourceApplicationAttributeGroupAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := state.parseId()
	if err != nil {
		resp.Diagnostics.AddError("failed to parse id", err.Error())
		return
	}

	exists, err := findApplicationAttributeGroupAssociationByID(ctx, conn, state.Application.ValueString(), state.AttributeGroup.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionSetting, ResNameApplicationAttributeGroupAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	if exists != nil && !*exists {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceApplicationAttributeGroupAssociation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceApplicationAttributeGroupAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ServiceCatalogAppRegistryClient(ctx)

	var state resourceApplicationAttributeGroupAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &servicecatalogappregistry.DisassociateAttributeGroupInput{
		Application:    aws.String(state.Application.ValueString()),
		AttributeGroup: aws.String(state.AttributeGroup.ValueString()),
	}

	_, err := conn.DisassociateAttributeGroup(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionDeleting, ResNameApplicationAttributeGroupAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceApplicationAttributeGroupAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func findApplicationAttributeGroupAssociationByID(ctx context.Context, conn *servicecatalogappregistry.Client, applicationId string, attributeGroupId string) (*bool, error) {
	in := &servicecatalogappregistry.ListAssociatedAttributeGroupsInput{
		Application: aws.String(applicationId),
	}
	exists := false

	paginator := servicecatalogappregistry.NewListAssociatedAttributeGroupsPaginator(conn, in)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, item := range page.AttributeGroups {
			if item == attributeGroupId {
				exists = true
				return &exists, nil
			}
		}
	}

	return &exists, nil
}

type resourceApplicationAttributeGroupAssociationData struct {
	ID             types.String `tfsdk:"id"`
	Application    types.String `tfsdk:"application_id"`
	AttributeGroup types.String `tfsdk:"attribute_group_id"`
}

func (data *resourceApplicationAttributeGroupAssociationData) setId() {
	data.ID = types.StringValue(errs.Must(flex2.FlattenResourceId([]string{data.Application.ValueString(), data.AttributeGroup.ValueString()}, applicationAttributeGroupAssociationIDParts, false)))
}

func (data *resourceApplicationAttributeGroupAssociationData) parseId() error {
	id := data.ID.ValueString()
	parts, err := flex2.ExpandResourceId(id, applicationAttributeGroupAssociationIDParts, false)

	if err != nil {
		return err
	}

	data.Application = types.StringValue(parts[0])
	data.AttributeGroup = types.StringValue(parts[1])

	return nil
}
