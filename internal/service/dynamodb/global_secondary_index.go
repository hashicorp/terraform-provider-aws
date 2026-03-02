// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package dynamodb

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

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
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
	semconv "go.opentelemetry.io/otel/semconv/v1.38.0"
)

const (
	minNumberOfHashes = 1
	maxNumberOfHashes = 4

	maxNumberOfRanges = 4
)

// @FrameworkResource("aws_dynamodb_global_secondary_index", name="Global Secondary Index")
// @IdentityAttribute("table_name")
// @IdentityAttribute("index_name")
// @ImportIDHandler("globalSecondaryIndexImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/dynamodb/types;awstypes;awstypes.GlobalSecondaryIndexDescription")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc=testAccGlobalSecondaryIndexImportStateIdFunc)
// @Testing(importStateIdAttribute="arn")
// @Testing(requireEnvVar="TF_AWS_EXPERIMENT_dynamodb_global_secondary_index")
func newResourceGlobalSecondaryIndex(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceGlobalSecondaryIndex{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(60 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

var (
	_ resource.ResourceWithValidateConfig = &resourceGlobalSecondaryIndex{}
)

type resourceGlobalSecondaryIndex struct {
	framework.ResourceWithModel[resourceGlobalSecondaryIndexModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
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
			// TODO: Once Protocol v6 is supported, convert this to a `schema.SingleNestedAttribute` with full schema information
			"warm_throughput": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[warmThroughputModel](ctx),
				Optional:   true,
				Computed:   true,
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
					globalSecondaryIndexKeySchemaListValidator{},
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}

	response.Schema = s
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
	response.Diagnostics.Append(data.KeySchema.ElementsAs(ctx, &ksms, false)...)
	if response.Diagnostics.HasError() {
		return
	}
	for _, ks := range ksms {
		if _, exists := knownAttributes[ks.AttributeName.ValueString()]; exists {
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

	index, err := waitGSIActive(ctx, conn, data.TableName.ValueString(), data.IndexName.ValueString(), createTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(`Error while waiting for GSI "%s" on table "%s" to be active`, data.IndexName.ValueString(), data.TableName.ValueString()),
			err.Error(),
		)

		return
	}

	response.Diagnostics.Append(flattenGlobalSecondaryIndex(ctx, &data, index, table)...)
	if response.Diagnostics.HasError() {
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
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrTableName, data.TableName.ValueString(), "index_name", data.IndexName.ValueString())
		return
	}

	index, err := findGSIFromTable(table, indexName)
	if err != nil || index == nil {
		response.Diagnostics.Append(
			fwdiag.NewResourceNotFoundWarningDiagnostic(
				fmt.Errorf(`unable to find global secondary index with arn "%s", treating it as new`, data.ARN.ValueString()),
			),
		)
		response.State.RemoveResource(ctx)

		return
	}

	response.Diagnostics.Append(flattenGlobalSecondaryIndex(ctx, &data, index, table)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceGlobalSecondaryIndex) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan, config resourceGlobalSecondaryIndexModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	updateProvisionedThroughput := !plan.ProvisionedThroughput.Equal(state.ProvisionedThroughput)
	updateOnDemandThroughput := !plan.OnDemandThroughput.Equal(state.OnDemandThroughput)
	// Need to ignore `warm_throughput` when it's not set in config
	updateWarmThroughput := !config.WarmThroughput.IsNull() && !plan.WarmThroughput.Equal(state.WarmThroughput)

	if updateOnDemandThroughput {
		newOnDemandThroughput, d := plan.OnDemandThroughput.ToPtr(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}

		if newOnDemandThroughput.MaxReadRequestUnits.IsNull() || newOnDemandThroughput.MaxReadRequestUnits.IsUnknown() {
			newOnDemandThroughput.MaxReadRequestUnits = types.Int64Value(-1)
		}
		if newOnDemandThroughput.MaxWriteRequestUnits.IsNull() || newOnDemandThroughput.MaxWriteRequestUnits.IsUnknown() {
			newOnDemandThroughput.MaxWriteRequestUnits = types.Int64Value(-1)
		}

		plan.OnDemandThroughput = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, newOnDemandThroughput)
	}

	if updateProvisionedThroughput || updateOnDemandThroughput || updateWarmThroughput {
		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		conn := r.Meta().DynamoDBClient(ctx)

		table, err := waitAllGSIActive(ctx, conn, plan.TableName.ValueString(), updateTimeout)
		if err != nil {
			if retry.NotFound(err) {
				response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
				response.State.RemoveResource(ctx)
			} else {
				response.Diagnostics.AddError(
					fmt.Sprintf(`Unable to read table "%s"`, plan.TableName.ValueString()),
					err.Error(),
				)
			}

			return
		}

		response.Diagnostics.Append(validateNewGSIAttributes(ctx, plan, table)...)
		if response.Diagnostics.HasError() {
			return
		}

		input := dynamodb.UpdateTableInput{
			TableName:                   plan.TableName.ValueStringPointer(),
			GlobalSecondaryIndexUpdates: make([]awstypes.GlobalSecondaryIndexUpdate, 1),
		}

		var action awstypes.UpdateGlobalSecondaryIndexAction
		response.Diagnostics.Append(fwflex.Expand(ctx, plan, &action)...)
		if response.Diagnostics.HasError() {
			return
		}

		if updateProvisionedThroughput || updateOnDemandThroughput {
			innerAction := action
			innerAction.WarmThroughput = nil

			input.GlobalSecondaryIndexUpdates[0].Update = &innerAction

			_, err := conn.UpdateTable(ctx, &input)
			if err != nil {
				smerr.AddError(ctx, &response.Diagnostics, err, names.AttrTableName, plan.TableName.ValueString(), "index_name", plan.IndexName.ValueString())
				return
			}
		}

		if updateWarmThroughput {
			innerAction := action
			innerAction.OnDemandThroughput = nil
			innerAction.ProvisionedThroughput = nil

			input.GlobalSecondaryIndexUpdates[0].Update = &innerAction

			_, err := conn.UpdateTable(ctx, &input)
			if err != nil {
				smerr.AddError(ctx, &response.Diagnostics, err, names.AttrTableName, plan.TableName.ValueString(), "index_name", plan.IndexName.ValueString())
				return
			}
		}

		if table, err = waitTableActive(ctx, conn, plan.TableName.ValueString(), updateTimeout); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, names.AttrTableName, plan.TableName.ValueString(), "index_name", plan.IndexName.ValueString())
			return
		}

		if err = waitGSIWarmThroughputActive(ctx, conn, plan.TableName.ValueString(), plan.IndexName.ValueString(), updateTimeout); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, names.AttrTableName, plan.TableName.ValueString(), "index_name", plan.IndexName.ValueString())
			return
		}

		index, err := waitGSIActive(ctx, conn, plan.TableName.ValueString(), plan.IndexName.ValueString(), updateTimeout)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, names.AttrTableName, plan.TableName.ValueString(), "index_name", plan.IndexName.ValueString())
			return
		}

		response.Diagnostics.Append(flattenGlobalSecondaryIndex(ctx, &plan, index, table)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
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
		GlobalSecondaryIndexUpdates: []awstypes.GlobalSecondaryIndexUpdate{
			{
				Delete: &awstypes.DeleteGlobalSecondaryIndexAction{
					IndexName: data.IndexName.ValueStringPointer(),
				},
			},
		},
	}

	if res, err := conn.UpdateTable(ctx, &input); err != nil {
		// TODO: not sure this is possible when err != nil
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

const (
	globalSecondaryIndexExperimentalFlagEnvVar = "TF_AWS_EXPERIMENT_dynamodb_global_secondary_index"
)

func (r *resourceGlobalSecondaryIndex) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	if flag := os.Getenv(globalSecondaryIndexExperimentalFlagEnvVar); flag != "" {
		response.Diagnostics.AddWarning(
			"Experimental Resource Type Enabled: \"aws_dynamodb_global_secondary_index\"",
			fmt.Sprintf("The experimental resource type \"aws_dynamodb_global_secondary_index\" has been enabled by setting the environment variable \""+globalSecondaryIndexExperimentalFlagEnvVar+"\" to %q.", flag)+
				"\n\nPlease be aware that experimental features are not covered by any SLA or support agreement, and may not be suitable for production workloads. "+
				"Experimental features may change without notice, including removal from the provider.",
		)
		tflog.Info(ctx, "Experimental resource type enabled", map[string]any{
			string(semconv.FeatureFlagKeyKey):           globalSecondaryIndexExperimentalFlagEnvVar,
			string(semconv.FeatureFlagResultValueKey):   flag,
			string(semconv.FeatureFlagResultVariantKey): "enabled", // nosemgrep:ci.literal-enabled-string-constant
		})
	} else {
		response.Diagnostics.AddError(
			"Experimental Resource Type Not Enabled: \"aws_dynamodb_global_secondary_index\"",
			"The experimental resource type \"aws_dynamodb_global_secondary_index\" is not enabled. "+
				"To enable this resource type, set the environment variable \""+globalSecondaryIndexExperimentalFlagEnvVar+"\" to any value before running Terraform."+
				"\n\nPlease be aware that experimental features are not covered by any SLA or support agreement, and may not be suitable for production workloads. "+
				"Experimental features may change without notice, including removal from the provider.",
		)
		tflog.Error(ctx, "Experimental resource type not enabled", map[string]any{
			string(semconv.FeatureFlagKeyKey):                  globalSecondaryIndexExperimentalFlagEnvVar,
			string(semconv.FeatureFlagResultValueKey):          nil,
			string(semconv.FeatureFlagResultVariantKey):        "disabled",
			string(semconv.FeatureFlagResultReasonDefault.Key): semconv.FeatureFlagResultReasonDefault.Value,
		})
	}
}

