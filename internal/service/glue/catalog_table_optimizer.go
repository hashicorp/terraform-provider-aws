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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
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

	idParts = 4
)

type resourceCatalogTableOptimizer struct {
	framework.ResourceWithConfigure
}

func (r *resourceCatalogTableOptimizer) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCatalogID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDatabaseName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTableName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TableOptimizerType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[configurationData](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrEnabled: schema.BoolAttribute{
							Required: true,
						},
						names.AttrRoleARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"retention_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[retentionConfigurationData](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"iceberg_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[icebergRetentionConfigurationData](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"snapshot_retention_period_in_days": schema.Int32Attribute{
													Optional: true,
												},
												"number_of_snapshots_to_retain": schema.Int32Attribute{
													Optional: true,
												},
												"clean_expired_files": schema.BoolAttribute{
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"orphan_file_deletion_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[orphanFileDeletionConfigurationData](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"iceberg_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[icebergOrphanFileDeletionConfigurationData](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"orphan_file_retention_period_in_days": schema.Int32Attribute{
													Optional: true,
												},
												names.AttrLocation: schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
								},
							},
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

	if tfresource.TimedOut(err) {
		_, err = conn.CreateTableOptimizer(ctx, &input)
	}

	if err != nil {
		id, _ := flex.FlattenResourceId([]string{
			plan.CatalogID.ValueString(),
			plan.DatabaseName.ValueString(),
			plan.TableName.ValueString(),
			plan.Type.ValueString(),
		}, idParts, false)

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

	output, err := findCatalogTableOptimizer(ctx, conn, data.CatalogID.ValueString(), data.DatabaseName.ValueString(), data.TableName.ValueString(), data.Type.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		id, _ := flex.FlattenResourceId([]string{
			data.CatalogID.ValueString(),
			data.DatabaseName.ValueString(),
			data.TableName.ValueString(),
			data.Type.ValueString(),
		}, idParts, false)

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
			}, idParts, false)

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

	tflog.Debug(ctx, "deleting Glue Catalog Table Optimizer", map[string]any{
		names.AttrCatalogID:    data.CatalogID.ValueString(),
		names.AttrDatabaseName: data.DatabaseName.ValueString(),
		names.AttrTableName:    data.TableName.ValueString(),
		names.AttrType:         data.Type.ValueString(),
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
		}, idParts, false)

		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Glue, create.ErrActionDeleting, ResNameCatalogTableOptimizer, id, err),
			err.Error(),
		)
		return
	}
}

func (r *resourceCatalogTableOptimizer) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts, err := flex.ExpandResourceId(request.ID, idParts, false)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Glue, create.ErrActionImporting, ResNameCatalogTableOptimizer, request.ID, err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrCatalogID), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrDatabaseName), parts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTableName), parts[2])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrType), parts[3])...)
}

type resourceCatalogTableOptimizerData struct {
	CatalogID     types.String                                       `tfsdk:"catalog_id"`
	Configuration fwtypes.ListNestedObjectValueOf[configurationData] `tfsdk:"configuration"`
	DatabaseName  types.String                                       `tfsdk:"database_name"`
	TableName     types.String                                       `tfsdk:"table_name"`
	Type          fwtypes.StringEnum[awstypes.TableOptimizerType]    `tfsdk:"type"`
}

type configurationData struct {
	Enabled                         types.Bool                                                           `tfsdk:"enabled"`
	RoleARN                         fwtypes.ARN                                                          `tfsdk:"role_arn"`
	RetentionConfiguration          fwtypes.ListNestedObjectValueOf[retentionConfigurationData]          `tfsdk:"retention_configuration"`
	OrphanFileDeletionConfiguration fwtypes.ListNestedObjectValueOf[orphanFileDeletionConfigurationData] `tfsdk:"orphan_file_deletion_configuration"`
}

type retentionConfigurationData struct {
	IcebergConfiguration fwtypes.ListNestedObjectValueOf[icebergRetentionConfigurationData] `tfsdk:"iceberg_configuration"`
}

type icebergRetentionConfigurationData struct {
	SnapshotRetentionPeriodInDays types.Int32 `tfsdk:"snapshot_retention_period_in_days"`
	NumberOfSnapshotsToRetain     types.Int32 `tfsdk:"number_of_snapshots_to_retain"`
	CleanExpiredFiles             types.Bool  `tfsdk:"clean_expired_files"`
}

type orphanFileDeletionConfigurationData struct {
	IcebergConfiguration fwtypes.ListNestedObjectValueOf[icebergOrphanFileDeletionConfigurationData] `tfsdk:"iceberg_configuration"`
}

type icebergOrphanFileDeletionConfigurationData struct {
	OrphanFileRetentionPeriodInDays types.Int32  `tfsdk:"orphan_file_retention_period_in_days"`
	Location                        types.String `tfsdk:"location"`
}

func findCatalogTableOptimizer(ctx context.Context, conn *glue.Client, catalogID, dbName, tableName, optimizerType string) (*glue.GetTableOptimizerOutput, error) {
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
