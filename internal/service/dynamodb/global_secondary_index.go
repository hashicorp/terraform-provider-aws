// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	gocmp "github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	warmThoughtPutMinReadUnitsPerSecond  = 12000
	warmThoughtPutMinWriteUnitsPerSecond = 4000
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
	framework.WithImportByID
}

func (r *resourceGlobalSecondaryIndex) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"hash_key": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hash_key_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[ddbtypes.ScalarAttributeType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
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
					setplanmodifier.RequiresReplaceIf(func(ctx context.Context, request planmodifier.SetRequest, response *setplanmodifier.RequiresReplaceIfFuncResponse) {
						var old []string
						var new []string

						request.StateValue.ElementsAs(ctx, &old, false)
						request.PlanValue.ElementsAs(ctx, &new, false)

						if len(old) == 0 && len(new) == 0 {
							return
						}

						if old != nil {
							slices.Sort(old)
						}

						if new != nil {
							slices.Sort(new)
						}

						response.RequiresReplace = !gocmp.Equal(old, new)
					}, "", ""),
				},
			},
			"projection_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[ddbtypes.ProjectionType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"range_key": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"range_key_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[ddbtypes.ScalarAttributeType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"read_capacity": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"table": schema.StringAttribute{
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
		Version: 1,
	}

	if s.Blocks == nil {
		s.Blocks = make(map[string]schema.Block)
	}
	s.Blocks[names.AttrTimeouts] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
		Update: true,
		Delete: true,
	})

	response.Schema = s
}

