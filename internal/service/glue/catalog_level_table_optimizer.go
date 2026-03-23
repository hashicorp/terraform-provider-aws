// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_glue_catalog_level_table_optimizer", name="Catalog Level Table Optimizer")
func newCatalogLevelTableOptimizerResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &catalogLevelTableOptimizerResource{}
	return r, nil
}

const (
	ResNameCatalogLevelTableOptimizer = "Catalog Level Table Optimizer"
)

type catalogLevelTableOptimizerResource struct {
	framework.ResourceWithModel[catalogLevelTableOptimizerResourceModel]
}

func (r *catalogLevelTableOptimizerResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCatalogID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"iceberg_optimization": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[icebergOptimizationData](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrRoleARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
						"compaction": schema.MapAttribute{
							ElementType: types.StringType,
							Optional:    true,
						},
						"retention": schema.MapAttribute{
							ElementType: types.StringType,
							Optional:    true,
						},
						"orphan_file_deletion": schema.MapAttribute{
							ElementType: types.StringType,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (r *catalogLevelTableOptimizerResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().GlueClient(ctx)
	var plan catalogLevelTableOptimizerResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	icebergOpts, diags := plan.IcebergOptimization.ToPtr(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	compaction := fwflex.ExpandFrameworkStringValueMap(ctx, icebergOpts.Compaction)
	retention := fwflex.ExpandFrameworkStringValueMap(ctx, icebergOpts.Retention)
	orphan := fwflex.ExpandFrameworkStringValueMap(ctx, icebergOpts.OrphanFileDeletion)
	if response.Diagnostics.HasError() {
		return
	}

	input := &glue.UpdateCatalogInput{
		CatalogId: aws.String(plan.CatalogID.ValueString()),
		CatalogInput: &awstypes.CatalogInput{
			CatalogProperties: &awstypes.CatalogProperties{
				IcebergOptimizationProperties: &awstypes.IcebergOptimizationProperties{
					RoleArn:            aws.String(icebergOpts.RoleARN.ValueString()),
					Compaction:         compaction,
					Retention:          retention,
					OrphanFileDeletion: orphan,
				},
			},
		},
	}

	_, err := conn.UpdateCatalog(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Glue, create.ErrActionCreating, ResNameCatalogLevelTableOptimizer, plan.CatalogID.ValueString(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *catalogLevelTableOptimizerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().GlueClient(ctx)
	var data catalogLevelTableOptimizerResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findCatalogLevelTableOptimizer(ctx, conn, data.CatalogID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Glue, create.ErrActionReading, ResNameCatalogLevelTableOptimizer, data.CatalogID.ValueString(), err),
			err.Error(),
		)
		return
	}

	props := output.Catalog.CatalogProperties
	if props == nil || props.IcebergOptimizationProperties == nil {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	ice := props.IcebergOptimizationProperties

	compaction := fwflex.FlattenFrameworkStringValueMap(ctx, ice.Compaction)
	retention := fwflex.FlattenFrameworkStringValueMap(ctx, ice.Retention)
	orphan := fwflex.FlattenFrameworkStringValueMap(ctx, ice.OrphanFileDeletion)
	if response.Diagnostics.HasError() {
		return
	}

	iceData := &icebergOptimizationData{
		RoleARN:            fwtypes.ARNValue(aws.ToString(ice.RoleArn)),
		Compaction:         compaction,
		Retention:          retention,
		OrphanFileDeletion: orphan,
	}

	iceList, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, iceData)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}
	if response.Diagnostics.HasError() {
		return
	}

	data.IcebergOptimization = iceList

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *catalogLevelTableOptimizerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().GlueClient(ctx)
	var plan, state catalogLevelTableOptimizerResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.IcebergOptimization.Equal(state.IcebergOptimization) {
		icebergOpts, diags := plan.IcebergOptimization.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		compaction := fwflex.ExpandFrameworkStringValueMap(ctx, icebergOpts.Compaction)
		retention := fwflex.ExpandFrameworkStringValueMap(ctx, icebergOpts.Retention)
		orphan := fwflex.ExpandFrameworkStringValueMap(ctx, icebergOpts.OrphanFileDeletion)
		if response.Diagnostics.HasError() {
			return
		}

		input := &glue.UpdateCatalogInput{
			CatalogId: aws.String(plan.CatalogID.ValueString()),
			CatalogInput: &awstypes.CatalogInput{
				CatalogProperties: &awstypes.CatalogProperties{
					IcebergOptimizationProperties: &awstypes.IcebergOptimizationProperties{
						RoleArn:            aws.String(icebergOpts.RoleARN.ValueString()),
						Compaction:         compaction,
						Retention:          retention,
						OrphanFileDeletion: orphan,
					},
				},
			},
		}

		_, err := conn.UpdateCatalog(ctx, input)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Glue, create.ErrActionUpdating, ResNameCatalogLevelTableOptimizer, plan.CatalogID.ValueString(), err),
				err.Error(),
			)
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *catalogLevelTableOptimizerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().GlueClient(ctx)
	var data catalogLevelTableOptimizerResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting Glue Catalog Level Table Optimizer", map[string]any{
		names.AttrCatalogID: data.CatalogID.ValueString(),
	})

	// Deletion clears the IcebergOptimizationProperties by setting them to empty
	_, err := conn.UpdateCatalog(ctx, &glue.UpdateCatalogInput{
		CatalogId: aws.String(data.CatalogID.ValueString()),
		CatalogInput: &awstypes.CatalogInput{
			CatalogProperties: &awstypes.CatalogProperties{
				IcebergOptimizationProperties: &awstypes.IcebergOptimizationProperties{},
			},
		},
	})

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Glue, create.ErrActionDeleting, ResNameCatalogLevelTableOptimizer, data.CatalogID.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func (r *catalogLevelTableOptimizerResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// Import by catalog_id
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrCatalogID), request.ID)...)
}

func findCatalogLevelTableOptimizer(ctx context.Context, conn *glue.Client, catalogID string) (*glue.GetCatalogOutput, error) {
	input := &glue.GetCatalogInput{
		CatalogId: aws.String(catalogID),
	}

	output, err := conn.GetCatalog(ctx, input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Catalog == nil {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

type catalogLevelTableOptimizerResourceModel struct {
	framework.WithRegionModel
	CatalogID           types.String                                                `tfsdk:"catalog_id"`
	IcebergOptimization fwtypes.ListNestedObjectValueOf[icebergOptimizationData]    `tfsdk:"iceberg_optimization"`
}

type icebergOptimizationData struct {
	RoleARN            fwtypes.ARN       `tfsdk:"role_arn"`
	Compaction         types.Map         `tfsdk:"compaction"`
	Retention          types.Map         `tfsdk:"retention"`
	OrphanFileDeletion types.Map         `tfsdk:"orphan_file_deletion"`
}
