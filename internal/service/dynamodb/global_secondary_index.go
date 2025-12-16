// Copyright IBM Corp. 2014, 2025
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
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
			names.AttrTableName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
				CustomType: fwtypes.NewListNestedObjectTypeOf[onDemandThroughputModel](ctx),
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
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"projection": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[projectionModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
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
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
					listplanmodifier.RequiresReplace(),
				},
			},
			"provisioned_throughput": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[provisionedThroughputModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"read_capacity_units": schema.Int64Attribute{
							Optional: true,
							Computed: true,
						},
						"write_capacity_units": schema.Int64Attribute{
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
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

	response.Diagnostics.Append(validateNewGSIAttributes(ctx, data, table)...)
	if response.Diagnostics.HasError() {
		return
	}

	billingMode := awstypes.BillingModeProvisioned
	if table.BillingModeSummary != nil {
		billingMode = table.BillingModeSummary.BillingMode
	}

	if billingMode == awstypes.BillingModeProvisioned {
		if data.ProvisionedThroughput.IsNull() || data.ProvisionedThroughput.IsUnknown() {
			response.Diagnostics.Append(
				validatordiag.InvalidAttributeCombinationDiagnostic(
					path.Root("provisioned_throughput"),
					fmt.Sprintf("Attribute %q must be specified when the associated table's attribute \"billing_mode\" is %q.",
						path.Root("provisioned_throughput").String(),
						awstypes.BillingModeProvisioned,
					),
				),
			)
		}
	} else {
		if !data.ProvisionedThroughput.IsNull() && !data.ProvisionedThroughput.IsUnknown() {
			response.Diagnostics.Append(
				validatordiag.InvalidAttributeCombinationDiagnostic(
					path.Root("provisioned_throughput"),
					fmt.Sprintf("Attribute %q cannot be specified when the associated table's attribute \"billing_mode\" is %q.",
						path.Root("provisioned_throughput").String(),
						billingMode,
					),
				),
			)
		}
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

		input.AttributeDefinitions = append(input.AttributeDefinitions, awstypes.AttributeDefinition{
			AttributeName: ks.AttributeName.ValueStringPointer(),
			AttributeType: ks.AttributeType.ValueEnum(),
		})
	}

	var action awstypes.CreateGlobalSecondaryIndexAction
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &action)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.GlobalSecondaryIndexUpdates = []awstypes.GlobalSecondaryIndexUpdate{
		{
			Create: &action,
		},
	}

	if utRes, err := conn.UpdateTable(ctx, &input); err != nil {
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

	response.Diagnostics.Append(validateNewGSIAttributes(ctx, new, table)...)
	if response.Diagnostics.HasError() {
		return
	}

	var action awstypes.UpdateGlobalSecondaryIndexAction
	response.Diagnostics.Append(fwflex.Expand(ctx, new, &action)...)
	if response.Diagnostics.HasError() {
		return
	}

	hasUpdate := !new.ProvisionedThroughput.Equal(old.ProvisionedThroughput) ||
		!new.OnDemandThroughputs.Equal(old.OnDemandThroughputs) ||
		!new.WarmThroughputs.Equal(old.WarmThroughputs)

	input := dynamodb.UpdateTableInput{
		TableName: new.TableName.ValueStringPointer(),
		GlobalSecondaryIndexUpdates: []awstypes.GlobalSecondaryIndexUpdate{
			{
				Update: &action,
			},
		},
	}

	if hasUpdate {
		if utRes, err := conn.UpdateTable(ctx, &input); err != nil {
			response.Diagnostics.AddError(
				fmt.Sprintf(`Unable to update index "%s" on table "%s"`, new.IndexName.ValueString(), new.TableName.ValueString()),
				err.Error(),
			)

			return
		} else {
			g, err := findGSIFromTable(utRes.TableDescription, new.IndexName.ValueString())
			table = utRes.TableDescription

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
		}

		if _, err = waitTableActive(ctx, conn, new.TableName.ValueString(), updateTimeout); err != nil {
			response.Diagnostics.AddError(
				fmt.Sprintf(`Error while waiting for table "%s" to be active`, new.TableName.ValueString()),
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

		if g, err := waitGSIActive(ctx, conn, new.TableName.ValueString(), new.IndexName.ValueString(), updateTimeout); err != nil {
			response.Diagnostics.AddError(
				fmt.Sprintf(`Error while waiting for GSI "%s" on table "%s" to be active`, new.IndexName.ValueString(), new.TableName.ValueString()),
				err.Error(),
			)

			return
		} else {
			response.Diagnostics.Append(flattenGlobalSecondaryIndex(ctx, &new, *g, table)...)
			if response.Diagnostics.HasError() {
				return
			}
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

	input := dynamodb.UpdateTableInput{
		TableName: data.TableName.ValueStringPointer(),
		GlobalSecondaryIndexUpdates: []awstypes.GlobalSecondaryIndexUpdate{
			{
				Delete: &awstypes.DeleteGlobalSecondaryIndexAction{
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

	ARN                   types.String                                                `tfsdk:"arn"`
	IndexName             types.String                                                `tfsdk:"index_name"`
	KeySchema             fwtypes.ListNestedObjectValueOf[keySchemaModel]             `tfsdk:"key_schema"`
	TableName             types.String                                                `tfsdk:"table_name"`
	OnDemandThroughputs   fwtypes.ListNestedObjectValueOf[onDemandThroughputModel]    `tfsdk:"on_demand_throughput"`
	Projection            fwtypes.ListNestedObjectValueOf[projectionModel]            `tfsdk:"projection"`
	ProvisionedThroughput fwtypes.ListNestedObjectValueOf[provisionedThroughputModel] `tfsdk:"provisioned_throughput"`
	WarmThroughputs       fwtypes.ListNestedObjectValueOf[warmThroughputModel]        `tfsdk:"warm_throughput"`

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

	var projection projectionModel
	d := fwflex.Flatten(ctx, g.Projection, &projection)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	data.Projection = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []projectionModel{projection})

	if g.ProvisionedThroughput == nil {
		data.ProvisionedThroughput = fwtypes.NewListNestedObjectValueOfNull[provisionedThroughputModel](ctx)
	} else {
		var ptM provisionedThroughputModel
		d := fwflex.Flatten(ctx, g.ProvisionedThroughput, &ptM)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		if ptM.IsZero() {
			data.ProvisionedThroughput = fwtypes.NewListNestedObjectValueOfNull[provisionedThroughputModel](ctx)
		} else {
			data.ProvisionedThroughput = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []provisionedThroughputModel{ptM})
		}
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

type projectionModel struct {
	NonKeyAttributes fwtypes.SetOfString                         `tfsdk:"non_key_attributes"`
	ProjectionType   fwtypes.StringEnum[awstypes.ProjectionType] `tfsdk:"projection_type"`
}

type provisionedThroughputModel struct {
	ReadCapacityUnits  types.Int64 `tfsdk:"read_capacity_units"`
	WriteCapacityUnits types.Int64 `tfsdk:"write_capacity_units"`
}

func (m provisionedThroughputModel) IsZero() bool {
	return int64IsZero(m.ReadCapacityUnits) && int64IsZero(m.WriteCapacityUnits)
}

func int64IsZero(v types.Int64) bool {
	return v.IsNull() || v.ValueInt64() == 0
}

func (m provisionedThroughputModel) Equal(o provisionedThroughputModel) bool {
	return m.ReadCapacityUnits.Equal(o.ReadCapacityUnits) && m.WriteCapacityUnits.Equal(o.WriteCapacityUnits)
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

func validateNewGSIAttributes(ctx context.Context, data resourceGlobalSecondaryIndexModel, table *awstypes.TableDescription) diag.Diagnostics {
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
