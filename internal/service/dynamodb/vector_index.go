// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_dynamodb_vector_index", name="Vector Index")
// @IdentityAttribute("table_name")
// @IdentityAttribute("index_name")
// @ImportIDHandler("vectorIndexImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/dynamodb/types;awstypes;awstypes.VectorIndexDescription")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc=testAccVectorIndexImportStateIdFunc)
// @Testing(importStateIdAttribute="arn")
func newVectorIndexResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &vectorIndexResource{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type vectorIndexResource struct {
	framework.ResourceWithModel[vectorIndexResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
	framework.WithNoUpdate
}

func (r *vectorIndexResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"index_name": schema.StringAttribute{
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
			"dimensions": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"distance_function": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.VectorDistanceFunction](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// TODO: Once Protocol v6 is supported, convert this to a `schema.SingleNestedAttribute` with full schema information
			"vector_attribute": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[vectorAttributeModel](ctx),
				Required:   true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"projection": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[projectionModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"non_key_attributes": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Optional:    true,
						},
						"projection_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ProjectionType](),
							Required:   true,
						},
					},
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"search_schema": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[searchSchemaElementModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"attribute_name": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"attribute_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ScalarAttributeType](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.SearchSchemaElementType](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}

	response.Schema = s
}

func (r *vectorIndexResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data vectorIndexResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	conn := r.Meta().DynamoDBClient(ctx)

	table, err := waitAllVectorIndexesActive(ctx, conn, data.TableName.ValueString(), createTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for all vector indexes on table "%s" to be active`, data.TableName.ValueString()),
			err.Error(),
		)

		return
	}

	input := dynamodb.UpdateTableInput{
		TableName:            data.TableName.ValueStringPointer(),
		AttributeDefinitions: []awstypes.AttributeDefinition{},
	}

	knownAttributes := map[string]awstypes.ScalarAttributeType{}

	for _, ad := range table.AttributeDefinitions {
		input.AttributeDefinitions = append(input.AttributeDefinitions, ad)
		knownAttributes[aws.ToString(ad.AttributeName)] = ad.AttributeType
	}

	var sses []searchSchemaElementModel
	response.Diagnostics.Append(data.SearchSchema.ElementsAs(ctx, &sses, false)...)
	if response.Diagnostics.HasError() {
		return
	}
	for _, sse := range sses {
		attributeName := sse.AttributeName.ValueString()
		if _, exists := knownAttributes[attributeName]; exists {
			continue
		}

		attributeType := sse.AttributeType.ValueEnum()
		input.AttributeDefinitions = append(input.AttributeDefinitions, awstypes.AttributeDefinition{
			AttributeName: sse.AttributeName.ValueStringPointer(),
			AttributeType: attributeType,
		})
		knownAttributes[attributeName] = attributeType
	}

	var action awstypes.CreateVectorIndexAction
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &action)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.VectorIndexUpdates = []awstypes.VectorIndexUpdate{
		{
			Create: &action,
		},
	}

	_, err = conn.UpdateTable(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrTableName, data.TableName.ValueString(), "index_name", data.IndexName.ValueString())
		return
	}

	if table, err = waitTableActive(ctx, conn, data.TableName.ValueString(), createTimeout); err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for table "%s" to be active`, data.TableName.ValueString()),
			err.Error(),
		)

		return
	}

	index, err := waitVectorIndexActive(ctx, conn, data.TableName.ValueString(), data.IndexName.ValueString(), createTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for vector index "%s" on table "%s" to be active`, data.IndexName.ValueString(), data.TableName.ValueString()),
			err.Error(),
		)

		return
	}

	response.Diagnostics.Append(flattenVectorIndex(ctx, &data, index, table)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vectorIndexResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data vectorIndexResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	tableName := data.TableName.ValueString()
	indexName := data.IndexName.ValueString()

	conn := r.Meta().DynamoDBClient(ctx)

	table, err := findTableByName(ctx, conn, tableName)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrTableName, tableName, "index_name", indexName)
		return
	}

	index, err := findVectorIndexFromTable(table, indexName)
	if err != nil || index == nil {
		response.Diagnostics.Append(
			fwdiag.NewResourceNotFoundWarningDiagnostic(
				fmt.Errorf(`unable to find vector index with name "%s" on table "%s", treating it as new`, indexName, tableName),
			),
		)
		response.State.RemoveResource(ctx)

		return
	}

	response.Diagnostics.Append(flattenVectorIndex(ctx, &data, index, table)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// Update is never invoked: every attribute forces replacement, since the underlying
// UpdateTable VectorIndexUpdates API only supports Create/Delete, not an in-place
// update of an existing vector index's configuration. framework.WithNoUpdate handles
// the case with an error diagnostic rather than silently persisting the plan.

func (r *vectorIndexResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data vectorIndexResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	deleteTimeout := r.DeleteTimeout(ctx, data.Timeouts)
	conn := r.Meta().DynamoDBClient(ctx)

	table, err := waitAllVectorIndexesActive(ctx, conn, data.TableName.ValueString(), deleteTimeout)
	if retry.NotFound(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrTableName, data.TableName.ValueString(), "index_name", data.IndexName.ValueString())
		return
	}

	// If owning table is already deleting, exit
	if table != nil && table.TableStatus == awstypes.TableStatusDeleting {
		return
	}

	input := dynamodb.UpdateTableInput{
		TableName: data.TableName.ValueStringPointer(),
		VectorIndexUpdates: []awstypes.VectorIndexUpdate{
			{
				Delete: &awstypes.DeleteVectorIndexAction{
					IndexName: data.IndexName.ValueStringPointer(),
				},
			},
		},
	}

	if res, err := conn.UpdateTable(ctx, &input); err != nil {
		// exit if owning table is already in deleting state
		if res != nil && res.TableDescription != nil && res.TableDescription.TableStatus == awstypes.TableStatusDeleting {
			return
		}

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		// exit if error says the table is being deleted
		if err, ok := errors.AsType[*awstypes.ResourceInUseException](err); ok && err != nil && strings.Contains(err.Error(), "Table is being deleted") {
			return
		}

		response.Diagnostics.AddError(
			fmt.Sprintf(`Unable to delete vector index "%s" on table "%s"`, data.IndexName.ValueString(), data.TableName.ValueString()),
			err.Error(),
		)

		return
	}

	if _, err := waitVectorIndexDeleted(ctx, conn, data.TableName.ValueString(), data.IndexName.ValueString(), deleteTimeout); err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for vector index "%s" on table "%s" to be deleted`, data.IndexName.ValueString(), data.TableName.ValueString()),
			err.Error(),
		)
	}

	if _, err := waitTableActive(ctx, conn, data.TableName.ValueString(), deleteTimeout); err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for "%s" to be active after delete`, data.TableName.ValueString()),
			err.Error(),
		)
	}
}

func flattenVectorIndex(ctx context.Context, data *vectorIndexResourceModel, index *awstypes.VectorIndexDescription, table *awstypes.TableDescription) diag.Diagnostics { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, index, data,
		fwflex.WithFieldNamePrefix("Index"),
		fwflex.WithIgnoredFieldNamesAppend("SearchSchema"),
	)...)
	if diags.HasError() {
		return diags
	}

	data.TableName = fwflex.StringToFramework(ctx, table.TableName)

	attributeTypes := make(map[string]awstypes.ScalarAttributeType, len(table.AttributeDefinitions))
	for _, attribute := range table.AttributeDefinitions {
		attributeTypes[aws.ToString(attribute.AttributeName)] = attribute.AttributeType
	}

	elements := make([]searchSchemaElementModel, len(index.SearchSchema))
	for i, sse := range index.SearchSchema {
		elements[i] = searchSchemaElementModel{
			AttributeName:           fwflex.StringToFramework(ctx, sse.AttributeName),
			AttributeType:           fwtypes.StringEnumValue(attributeTypes[aws.ToString(sse.AttributeName)]),
			SearchSchemaElementType: fwtypes.StringEnumValue(sse.SearchSchemaElementType),
		}
	}
	data.SearchSchema = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, elements)

	return diags
}

type vectorIndexResourceModel struct {
	framework.WithRegionModel

	ARN              types.String                                              `tfsdk:"arn"`
	Dimensions       types.Int64                                               `tfsdk:"dimensions"`
	DistanceFunction fwtypes.StringEnum[awstypes.VectorDistanceFunction]       `tfsdk:"distance_function"`
	IndexName        types.String                                              `tfsdk:"index_name"`
	Projection       fwtypes.ListNestedObjectValueOf[projectionModel]          `tfsdk:"projection"`
	SearchSchema     fwtypes.ListNestedObjectValueOf[searchSchemaElementModel] `tfsdk:"search_schema"`
	TableName        types.String                                              `tfsdk:"table_name"`
	VectorAttribute  fwtypes.ObjectValueOf[vectorAttributeModel]               `tfsdk:"vector_attribute"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type vectorAttributeModel struct {
	AttributeName types.String `tfsdk:"attribute_name"`
}

type searchSchemaElementModel struct {
	AttributeName           types.String                                         `tfsdk:"attribute_name"`
	AttributeType           fwtypes.StringEnum[awstypes.ScalarAttributeType]     `tfsdk:"attribute_type"`
	SearchSchemaElementType fwtypes.StringEnum[awstypes.SearchSchemaElementType] `tfsdk:"type"`
}

var (
	_ inttypes.ImportIDParser = vectorIndexImportID{}
)

type vectorIndexImportID struct{}

func (vectorIndexImportID) Parse(id string) (string, map[string]any, error) {
	tableName, indexName, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("Import ID \"%s\" should be in the format <table-name>"+intflex.ResourceIdSeparator+"<index-name>", id)
	}

	result := map[string]any{
		names.AttrTableName: tableName,
		"index_name":        indexName,
	}

	return id, result, nil
}