type resourceGlobalSecondaryIndexModel struct {
	framework.WithRegionModel

	ARN                   types.String                                                `tfsdk:"arn"`
	IndexName             types.String                                                `tfsdk:"index_name"`
	KeySchema             fwtypes.ListNestedObjectValueOf[keySchemaModel]             `tfsdk:"key_schema"`
	TableName             types.String                                                `tfsdk:"table_name"`
	OnDemandThroughput    fwtypes.ListNestedObjectValueOf[onDemandThroughputModel]    `tfsdk:"on_demand_throughput"`
	Projection            fwtypes.ListNestedObjectValueOf[projectionModel]            `tfsdk:"projection"`
	ProvisionedThroughput fwtypes.ListNestedObjectValueOf[provisionedThroughputModel] `tfsdk:"provisioned_throughput"`
	WarmThroughput        fwtypes.ObjectValueOf[warmThroughputModel]                  `tfsdk:"warm_throughput"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func flattenGlobalSecondaryIndex(ctx context.Context, data *resourceGlobalSecondaryIndexModel, index *awstypes.GlobalSecondaryIndexDescription, table *awstypes.TableDescription) diag.Diagnostics { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, index, data,
		fwflex.WithFieldNamePrefix("Index"),
		fwflex.WithIgnoredFieldNamesAppend("KeySchema"),
		fwflex.WithIgnoredFieldNamesAppend("ProvisionedThroughput"),
	)...)

	data.TableName = fwflex.StringToFramework(ctx, table.TableName)

	attributeTypes := make(map[string]awstypes.ScalarAttributeType, len(table.AttributeDefinitions))
	for _, attribute := range table.AttributeDefinitions {
		attributeTypes[aws.ToString(attribute.AttributeName)] = attribute.AttributeType
	}

	keyModels := make([]keySchemaModel, len(index.KeySchema))
	for i, ks := range index.KeySchema {
		keyModels[i] = keySchemaModel{
			AttributeName: fwflex.StringToFramework(ctx, ks.AttributeName),
			AttributeType: fwtypes.StringEnumValue(attributeTypes[aws.ToString(ks.AttributeName)]),
			KeyType:       fwtypes.StringEnumValue(ks.KeyType),
		}
	}

	data.KeySchema = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, keyModels)

	if index.ProvisionedThroughput != nil {
		var ptM provisionedThroughputModel
		d := fwflex.Flatten(ctx, index.ProvisionedThroughput, &ptM)
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

	keySchemaRefCounts := map[string]int{}
	for _, ks := range table.KeySchema {
		keySchemaRefCounts[aws.ToString(ks.AttributeName)]++
	}

	for _, l := range table.LocalSecondaryIndexes {
		for _, ks := range l.KeySchema {
			keySchemaRefCounts[aws.ToString(ks.AttributeName)]++
		}
	}

	for _, g := range table.GlobalSecondaryIndexes {
		if aws.ToString(g.IndexName) == data.IndexName.ValueString() {
			continue
		}

		for _, ks := range g.KeySchema {
			keySchemaRefCounts[aws.ToString(ks.AttributeName)]++
		}
	}

	keySchemaPath := path.Root("key_schema")

	for i, keySchema := range fwdiag.Must(data.KeySchema.ToSlice(ctx)) {
		attributeName := keySchema.AttributeName.ValueString()
		gsiAttributeType := keySchema.AttributeType.ValueEnum()

		var tableAttributeType awstypes.ScalarAttributeType
		for _, tableAttribute := range table.AttributeDefinitions {
			if aws.ToString(tableAttribute.AttributeName) == attributeName {
				tableAttributeType = tableAttribute.AttributeType
			}
		}

		if tableAttributeType != gsiAttributeType && keySchemaRefCounts[attributeName] > 0 {
			diags.Append(diag.NewAttributeErrorDiagnostic(
				keySchemaPath.AtListIndex(i).AtMapKey("attribute_type"),
				"Invalid Key Schema Type Change",
				fmt.Sprintf(`The "attribute_type" of the key schema attribute %q was previously defined as %q. It cannot be redefined as %q.`, attributeName, tableAttributeType, gsiAttributeType),
			))
		}
	}

	return diags
}

var (
	_ inttypes.ImportIDParser = globalSecondaryIndexImportID{}
)

type globalSecondaryIndexImportID struct{}

func (globalSecondaryIndexImportID) Parse(id string) (string, map[string]any, error) {
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
