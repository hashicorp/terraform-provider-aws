// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Import Table")
func newResourceImportTable(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceImportTable{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameImportTable = "Import Table"
)

type resourceImportTable struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceImportTable) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_dynamodb_import_table"
}

func (r *resourceImportTable) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"id":  framework.IDAttribute(),
			"input_compression_type": schema.StringAttribute{
				Description: "Type of compression to be used on the input coming from the imported table.",
				Optional:    true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.InputCompressionType](),
				},
			},
			"input_format": schema.StringAttribute{
				Description: "The format of the source data.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.InputFormat](),
				},
			},
			"table_id": framework.IDAttribute(),
		},
	}

	s.Blocks = map[string]schema.Block{
		"input_format_options": schema.ListNestedBlock{
			Description: "Additional properties that specify how the input is formatted",
			Validators: []validator.List{
				listvalidator.SizeAtMost(1),
			},
			NestedObject: schema.NestedBlockObject{
				Blocks: map[string]schema.Block{
					"csv": schema.ListNestedBlock{
						Description: "The options for imported source files in CSV format.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"delimiter": schema.StringAttribute{
									Description: "The delimiter used for separating items in the CSV file being imported.",
									Optional:    true,
								},
								"header_list": schema.SetAttribute{
									ElementType: types.StringType,
									Description: "List of the headers used to specify a common header for all source CSV files being imported.",
									Optional:    true,
								},
							},
						},
					},
				},
			},
		},
		"s3_bucket_source": schema.ListNestedBlock{
			Description: "The S3 bucket that provides the source for the import.",
			Validators: []validator.List{
				listvalidator.IsRequired(),
				listvalidator.SizeAtMost(1),
			},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Description: "The S3 bucket that is being imported from.",
						Required:    true,
					},
					"bucket_owner": schema.StringAttribute{
						Description: "The account number of the S3 bucket that is being imported from.",
						Optional:    true,
					},
					"prefix": schema.StringAttribute{
						Description: "The key prefix shared by all S3 Objects that are being imported.",
						Optional:    true,
					},
				},
			},
		},
		"table_creation_parameters": schema.ListNestedBlock{
			Description: "Parameters for the table to import the data into.",
			Validators: []validator.List{
				listvalidator.IsRequired(),
				listvalidator.SizeAtMost(1),
			},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"billing_mode": schema.StringAttribute{
						Description: "The billing mode for provisioning the table created as part of the import operation",
						Required:    true,
						Validators: []validator.String{
							enum.FrameworkValidate[awstypes.BillingMode](),
						},
					},
					"name": schema.StringAttribute{
						Description: "The name of the table created as part of the import operation",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
				Blocks: map[string]schema.Block{
					"attribute_definition": schema.SetNestedBlock{
						Description: "The attributes of the table created as part of the import operation",
						Validators: []validator.Set{
							setvalidator.IsRequired(),
						},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "A name for the attribute",
									Required:    true,
								},
								"type": schema.StringAttribute{
									Description: "The data type for the attribute",
									Required:    true,
									Validators: []validator.String{
										enum.FrameworkValidate[awstypes.ScalarAttributeType](),
									},
								},
							},
						},
					},
					"key_schema": schema.SetNestedBlock{
						Description: "The complete key schema for a global secondary index",
						Validators: []validator.Set{
							setvalidator.IsRequired(),
						},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"attribute_name": schema.StringAttribute{
									Description: "The name of a key attribute",
									Required:    true,
								},
								"type": schema.StringAttribute{
									Description: "he role that this key attribute will assum",
									Required:    true,
									Validators: []validator.String{
										enum.FrameworkValidate[awstypes.KeyType](),
									},
								},
							},
						},
					},
					"global_secondary_index": schema.SetNestedBlock{
						Description: "The Global Secondary Indexes (GSI) of the table to be created as part of the import operation",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "The name of the global secondary index.",
									Required:    true,
								},
							},
							Blocks: map[string]schema.Block{
								"key_schema": schema.SetNestedBlock{
									Description: "The complete key schema for a global secondary index",
									Validators: []validator.Set{
										setvalidator.IsRequired(),
									},
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"attribute_name": schema.StringAttribute{
												Description: "The name of a key attribute",
												Required:    true,
											},
											"type": schema.StringAttribute{
												Description: "he role that this key attribute will assum",
												Required:    true,
												Validators: []validator.String{
													enum.FrameworkValidate[awstypes.KeyType](),
												},
											},
										},
									},
								},
								"provisioned_throughput": schema.ListNestedBlock{
									Description: "Represents the provisioned throughput settings for a specified table or index",
									Validators: []validator.List{
										listvalidator.SizeAtMost(1),
									},
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"read_capacity_units": schema.Int64Attribute{
												Description: "The maximum number of strongly consistent reads consumed per second before DynamoDB returns a ThrottlingException",
												Required:    true,
											},
											"write_capacity_units": schema.Int64Attribute{
												Description: "The maximum number of writes consumed per second before DynamoDB returns a ThrottlingException",
												Required:    true,
											},
										},
									},
								},
								"sse_specification": schema.ListNestedBlock{
									Description: "Represents the settings used to enable server-side encryption",
									Validators: []validator.List{
										listvalidator.SizeAtMost(1),
									},
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"enabled": schema.BoolAttribute{
												Description: "Indicates whether server-side encryption is done using an Amazon Web Services managed key or an Amazon Web Services owned key.",
												Optional:    true,
												Computed:    true,
											},
											"kms_master_key_id": schema.StringAttribute{
												Description: "The KMS key that should be used for the KMS encryption.",
												Optional:    true,
											},
											"type": schema.StringAttribute{
												Description: "Server-side encryption type",
												Optional:    true,
												Validators: []validator.String{
													enum.FrameworkValidate[awstypes.SSEType](),
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

	s.Blocks["timeouts"] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
		Update: true,
		Delete: true,
	})

	resp.Schema = s
}

