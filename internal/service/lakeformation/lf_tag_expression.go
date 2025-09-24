// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lakeformation_lf_tag_expression", name="LF Tag Expression")
func newLFTagExpressionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &lfTagExpressionResource{}, nil
}

const (
	ResNameLFTagExpression = "LF Tag Expression"
)

type lfTagExpressionResource struct {
	framework.ResourceWithModel[lfTagExpressionResourceModel]
}

func (r *lfTagExpressionResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Manages an AWS Lake Formation Tag Expression.",
		Attributes: map[string]schema.Attribute{
			names.AttrCatalogID: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The ID of the Data Catalog.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The name of the LF-Tag Expression.",
			},
			names.AttrDescription: schema.StringAttribute{
				Optional:    true,
				Description: "A description of the LF-Tag Expression.",
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrExpression: schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[expressionLfTag](ctx),
				Validators: []validator.Set{
					setvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"tag_key": schema.StringAttribute{
							Required: true,
						},
						"tag_values": schema.SetAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *lfTagExpressionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	var data lfTagExpressionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if data.CatalogId.IsNull() || data.CatalogId.IsUnknown() {
		data.CatalogId = fwflex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	}

	input := lakeformation.CreateLFTagExpressionInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.CreateLFTagExpression(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionCreating, ResNameLFTagExpression, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *lfTagExpressionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	var data lfTagExpressionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findLFTagExpression(ctx, conn, data.Name.ValueString(), data.CatalogId.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionReading, ResNameLFTagExpression, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *lfTagExpressionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	var plan, state lfTagExpressionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input lakeformation.UpdateLFTagExpressionInput
		response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateLFTagExpression(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LakeFormation, create.ErrActionUpdating, ResNameLFTagExpression, plan.Name.String(), err),
				err.Error(),
			)
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *lfTagExpressionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	var state lfTagExpressionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := lakeformation.DeleteLFTagExpressionInput{
		CatalogId: state.CatalogId.ValueStringPointer(),
		Name:      state.Name.ValueStringPointer(),
	}

	_, err := conn.DeleteLFTagExpression(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionDeleting, ResNameLFTagExpression, state.Name.String(), err),
			err.Error(),
		)
		return
	}
}

const (
	lfTagExpressionIDPartCount = 2
)

func (r *lfTagExpressionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(request.ID, lfTagExpressionIDPartCount, false)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionImporting, ResNameLFTagExpression, request.ID, err),
			err.Error(),
		)
		return
	}

	name := parts[0]
	catalogId := parts[1]
	// Set the parsed values in state
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrName), name)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrCatalogID), catalogId)...)
	if response.Diagnostics.HasError() {
		return
	}
}

type lfTagExpressionResourceModel struct {
	framework.WithRegionModel
	CatalogId   types.String                                    `tfsdk:"catalog_id"`
	Description types.String                                    `tfsdk:"description"`
	Name        types.String                                    `tfsdk:"name"`
	Expression  fwtypes.SetNestedObjectValueOf[expressionLfTag] `tfsdk:"expression"`
}

type expressionLfTag struct {
	TagKey    types.String        `tfsdk:"tag_key"`
	TagValues fwtypes.SetOfString `tfsdk:"tag_values"`
}

func findLFTagExpression(ctx context.Context, conn *lakeformation.Client, name, catalogId string) (*lakeformation.GetLFTagExpressionOutput, error) {
	input := lakeformation.GetLFTagExpressionInput{
		CatalogId: aws.String(catalogId),
		Name:      aws.String(name),
	}

	output, err := conn.GetLFTagExpression(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Expression == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
