// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_glue_catalog_table_optimizer",name="Catalog Table Optimizer")
func newResourceCatalogTableOptimizer(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCatalogTableOptimizer{}

	return r, nil
}

const (
	ResNameCatalogTableOptimizer = "Catalog Table Optimizer"
)

type resourceCatalogTableOptimizer struct {
	framework.ResourceWithConfigure
}

func (r *resourceCatalogTableOptimizer) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_glue_catalog_table_optimizer"
}

func (r *resourceCatalogTableOptimizer) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"catalog_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"table_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TableOptimizerType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[configurationData](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							Required: true,
						},
						"role_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
					},
				},
			},
		},
	}

	response.Schema = s
}

func (r *resourceCatalogTableOptimizer) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().GlueClient(ctx)
	var plan resourceCatalogTableOptimizerData

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	input := glue.CreateTableOptimizerInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("TableOptimizer"))...)

	if response.Diagnostics.HasError() {
		return
	}

	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		_, err := conn.CreateTableOptimizer(ctx, &input)
		if err != nil {
			// Retry IAM propagation errors
			if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "does not have the correct trust policies and is unable to be assumed by our service") {
				return retry.RetryableError(err)
			}
			if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "does not have the proper IAM permissions to call Glue APIs") {
				return retry.RetryableError(err)
			}
			if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "is not authorized to perform") {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		id, _ := flex.FlattenResourceId([]string{
			plan.CatalogID.ValueString(),
			plan.DatabaseName.ValueString(),
			plan.TableName.ValueString(),
			plan.Type.ValueString(),
		}, 4, false)

		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Glue, create.ErrActionCreating, ResNameCatalogTableOptimizer, id, err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceCatalogTableOptimizer) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().GlueClient(ctx)
	var data resourceCatalogTableOptimizerData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	output, err := findTableOptimizer(ctx, conn, data.CatalogID.ValueString(), data.DatabaseName.ValueString(), data.TableName.ValueString(), data.Type.ValueString())

	if err != nil {
		id, _ := flex.FlattenResourceId([]string{
			data.CatalogID.ValueString(),
			data.DatabaseName.ValueString(),
			data.TableName.ValueString(),
			data.Type.ValueString(),
		}, 4, false)

		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Glue, create.ErrActionReading, ResNameCatalogTableOptimizer, id, err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.TableOptimizer, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceCatalogTableOptimizer) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().GlueClient(ctx)

	var plan, state resourceCatalogTableOptimizerData
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Configuration.Equal(state.Configuration) {
		input := glue.UpdateTableOptimizerInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("TableOptimizer"))...)

		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateTableOptimizer(ctx, &input)

		if err != nil {
			id, _ := flex.FlattenResourceId([]string{
				state.CatalogID.ValueString(),
				state.DatabaseName.ValueString(),
				state.TableName.ValueString(),
				state.Type.ValueString(),
			}, 4, false)

			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Glue, create.ErrActionUpdating, ResNameCatalogTableOptimizer, id, err),
				err.Error(),
			)
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceCatalogTableOptimizer) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().GlueClient(ctx)
	var data resourceCatalogTableOptimizerData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting Glue Catalog Table Optimizer", map[string]interface{}{
		"catalogId":    data.CatalogID.ValueString(),
		"databaseName": data.DatabaseName.ValueString(),
		"tableName":    data.TableName.ValueString(),
		"type":         data.Type.ValueString(),
	})

	_, err := conn.DeleteTableOptimizer(ctx, &glue.DeleteTableOptimizerInput{
		CatalogId:    data.CatalogID.ValueStringPointer(),
		DatabaseName: data.DatabaseName.ValueStringPointer(),
		TableName:    data.TableName.ValueStringPointer(),
		Type:         data.Type.ValueEnum(),
	})

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return
	}

	if err != nil {
		id, _ := flex.FlattenResourceId([]string{
			data.CatalogID.ValueString(),
			data.DatabaseName.ValueString(),
			data.TableName.ValueString(),
			data.Type.ValueString(),
		}, 4, false)

		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Glue, create.ErrActionDeleting, ResNameCatalogTableOptimizer, id, err),
			err.Error(),
		)
		return
	}
}

func (r *resourceCatalogTableOptimizer) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts, err := flex.ExpandResourceId(request.ID, 4, false)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Glue, create.ErrActionImporting, ResNameCatalogTableOptimizer, request.ID, err),
			err.Error(),
		)
		return
	}

	state := resourceCatalogTableOptimizerData{
		CatalogID:    fwflex.StringValueToFramework(ctx, parts[0]),
		DatabaseName: fwflex.StringValueToFramework(ctx, parts[1]),
		TableName:    fwflex.StringValueToFramework(ctx, parts[2]),
		Type:         fwtypes.StringEnumValue(awstypes.TableOptimizerType(parts[3])),
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}
}

type resourceCatalogTableOptimizerData struct {
	CatalogID     types.String                                       `tfsdk:"catalog_id"`
	Configuration fwtypes.ListNestedObjectValueOf[configurationData] `tfsdk:"configuration"`
	DatabaseName  types.String                                       `tfsdk:"database_name"`
	TableName     types.String                                       `tfsdk:"table_name"`
	Type          fwtypes.StringEnum[awstypes.TableOptimizerType]    `tfsdk:"type"`
}

type configurationData struct {
	Enabled bool        `tfsdk:"enabled"`
	RoleARN fwtypes.ARN `tfsdk:"role_arn"`
}

func findTableOptimizer(ctx context.Context, conn *glue.Client, catalogID, dbName, tableName, optimizerType string) (*glue.GetTableOptimizerOutput, error) {
	input := &glue.GetTableOptimizerInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableName:    aws.String(tableName),
		Type:         awstypes.TableOptimizerType(optimizerType),
	}

	output, err := conn.GetTableOptimizer(ctx, input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