func (r *resourceGlobalSecondaryIndex) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceGlobalSecondaryIndexModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DynamoDBClient(ctx)
	if _, err := waitAllGSIActive(ctx, conn, data.Table.ValueString(), r.ReadTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for all GSIs on table "%s" to be active`, data.Table.ValueString()),
			err.Error(),
		)

		return
	}

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)

	table, err := findTableByName(ctx, conn, data.Table.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Unable to read table "%s"`, data.Table.ValueString()),
			err.Error(),
		)

		return
	}

	knownAttributes := map[string]ddbtypes.ScalarAttributeType{}
	ut := &dynamodb.UpdateTableInput{
		TableName:            data.Table.ValueStringPointer(),
		AttributeDefinitions: []ddbtypes.AttributeDefinition{},
	}

	for _, ad := range table.AttributeDefinitions {
		ut.AttributeDefinitions = append(ut.AttributeDefinitions, ad)
		knownAttributes[aws.ToString(ad.AttributeName)] = ad.AttributeType
	}

	keySchema := []ddbtypes.KeySchemaElement{}
	hashKey := data.HashKey.ValueString()
	if hashKey == "" {
		response.Diagnostics.AddError(
			"Cannot create GSI without a hash key",
			fmt.Sprintf(`GSI named "%s" is missing a hash key`, data.Name.ValueString()),
		)

		return
	}

	if data.HashKeyType.ValueString() == "" {
		if hashKeyType, found := knownAttributes[hashKey]; found {
			data.HashKeyType = fwtypes.StringEnumValue(hashKeyType)
		} else {
			response.Diagnostics.AddError(
				`"hash_key_type" must be set in this configuration`,
				`When using "hash_key" that is not defined in the attributes list of the table, you must specify the "hash_key_type"`,
			)

			return
		}
	} else {
		ut.AttributeDefinitions = append(ut.AttributeDefinitions, ddbtypes.AttributeDefinition{
			AttributeName: aws.String(hashKey),
			AttributeType: data.HashKeyType.ValueEnum(),
		})
	}
	keySchema = append(keySchema, ddbtypes.KeySchemaElement{
		AttributeName: aws.String(hashKey),
		KeyType:       ddbtypes.KeyTypeHash,
	})

	rangeKey := data.RangeKey.ValueString()
	if rangeKey != "" {
		if data.RangeKeyType.ValueString() == "" {
			if rangeKeyType, found := knownAttributes[rangeKey]; found {
				data.RangeKeyType = fwtypes.StringEnumValue(rangeKeyType)
			} else {
				response.Diagnostics.AddError(
					`"range_key_type" must be set in this configuration`,
					`When using "range_key" that is not defined in the attributes list of the table, you must specify the "range_key_type"`,
				)

				return
			}
		} else {
			ut.AttributeDefinitions = append(ut.AttributeDefinitions, ddbtypes.AttributeDefinition{
				AttributeName: aws.String(rangeKey),
				AttributeType: data.RangeKeyType.ValueEnum(),
			})
		}

		keySchema = append(keySchema, ddbtypes.KeySchemaElement{
			AttributeName: aws.String(rangeKey),
			KeyType:       ddbtypes.KeyTypeRange,
		})
	}

	projection := &ddbtypes.Projection{
		ProjectionType: data.ProjectionType.ValueEnum(),
	}

	if !data.NonKeyAttributes.IsNull() && !data.NonKeyAttributes.IsUnknown() {
		response.Diagnostics.Append(
			data.NonKeyAttributes.ElementsAs(ctx, &projection.NonKeyAttributes, false)...,
		)
	}

	action := &ddbtypes.CreateGlobalSecondaryIndexAction{
		IndexName:  data.Name.ValueStringPointer(),
		KeySchema:  keySchema,
		Projection: projection,
	}

	if data.ReadCapacity.IsNull() || data.ReadCapacity.IsUnknown() {
		data.ReadCapacity = types.Int64Value(0)
	}

	if data.WriteCapacity.IsNull() || data.WriteCapacity.IsUnknown() {
		data.WriteCapacity = types.Int64Value(0)
	}

	billingMode := ddbtypes.BillingModeProvisioned
	if table.BillingModeSummary != nil {
		billingMode = table.BillingModeSummary.BillingMode
	}

	if billingMode == ddbtypes.BillingModeProvisioned {
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

	ut.GlobalSecondaryIndexUpdates = []ddbtypes.GlobalSecondaryIndexUpdate{
		{
			Create: action,
		},
	}

	response.Diagnostics.Append(validateNewGSIAttributes(data, *table)...)

	if response.Diagnostics.HasError() {
		return
	}

	if utRes, err := conn.UpdateTable(ctx, ut); err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Unable to create index "%s" on table "%s"`, data.Name.ValueString(), data.Table.ValueString()),
			err.Error(),
		)

		return
	} else {
		for _, g := range utRes.TableDescription.GlobalSecondaryIndexes {
			if aws.ToString(g.IndexName) == data.Name.ValueString() {
				flattenGSI(ctx, *table, g, &data)

				break
			}
		}
	}

	if _, err = waitTableActive(ctx, conn, data.Table.ValueString(), createTimeout); err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for table "%s" to be active`, data.Table.ValueString()),
			err.Error(),
		)

		return
	}

	if _, err := waitGSIActive(ctx, conn, data.Table.ValueString(), data.Name.ValueString(), r.UpdateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for GSI "%s" on table "%s" to be active`, data.Name.ValueString(), data.Table.ValueString()),
			err.Error(),
		)
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
	grps := re.FindStringSubmatch(data.ID.ValueString())
	var tableName string
	if len(grps) == 3 {
		tableName = grps[1]
	} else {
		tableName = data.Table.ValueString()
	}

	conn := r.Meta().DynamoDBClient(ctx)

	table, err := findTableByName(ctx, conn, tableName)
	if err != nil {
		if tfresource.NotFound(err) {
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

	found := false
	for _, g := range table.GlobalSecondaryIndexes {
		if aws.ToString(g.IndexArn) == data.ID.ValueString() {
			found = true

			flattenGSI(ctx, *table, g, &data)

			break
		}
	}

	if !found {
		response.Diagnostics.Append(
			fwdiag.NewResourceNotFoundWarningDiagnostic(
				fmt.Errorf(`unable to find global secondary index with id "%s", treating it as new`, data.ID.ValueString()),
			),
		)
		response.State.RemoveResource(ctx)
	}

	if !data.WarmThroughputs.IsNull() && !data.WarmThroughputs.IsUnknown() {
		var wt []warmThroughputModel

		response.Diagnostics.Append(data.WarmThroughputs.ElementsAs(ctx, &wt, false)...)

		if !response.Diagnostics.HasError() && len(wt) > 0 {
			rups := wt[0].ReadUnitsPerSecond.ValueInt64()
			wups := wt[0].WriteUnitsPerSecond.ValueInt64()

			if (rups < warmThoughtPutMinReadUnitsPerSecond && wups < warmThoughtPutMinWriteUnitsPerSecond) || (rups == warmThoughtPutMinReadUnitsPerSecond && wups == warmThoughtPutMinWriteUnitsPerSecond) {
				data.WarmThroughputs = fwtypes.NewListNestedObjectValueOfValueSliceMust[warmThroughputModel](ctx, []warmThroughputModel{})
			}
		}
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
	readTimeout := r.ReadTimeout(ctx, new.Timeouts)
	conn := r.Meta().DynamoDBClient(ctx)

	if _, err := waitAllGSIActive(ctx, conn, new.Table.ValueString(), readTimeout); err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for all GSIs on table "%s" to be active`, new.Table.ValueString()),
			err.Error(),
		)

		return
	}

	table, err := findTableByName(ctx, conn, new.Table.ValueString())
	if err != nil {
		if tfresource.NotFound(err) {
			response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
			response.State.RemoveResource(ctx)
		} else {
			response.Diagnostics.AddError(
				fmt.Sprintf(`Unable to read table "%s"`, new.Table.ValueString()),
				err.Error(),
			)
		}

		return
	}

	action := &ddbtypes.UpdateGlobalSecondaryIndexAction{
		IndexName: new.Name.ValueStringPointer(),
	}

	billingMode := ddbtypes.BillingModeProvisioned
	if table.BillingModeSummary != nil {
		billingMode = table.BillingModeSummary.BillingMode
	}

	hasUpdate := false

	if billingMode == ddbtypes.BillingModeProvisioned {
		provisionedThroughputData := map[string]any{
			"read_capacity":  int(new.ReadCapacity.ValueInt64()),
			"write_capacity": int(new.WriteCapacity.ValueInt64()),
		}
		newProvisionedThroughput := expandProvisionedThroughput(provisionedThroughputData, billingMode)

		filteredGSIs := tfslices.Filter(table.GlobalSecondaryIndexes, func(description ddbtypes.GlobalSecondaryIndexDescription) bool {
			return aws.ToString(description.IndexArn) == old.ID.ValueString()
		})

		if len(filteredGSIs) == 0 {
			response.Diagnostics.AddError(
				"Unable to find remote GSI to update",
				fmt.Sprintf(
					`GSI with name "%s" (arn: "%s") was not found in table "%s"`,
					new.Name.ValueString(),
					new.ID.ValueString(),
					new.Table.ValueString(),
				),
			)

			return
		}

		g := filteredGSIs[0]

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
		TableName: new.Table.ValueStringPointer(),
		GlobalSecondaryIndexUpdates: []ddbtypes.GlobalSecondaryIndexUpdate{
			{
				Update: action,
			},
		},
	}

	response.Diagnostics.Append(validateNewGSIAttributes(new, *table)...)

	if hasUpdate && !response.Diagnostics.HasError() {
		if utRes, err := conn.UpdateTable(ctx, ut); err != nil {
			response.Diagnostics.AddError(
				fmt.Sprintf(`Unable to update index "%s" on table "%s"`, new.Name.ValueString(), new.Table.ValueString()),
				err.Error(),
			)

			return
		} else {
			for _, g := range utRes.TableDescription.GlobalSecondaryIndexes {
				if aws.ToString(g.IndexName) == new.Name.ValueString() {
					flattenGSI(ctx, *table, g, &new)

					break
				}
			}
		}

		if _, err = waitTableActive(ctx, conn, new.Table.ValueString(), updateTimeout); err != nil {
			response.Diagnostics.AddError(
				fmt.Sprintf(`Error while waiting for table "%s" to be active`, new.Table.ValueString()),
				err.Error(),
			)

			return
		}

		if _, err := waitGSIActive(ctx, conn, new.Table.ValueString(), new.Name.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(
				fmt.Sprintf(`Error while waiting for GSI "%s" on table "%s" to be active`, new.Name.ValueString(), new.Table.ValueString()),
				err.Error(),
			)
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
	readTimeout := r.ReadTimeout(ctx, data.Timeouts)
	conn := r.Meta().DynamoDBClient(ctx)

	if _, err := waitAllGSIActive(ctx, conn, data.Table.ValueString(), readTimeout); err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for all GSIs on table "%s" to be active`, data.Table.ValueString()),
			err.Error(),
		)

		return
	}

	table, err := findTableByName(ctx, conn, data.Table.ValueString())
	if err != nil {
		if tfresource.NotFound(err) {
			response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
			response.State.RemoveResource(ctx)
		} else {
			response.Diagnostics.AddError(
				fmt.Sprintf(`Unable to read table "%s"`, data.Table.ValueString()),
				err.Error(),
			)
		}

		return
	}

	knownAttributes := map[string]int{}
	for _, l := range table.LocalSecondaryIndexes {
		for _, ks := range l.KeySchema {
			knownAttributes[aws.ToString(ks.AttributeName)] = knownAttributes[aws.ToString(ks.AttributeName)] + 1
		}
	}

	for _, g := range table.GlobalSecondaryIndexes {
		if data.ID.ValueString() != aws.ToString(g.IndexArn) {
			for _, ks := range g.KeySchema {
				knownAttributes[aws.ToString(ks.AttributeName)] = knownAttributes[aws.ToString(ks.AttributeName)] + 1
			}
		}
	}

	ut := &dynamodb.UpdateTableInput{
		TableName:            data.Table.ValueStringPointer(),
		AttributeDefinitions: []ddbtypes.AttributeDefinition{},
		GlobalSecondaryIndexUpdates: []ddbtypes.GlobalSecondaryIndexUpdate{
			{
				Delete: &ddbtypes.DeleteGlobalSecondaryIndexAction{
					IndexName: data.Name.ValueStringPointer(),
				},
			},
		},
	}

	for _, ad := range table.AttributeDefinitions {
		if knownAttributes[aws.ToString(ad.AttributeName)] > 0 {
			ut.AttributeDefinitions = append(ut.AttributeDefinitions, ad)
		}
	}

	if utRes, err := conn.UpdateTable(ctx, ut); err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Unable to delete index "%s" on table "%s"`, data.Name.ValueString(), data.Table.ValueString()),
			err.Error(),
		)

		return
	} else {
		for _, gsi := range utRes.TableDescription.GlobalSecondaryIndexes {
			if aws.ToString(gsi.IndexName) == data.Name.ValueString() {
				data.ID = types.StringValue(aws.ToString(gsi.IndexArn))
			}
		}
	}

	if _, err = waitTableActive(ctx, conn, data.Table.ValueString(), deleteTimeout); err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for table "%s" to be active`, data.Table.ValueString()),
			err.Error(),
		)

		return
	}

	if _, err := waitGSIDeleted(ctx, conn, data.Table.ValueString(), data.Name.ValueString(), deleteTimeout); err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for GSI "%s" on table "%s" to be deleted`, data.Name.ValueString(), data.Table.ValueString()),
			err.Error(),
		)
	}
}

