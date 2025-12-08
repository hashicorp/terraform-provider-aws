// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	warmThoughtPutMinReadUnitsPerSecond  = 12000
	warmThoughtPutMinWriteUnitsPerSecond = 4000

	minNumberOfHashes = 1
	maxNumberOfHashes = 4

	minNumberOfRanges = 4
	maxNumberOfRanges = 4
)

// @FrameworkResource("aws_dynamodb_global_secondary_index", name="Global Secondary Index")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/dynamodb/types;types.TableDescription")
func newResourceGlobalSecondaryIndex(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceGlobalSecondaryIndex{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(60 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type resourceGlobalSecondaryIndex struct {
	framework.ResourceWithModel[resourceGlobalSecondaryIndexModel]
	framework.WithTimeouts
}

func (r *resourceGlobalSecondaryIndex) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"index_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"non_key_attributes": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
					setplanmodifier.RequiresReplace(),
				},
			},
			"projection_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ProjectionType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"read_capacity": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			names.AttrTableName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"write_capacity": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"key_schema": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[keySchemaModel](ctx),
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
						"key_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.KeyType](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 8),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"on_demand_throughput": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"max_read_request_units": schema.Int64Attribute{
							Optional: true,
							Computed: true,
						},
						"max_write_request_units": schema.Int64Attribute{
							Optional: true,
							Computed: true,
						},
					},
				},
				CustomType: fwtypes.NewListNestedObjectTypeOf[onDemandThroughputModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"warm_throughput": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[warmThroughputModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"read_units_per_second": schema.Int64Attribute{
							Optional: true,
							Computed: true,
						},
						"write_units_per_second": schema.Int64Attribute{
							Optional: true,
							Computed: true,
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Version: 0,
	}

	response.Schema = s
}

func (r *resourceGlobalSecondaryIndex) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), request, response)
}

