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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Application")
func newResourceApplication(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceApplication{}, nil
}

const (
	ResNameApplication = "Application"
)

type resourceApplication struct {
	framework.ResourceWithConfigure
}

func (r *resourceApplication) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_servicecatalogappregistry_application"
}

func (r *resourceApplication) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				// Name cannot be updated despite being present in the update API struct.
				//
				// ValidationException: Application name update feature has been deprecated.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application_tag": schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (r *resourceApplication) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ServiceCatalogAppRegistryClient(ctx)

	var plan resourceApplicationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &servicecatalogappregistry.CreateApplicationInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateApplication(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionCreating, ResNameApplication, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Application == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionCreating, ResNameApplication, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.Application, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceApplication) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ServiceCatalogAppRegistryClient(ctx)

	var state resourceApplicationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findApplicationByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionSetting, ResNameApplication, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceApplication) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ServiceCatalogAppRegistryClient(ctx)

	var plan, state resourceApplicationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) {
		in := &servicecatalogappregistry.UpdateApplicationInput{}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.Application = aws.String(plan.ID.ValueString()) // Set manually, AWS naming is inconsistent

		out, err := conn.UpdateApplication(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionUpdating, ResNameApplication, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Application == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionUpdating, ResNameApplication, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out.Application, &plan)...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceApplication) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ServiceCatalogAppRegistryClient(ctx)

	var state resourceApplicationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &servicecatalogappregistry.DeleteApplicationInput{
		Application: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteApplication(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionDeleting, ResNameApplication, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceApplication) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func findApplicationByID(ctx context.Context, conn *servicecatalogappregistry.Client, id string) (*servicecatalogappregistry.GetApplicationOutput, error) {
	in := &servicecatalogappregistry.GetApplicationInput{
		Application: aws.String(id),
	}

	out, err := conn.GetApplication(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceApplicationData struct {
	ARN            types.String `tfsdk:"arn"`
	Description    types.String `tfsdk:"description"`
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ApplicationTag types.Map    `tfsdk:"application_tag"`
}
