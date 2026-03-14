// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package servicecatalogappregistry

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_servicecatalogappregistry_application", name="Application")
// @Tags(identifierAttribute="arn")
func newApplicationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &applicationResource{}, nil
}

const (
	ResNameApplication = "Application"
)

type applicationResource struct {
	framework.ResourceWithModel[applicationResourceModel]
	framework.WithImportByID
}

func (r *applicationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *applicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ServiceCatalogAppRegistryClient(ctx)

	var plan applicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &servicecatalogappregistry.CreateApplicationInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in.Tags = getTagsIn(ctx)

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

func (r *applicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ServiceCatalogAppRegistryClient(ctx)

	var state applicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findApplicationByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
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

func (r *applicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ServiceCatalogAppRegistryClient(ctx)

	var plan, state applicationResourceModel
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
		in.Application = plan.ID.ValueStringPointer() // Set manually, AWS naming is inconsistent

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

func (r *applicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ServiceCatalogAppRegistryClient(ctx)

	var state applicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &servicecatalogappregistry.DeleteApplicationInput{
		Application: state.ID.ValueStringPointer(),
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

func findApplicationByID(ctx context.Context, conn *servicecatalogappregistry.Client, id string) (*servicecatalogappregistry.GetApplicationOutput, error) {
	in := &servicecatalogappregistry.GetApplicationInput{
		Application: aws.String(id),
	}

	out, err := conn.GetApplication(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

type applicationResourceModel struct {
	framework.WithRegionModel
	ApplicationTag fwtypes.MapOfString `tfsdk:"application_tag"`
	ARN            types.String        `tfsdk:"arn"`
	Description    types.String        `tfsdk:"description"`
	ID             types.String        `tfsdk:"id"`
	Name           types.String        `tfsdk:"name"`
	Tags           tftags.Map          `tfsdk:"tags"`
	TagsAll        tftags.Map          `tfsdk:"tags_all"`
}
