// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package databrew

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databrew"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databrew/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_databrew_dataset", name="Dataset")
func newResourceDataset(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDataset{}

	return r, nil
}

const (
	ResNameDataset = "Dataset"
)

type resourceDataset struct {
	framework.ResourceWithConfigure
}

func (r *resourceDataset) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_databrew_dataset"
}

func (r *resourceDataset) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrFormat: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.InputFormat](),
				Optional:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"format_options": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[formatOptionsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"csv": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[csvFormatOptionsModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"delimiter": schema.BoolAttribute{
										Optional: true,
									},
									"header_row": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
						"excel": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[excelFormatOptionsModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"header_row": schema.StringAttribute{
										Optional: true,
									},
									"sheet_indexes": schema.StringAttribute{
										Optional: true,
									},
									"sheet_names": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
						names.AttrJSON: schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[jsonFormatOptionsModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"multi_line": schema.BoolAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"input": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[inputModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"metadata": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[metadataModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"source_arn": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
						"database_input_definition": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[databaseInputDefinitionModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"glue_connection_name": schema.StringAttribute{
										Required: true,
									},
									"database_table_name": schema.StringAttribute{
										Optional: true,
									},
									"query_string": schema.StringAttribute{
										Optional: true,
									},
								},
								Blocks: map[string]schema.Block{
									"temp_directory": schema.SetNestedBlock{
										CustomType: fwtypes.NewSetNestedObjectTypeOf[s3LocationModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrBucket: schema.StringAttribute{
													Required: true,
												},
												"bucket_owner": schema.StringAttribute{
													Optional: true,
												},
												names.AttrKey: schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"data_catalog_input_definition": schema.SetNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrCatalogID: schema.StringAttribute{
										Optional: true,
									},
									names.AttrDatabaseName: schema.StringAttribute{
										Required: true,
									},
									names.AttrTableName: schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"temp_directory": schema.SetNestedBlock{
										CustomType: fwtypes.NewSetNestedObjectTypeOf[s3LocationModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrBucket: schema.StringAttribute{
													Required: true,
												},
												"bucket_owner": schema.StringAttribute{
													Optional: true,
												},
												names.AttrKey: schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"s3_input_definition": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[s3LocationModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrBucket: schema.StringAttribute{
										Required: true,
									},
									"bucket_owner": schema.StringAttribute{
										Optional: true,
									},
									names.AttrKey: schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"path_options": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[pathOptionsModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"files_limit": schema.SetNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"max_files": schema.Int64Attribute{
										Required: true,
									},
									"order": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.Order](),
										Optional:   true,
									},
									"ordered_by": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.OrderedBy](),
										Optional:   true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceDataset) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceDatasetModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataBrewClient(ctx)

	in := &databrew.CreateDatasetInput{}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)

	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateDataset(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataBrew, create.ErrActionCreating, "Create Dataset", plan.Name.String(), nil),
			err.Error(),
		)
		return
	}

	var state resourceDatasetModel
	dataset, _ := findDatasetByName(ctx, conn, *out.Name)

	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = flex.StringToFramework(ctx, dataset.Name)
	resp.Diagnostics.Append(flex.Flatten(ctx, dataset, &state)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDataset) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataBrewClient(ctx)

	var data resourceDatasetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataset, err := findDatasetByName(ctx, conn, data.Name.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataBrew, create.ErrActionSetting, ResNameDataset, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	var readData resourceDatasetModel

	readData.ID = flex.StringToFramework(ctx, dataset.Name)
	resp.Diagnostics.Append(flex.Flatten(ctx, dataset, &readData)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &readData)...)
}

func (r *resourceDataset) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataBrewClient(ctx)

	var plan, state resourceDatasetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) {
		in := &databrew.UpdateDatasetInput{
			Name: aws.String(plan.Name.ValueString()),
		}

		out, err := conn.UpdateDataset(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataBrew, create.ErrActionUpdating, ResNameDataset, plan.Name.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Name == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataBrew, create.ErrActionUpdating, ResNameDataset, plan.Name.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.Name = flex.StringToFramework(ctx, out.Name)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDataset) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataBrewClient(ctx)

	var state resourceDatasetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &databrew.DeleteDatasetInput{
		Name: aws.String(state.Name.ValueString()),
	}

	_, err := conn.DeleteDataset(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataBrew, create.ErrActionDeleting, ResNameDataset, state.Name.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceDataset) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), req, resp)
}

func findDatasetByName(ctx context.Context, conn *databrew.Client, name string) (*awstypes.Dataset, error) {
	in := &databrew.DescribeDatasetInput{
		Name: aws.String(name),
	}

	out, err := conn.DescribeDataset(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Name == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	dataset := &awstypes.Dataset{}
	dataset.Input = out.Input
	dataset.PathOptions = out.PathOptions
	dataset.FormatOptions = out.FormatOptions
	dataset.Format = out.Format
	dataset.Name = out.Name

	return dataset, nil
}

type s3LocationModel struct {
	Bucket      types.String `tfsdk:"bucket"`
	BucketOwner types.String `tfsdk:"bucket_owner"`
	Key         types.String `tfsdk:"key"`
}

type dataCatalogInputDefinitionModel struct {
	CatalogId     types.String                                    `tfsdk:"catalog_id"`
	DatabaseName  types.String                                    `tfsdk:"database_name"`
	TableName     types.String                                    `tfsdk:"table_name"`
	TempDirectory fwtypes.SetNestedObjectValueOf[s3LocationModel] `tfsdk:"temp_directory"`
}

type databaseInputDefinitionModel struct {
	GlueConnectionName types.String                                    `tfsdk:"glue_connection_name"`
	DatabaseTableName  types.String                                    `tfsdk:"database_table_name"`
	QueryString        types.String                                    `tfsdk:"query_string"`
	TempDirectory      fwtypes.SetNestedObjectValueOf[s3LocationModel] `tfsdk:"temp_directory"`
}

type metadataModel struct {
	SourceArn types.String `tfsdk:"source_arn"`
}

type inputModel struct {
	S3InputDefinition          fwtypes.SetNestedObjectValueOf[s3LocationModel]                 `tfsdk:"s3_input_definition"`
	DataCatalogInputDefinition fwtypes.SetNestedObjectValueOf[dataCatalogInputDefinitionModel] `tfsdk:"data_catalog_input_definition"`
	DatabaseInputDefinition    fwtypes.SetNestedObjectValueOf[databaseInputDefinitionModel]    `tfsdk:"database_input_definition"`
	Metadata                   fwtypes.SetNestedObjectValueOf[metadataModel]                   `tfsdk:"metadata"`
}

type formatOptionsModel struct {
	CSV   fwtypes.SetNestedObjectValueOf[csvFormatOptionsModel]   `tfsdk:"csv"`
	JSON  fwtypes.SetNestedObjectValueOf[jsonFormatOptionsModel]  `tfsdk:"json"`
	Excel fwtypes.SetNestedObjectValueOf[excelFormatOptionsModel] `tfsdk:"excel"`
}

type csvFormatOptionsModel struct {
	Delimiter types.Bool   `tfsdk:"delimiter"`
	HeaderRow types.String `tfsdk:"header_row"`
}

type excelFormatOptionsModel struct {
	SheetIndexes types.String `tfsdk:"sheet_indexes"`
	SheetNames   types.String `tfsdk:"sheet_names"`
	HeaderRow    types.String `tfsdk:"header_row"`
}

type jsonFormatOptionsModel struct {
	MultiLine types.Bool `tfsdk:"multi_line"`
}

type pathOptionsModel struct {
	FilesLimit fwtypes.SetNestedObjectValueOf[filesLimitPathOptionsModel] `tfsdk:"files_limit"`
}

type filesLimitPathOptionsModel struct {
	MaxFiles  types.Int64                            `tfsdk:"max_files"`
	Order     fwtypes.StringEnum[awstypes.Order]     `tfsdk:"order"`
	OrderedBy fwtypes.StringEnum[awstypes.OrderedBy] `tfsdk:"ordered_by"`
}

type resourceDatasetModel struct {
	ID            types.String                                        `tfsdk:"id"`
	Name          types.String                                        `tfsdk:"name"`
	Format        fwtypes.StringEnum[awstypes.InputFormat]            `tfsdk:"format"`
	Input         fwtypes.ListNestedObjectValueOf[inputModel]         `tfsdk:"input"`
	FormatOptions fwtypes.ListNestedObjectValueOf[formatOptionsModel] `tfsdk:"format_options"`
	PathOptions   fwtypes.ListNestedObjectValueOf[pathOptionsModel]   `tfsdk:"path_options"`
}