func (r *resourceGlobalSecondaryIndex) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceGlobalSecondaryIndexModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	conn := r.Meta().DynamoDBClient(ctx)

	table, err := waitAllGSIActive(ctx, conn, data.TableName.ValueString(), createTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for all GSIs on table "%s" to be active`, data.TableName.ValueString()),
			err.Error(),
		)

		return
	}

	knownAttributes := map[string]awstypes.ScalarAttributeType{}
	ut := &dynamodb.UpdateTableInput{
		TableName:            data.TableName.ValueStringPointer(),
		AttributeDefinitions: []awstypes.AttributeDefinition{},
	}

	for _, ad := range table.AttributeDefinitions {
		ut.AttributeDefinitions = append(ut.AttributeDefinitions, ad)
		knownAttributes[aws.ToString(ad.AttributeName)] = ad.AttributeType
	}

	var ksms []keySchemaModel
	var keySchema []awstypes.KeySchemaElement
	response.Diagnostics.Append(data.KeySchema.ElementsAs(ctx, &ksms, false)...)
	if response.Diagnostics.HasError() {
		return
	}
	for _, ks := range ksms {
		keySchema = append(keySchema, awstypes.KeySchemaElement{
			AttributeName: ks.AttributeName.ValueStringPointer(),
			KeyType:       ks.KeyType.ValueEnum(),
		})

		typ, exists := knownAttributes[ks.AttributeName.ValueString()]
		if exists && typ == ks.AttributeType.ValueEnum() {
			continue
		}

		ut.AttributeDefinitions = append(ut.AttributeDefinitions, awstypes.AttributeDefinition{
			AttributeName: ks.AttributeName.ValueStringPointer(),
			AttributeType: ks.AttributeType.ValueEnum(),
		})
	}

	projection := &awstypes.Projection{
		ProjectionType: data.ProjectionType.ValueEnum(),
	}

	if !data.NonKeyAttributes.IsNull() && !data.NonKeyAttributes.IsUnknown() {
		response.Diagnostics.Append(
			data.NonKeyAttributes.ElementsAs(ctx, &projection.NonKeyAttributes, false)...,
		)
	}

	action := &awstypes.CreateGlobalSecondaryIndexAction{
		IndexName:  data.IndexName.ValueStringPointer(),
		KeySchema:  keySchema,
		Projection: projection,
	}

	if data.ReadCapacity.IsNull() || data.ReadCapacity.IsUnknown() {
		data.ReadCapacity = types.Int64Value(0)
	}

	if data.WriteCapacity.IsNull() || data.WriteCapacity.IsUnknown() {
		data.WriteCapacity = types.Int64Value(0)
	}

	billingMode := awstypes.BillingModeProvisioned
	if table.BillingModeSummary != nil {
		billingMode = table.BillingModeSummary.BillingMode
	}

	if billingMode == awstypes.BillingModeProvisioned {
		rc := int(data.ReadCapacity.ValueInt64())
		wc := int(data.WriteCapacity.ValueInt64())

		if rc == 0 && table.ProvisionedThroughput != nil {
			rc = int(aws.ToInt64(table.ProvisionedThroughput.ReadCapacityUnits))
		}
		if wc == 0 && table.ProvisionedThroughput != nil {
			wc = int(aws.ToInt64(table.ProvisionedThroughput.WriteCapacityUnits))
		}

		provisionedThroughputData := map[string]any{
			"read_capacity":  rc,
			"write_capacity": wc,
		}
		action.ProvisionedThroughput = expandProvisionedThroughput(provisionedThroughputData, billingMode)
	} else if !data.OnDemandThroughputs.IsNull() && !data.OnDemandThroughputs.IsUnknown() {
		var odts []onDemandThroughputModel
		response.Diagnostics.Append(data.OnDemandThroughputs.ElementsAs(ctx, &odts, false)...)
		if len(odts) > 0 {
			v := map[string]any{
				"max_read_request_units":  int(odts[0].MaxReadRequestUnits.ValueInt64()),
				"max_write_request_units": int(odts[0].MaxWriteRequestUnits.ValueInt64()),
			}
			action.OnDemandThroughput = expandOnDemandThroughput(v)
		}
	}

	var wts []warmThroughputModel
	response.Diagnostics.Append(data.WarmThroughputs.ElementsAs(ctx, &wts, false)...)
	if len(wts) > 0 {
		action.WarmThroughput = expandWarmThroughput(map[string]any{
			"read_units_per_second":  int(wts[0].ReadUnitsPerSecond.ValueInt64()),
			"write_units_per_second": int(wts[0].WriteUnitsPerSecond.ValueInt64()),
		})
	}

	ut.GlobalSecondaryIndexUpdates = []awstypes.GlobalSecondaryIndexUpdate{
		{
			Create: action,
		},
	}

	response.Diagnostics.Append(validateNewGSIAttributes(ctx, data, *table)...)
	if response.Diagnostics.HasError() {
		return
	}

	if utRes, err := conn.UpdateTable(ctx, ut); err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Unable to create index "%s" on table "%s"`, data.IndexName.ValueString(), data.TableName.ValueString()),
			err.Error(),
		)

		return
	} else {
		g, err := findGSIFromTable(utRes.TableDescription, data.IndexName.ValueString())
		if err != nil || g == nil {
			response.Diagnostics.AddError(
				"Unable to find remote GSI after create",
				fmt.Sprintf(
					`GSI with name "%s" (arn: "%s") was not found in table "%s"`,
					data.IndexName.ValueString(),
					data.ARN.ValueString(),
					data.TableName.ValueString(),
				),
			)

			return
		}

		response.Diagnostics.Append(flattenGlobalSecondaryIndex(ctx, &data, *g, utRes.TableDescription)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	if _, err = waitTableActive(ctx, conn, data.TableName.ValueString(), createTimeout); err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for table "%s" to be active`, data.TableName.ValueString()),
			err.Error(),
		)

		return
	}

	if _, err := waitGSIActive(ctx, conn, data.TableName.ValueString(), data.IndexName.ValueString(), createTimeout); err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for GSI "%s" on table "%s" to be active`, data.IndexName.ValueString(), data.TableName.ValueString()),
			err.Error(),
		)

		return
	}

	if err = waitGSIWarmThroughputActive(ctx, conn, data.TableName.ValueString(), data.IndexName.ValueString(), createTimeout); err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for warm throughput on GSI "%s" on table "%s" to be active`, data.IndexName.ValueString(), data.TableName.ValueString()),
			err.Error(),
		)

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceGlobalSecondaryIndex) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceGlobalSecondaryIndexModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	re := regexache.MustCompile(":table/([^\\/]+)/index/(.+)")
	grps := re.FindStringSubmatch(data.ARN.ValueString())
	var tableName string
	var indexName string
	if len(grps) == 3 {
		tableName = grps[1]
		indexName = grps[2]
	} else {
		tableName = data.TableName.ValueString()
		indexName = data.IndexName.ValueString()
	}

	conn := r.Meta().DynamoDBClient(ctx)

	table, err := findTableByName(ctx, conn, tableName)
	if err != nil {
		if retry.NotFound(err) {
			response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
			response.State.RemoveResource(ctx)
		} else {
			response.Diagnostics.AddError(
				fmt.Sprintf(`Unable to read table "%s"`, tableName),
				err.Error(),
			)
		}

		return
	}

	g, err := findGSIFromTable(table, indexName)
	if err != nil || g == nil {
		response.Diagnostics.Append(
			fwdiag.NewResourceNotFoundWarningDiagnostic(
				fmt.Errorf(`unable to find global secondary index with arn "%s", treating it as new`, data.ARN.ValueString()),
			),
		)
		response.State.RemoveResource(ctx)

		return
	}

	response.Diagnostics.Append(flattenGlobalSecondaryIndex(ctx, &data, *g, table)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceGlobalSecondaryIndex) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new resourceGlobalSecondaryIndexModel

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)

	if response.Diagnostics.HasError() {
		return
	}

	updateTimeout := r.UpdateTimeout(ctx, new.Timeouts)
	conn := r.Meta().DynamoDBClient(ctx)

	table, err := waitAllGSIActive(ctx, conn, new.TableName.ValueString(), updateTimeout)
	if err != nil {
		if retry.NotFound(err) {
			response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
			response.State.RemoveResource(ctx)
		} else {
			response.Diagnostics.AddError(
				fmt.Sprintf(`Unable to read table "%s"`, new.TableName.ValueString()),
				err.Error(),
			)
		}

		return
	}

	action := &awstypes.UpdateGlobalSecondaryIndexAction{
		IndexName: new.IndexName.ValueStringPointer(),
	}

	billingMode := awstypes.BillingModeProvisioned
	if table.BillingModeSummary != nil {
		billingMode = table.BillingModeSummary.BillingMode
	}

	hasUpdate := false

	if billingMode == awstypes.BillingModeProvisioned {
		provisionedThroughputData := map[string]any{
			"read_capacity":  int(new.ReadCapacity.ValueInt64()),
			"write_capacity": int(new.WriteCapacity.ValueInt64()),
		}
		newProvisionedThroughput := expandProvisionedThroughput(provisionedThroughputData, billingMode)

		g, err := findGSIFromTable(table, old.IndexName.ValueString())
		if err != nil || g == nil {
			response.Diagnostics.AddError(
				"Unable to find remote GSI to update",
				fmt.Sprintf(
					`GSI with name "%s" (arn: "%s") was not found in table "%s"`,
					new.IndexName.ValueString(),
					new.ARN.ValueString(),
					new.TableName.ValueString(),
				),
			)

			return
		}

		if g.ProvisionedThroughput == nil {
			action.ProvisionedThroughput = newProvisionedThroughput
			hasUpdate = true
		} else if aws.ToInt64(g.ProvisionedThroughput.ReadCapacityUnits) != aws.ToInt64(newProvisionedThroughput.ReadCapacityUnits) || aws.ToInt64(g.ProvisionedThroughput.WriteCapacityUnits) != aws.ToInt64(newProvisionedThroughput.WriteCapacityUnits) {
			action.ProvisionedThroughput = newProvisionedThroughput
			hasUpdate = true
		}
	} else {
		var odts []onDemandThroughputModel
		response.Diagnostics.Append(new.OnDemandThroughputs.ElementsAs(ctx, &odts, false)...)

		if len(odts) > 0 {
			v := map[string]any{
				"max_read_request_units":  int(odts[0].MaxReadRequestUnits.ValueInt64()),
				"max_write_request_units": int(odts[0].MaxWriteRequestUnits.ValueInt64()),
			}
			action.OnDemandThroughput = expandOnDemandThroughput(v)
		}

		hasUpdate = !new.OnDemandThroughputs.Equal(old.OnDemandThroughputs)
	}

	var wts []warmThroughputModel
	response.Diagnostics.Append(new.WarmThroughputs.ElementsAs(ctx, &wts, false)...)
	if len(wts) > 0 {
		action.WarmThroughput = expandWarmThroughput(map[string]any{
			"read_units_per_second":  int(wts[0].ReadUnitsPerSecond.ValueInt64()),
			"write_units_per_second": int(wts[0].WriteUnitsPerSecond.ValueInt64()),
		})
		hasUpdate = hasUpdate || new.WarmThroughputs.Equal(old.WarmThroughputs)
	}

	ut := &dynamodb.UpdateTableInput{
		TableName: new.TableName.ValueStringPointer(),
		GlobalSecondaryIndexUpdates: []awstypes.GlobalSecondaryIndexUpdate{
			{
				Update: action,
			},
		},
	}

	response.Diagnostics.Append(validateNewGSIAttributes(ctx, new, *table)...)

	if hasUpdate && !response.Diagnostics.HasError() {
		if utRes, err := conn.UpdateTable(ctx, ut); err != nil {
			response.Diagnostics.AddError(
				fmt.Sprintf(`Unable to update index "%s" on table "%s"`, new.IndexName.ValueString(), new.TableName.ValueString()),
				err.Error(),
			)

			return
		} else {
			g, err := findGSIFromTable(utRes.TableDescription, new.IndexName.ValueString())
			if err != nil || g == nil {
				response.Diagnostics.AddError(
					"Unable to find remote GSI after update",
					fmt.Sprintf(
						`GSI with name "%s" (arn: "%s") was not found in table "%s"`,
						new.IndexName.ValueString(),
						new.ARN.ValueString(),
						new.TableName.ValueString(),
					),
				)

				return
			}

			response.Diagnostics.Append(flattenGlobalSecondaryIndex(ctx, &new, *g, utRes.TableDescription)...)
			if response.Diagnostics.HasError() {
				return
			}
		}

		if _, err = waitTableActive(ctx, conn, new.TableName.ValueString(), updateTimeout); err != nil {
			response.Diagnostics.AddError(
				fmt.Sprintf(`Error while waiting for table "%s" to be active`, new.TableName.ValueString()),
				err.Error(),
			)

			return
		}

		if _, err := waitGSIActive(ctx, conn, new.TableName.ValueString(), new.IndexName.ValueString(), updateTimeout); err != nil {
			response.Diagnostics.AddError(
				fmt.Sprintf(`Error while waiting for GSI "%s" on table "%s" to be active`, new.IndexName.ValueString(), new.TableName.ValueString()),
				err.Error(),
			)

			return
		}

		if err = waitGSIWarmThroughputActive(ctx, conn, new.TableName.ValueString(), new.IndexName.ValueString(), updateTimeout); err != nil {
			response.Diagnostics.AddError(
				fmt.Sprintf(`Error while waiting for warm throughput on GSI "%s" on table "%s" to be active`, new.IndexName.ValueString(), new.TableName.ValueString()),
				err.Error(),
			)

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *resourceGlobalSecondaryIndex) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceGlobalSecondaryIndexModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}
	deleteTimeout := r.DeleteTimeout(ctx, data.Timeouts)
	conn := r.Meta().DynamoDBClient(ctx)

	table, err := waitAllGSIActive(ctx, conn, data.TableName.ValueString(), deleteTimeout)
	// attempt early exit if owning table is already in deleting state
	if table != nil && table.TableStatus == awstypes.TableStatusDeleting {
		return
	}

	if err != nil {
		if retry.NotFound(err) {
			response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
			response.State.RemoveResource(ctx)
		} else {
			response.Diagnostics.AddError(
				fmt.Sprintf(`Unable to read table "%s"`, data.TableName.ValueString()),
				err.Error(),
			)
		}

		return
	}

	ut := &dynamodb.UpdateTableInput{
		TableName: data.TableName.ValueStringPointer(),
		GlobalSecondaryIndexUpdates: []awstypes.GlobalSecondaryIndexUpdate{
			{
				Delete: &awstypes.DeleteGlobalSecondaryIndexAction{
					IndexName: data.IndexName.ValueStringPointer(),
				},
			},
		},
	}

	if res, err := conn.UpdateTable(ctx, ut); err != nil {
		// exit if owning table is already in deleting state
		if res != nil && res.TableDescription != nil && res.TableDescription.TableStatus == awstypes.TableStatusDeleting {
			return
		}

		// exit if error says the table is being deleted
		if err, ok := errs.As[*awstypes.ResourceInUseException](err); ok && err != nil && strings.Contains(err.Error(), "Table is being deleted") {
			return
		}

		response.Diagnostics.AddError(
			fmt.Sprintf(`Unable to delete index "%s" on table "%s"`, data.IndexName.ValueString(), data.TableName.ValueString()),
			err.Error(),
		)

		return
	}

	if _, err := waitGSIDeleted(ctx, conn, data.TableName.ValueString(), data.IndexName.ValueString(), deleteTimeout); err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for GSI "%s" on table "%s" to be deleted`, data.IndexName.ValueString(), data.TableName.ValueString()),
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