func (r *resourceImportTable) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DynamoDBClient(ctx)

	var plan resourceImportTableData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bs := make([]s3BucketSource, 1)
	resp.Diagnostics.Append(plan.S3BucketSource.ElementsAs(ctx, &bs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucketSource, diagsErr := expandS3BucketSource(ctx, bs)
	if diagsErr.HasError() {
		resp.Diagnostics.Append(diagsErr...)
		return
	}

	tcp := make([]tableCreationParameters, 1)
	resp.Diagnostics.Append(plan.TableCreationParameters.ElementsAs(ctx, &tcp, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tableCreationParams, diagsErr := expandTableCreationParameters(ctx, tcp)
	if diagsErr.HasError() {
		resp.Diagnostics.Append(diagsErr...)
		return
	}

	in := &dynamodb.ImportTableInput{
		ClientToken:             aws.String(id.UniqueId()),
		S3BucketSource:          bucketSource,
		TableCreationParameters: tableCreationParams,
	}

	resp.Diagnostics.Append(flex.Expand(ctx, &plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.ImportTable(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DynamoDB, create.ErrActionCreating, ResNameImportTable, plan.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	if out == nil || out.ImportTableDescription == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DynamoDB, create.ErrActionCreating, ResNameImportTable, plan.ID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	table := out.ImportTableDescription
	plan.ARN = flex.StringToFramework(ctx, table.ImportArn)
	plan.ID = flex.StringToFramework(ctx, table.ImportArn)
	plan.TableID = flex.StringToFramework(ctx, table.TableId)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	outRaw, err := waitImportTableCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DynamoDB, create.ErrActionWaitingForCreation, ResNameImportTable, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	//set state for unknowns
	resp.Diagnostics.Append(flex.Flatten(ctx, outRaw, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceImportTable) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DynamoDBClient(ctx)

	var state resourceImportTableData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findImportTableByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DynamoDB, create.ErrActionSetting, ResNameImportTable, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceImportTable) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//conn := r.Meta().DynamoDBClient(ctx)
	//
	var plan, state resourceImportTableData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	//
	//if !plan.Name.Equal(state.Name) ||
	//	!plan.Description.Equal(state.Description) ||
	//	!plan.ComplexArgument.Equal(state.ComplexArgument) ||
	//	!plan.Type.Equal(state.Type) {
	//
	//	in := &dynamodb.UpdateImportTableInput{
	//		ImportTableId:   aws.String(plan.ID.ValueString()),
	//		ImportTableName: aws.String(plan.Name.ValueString()),
	//		ImportTableType: aws.String(plan.Type.ValueString()),
	//	}
	//
	//	if !plan.Description.IsNull() {
	//		in.Description = aws.String(plan.Description.ValueString())
	//	}
	//	if !plan.ComplexArgument.IsNull() {
	//		var tfList []complexArgumentData
	//		resp.Diagnostics.Append(plan.ComplexArgument.ElementsAs(ctx, &tfList, false)...)
	//		if resp.Diagnostics.HasError() {
	//			return
	//		}
	//
	//		in.ComplexArgument = expandComplexArgument(tfList)
	//	}
	//
	//	out, err := conn.UpdateImportTableWithContext(ctx, in)
	//	if err != nil {
	//		resp.Diagnostics.AddError(
	//			create.ProblemStandardMessage(names.DynamoDB, create.ErrActionUpdating, ResNameImportTable, plan.ID.String(), err),
	//			err.Error(),
	//		)
	//		return
	//	}
	//	if out == nil || out.ImportTable == nil {
	//		resp.Diagnostics.AddError(
	//			create.ProblemStandardMessage(names.DynamoDB, create.ErrActionUpdating, ResNameImportTable, plan.ID.String(), nil),
	//			errors.New("empty output").Error(),
	//		)
	//		return
	//	}
	//
	//	plan.ARN = flex.StringToFramework(ctx, out.ImportTable.Arn)
	//	plan.ID = flex.StringToFramework(ctx, out.ImportTable.ImportTableId)
	//}
	//
	//updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	//_, err := waitImportTableUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	//if err != nil {
	//	resp.Diagnostics.AddError(
	//		create.ProblemStandardMessage(names.DynamoDB, create.ErrActionWaitingForUpdate, ResNameImportTable, plan.ID.String(), err),
	//		err.Error(),
	//	)
	//	return
	//}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceImportTable) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DynamoDBClient(ctx)

	var state resourceImportTableData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tcp []tableCreationParameters
	resp.Diagnostics.Append(state.TableCreationParameters.ElementsAs(ctx, &tcp, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tableCreationParams, diagsErr := expandTableCreationParameters(ctx, tcp)
	if diagsErr.HasError() {
		resp.Diagnostics.Append(diagsErr...)
		return
	}

	in := &dynamodb.DeleteTableInput{
		TableName: tableCreationParams.TableName,
	}

	_, err := conn.DeleteTable(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DynamoDB, create.ErrActionDeleting, ResNameImportTable, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitImportTableDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DynamoDB, create.ErrActionWaitingForDeletion, ResNameImportTable, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

//func (r *resourceImportTable) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
//	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
//}

func waitImportTableCreated(ctx context.Context, conn *dynamodb.Client, id string, timeout time.Duration) (*awstypes.ImportTableDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ImportStatusInProgress),
		Target:                    enum.Slice(awstypes.ImportStatusCompleted),
		Refresh:                   statusImportTable(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ImportTableDescription); ok {
		return out, err
	}

	return nil, err
}

//func waitImportTableUpdated(ctx context.Context, conn *dynamodb.DynamoDB, id string, timeout time.Duration) (*dynamodb.ImportTable, error) {
//	stateConf := &retry.StateChangeConf{
//		Pending:                   []string{statusChangePending},
//		Target:                    []string{statusUpdated},
//		Refresh:                   statusImportTable(ctx, conn, id),
//		Timeout:                   timeout,
//		NotFoundChecks:            20,
//		ContinuousTargetOccurence: 2,
//	}
//
//	outputRaw, err := stateConf.WaitForStateContext(ctx)
//	if out, ok := outputRaw.(*dynamodb.ImportTableDescription); ok {
//		return out, err
//	}
//
//	return nil, err
//}

func waitImportTableDeleted(ctx context.Context, conn *dynamodb.Client, id string, timeout time.Duration) (*awstypes.ImportTableDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{},
		Refresh: statusImportTable(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ImportTableDescription); ok {
		return out, err
	}

	return nil, err
}

func statusImportTable(ctx context.Context, conn *dynamodb.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findImportTableByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.ImportStatus), nil
	}
}

func findImportTableByID(ctx context.Context, conn *dynamodb.Client, id string) (*awstypes.ImportTableDescription, error) {
	in := &dynamodb.DescribeImportInput{
		ImportArn: aws.String(id),
	}

	out, err := conn.DescribeImport(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.ImportTableDescription == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.ImportTableDescription, nil
}

func expandInputFormatOptions(ctx context.Context, tfList []inputFormatOptions) (*awstypes.InputFormatOptions, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}

	ifo := tfList[0]

	var csv []csvOptions
	err := ifo.CSV.ElementsAs(ctx, &csv, false)
	if err.HasError() {
		diags.Append(err...)
		return nil, diags
	}

	out := &awstypes.InputFormatOptions{
		Csv: expandCSVOptions(ctx, csv),
	}

	return out, nil
}

func expandCSVOptions(ctx context.Context, tfList []csvOptions) *awstypes.CsvOptions {
	if len(tfList) == 0 {
		return nil
	}

	options := tfList[0]

	out := &awstypes.CsvOptions{
		Delimiter:  flex.StringFromFramework(ctx, options.Delimiter),
		HeaderList: flex.ExpandFrameworkStringValueSet(ctx, options.HeaderList),
	}
	return out
}

func expandS3BucketSource(ctx context.Context, tfList []s3BucketSource) (*awstypes.S3BucketSource, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}

	bs := tfList[0]

	out := &awstypes.S3BucketSource{
		S3Bucket: flex.StringFromFramework(ctx, bs.Bucket),
	}

	if !bs.BucketOwner.IsNull() {
		out.S3BucketOwner = flex.StringFromFramework(ctx, bs.BucketOwner)
	}

	if !bs.Prefix.IsNull() {
		out.S3KeyPrefix = flex.StringFromFramework(ctx, bs.Prefix)
	}

	return out, nil
}

func expandTableCreationParameters(ctx context.Context, tfList []tableCreationParameters) (*awstypes.TableCreationParameters, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}

	tcp := tfList[0]

	out := &awstypes.TableCreationParameters{
		BillingMode: flex.StringTypeFromFramework[awstypes.BillingMode](ctx, tcp.BillingMode),
		TableName:   flex.StringFromFramework(ctx, tcp.Name),
	}

	var attrDefinitions []attributeDefinition
	err := tcp.AttributeDefinition.ElementsAs(ctx, &attrDefinitions, false)
	if err.HasError() {
		diags.Append(err...)
		return nil, diags
	}

	ad, err := expandAttributeDefinitions(ctx, attrDefinitions)
	if err.HasError() {
		diags.Append(err...)
		return nil, diags
	}
	out.AttributeDefinitions = ad

	var keySchemas []keySchema
	err = tcp.KeySchema.ElementsAs(ctx, &keySchemas, false)
	if err.HasError() {
		diags.Append(err...)
		return nil, diags
	}

	ks, err := expandImportKeySchema(ctx, keySchemas)
	if err.HasError() {
		diags.Append(err...)
		return nil, diags
	}
	out.KeySchema = ks

	return out, nil
}

func expandAttributeDefinitions(ctx context.Context, tfList []attributeDefinition) ([]awstypes.AttributeDefinition, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}

	var out []awstypes.AttributeDefinition
	for _, v := range tfList {
		d := awstypes.AttributeDefinition{
			AttributeName: flex.StringFromFramework(ctx, v.Name),
			AttributeType: flex.StringTypeFromFramework[awstypes.ScalarAttributeType](ctx, v.Type),
		}

		out = append(out, d)
	}

	return out, nil
}

func expandImportKeySchema(ctx context.Context, tfList []keySchema) ([]awstypes.KeySchemaElement, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}

	var out []awstypes.KeySchemaElement
	for _, v := range tfList {
		d := awstypes.KeySchemaElement{
			AttributeName: flex.StringFromFramework(ctx, v.AttributeName),
			KeyType:       flex.StringTypeFromFramework[awstypes.KeyType](ctx, v.Type),
		}

		out = append(out, d)
	}

	return out, nil
}

func flattenS3BucketSource(ctx context.Context, options *awstypes.S3BucketSource) types.List {
	elemType := types.ObjectType{AttrTypes: s3BucketSourceAttrTypes}

	if options == nil {
		return types.ListNull(elemType)
	}

	attrs := map[string]attr.Value{}
	attrs["bucket"] = flex.StringToFramework(ctx, options.S3Bucket)
	attrs["bucket_owner"] = flex.StringToFramework(ctx, options.S3BucketOwner)
	attrs["prefix"] = flex.StringToFramework(ctx, options.S3KeyPrefix)

	values := types.ObjectValueMust(s3BucketSourceAttrTypes, attrs)

	return types.ListValueMust(elemType, []attr.Value{values})
}

func flattenInputFormatOptions(ctx context.Context, options *awstypes.InputFormatOptions) types.List {
	elemType := types.ObjectType{AttrTypes: inputFormatOptionsAttrTypes}

	if options == nil {
		return types.ListNull(elemType)
	}

	attrs := map[string]attr.Value{}
	attrs["csv"] = flattenCSVOptions(ctx, options.Csv)

	values := types.ObjectValueMust(inputFormatOptionsAttrTypes, attrs)

	return types.ListValueMust(elemType, []attr.Value{values})
}

func flattenCSVOptions(ctx context.Context, options *awstypes.CsvOptions) types.List {
	elemType := types.ObjectType{AttrTypes: csvOptionsAttrTypes}

	if options == nil {
		return types.ListNull(elemType)
	}

	attrs := map[string]attr.Value{}
	attrs["delimiter"] = flex.StringToFramework(ctx, options.Delimiter)
	attrs["header_list"] = flex.FlattenFrameworkStringValueSet(ctx, options.HeaderList)

	values := types.ObjectValueMust(csvOptionsAttrTypes, attrs)

	return types.ListValueMust(elemType, []attr.Value{values})
}

type resourceImportTableData struct {
	ARN                     types.String   `tfsdk:"arn"`
	ID                      types.String   `tfsdk:"id"`
	InputCompressionType    types.String   `tfsdk:"input_compression_type"`
	InputFormat             types.String   `tfsdk:"input_format"`
	InputFormatOptions      types.List     `tfsdk:"input_format_options"`
	S3BucketSource          types.List     `tfsdk:"s3_bucket_source"`
	TableCreationParameters types.List     `tfsdk:"table_creation_parameters"`
	TableID                 types.String   `tfsdk:"table_id"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
}

type inputFormatOptions struct {
	CSV types.List `tfsdk:"csv"`
}

type csvOptions struct {
	Delimiter  types.String `tfsdk:"delimiter"`
	HeaderList types.Set    `tfsdk:"header_list"`
}

type s3BucketSource struct {
	Bucket      types.String `tfsdk:"bucket"`
	BucketOwner types.String `tfsdk:"bucket_owner"`
	Prefix      types.String `tfsdk:"prefix"`
}

type tableCreationParameters struct {
	BillingMode          types.String `tfsdk:"billing_mode"`
	Name                 types.String `tfsdk:"name"`
	AttributeDefinition  types.Set    `tfsdk:"attribute_definition"`
	KeySchema            types.Set    `tfsdk:"key_schema"`
	GlobalSecondaryIndex types.Set    `tfsdk:"global_secondary_index"`
}

type attributeDefinition struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

type keySchema struct {
	AttributeName types.String `tfsdk:"attribute_name"`
	Type          types.String `tfsdk:"type"`
}

var inputFormatOptionsAttrTypes = map[string]attr.Type{
	"csv": types.ListType{ElemType: types.ObjectType{AttrTypes: csvOptionsAttrTypes}},
}

var csvOptionsAttrTypes = map[string]attr.Type{
	"delimiter":   types.StringType,
	"header_list": types.SetType{ElemType: types.StringType},
}

var s3BucketSourceAttrTypes = map[string]attr.Type{
	"bucket":       types.StringType,
	"bucket_owner": types.StringType,
	"prefix":       types.StringType,
}

var tableCreationParametersAttrTypes = map[string]attr.Type{
	"billing_mode":         types.StringType,
	"name":                 types.StringType,
	"attribute_definition": types.SetType{ElemType: types.ObjectType{AttrTypes: attributeDefinitionAttrTypes}},
	"key_schema":           types.SetType{ElemType: types.ObjectType{AttrTypes: keySchemaAttrTypes}},
}

var attributeDefinitionAttrTypes = map[string]attr.Type{
	"name": types.StringType,
	"type": types.StringType,
}

var keySchemaAttrTypes = map[string]attr.Type{
	"attribute_name": types.StringType,
	"type":           types.StringType,
}