type resourceGlobalSecondaryIndexModel struct {
	framework.WithRegionModel

	HashKey             types.String                                             `tfsdk:"hash_key"`
	HashKeyType         fwtypes.StringEnum[ddbtypes.ScalarAttributeType]         `tfsdk:"hash_key_type"`
	ID                  types.String                                             `tfsdk:"id"`
	Name                types.String                                             `tfsdk:"name"`
	NonKeyAttributes    fwtypes.SetOfString                                      `tfsdk:"non_key_attributes"`
	ProjectionType      fwtypes.StringEnum[ddbtypes.ProjectionType]              `tfsdk:"projection_type"`
	RangeKey            types.String                                             `tfsdk:"range_key"`
	RangeKeyType        fwtypes.StringEnum[ddbtypes.ScalarAttributeType]         `tfsdk:"range_key_type"`
	ReadCapacity        types.Int64                                              `tfsdk:"read_capacity"`
	Table               types.String                                             `tfsdk:"table"`
	WriteCapacity       types.Int64                                              `tfsdk:"write_capacity"`
	OnDemandThroughputs fwtypes.ListNestedObjectValueOf[onDemandThroughputModel] `tfsdk:"on_demand_throughput"`
	WarmThroughputs     fwtypes.ListNestedObjectValueOf[warmThroughputModel]     `tfsdk:"warm_throughput"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type onDemandThroughputModel struct {
	MaxReadRequestUnits  types.Int64 `tfsdk:"max_read_request_units"`
	MaxWriteRequestUnits types.Int64 `tfsdk:"max_write_request_units"`
}

type warmThroughputModel struct {
	ReadUnitsPerSecond  types.Int64 `tfsdk:"read_units_per_second"`
	WriteUnitsPerSecond types.Int64 `tfsdk:"write_units_per_second"`
}

func flattenGSI(ctx context.Context, table ddbtypes.TableDescription, g ddbtypes.GlobalSecondaryIndexDescription, data *resourceGlobalSecondaryIndexModel) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	data.ID = types.StringValue(aws.ToString(g.IndexArn))

	data.Name = types.StringValue(aws.ToString(g.IndexName))
	data.Table = types.StringValue(aws.ToString(table.TableName))
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

		data.OnDemandThroughputs = fwtypes.NewListNestedObjectValueOfValueSliceMust[onDemandThroughputModel](ctx, odtms)
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

		data.WarmThroughputs = fwtypes.NewListNestedObjectValueOfValueSliceMust[warmThroughputModel](ctx, wtms)
	}

	for _, ks := range g.KeySchema {
		if ks.KeyType == ddbtypes.KeyTypeHash && (data.HashKey.IsNull() || data.HashKey.IsUnknown()) {
			data.HashKey = types.StringValue(aws.ToString(ks.AttributeName))
		}

		if ks.KeyType == ddbtypes.KeyTypeRange && (data.RangeKey.IsNull() || data.RangeKey.IsUnknown()) {
			data.RangeKey = types.StringValue(aws.ToString(ks.AttributeName))
		}
	}

	for _, attr := range table.AttributeDefinitions {
		if data.HashKeyType.ValueString() == "" && data.HashKey.ValueString() == aws.ToString(attr.AttributeName) {
			data.HashKeyType = fwtypes.StringEnumValue(attr.AttributeType)
		}
		if data.RangeKeyType.ValueString() == "" && data.RangeKey.ValueString() == aws.ToString(attr.AttributeName) {
			data.RangeKeyType = fwtypes.StringEnumValue(attr.AttributeType)
		}
	}
}

func validateNewGSIAttributes(data resourceGlobalSecondaryIndexModel, table ddbtypes.TableDescription) diag.Diagnostics {
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
		if aws.ToString(g.IndexName) == data.Name.ValueString() {
			continue
		}

		for _, ks := range g.KeySchema {
			counts[aws.ToString(ks.AttributeName)] = counts[aws.ToString(ks.AttributeName)] + 1
		}
	}

	keys := []string{
		data.HashKey.ValueString(),
		data.HashKeyType.ValueString(),
		data.RangeKey.ValueString(),
		data.RangeKeyType.ValueString(),
	}

	for idx := range len(keys) / 2 {
		pos := idx * 2
		name := keys[pos]
		typ := keys[pos+1]
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

		if existing != typ && counts[name] > 0 {
			diags.AddError(
				"Changing already existing attribute",
				fmt.Sprintf(
					`creation of index "%s" on table "%s" is attempting to change already existing attribute "%s" from type "%s" to "%s"`,
					data.Name.ValueString(),
					data.Table.ValueString(),
					name,
					existing,
					typ,
				),
			)
		}
	}

	return diags
}