type resourceGlobalSecondaryIndexModel struct {
	framework.WithRegionModel

	ARN                 types.String                                             `tfsdk:"arn"`
	IndexName           types.String                                             `tfsdk:"index_name"`
	KeySchema           fwtypes.ListNestedObjectValueOf[keySchemaModel]          `tfsdk:"key_schema"`
	NonKeyAttributes    fwtypes.SetOfString                                      `tfsdk:"non_key_attributes"`
	ProjectionType      fwtypes.StringEnum[awstypes.ProjectionType]              `tfsdk:"projection_type"`
	ReadCapacity        types.Int64                                              `tfsdk:"read_capacity"`
	TableName           types.String                                             `tfsdk:"table_name"`
	WriteCapacity       types.Int64                                              `tfsdk:"write_capacity"`
	OnDemandThroughputs fwtypes.ListNestedObjectValueOf[onDemandThroughputModel] `tfsdk:"on_demand_throughput"`
	WarmThroughputs     fwtypes.ListNestedObjectValueOf[warmThroughputModel]     `tfsdk:"warm_throughput"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func flattenGlobalSecondaryIndex(ctx context.Context, data *resourceGlobalSecondaryIndexModel, g awstypes.GlobalSecondaryIndexDescription, table *awstypes.TableDescription) diag.Diagnostics { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	if table == nil {
		diags.AddError(
			"Bad argument",
			"Table description is nil",
		)

		return diags
	}

	data.ARN = fwflex.StringToFramework(ctx, g.IndexArn)
	data.IndexName = fwflex.StringToFramework(ctx, g.IndexName)
	data.TableName = fwflex.StringToFramework(ctx, table.TableName)

	var kss []keySchemaModel
	adm := map[string]awstypes.ScalarAttributeType{}
	for _, ad := range table.AttributeDefinitions {
		adm[aws.ToString(ad.AttributeName)] = ad.AttributeType
	}

	for _, ks := range g.KeySchema {
		kss = append(kss, keySchemaModel{
			AttributeName: fwflex.StringToFramework(ctx, ks.AttributeName),
			AttributeType: fwtypes.StringEnumValue(adm[aws.ToString(ks.AttributeName)]),
			KeyType:       fwtypes.StringEnumValue(ks.KeyType),
		})
	}

	if len(kss) > 0 {
		data.KeySchema = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, kss)
	} else {
		data.KeySchema = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []keySchemaModel{})
	}

	var nkas []attr.Value
	if g.Projection != nil {
		data.ProjectionType = fwtypes.StringEnumValue(g.Projection.ProjectionType)
		for _, n := range g.Projection.NonKeyAttributes {
			nkas = append(nkas, types.StringValue(n))
		}
	}

	if len(nkas) > 0 {
		data.NonKeyAttributes = fwtypes.NewSetValueOfMust[basetypes.StringValue](ctx, nkas)
	} else {
		data.NonKeyAttributes = fwtypes.NewSetValueOfNull[basetypes.StringValue](ctx)
	}

	if g.ProvisionedThroughput != nil {
		data.ReadCapacity = types.Int64Value(aws.ToInt64(g.ProvisionedThroughput.ReadCapacityUnits))
		data.WriteCapacity = types.Int64Value(aws.ToInt64(g.ProvisionedThroughput.WriteCapacityUnits))
	} else {
		data.ReadCapacity = types.Int64Value(0)
		data.WriteCapacity = types.Int64Value(0)
	}

	if g.OnDemandThroughput != nil {
		var odtms []onDemandThroughputModel
		mrru := aws.ToInt64(g.OnDemandThroughput.MaxReadRequestUnits)
		mwru := aws.ToInt64(g.OnDemandThroughput.MaxWriteRequestUnits)

		if mrru > 0 || mwru > 0 {
			odtms = append(odtms, onDemandThroughputModel{
				MaxReadRequestUnits:  types.Int64Value(mrru),
				MaxWriteRequestUnits: types.Int64Value(mwru),
			})
		}

		data.OnDemandThroughputs = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, odtms)
	}

	if g.WarmThroughput != nil {
		rups := max(aws.ToInt64(g.WarmThroughput.ReadUnitsPerSecond), warmThoughtPutMinReadUnitsPerSecond)
		wups := max(aws.ToInt64(g.WarmThroughput.WriteUnitsPerSecond), warmThoughtPutMinWriteUnitsPerSecond)

		var wtms []warmThroughputModel
		if rups > warmThoughtPutMinReadUnitsPerSecond || wups > warmThoughtPutMinWriteUnitsPerSecond {
			wtms = append(wtms, warmThroughputModel{
				ReadUnitsPerSecond:  types.Int64Value(rups),
				WriteUnitsPerSecond: types.Int64Value(wups),
			})
		}

		data.WarmThroughputs = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, wtms)
	} else {
		data.WarmThroughputs = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []warmThroughputModel{})
	}

	return diags
}

type onDemandThroughputModel struct {
	MaxReadRequestUnits  types.Int64 `tfsdk:"max_read_request_units"`
	MaxWriteRequestUnits types.Int64 `tfsdk:"max_write_request_units"`
}

type warmThroughputModel struct {
	ReadUnitsPerSecond  types.Int64 `tfsdk:"read_units_per_second"`
	WriteUnitsPerSecond types.Int64 `tfsdk:"write_units_per_second"`
}

type keySchemaModel struct {
	AttributeName types.String                                     `tfsdk:"attribute_name"`
	AttributeType fwtypes.StringEnum[awstypes.ScalarAttributeType] `tfsdk:"attribute_type"`
	KeyType       fwtypes.StringEnum[awstypes.KeyType]             `tfsdk:"key_type"`
}

func validateNewGSIAttributes(ctx context.Context, data resourceGlobalSecondaryIndexModel, table awstypes.TableDescription) diag.Diagnostics {
	var diags diag.Diagnostics

	counts := map[string]int{}
	for _, ks := range table.KeySchema {
		counts[aws.ToString(ks.AttributeName)] = counts[aws.ToString(ks.AttributeName)] + 1
	}

	for _, l := range table.LocalSecondaryIndexes {
		for _, ks := range l.KeySchema {
			counts[aws.ToString(ks.AttributeName)] = counts[aws.ToString(ks.AttributeName)] + 1
		}
	}

	for _, g := range table.GlobalSecondaryIndexes {
		if aws.ToString(g.IndexName) == data.IndexName.ValueString() {
			continue
		}

		for _, ks := range g.KeySchema {
			counts[aws.ToString(ks.AttributeName)] = counts[aws.ToString(ks.AttributeName)] + 1
		}
	}

	var kss []keySchemaModel
	diags.Append(data.KeySchema.ElementsAs(ctx, &kss, false)...)
	if diags.HasError() {
		return diags
	}

	var hashKeys []string
	var rangeKeys []string
	for _, ks := range kss {
		switch ks.KeyType.ValueEnum() {
		case awstypes.KeyTypeHash:
			hashKeys = append(hashKeys, ks.AttributeName.ValueString())
		case awstypes.KeyTypeRange:
			rangeKeys = append(rangeKeys, ks.AttributeName.ValueString())
		default:
			diags.AddError(
				"Unknown key type in key_schema",
				fmt.Sprintf(
					`Uknown value "%s" for key_type in key_schema with name "%s"`,
					ks.KeyType.ValueString(),
					ks.AttributeName.ValueString(),
				),
			)
		}
	}
	if len(hashKeys) < minNumberOfHashes || len(hashKeys) > maxNumberOfHashes {
		diags.AddError(
			"Unsupported number of hash keys",
			fmt.Sprintf(
				`Number of hash keys must be between %d and %d`,
				minNumberOfHashes,
				maxNumberOfHashes,
			),
		)
	}
	if len(rangeKeys) > maxNumberOfRanges {
		diags.AddError(
			"Unsupported number of range keys",
			fmt.Sprintf(
				`Number of range keys must be between %d and %d`,
				minNumberOfRanges,
				maxNumberOfRanges,
			),
		)
	}
	if diags.HasError() {
		return diags
	}

	var ads []keySchemaModel
	diags.Append(data.KeySchema.ElementsAs(ctx, &ads, false)...)
	if diags.HasError() {
		return diags
	}

	for _, ad := range ads {
		name := ad.AttributeName.ValueString()
		typ := ad.AttributeType.ValueEnum()
		if name == "" {
			continue
		}

		existing := ""
		for _, ad := range table.AttributeDefinitions {
			if aws.ToString(ad.AttributeName) == name {
				existing = string(ad.AttributeType)
			}
		}

		if existing == "" {
			continue
		}

		if existing != string(typ) && counts[name] > 0 {
			diags.AddError(
				"Changing already existing attribute",
				fmt.Sprintf(
					`creation of index "%s" on table "%s" is attempting to change already existing attribute "%s" from type "%s" to "%s"`,
					data.IndexName.ValueString(),
					data.TableName.ValueString(),
					name,
					existing,
					typ,
				),
			)
		}
	}

	return diags
}
