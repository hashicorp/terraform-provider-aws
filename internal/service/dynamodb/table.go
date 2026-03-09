// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"slices"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfcty "github.com/hashicorp/terraform-provider-aws/internal/cty"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/sync"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	provisionedThroughputMinValue = 1
	resNameTable                  = "Table"
)

// @SDKResource("aws_dynamodb_table", name="Table")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/dynamodb/types;types.TableDescription")
func resourceTable() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceTableCreate,
		ReadWithoutTimeout:   resourceTableRead,
		UpdateWithoutTimeout: resourceTableUpdate,
		DeleteWithoutTimeout: resourceTableDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(createTableTimeout),
			Delete: schema.DefaultTimeout(deleteTableTimeout),
			Update: schema.DefaultTimeout(updateTableTimeoutTotal),
		},

		CustomizeDiff: customdiff.All(
			func(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
				return validateTableAttributes(ctx, diff, meta)
			},
			func(_ context.Context, diff *schema.ResourceDiff, meta any) error {
				if diff.Id() != "" && diff.HasChange("server_side_encryption") {
					o, n := diff.GetChange("server_side_encryption")
					if isTableOptionDisabled(o) && isTableOptionDisabled(n) {
						return diff.Clear("server_side_encryption")
					}
				}
				return nil
			},
			func(_ context.Context, diff *schema.ResourceDiff, meta any) error {
				if diff.Id() != "" && diff.HasChange("point_in_time_recovery") {
					o, n := diff.GetChange("point_in_time_recovery")
					if isTableOptionDisabled(o) && isTableOptionDisabled(n) {
						return diff.Clear("point_in_time_recovery")
					}
				}
				return nil
			},
			func(_ context.Context, diff *schema.ResourceDiff, meta any) error {
				if diff.Id() != "" && (diff.HasChange("stream_enabled") || (diff.Get("stream_view_type") != "" && diff.HasChange("stream_view_type"))) {
					if err := diff.SetNewComputed(names.AttrStreamARN); err != nil {
						return fmt.Errorf("setting stream_arn to computed: %w", err)
					}
				}
				return nil
			},
			customdiff.ForceNewIfChange("restore_source_name", func(_ context.Context, old, new, meta any) bool {
				// If they differ force new unless new is cleared
				// https://github.com/hashicorp/terraform-provider-aws/issues/25214
				return old.(string) != new.(string) && new.(string) != ""
			}),
			customdiff.ForceNewIfChange("restore_source_table_arn", func(_ context.Context, old, new, meta any) bool {
				return old.(string) != new.(string) && new.(string) != ""
			}),
			customdiff.ForceNewIfChange("warm_throughput.0.read_units_per_second", func(_ context.Context, old, new, meta any) bool {
				// warm_throughput can only be increased, not decreased
				// i.e., "api error ValidationException: One or more parameter values were invalid: Requested ReadUnitsPerSecond for WarmThroughput for table is lower than current WarmThroughput, decreasing WarmThroughput is not supported"
				if old, new := old.(int), new.(int); new != 0 && new < old {
					return true
				}

				return false
			}),
			customdiff.ForceNewIfChange("warm_throughput.0.write_units_per_second", func(_ context.Context, old, new, meta any) bool {
				// warm_throughput can only be increased, not decreased
				// i.e., "api error ValidationException: One or more parameter values were invalid: Requested ReadUnitsPerSecond for WarmThroughput for table is lower than current WarmThroughput, decreasing WarmThroughput is not supported"
				if old, new := old.(int), new.(int); new != 0 && new < old {
					return true
				}

				return false
			}),
			suppressTableWarmThroughputDefaults,
			customDiffGlobalSecondaryIndex,
			func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
				rs := diff.GetRawState()
				if rs.IsNull() {
					return nil
				}

				if diff.HasChange("attribute") && !diff.HasChange("global_secondary_index") {
					return diff.Clear("attribute")
				}

				return nil
			},
		),

		SchemaVersion: 1,
		MigrateState:  resourceTableMigrateState,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"attribute": {
					Type:     schema.TypeSet,
					Optional: true,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrName: {
								Type:     schema.TypeString,
								Required: true,
							},
							names.AttrType: {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.ScalarAttributeType](),
							},
						},
					},
					Set: sdkv2.SimpleSchemaSetFunc(names.AttrName),
				},
				"billing_mode": {
					Type:             schema.TypeString,
					Optional:         true,
					Default:          awstypes.BillingModeProvisioned,
					ValidateDiagFunc: enum.Validate[awstypes.BillingMode](),
				},
				"deletion_protection_enabled": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"global_secondary_index": {
					Type:     schema.TypeSet,
					Optional: true,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"hash_key": {
								Type:       schema.TypeString,
								Optional:   true,
								Computed:   true,
								Deprecated: "hash_key is deprecated. Use key_schema instead.",
							},
							"key_schema": {
								Type:     schema.TypeList,
								Optional: true,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"attribute_name": {
											Type:     schema.TypeString,
											Required: true,
										},
										"key_type": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[awstypes.KeyType](),
										},
									},
								},
							},
							names.AttrName: {
								Type:     schema.TypeString,
								Required: true,
							},
							"non_key_attributes": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"on_demand_throughput": onDemandThroughputSchema(),
							"projection_type": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.ProjectionType](),
							},
							"range_key": {
								Type:       schema.TypeString,
								Optional:   true,
								Deprecated: "range_key is deprecated. Use key_schema instead.",
							},
							"read_capacity": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"warm_throughput": warmThroughputSchema(),
							"write_capacity": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
						},
					},
				},
				"global_table_witness": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"region_name": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: verify.ValidRegionName,
							},
						},
					},
				},
				"hash_key": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
				"import_table": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{"restore_source_name", "restore_source_table_arn"},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"input_compression_type": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[awstypes.InputCompressionType](),
							},
							"input_format": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.InputFormat](),
							},
							"input_format_options": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"csv": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"delimiter": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"header_list": {
														Type:     schema.TypeSet,
														Optional: true,
														Elem:     &schema.Schema{Type: schema.TypeString},
													},
												},
											},
										},
									},
								},
							},
							"s3_bucket_source": {
								Type:     schema.TypeList,
								MaxItems: 1,
								Required: true,
								ForceNew: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrBucket: {
											Type:     schema.TypeString,
											Required: true,
										},
										"bucket_owner": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"key_prefix": {
											Type:     schema.TypeString,
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
				"local_secondary_index": {
					Type:     schema.TypeSet,
					Optional: true,
					ForceNew: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrName: {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
							},
							"non_key_attributes": {
								Type:     schema.TypeList,
								Optional: true,
								ForceNew: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"projection_type": {
								Type:             schema.TypeString,
								Required:         true,
								ForceNew:         true,
								ValidateDiagFunc: enum.Validate[awstypes.ProjectionType](),
							},
							"range_key": {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
							},
						},
					},
					Set: sdkv2.SimpleSchemaSetFunc(names.AttrName),
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"on_demand_throughput": onDemandThroughputSchema(),
				"point_in_time_recovery": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrEnabled: {
								Type:     schema.TypeBool,
								Required: true,
							},
							"recovery_period_in_days": {
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntBetween(1, 35),
								DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
									return !d.Get("point_in_time_recovery.0.enabled").(bool)
								},
							},
						},
					},
				},
				"range_key": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
				"read_capacity": {
					Type:          schema.TypeInt,
					Optional:      true,
					Computed:      true,
					ConflictsWith: []string{"on_demand_throughput"},
				},
				"replica": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrARN: {
								Type:     schema.TypeString,
								Computed: true,
							},
							"consistency_mode": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          awstypes.MultiRegionConsistencyEventual,
								ValidateDiagFunc: enum.Validate[awstypes.MultiRegionConsistency](),
							},
							"deletion_protection_enabled": {
								Type:     schema.TypeBool,
								Optional: true,
								Computed: true,
							},
							names.AttrKMSKeyARN: {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: verify.ValidARN,
								// update is equivalent of force a new *replica*, not table
							},
							"point_in_time_recovery": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
							},
							names.AttrPropagateTags: {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
							},
							"region_name": {
								Type:     schema.TypeString,
								Required: true,
								// update is equivalent of force a new *replica*, not table
							},
							names.AttrStreamARN: {
								Type:     schema.TypeString,
								Computed: true,
							},
							"stream_label": {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
				"restore_date_time": {
					Type:         schema.TypeString,
					Optional:     true,
					ForceNew:     true,
					ValidateFunc: verify.ValidUTCTimestamp,
				},
				"restore_source_table_arn": {
					Type:          schema.TypeString,
					Optional:      true,
					ValidateFunc:  verify.ValidARN,
					ConflictsWith: []string{"import_table", "restore_source_name"},
				},
				"restore_source_name": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"import_table", "restore_source_table_arn"},
				},
				"restore_to_latest_time": {
					Type:     schema.TypeBool,
					Optional: true,
					ForceNew: true,
				},
				"server_side_encryption": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrEnabled: {
								Type:     schema.TypeBool,
								Required: true,
							},
							names.AttrKMSKeyARN: {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
				names.AttrStreamARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"stream_enabled": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"stream_label": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"stream_view_type": {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					StateFunc:    sdkv2.ToUpperSchemaStateFunc,
					ValidateFunc: validation.StringInSlice(append(enum.Values[awstypes.StreamViewType](), ""), false),
				},
				"table_class": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  awstypes.TableClassStandard,
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						return old == "" && new == string(awstypes.TableClassStandard)
					},
					ValidateDiagFunc: enum.Validate[awstypes.TableClass](),
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"ttl": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"attribute_name": {
								Type:     schema.TypeString,
								Optional: true,
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									// AWS requires the attribute name to be set when disabling TTL but
									// does not return it so it causes a diff.
									if old == "" && new != "" && !d.Get("ttl.0.enabled").(bool) {
										return true
									}
									return false
								},
							},
							names.AttrEnabled: {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
							},
						},
					},
					DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				},
				"warm_throughput": warmThroughputSchema(),
				"write_capacity": {
					Type:          schema.TypeInt,
					Computed:      true,
					Optional:      true,
					ConflictsWith: []string{"on_demand_throughput"},
				},
			}
		},

		ValidateRawResourceConfigFuncs: []schema.ValidateRawResourceConfigFunc{
			validateGlobalSecondaryIndexes,
			validateStreamSpecification,
			validateProvisionedThroughputField(cty.GetAttrPath("read_capacity")),
			validateProvisionedThroughputField(cty.GetAttrPath("write_capacity")),
			validateTTLList,
		},
	}
}

func onDemandThroughputSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"max_read_request_units": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						return old == "0" && new == "-1"
					},
				},
				"max_write_request_units": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						return old == "0" && new == "-1"
					},
				},
			},
		},
	}
}

func warmThroughputSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"read_units_per_second": {
					Type:         schema.TypeInt,
					Optional:     true,
					Computed:     true,
					ValidateFunc: validation.IntAtLeast(12000),
				},
				"write_units_per_second": {
					Type:         schema.TypeInt,
					Optional:     true,
					Computed:     true,
					ValidateFunc: validation.IntAtLeast(4000),
				},
			},
		},
	}
}

func resourceTableCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	tableName := d.Get(names.AttrName).(string)
	keySchemaMap := map[string]any{
		"hash_key": d.Get("hash_key").(string),
	}
	if v, ok := d.GetOk("range_key"); ok {
		keySchemaMap["range_key"] = v.(string)
	}

	sourceName, nameOk := d.GetOk("restore_source_name")
	sourceArn, arnOk := d.GetOk("restore_source_table_arn")

	if nameOk || arnOk {
		input := &dynamodb.RestoreTableToPointInTimeInput{
			TargetTableName: aws.String(tableName),
		}

		if nameOk {
			input.SourceTableName = aws.String(sourceName.(string))
		}

		if arnOk {
			input.SourceTableArn = aws.String(sourceArn.(string))
		}

		if v, ok := d.GetOk("restore_date_time"); ok {
			t, _ := time.Parse(time.RFC3339, v.(string))
			input.RestoreDateTime = aws.Time(t)
		}

		if attr, ok := d.GetOk("restore_to_latest_time"); ok {
			input.UseLatestRestorableTime = aws.Bool(attr.(bool))
		}

		billingModeOverride := awstypes.BillingMode(d.Get("billing_mode").(string))

		if _, ok := d.GetOk("write_capacity"); ok {
			if _, ok := d.GetOk("read_capacity"); ok {
				capacityMap := map[string]any{
					"write_capacity": d.Get("write_capacity"),
					"read_capacity":  d.Get("read_capacity"),
				}
				input.ProvisionedThroughputOverride = expandProvisionedThroughput(capacityMap, billingModeOverride)
			}
		}

		if v, ok := d.GetOk("local_secondary_index"); ok {
			lsiSet := v.(*schema.Set)
			input.LocalSecondaryIndexOverride = expandLocalSecondaryIndexes(lsiSet.List(), keySchemaMap)
		}

		if v, ok := d.GetOk("global_secondary_index"); ok {
			globalSecondaryIndexes := []awstypes.GlobalSecondaryIndex{}
			gsiSet := v.(*schema.Set)

			for _, gsiObject := range gsiSet.List() {
				gsi := gsiObject.(map[string]any)
				if err := validateGSIProvisionedThroughput(gsi, billingModeOverride); err != nil {
					return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, d.Get(names.AttrName).(string), err)
				}

				gsiObject := expandGlobalSecondaryIndex(gsi, billingModeOverride)
				globalSecondaryIndexes = append(globalSecondaryIndexes, *gsiObject)
			}
			input.GlobalSecondaryIndexOverride = globalSecondaryIndexes
		}

		if v, ok := d.GetOk("server_side_encryption"); ok {
			input.SSESpecificationOverride = expandEncryptAtRestOptions(v.([]any))
		}

		_, err := tfresource.RetryWhen(ctx, createTableTimeout, func(ctx context.Context) (any, error) {
			return conn.RestoreTableToPointInTime(ctx, input)
		}, func(err error) (bool, error) {
			if tfawserr.ErrCodeEquals(err, errCodeThrottlingException) {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.LimitExceededException](err, "can be created, updated, or deleted simultaneously") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.LimitExceededException](err, "indexed tables that can be created simultaneously") {
				return true, err
			}

			return false, err
		})

		if err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, tableName, err)
		}
	} else if vit, ok := d.GetOk("import_table"); ok && len(vit.([]any)) > 0 && vit.([]any)[0] != nil {
		input := expandImportTable(vit.([]any)[0].(map[string]any))

		tcp := &awstypes.TableCreationParameters{
			TableName:   aws.String(tableName),
			BillingMode: awstypes.BillingMode(d.Get("billing_mode").(string)),
			KeySchema:   expandKeySchema(keySchemaMap),
		}

		capacityMap := map[string]any{
			"write_capacity": d.Get("write_capacity"),
			"read_capacity":  d.Get("read_capacity"),
		}

		billingMode := awstypes.BillingMode(d.Get("billing_mode").(string))

		tcp.ProvisionedThroughput = expandProvisionedThroughput(capacityMap, billingMode)

		if v, ok := d.GetOk("attribute"); ok {
			aSet := v.(*schema.Set)
			tcp.AttributeDefinitions = expandAttributes(aSet.List())
		}

		if v, ok := d.GetOk("server_side_encryption"); ok {
			tcp.SSESpecification = expandEncryptAtRestOptions(v.([]any))
		}

		if v, ok := d.GetOk("global_secondary_index"); ok {
			globalSecondaryIndexes := []awstypes.GlobalSecondaryIndex{}
			gsiSet := v.(*schema.Set)

			for _, gsiObject := range gsiSet.List() {
				gsi := gsiObject.(map[string]any)
				if err := validateGSIProvisionedThroughput(gsi, billingMode); err != nil {
					return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, d.Get(names.AttrName).(string), err)
				}

				gsiObject := expandGlobalSecondaryIndex(gsi, billingMode)
				globalSecondaryIndexes = append(globalSecondaryIndexes, *gsiObject)
			}
			tcp.GlobalSecondaryIndexes = globalSecondaryIndexes
		}

		if v, ok := d.GetOk("on_demand_throughput"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			tcp.OnDemandThroughput = expandOnDemandThroughput(v.([]any)[0].(map[string]any))
		}

		input.TableCreationParameters = tcp

		importTableOutput, err := tfresource.RetryWhen(ctx, createTableTimeout, func(ctx context.Context) (any, error) {
			return conn.ImportTable(ctx, input)
		}, func(err error) (bool, error) {
			if tfawserr.ErrCodeEquals(err, errCodeThrottlingException) {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.LimitExceededException](err, "can be created, updated, or deleted simultaneously") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.LimitExceededException](err, "indexed tables that can be created simultaneously") {
				return true, err
			}

			return false, err
		})

		if err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, tableName, err)
		}

		importARN := importTableOutput.(*dynamodb.ImportTableOutput).ImportTableDescription.ImportArn
		if _, err := waitImportComplete(ctx, conn, aws.ToString(importARN), d.Timeout(schema.TimeoutCreate)); err != nil {
			d.SetId(tableName)
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, tableName, err)
		}
	} else {
		input := &dynamodb.CreateTableInput{
			BillingMode: awstypes.BillingMode(d.Get("billing_mode").(string)),
			KeySchema:   expandKeySchema(keySchemaMap),
			TableName:   aws.String(tableName),
			Tags:        getTagsIn(ctx),
		}

		billingMode := awstypes.BillingMode(d.Get("billing_mode").(string))

		capacityMap := map[string]any{
			"write_capacity": d.Get("write_capacity"),
			"read_capacity":  d.Get("read_capacity"),
		}

		input.ProvisionedThroughput = expandProvisionedThroughput(capacityMap, billingMode)

		if v, ok := d.GetOk("attribute"); ok {
			aSet := v.(*schema.Set)
			input.AttributeDefinitions = expandAttributes(aSet.List())
		}

		if v, ok := d.GetOk("deletion_protection_enabled"); ok {
			input.DeletionProtectionEnabled = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("local_secondary_index"); ok {
			lsiSet := v.(*schema.Set)
			input.LocalSecondaryIndexes = expandLocalSecondaryIndexes(lsiSet.List(), keySchemaMap)
		}

		if v, ok := d.GetOk("global_secondary_index"); ok {
			globalSecondaryIndexes := []awstypes.GlobalSecondaryIndex{}
			gsiSet := v.(*schema.Set)

			for _, gsiObject := range gsiSet.List() {
				gsi := gsiObject.(map[string]any)
				if err := validateGSIProvisionedThroughput(gsi, billingMode); err != nil {
					return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, tableName, err)
				}

				gsiObject := expandGlobalSecondaryIndex(gsi, billingMode)
				globalSecondaryIndexes = append(globalSecondaryIndexes, *gsiObject)
			}
			input.GlobalSecondaryIndexes = globalSecondaryIndexes
		}

		if v, ok := d.GetOk("on_demand_throughput"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.OnDemandThroughput = expandOnDemandThroughput(v.([]any)[0].(map[string]any))
		}

		if v, ok := d.GetOk("stream_enabled"); ok {
			input.StreamSpecification = &awstypes.StreamSpecification{
				StreamEnabled:  aws.Bool(v.(bool)),
				StreamViewType: awstypes.StreamViewType(d.Get("stream_view_type").(string)),
			}
		}

		if v, ok := d.GetOk("server_side_encryption"); ok {
			input.SSESpecification = expandEncryptAtRestOptions(v.([]any))
		}

		if v, ok := d.GetOk("table_class"); ok {
			input.TableClass = awstypes.TableClass(v.(string))
		}

		if v, ok := d.GetOk("warm_throughput"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.WarmThroughput = expandWarmThroughput(v.([]any)[0].(map[string]any))
		}

		_, err := tfresource.RetryWhen(ctx, createTableTimeout, func(ctx context.Context) (any, error) {
			return conn.CreateTable(ctx, input)
		}, func(err error) (bool, error) {
			if tfawserr.ErrCodeEquals(err, errCodeThrottlingException) {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.LimitExceededException](err, "can be created, updated, or deleted simultaneously") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.LimitExceededException](err, "indexed tables that can be created simultaneously") {
				return true, err
			}
			return false, err
		})

		if err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, tableName, err)
		}
	}

	d.SetId(tableName)

	var output *awstypes.TableDescription
	var err error
	if output, err = waitTableActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForCreation, resNameTable, d.Id(), err)
	}
	if err := waitTableWarmThroughputActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForUpdate, resNameTable, d.Id(), err)
	}

	if v, ok := d.GetOk("global_secondary_index"); ok {
		gsiSet := v.(*schema.Set)

		for _, gsiObject := range gsiSet.List() {
			gsi := gsiObject.(map[string]any)

			if _, err := waitGSIActive(ctx, conn, d.Id(), gsi[names.AttrName].(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForCreation, resNameTable, d.Id(), fmt.Errorf("GSI (%s): %w", gsi[names.AttrName].(string), err))
			}
		}
	}

	if d.Get("ttl.0.enabled").(bool) {
		if err := updateTimeToLive(ctx, conn, d.Id(), d.Get("ttl").([]any), d.Timeout(schema.TimeoutCreate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, d.Id(), fmt.Errorf("enabling TTL: %w", err))
		}
	}

	if d.Get("point_in_time_recovery.0.enabled").(bool) {
		if err := updatePITR(ctx, conn, d.Id(), true, aws.Int32(int32(d.Get("point_in_time_recovery.0.recovery_period_in_days").(int))), meta.(*conns.AWSClient).Region(ctx), d.Timeout(schema.TimeoutCreate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, d.Id(), fmt.Errorf("enabling point in time recovery: %w", err))
		}
	}

	if v := d.Get("replica").(*schema.Set); v.Len() > 0 {
		if err := createReplicas(ctx, conn, d.Id(), v.List(), expandGlobalTableWitness(d.Get("global_table_witness")), true, d.Timeout(schema.TimeoutCreate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, d.Id(), fmt.Errorf("replicas: %w", err))
		}

		if err := updateReplicaTags(ctx, conn, aws.ToString(output.TableArn), v.List(), keyValueTags(ctx, getTagsIn(ctx))); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, d.Id(), fmt.Errorf("replica tags: %w", err))
		}
	}

	return append(diags, resourceTableRead(ctx, d, meta)...)
}

func resourceTableRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.DynamoDBClient(ctx)

	table, err := findTableByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		create.LogNotFoundRemoveState(names.DynamoDB, create.ErrActionReading, resNameTable, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, resNameTable, d.Id(), err)
	}

	d.Set(names.AttrARN, table.TableArn)
	d.Set(names.AttrName, table.TableName)

	if table.BillingModeSummary != nil {
		d.Set("billing_mode", table.BillingModeSummary.BillingMode)
	} else {
		d.Set("billing_mode", awstypes.BillingModeProvisioned)
	}

	d.Set("deletion_protection_enabled", table.DeletionProtectionEnabled)

	if table.ProvisionedThroughput != nil {
		d.Set("write_capacity", table.ProvisionedThroughput.WriteCapacityUnits)
		d.Set("read_capacity", table.ProvisionedThroughput.ReadCapacityUnits)
	}

	if err := d.Set("attribute", flattenTableAttributeDefinitions(table.AttributeDefinitions)); err != nil {
		return create.AppendDiagSettingError(diags, names.DynamoDB, resNameTable, d.Id(), "attribute", err)
	}

	for _, attribute := range table.KeySchema {
		if attribute.KeyType == awstypes.KeyTypeHash {
			d.Set("hash_key", attribute.AttributeName)
		}

		if attribute.KeyType == awstypes.KeyTypeRange {
			d.Set("range_key", attribute.AttributeName)
		}
	}

	if err := d.Set("local_secondary_index", flattenTableLocalSecondaryIndex(table.LocalSecondaryIndexes)); err != nil {
		return create.AppendDiagSettingError(diags, names.DynamoDB, resNameTable, d.Id(), "local_secondary_index", err)
	}

	if err := d.Set("global_secondary_index", flattenTableGlobalSecondaryIndex(table.GlobalSecondaryIndexes)); err != nil {
		return create.AppendDiagSettingError(diags, names.DynamoDB, resNameTable, d.Id(), "global_secondary_index", err)
	}

	if err := d.Set("global_table_witness", flattenGlobalTableWitnesses(table.GlobalTableWitnesses)); err != nil {
		return create.AppendDiagSettingError(diags, names.DynamoDB, resNameTable, d.Id(), "global_table_witness", err)
	}

	if err := d.Set("on_demand_throughput", flattenOnDemandThroughput(table.OnDemandThroughput)); err != nil {
		return create.AppendDiagSettingError(diags, names.DynamoDB, resNameTable, d.Id(), "on_demand_throughput", err)
	}

	if table.StreamSpecification != nil {
		d.Set("stream_enabled", table.StreamSpecification.StreamEnabled)
		d.Set("stream_view_type", table.StreamSpecification.StreamViewType)
	} else {
		d.Set("stream_enabled", false)
		d.Set("stream_view_type", d.Get("stream_view_type").(string))
	}

	d.Set(names.AttrStreamARN, table.LatestStreamArn)
	d.Set("stream_label", table.LatestStreamLabel)

	sse := flattenTableServerSideEncryption(table.SSEDescription)
	sse = clearSSEDefaultKey(ctx, c, sse)

	if err := d.Set("server_side_encryption", sse); err != nil {
		return create.AppendDiagSettingError(diags, names.DynamoDB, resNameTable, d.Id(), "server_side_encryption", err)
	}

	replicas := flattenReplicaDescriptions(table.Replicas)

	if replicas, err = addReplicaPITRs(ctx, conn, d.Id(), replicas); err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, resNameTable, d.Id(), err)
	}

	if replicas, err = enrichReplicas(ctx, conn, aws.ToString(table.TableArn), d.Id(), replicas); err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, resNameTable, d.Id(), err)
	}

	replicas = addReplicaTagPropagates(d.Get("replica").(*schema.Set), replicas)
	replicas = clearReplicaDefaultKeys(ctx, c, replicas)

	mrc := awstypes.MultiRegionConsistencyEventual
	if table.MultiRegionConsistency != "" {
		mrc = table.MultiRegionConsistency
	}
	replicas = addReplicaMultiRegionConsistency(replicas, mrc)

	if err := d.Set("replica", replicas); err != nil {
		return create.AppendDiagSettingError(diags, names.DynamoDB, resNameTable, d.Id(), "replica", err)
	}

	if table.TableClassSummary != nil {
		d.Set("table_class", table.TableClassSummary.TableClass)
	} else {
		d.Set("table_class", awstypes.TableClassStandard)
	}

	if err := d.Set("warm_throughput", flattenTableWarmThroughput(table.WarmThroughput)); err != nil {
		return create.AppendDiagSettingError(diags, names.DynamoDB, resNameTable, d.Id(), "warm_throughput", err)
	}

	describeBackupsInput := dynamodb.DescribeContinuousBackupsInput{
		TableName: aws.String(d.Id()),
	}
	pitrOut, err := conn.DescribeContinuousBackups(ctx, &describeBackupsInput)

	// When a Table is `ARCHIVED`, DescribeContinuousBackups returns `TableNotFoundException`
	if err != nil && !tfawserr.ErrCodeEquals(err, errCodeUnknownOperationException, errCodeTableNotFoundException) {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, resNameTable, d.Id(), fmt.Errorf("continuous backups: %w", err))
	}

	if err := d.Set("point_in_time_recovery", flattenPITR(pitrOut)); err != nil {
		return create.AppendDiagSettingError(diags, names.DynamoDB, resNameTable, d.Id(), "point_in_time_recovery", err)
	}

	describeTTLInput := dynamodb.DescribeTimeToLiveInput{
		TableName: aws.String(d.Id()),
	}
	ttlOut, err := conn.DescribeTimeToLive(ctx, &describeTTLInput)

	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, resNameTable, d.Id(), fmt.Errorf("TTL: %w", err))
	}

	if err := d.Set("ttl", flattenTTL(ttlOut)); err != nil {
		return create.AppendDiagSettingError(diags, names.DynamoDB, resNameTable, d.Id(), "ttl", err)
	}

	return diags
}

func resourceTableUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	o, n := d.GetChange("billing_mode")
	newBillingMode, oldBillingMode := awstypes.BillingMode(n.(string)), awstypes.BillingMode(o.(string))

	// Global Secondary Index operations must occur in multiple phases
	// to prevent various error scenarios. If there are no detected required
	// updates in the Terraform configuration, later validation or API errors
	// will signal the problems.
	var gsiUpdates []awstypes.GlobalSecondaryIndexUpdate

	if d.HasChange("global_secondary_index") {
		var err error
		o, n := d.GetChange("global_secondary_index")
		gsiUpdates, err = updateDiffGSI(o.(*schema.Set).List(), n.(*schema.Set).List(), newBillingMode)

		if err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTable, d.Id(), fmt.Errorf("computing GSI difference: %w", err))
		}
	}

	// Phase 1 of Global Secondary Index Operations: Delete Only
	//  * Delete indexes first to prevent error when simultaneously updating
	//    BillingMode to PROVISIONED, which requires updating index
	//    ProvisionedThroughput first, but we have no definition
	//  * Only 1 online index can be deleted simultaneously per table
	for _, gsiUpdate := range gsiUpdates {
		if gsiUpdate.Delete == nil {
			continue
		}

		idxName := aws.ToString(gsiUpdate.Delete.IndexName)
		input := &dynamodb.UpdateTableInput{
			GlobalSecondaryIndexUpdates: []awstypes.GlobalSecondaryIndexUpdate{gsiUpdate},
			TableName:                   aws.String(d.Id()),
		}

		if _, err := conn.UpdateTable(ctx, input); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionDeleting, resNameTable, d.Id(), fmt.Errorf("GSI (%s): %w", idxName, err))
		}

		if _, err := waitGSIDeleted(ctx, conn, d.Id(), idxName, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForDeletion, resNameTable, d.Id(), fmt.Errorf("GSI (%s): %w", idxName, err))
		}
	}

	// Table Class cannot be changed concurrently with other values
	if d.HasChange("table_class") {
		input := dynamodb.UpdateTableInput{
			TableClass: awstypes.TableClass(d.Get("table_class").(string)),
			TableName:  aws.String(d.Id()),
		}
		_, err := conn.UpdateTable(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DynamoDB Table (%s) table class: %s", d.Id(), err)
		}
		if _, err := waitTableActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DynamoDB Table (%s) table class: waiting for completion: %s", d.Id(), err)
		}
	}

	hasTableUpdate := false
	input := &dynamodb.UpdateTableInput{
		TableName: aws.String(d.Id()),
	}

	if d.HasChanges("billing_mode", "read_capacity", "write_capacity") {
		hasTableUpdate = true

		capacityMap := map[string]any{
			"write_capacity": d.Get("write_capacity"),
			"read_capacity":  d.Get("read_capacity"),
		}

		input.BillingMode = newBillingMode
		input.ProvisionedThroughput = expandProvisionedThroughputUpdate(d.Id(), capacityMap, newBillingMode, oldBillingMode)
	}

	if d.HasChange("deletion_protection_enabled") {
		hasTableUpdate = true
		input.DeletionProtectionEnabled = aws.Bool(d.Get("deletion_protection_enabled").(bool))
	}

	if d.HasChange("on_demand_throughput") {
		hasTableUpdate = true
		if v, ok := d.GetOk("on_demand_throughput"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.OnDemandThroughput = expandOnDemandThroughput(v.([]any)[0].(map[string]any))
		}
	}

	// make change when
	//   stream_enabled has change (below) OR
	//   stream_view_type has change and stream_enabled is true (special case)
	if !d.HasChange("stream_enabled") && d.HasChange("stream_view_type") {
		if v, ok := d.Get("stream_enabled").(bool); ok && v {
			// in order to change stream view type:
			//   1) stream have already been enabled, and
			//   2) it must be disabled and then reenabled (otherwise, ValidationException: Table already has an enabled stream)
			if err := cycleStreamEnabled(ctx, conn, d.Id(), awstypes.StreamViewType(d.Get("stream_view_type").(string)), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTable, d.Id(), err)
			}
		}
	}

	if d.HasChange("stream_enabled") {
		hasTableUpdate = true

		input.StreamSpecification = &awstypes.StreamSpecification{
			StreamEnabled: aws.Bool(d.Get("stream_enabled").(bool)),
		}
		if d.Get("stream_enabled").(bool) {
			input.StreamSpecification.StreamViewType = awstypes.StreamViewType(d.Get("stream_view_type").(string))
		}
	}

	// Phase 2 of Global Secondary Index Operations: Update Only
	// Cannot create or delete index while updating table ProvisionedThroughput
	// Must skip all index updates when switching BillingMode from PROVISIONED to PAY_PER_REQUEST
	// Must update all indexes when switching BillingMode from PAY_PER_REQUEST to PROVISIONED
	if newBillingMode == awstypes.BillingModeProvisioned {
		for _, gsiUpdate := range gsiUpdates {
			if gsiUpdate.Update == nil || (gsiUpdate.Update != nil && gsiUpdate.Update.WarmThroughput != nil) {
				continue
			}

			hasTableUpdate = true
			input.GlobalSecondaryIndexUpdates = append(input.GlobalSecondaryIndexUpdates, gsiUpdate)
		}
	}

	// update only on-demand throughput indexes when switching to PAY_PER_REQUEST in Phase 2a
	if newBillingMode == awstypes.BillingModePayPerRequest {
		for _, gsiUpdate := range gsiUpdates {
			if gsiUpdate.Update == nil || (gsiUpdate.Update != nil && gsiUpdate.Update.OnDemandThroughput == nil) {
				continue
			}

			hasTableUpdate = true
			input.GlobalSecondaryIndexUpdates = append(input.GlobalSecondaryIndexUpdates, gsiUpdate)
		}
	}

	if hasTableUpdate {
		_, err := conn.UpdateTable(ctx, input)

		if err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTable, d.Id(), err)
		}

		if _, err := waitTableActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForUpdate, resNameTable, d.Id(), err)
		}

		if err := waitTableWarmThroughputActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForUpdate, resNameTable, d.Id(), err)
		}

		for _, gsiUpdate := range gsiUpdates {
			if gsiUpdate.Update == nil {
				continue
			}

			idxName := aws.ToString(gsiUpdate.Update.IndexName)

			if _, err := waitGSIActive(ctx, conn, d.Id(), idxName, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForUpdate, resNameTable, d.Id(), fmt.Errorf("GSI (%s): %w", idxName, err))
			}

			if err := waitGSIWarmThroughputActive(ctx, conn, d.Id(), idxName, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForUpdate, resNameTable, d.Id(), fmt.Errorf("GSI (%s): %w", idxName, err))
			}
		}
	}

	// Phase 2b: update indexes in two steps when warm throughput is set
	for _, gsiUpdate := range gsiUpdates {
		if gsiUpdate.Update == nil || (gsiUpdate.Update != nil && gsiUpdate.Update.WarmThroughput == nil) {
			continue
		}

		idxName := aws.ToString(gsiUpdate.Update.IndexName)
		input := &dynamodb.UpdateTableInput{
			GlobalSecondaryIndexUpdates: []awstypes.GlobalSecondaryIndexUpdate{gsiUpdate},
			TableName:                   aws.String(d.Id()),
		}

		if _, err := conn.UpdateTable(ctx, input); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTable, d.Id(), fmt.Errorf("updating GSI for warm throughput (%s): %w", idxName, err))
		}

		if _, err := waitGSIActive(ctx, conn, d.Id(), idxName, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTable, d.Id(), fmt.Errorf("%s GSI (%s): %w", create.ErrActionWaitingForCreation, idxName, err))
		}

		if err := waitGSIWarmThroughputActive(ctx, conn, d.Id(), idxName, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForUpdate, resNameTable, d.Id(), fmt.Errorf("GSI (%s): %w", idxName, err))
		}
	}

	// Phase 3 of Global Secondary Index Operations: Create Only
	// Only 1 online index can be created simultaneously per table
	for _, gsiUpdate := range gsiUpdates {
		if gsiUpdate.Create == nil {
			continue
		}

		idxName := aws.ToString(gsiUpdate.Create.IndexName)
		input := &dynamodb.UpdateTableInput{
			AttributeDefinitions:        expandAttributes(d.Get("attribute").(*schema.Set).List()),
			GlobalSecondaryIndexUpdates: []awstypes.GlobalSecondaryIndexUpdate{gsiUpdate},
			TableName:                   aws.String(d.Id()),
		}

		if _, err := conn.UpdateTable(ctx, input); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTable, d.Id(), fmt.Errorf("creating GSI (%s): %w", idxName, err))
		}

		if _, err := waitGSIActive(ctx, conn, d.Id(), idxName, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTable, d.Id(), fmt.Errorf("%s GSI (%s): %w", create.ErrActionWaitingForCreation, idxName, err))
		}

		if err := waitGSIWarmThroughputActive(ctx, conn, d.Id(), idxName, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForUpdate, resNameTable, d.Id(), fmt.Errorf("GSI (%s): %w", idxName, err))
		}
	}

	if d.HasChange("server_side_encryption") {
		if replicas, sseSpecification := d.Get("replica").(*schema.Set), expandEncryptAtRestOptions(d.Get("server_side_encryption").([]any)); replicas.Len() > 0 && sseSpecification.KMSMasterKeyId != nil {
			log.Printf("[DEBUG] Using SSE update on replicas")
			var replicaUpdates []awstypes.ReplicationGroupUpdate
			for _, replica := range replicas.List() {
				tfMap, ok := replica.(map[string]any)
				if !ok {
					continue
				}

				region, ok := tfMap["region_name"].(string)
				if !ok {
					continue
				}

				key, ok := tfMap[names.AttrKMSKeyARN].(string)
				if !ok || key == "" {
					continue
				}

				var input = &awstypes.UpdateReplicationGroupMemberAction{
					RegionName:     aws.String(region),
					KMSMasterKeyId: aws.String(key),
				}
				var update = awstypes.ReplicationGroupUpdate{Update: input}
				replicaUpdates = append(replicaUpdates, update)
			}
			var updateAction = awstypes.UpdateReplicationGroupMemberAction{
				KMSMasterKeyId: sseSpecification.KMSMasterKeyId,
				RegionName:     aws.String(meta.(*conns.AWSClient).Region(ctx)),
			}
			var update = awstypes.ReplicationGroupUpdate{
				Update: &updateAction,
			}
			replicaUpdates = append(replicaUpdates, update)
			input := dynamodb.UpdateTableInput{
				TableName:      aws.String(d.Id()),
				ReplicaUpdates: replicaUpdates,
			}
			_, err := conn.UpdateTable(ctx, &input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating DynamoDB Table (%s) SSE: %s", d.Id(), err)
			}
		} else {
			log.Printf("[DEBUG] Using normal update for SSE")
			input := dynamodb.UpdateTableInput{
				TableName:        aws.String(d.Id()),
				SSESpecification: sseSpecification,
			}
			_, err := conn.UpdateTable(ctx, &input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating DynamoDB Table (%s) SSE: %s", d.Id(), err)
			}
		}

		// since we don't update replicas unless there is a KMS key, we need to wait for replica
		// updates for the scenario where 1) there are replicas, 2) we are updating SSE (such as
		// disabling), and 3) we have no KMS key
		if replicas := d.Get("replica").(*schema.Set); replicas.Len() > 0 {
			var replicaRegions []string
			for _, replica := range replicas.List() {
				tfMap, ok := replica.(map[string]any)
				if !ok {
					continue
				}
				if v, ok := tfMap["region_name"].(string); ok {
					replicaRegions = append(replicaRegions, v)
				}
			}
			for _, region := range replicaRegions {
				if _, err := waitReplicaSSEUpdated(ctx, conn, region, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for DynamoDB Table (%s) replica SSE update in region %q: %s", d.Id(), region, err)
				}
			}
		}

		if _, err := waitSSEUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for DynamoDB Table (%s) SSE update: %s", d.Id(), err)
		}
	}

	if d.HasChange("ttl") {
		if err := updateTimeToLive(ctx, conn, d.Id(), d.Get("ttl").([]any), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTable, d.Id(), err)
		}
	}

	replicaTagsChange := false
	if d.HasChange("replica") {
		replicaTagsChange = true

		if err := updateReplica(ctx, conn, d); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTable, d.Id(), err)
		}
	}

	if d.HasChange(names.AttrTagsAll) {
		replicaTagsChange = true
	}

	if replicaTagsChange {
		if v, ok := d.Get("replica").(*schema.Set); ok && v.Len() > 0 {
			if err := updateReplicaTags(ctx, conn, d.Get(names.AttrARN).(string), v.List(), d.Get(names.AttrTagsAll)); err != nil {
				return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTable, d.Id(), err)
			}
		}
	}

	if d.HasChange("point_in_time_recovery") {
		if err := updatePITR(ctx, conn, d.Id(), d.Get("point_in_time_recovery.0.enabled").(bool), aws.Int32(int32(d.Get("point_in_time_recovery.0.recovery_period_in_days").(int))), meta.(*conns.AWSClient).Region(ctx), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTable, d.Id(), err)
		}
	}

	if d.HasChange("warm_throughput") {
		if err := updateWarmThroughput(ctx, conn, d.Get("warm_throughput").([]any), d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTable, d.Id(), err)
		}
	}

	return append(diags, resourceTableRead(ctx, d, meta)...)
}

func resourceTableDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	if replicas := d.Get("replica").(*schema.Set).List(); len(replicas) > 0 {
		log.Printf("[DEBUG] Deleting DynamoDB Table replicas: %s", d.Id())
		if err := deleteReplicas(ctx, conn, d.Id(), replicas, expandGlobalTableWitness(d.Get("global_table_witness")), d.Timeout(schema.TimeoutDelete)); err != nil {
			// ValidationException: Replica specified in the Replica Update or Replica Delete action of the request was not found.
			// ValidationException: Cannot add, delete, or update the local region through ReplicaUpdates. Use CreateTable, DeleteTable, or UpdateTable as required.
			if !tfawserr.ErrMessageContains(err, errCodeValidationException, "request was not found") &&
				!tfawserr.ErrMessageContains(err, errCodeValidationException, "MultiRegionConsistency must be set as STRONG when GlobalTableWitnessUpdates parameter is present") &&
				!tfawserr.ErrMessageContains(err, errCodeValidationException, "Cannot add, delete, or update the local region through ReplicaUpdates") {
				return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionDeleting, resNameTable, d.Id(), err)
			}
		}
	}

	log.Printf("[DEBUG] Deleting DynamoDB Table: %s", d.Id())
	err := deleteTable(ctx, conn, d.Id())

	if errs.IsAErrorMessageContains[*awstypes.ResourceNotFoundException](err, "Requested resource not found: Table: ") {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionDeleting, resNameTable, d.Id(), err)
	}

	if _, err := waitTableDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForDeletion, resNameTable, d.Id(), err)
	}

	return diags
}

// custom diff

func isTableOptionDisabled(v any) bool {
	options := v.([]any)
	if len(options) == 0 {
		return true
	}
	e := options[0].(map[string]any)[names.AttrEnabled]
	return !e.(bool)
}

// CRUD helpers

// cycleStreamEnabled disables the stream and then re-enables it with streamViewType
func cycleStreamEnabled(ctx context.Context, conn *dynamodb.Client, id string, streamViewType awstypes.StreamViewType, timeout time.Duration) error {
	input := &dynamodb.UpdateTableInput{
		TableName: aws.String(id),
	}
	input.StreamSpecification = &awstypes.StreamSpecification{
		StreamEnabled: aws.Bool(false),
	}

	_, err := conn.UpdateTable(ctx, input)

	if err != nil {
		return fmt.Errorf("cycling stream enabled: %w", err)
	}

	if _, err := waitTableActive(ctx, conn, id, timeout); err != nil {
		return fmt.Errorf("waiting for stream cycle: %w", err)
	}

	input.StreamSpecification = &awstypes.StreamSpecification{
		StreamEnabled:  aws.Bool(true),
		StreamViewType: streamViewType,
	}

	_, err = conn.UpdateTable(ctx, input)

	if err != nil {
		return fmt.Errorf("cycling stream enabled: %w", err)
	}

	if _, err := waitTableActive(ctx, conn, id, timeout); err != nil {
		return fmt.Errorf("waiting for stream cycle: %w", err)
	}

	return nil
}

func createReplicas(ctx context.Context, conn *dynamodb.Client, tableName string, tfList []any, globalTableWitnessRegionName string, create bool, timeout time.Duration) error {
	// Duplicating this for MRSC Adoption. If using MRSC and CreateReplicationGroupMemberAction list isn't initiated for at least 2 replicas
	// then the update table action will fail with
	// "Unsupported table replica count for global tables with MultiRegionConsistency set to STRONG"
	// If this logic can be consolidated for regular Replica creation then this can be refactored
	numReplicas := len(tfList)
	numReplicasMRSC := 0
	useMRSC := false
	mrscInput := awstypes.MultiRegionConsistencyStrong

	for _, tfMapRaw := range tfList {
		tfMap, _ := tfMapRaw.(map[string]any)

		if v, ok := tfMap["consistency_mode"].(string); ok {
			if awstypes.MultiRegionConsistency(v) == awstypes.MultiRegionConsistencyStrong {
				numReplicasMRSC += 1
			}
		}
	}

	if numReplicasMRSC > 0 {
		mrscErrorMsg := "creating replicas: Using MultiRegionStrongConsistency requires exactly 2 replicas, or 1 replica and 1 witness region."
		if numReplicasMRSC > 0 && numReplicasMRSC != numReplicas {
			return errors.New(mrscErrorMsg)
		}
		if numReplicasMRSC == 1 && globalTableWitnessRegionName == "" {
			return fmt.Errorf("%s Only MRSC Replica count of 1 was provided but no Witness region was provided", mrscErrorMsg)
		}
		if numReplicasMRSC == 2 && (numReplicasMRSC == numReplicas && globalTableWitnessRegionName != "") {
			return fmt.Errorf("%s MRSC Replica count of 2 was provided and a Witness region was also provided", mrscErrorMsg)
		}
		if numReplicasMRSC > 2 {
			return fmt.Errorf("%s Too many Replicas were provided %d", mrscErrorMsg, numReplicasMRSC)
		}

		mrscInput = awstypes.MultiRegionConsistencyStrong
		useMRSC = true
	}

	// if MRSC or MREC is defined and meets the above criteria, then all replicas must be created in a single call to UpdateTable.
	if useMRSC {
		var replicaCreates []awstypes.ReplicationGroupUpdate
		for _, tfMapRaw := range tfList {
			tfMap, ok := tfMapRaw.(map[string]any)
			if !ok {
				continue
			}

			var replicaInput = &awstypes.CreateReplicationGroupMemberAction{}

			if v, ok := tfMap["region_name"].(string); ok && v != "" {
				replicaInput.RegionName = aws.String(v)
			}

			if v, ok := tfMap[names.AttrKMSKeyARN].(string); ok && v != "" {
				replicaInput.KMSMasterKeyId = aws.String(v)
			}

			replicaCreates = append(replicaCreates, awstypes.ReplicationGroupUpdate{
				Create: replicaInput,
			})
		}

		var gtgwu []awstypes.GlobalTableWitnessGroupUpdate
		if globalTableWitnessRegionName != "" {
			var cgtwgma = awstypes.CreateGlobalTableWitnessGroupMemberAction{
				RegionName: aws.String(globalTableWitnessRegionName),
			}
			gtgwu = append(gtgwu, awstypes.GlobalTableWitnessGroupUpdate{
				Create: &cgtwgma,
			})
		}
		input := dynamodb.UpdateTableInput{
			GlobalTableWitnessUpdates: gtgwu,
			MultiRegionConsistency:    mrscInput,
			ReplicaUpdates:            replicaCreates,
			TableName:                 aws.String(tableName),
		}

		err := tfresource.Retry(ctx, max(replicaUpdateTimeout, timeout), func(ctx context.Context) *tfresource.RetryError {
			_, err := conn.UpdateTable(ctx, &input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, errCodeThrottlingException) {
					return tfresource.RetryableError(err)
				}
				if errs.IsAErrorMessageContains[*awstypes.LimitExceededException](err, "can be created.") {
					return tfresource.NonRetryableError(err)
				}
				if tfawserr.ErrMessageContains(err, errCodeValidationException, "Replica specified in the Replica Update or Replica Delete action of the request was not found") {
					return tfresource.RetryableError(err)
				}
				if errs.IsA[*awstypes.ResourceInUseException](err) {
					return tfresource.RetryableError(err)
				}

				return tfresource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return err
		}

		for _, tfMapRaw := range tfList {
			tfMap, ok := tfMapRaw.(map[string]any)
			if !ok {
				continue
			}

			log.Printf("[DEBUG] Waiting for replica to be active in region (%s)\n", tfMap["region_name"])
			if _, err := waitReplicaActive(ctx, conn, tableName, tfMap["region_name"].(string), timeout, replicaDelayDefault); err != nil {
				return fmt.Errorf("waiting for replica (%s) creation: %w", tfMap["region_name"].(string), err)
			}

			// pitr
			if err = updatePITR(ctx, conn, tableName, tfMap["point_in_time_recovery"].(bool), nil, tfMap["region_name"].(string), timeout); err != nil {
				return fmt.Errorf("updating replica (%s) point in time recovery: %w", tfMap["region_name"].(string), err)
			}
		}
	} else {
		for _, tfMapRaw := range tfList {
			tfMap, ok := tfMapRaw.(map[string]any)

			if !ok {
				continue
			}
			var replicaInput = &awstypes.CreateReplicationGroupMemberAction{}

			if v, ok := tfMap["region_name"].(string); ok && v != "" {
				replicaInput.RegionName = aws.String(v)
			}

			if v, ok := tfMap[names.AttrKMSKeyARN].(string); ok && v != "" {
				replicaInput.KMSMasterKeyId = aws.String(v)
			}

			input := &dynamodb.UpdateTableInput{
				TableName: aws.String(tableName),
				ReplicaUpdates: []awstypes.ReplicationGroupUpdate{
					{
						Create: replicaInput,
					},
				},
			}

			// currently this would not be needed because (replica has these arguments):
			//   region_name can't be updated - new replica
			//   kms_key_arn can't be updated - remove/add replica
			//   propagate_tags - handled elsewhere
			//   point_in_time_recovery - handled elsewhere
			// if provisioned_throughput_override or table_class_override were added, they could be updated here
			if !create {
				var replicaInput = &awstypes.UpdateReplicationGroupMemberAction{}
				if v, ok := tfMap["region_name"].(string); ok && v != "" {
					replicaInput.RegionName = aws.String(v)
				}

				if v, ok := tfMap[names.AttrKMSKeyARN].(string); ok && v != "" {
					replicaInput.KMSMasterKeyId = aws.String(v)
				}

				input = &dynamodb.UpdateTableInput{
					TableName: aws.String(tableName),
					ReplicaUpdates: []awstypes.ReplicationGroupUpdate{
						{
							Update: replicaInput,
						},
					},
				}
			}

			err := tfresource.Retry(ctx, max(replicaUpdateTimeout, timeout), func(ctx context.Context) *tfresource.RetryError {
				_, err := conn.UpdateTable(ctx, input)
				if err != nil {
					if tfawserr.ErrCodeEquals(err, errCodeThrottlingException) {
						return tfresource.RetryableError(err)
					}
					if errs.IsAErrorMessageContains[*awstypes.LimitExceededException](err, "can be created, updated, or deleted simultaneously") {
						return tfresource.RetryableError(err)
					}
					if tfawserr.ErrMessageContains(err, errCodeValidationException, "Replica specified in the Replica Update or Replica Delete action of the request was not found") {
						return tfresource.RetryableError(err)
					}
					if errs.IsA[*awstypes.ResourceInUseException](err) {
						return tfresource.RetryableError(err)
					}

					return tfresource.NonRetryableError(err)
				}
				return nil
			})

			// An update that doesn't (makes no changes) returns ValidationException
			// (same region_name and kms_key_arn as currently) throws unhelpfully worded exception:
			// ValidationException: One or more parameter values were invalid: KMSMasterKeyId must be specified for each replica.

			if create && tfawserr.ErrMessageContains(err, errCodeValidationException, "already exist") {
				return createReplicas(ctx, conn, tableName, tfList, globalTableWitnessRegionName, false, timeout)
			}

			if err != nil && !tfawserr.ErrMessageContains(err, errCodeValidationException, "no actions specified") {
				return fmt.Errorf("creating replica (%s): %w", tfMap["region_name"].(string), err)
			}

			if _, err := waitReplicaActive(ctx, conn, tableName, tfMap["region_name"].(string), timeout, replicaDelayDefault); err != nil {
				return fmt.Errorf("waiting for replica (%s) creation: %w", tfMap["region_name"].(string), err)
			}

			// pitr
			if err = updatePITR(ctx, conn, tableName, tfMap["point_in_time_recovery"].(bool), nil, tfMap["region_name"].(string), timeout); err != nil {
				return fmt.Errorf("updating replica (%s) point in time recovery: %w", tfMap["region_name"].(string), err)
			}

			if v, ok := tfMap["deletion_protection_enabled"].(bool); ok {
				if err = updateReplicaDeletionProtection(ctx, conn, tableName, tfMap["region_name"].(string), v, timeout); err != nil {
					return fmt.Errorf("updating replica (%s) deletion protection: %w", tfMap["region_name"].(string), err)
				}
			}
		}
	}
	return nil
}

func updateReplicaTags(ctx context.Context, conn *dynamodb.Client, rn string, replicas []any, newTags any) error {
	for _, tfMapRaw := range replicas {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		region, ok := tfMap["region_name"].(string)

		if !ok || region == "" {
			continue
		}

		if v, ok := tfMap[names.AttrPropagateTags].(bool); ok && v {
			optFn := func(o *dynamodb.Options) {
				o.Region = region
			}

			repARN, err := arnForNewRegion(rn, region)
			if err != nil {
				return fmt.Errorf("per region ARN for replica (%s): %w", region, err)
			}

			oldTags, err := listTags(ctx, conn, repARN, optFn)
			if err != nil {
				return fmt.Errorf("listing tags (%s): %w", repARN, err)
			}

			if err := updateTags(ctx, conn, repARN, oldTags, newTags, optFn); err != nil {
				return fmt.Errorf("updating tags: %w", err)
			}
		}
	}

	return nil
}

func updateTimeToLive(ctx context.Context, conn *dynamodb.Client, tableName string, ttlList []any, timeout time.Duration) error {
	ttlMap := ttlList[0].(map[string]any)

	input := &dynamodb.UpdateTimeToLiveInput{
		TableName: aws.String(tableName),
		TimeToLiveSpecification: &awstypes.TimeToLiveSpecification{
			AttributeName: aws.String(ttlMap["attribute_name"].(string)),
			Enabled:       aws.Bool(ttlMap[names.AttrEnabled].(bool)),
		},
	}

	if _, err := conn.UpdateTimeToLive(ctx, input); err != nil {
		return fmt.Errorf("updating Time To Live: %w", err)
	}

	log.Printf("[DEBUG] Waiting for DynamoDB Table (%s) Time to Live update to complete", tableName)

	if _, err := waitTTLUpdated(ctx, conn, tableName, ttlMap[names.AttrEnabled].(bool), timeout); err != nil {
		return fmt.Errorf("waiting for Time To Live update: %w", err)
	}

	return nil
}

func updatePITR(ctx context.Context, conn *dynamodb.Client, tableName string, enabled bool, recoveryPeriodInDays *int32, region string, timeout time.Duration) error {
	// pitr must be modified from region where the main/replica resides
	log.Printf("[DEBUG] Updating DynamoDB point in time recovery status to %v (%s)", enabled, region)
	input := &dynamodb.UpdateContinuousBackupsInput{
		TableName: aws.String(tableName),
		PointInTimeRecoverySpecification: &awstypes.PointInTimeRecoverySpecification{
			PointInTimeRecoveryEnabled: aws.Bool(enabled),
		},
	}
	if enabled && aws.ToInt32(recoveryPeriodInDays) > 0 {
		input.PointInTimeRecoverySpecification.RecoveryPeriodInDays = recoveryPeriodInDays
	}

	optFn := func(o *dynamodb.Options) {
		o.Region = region
	}
	err := tfresource.Retry(ctx, updateTableContinuousBackupsTimeout, func(ctx context.Context) *tfresource.RetryError {
		_, err := conn.UpdateContinuousBackups(ctx, input, optFn)
		if err != nil {
			// Backups are still being enabled for this newly created table
			if errs.IsAErrorMessageContains[*awstypes.ContinuousBackupsUnavailableException](err, "Backups are being enabled") {
				return tfresource.RetryableError(err)
			}
			return tfresource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("updating PITR: %w", err)
	}

	if _, err := waitPITRUpdated(ctx, conn, tableName, enabled, timeout, optFn); err != nil {
		return fmt.Errorf("waiting for PITR update: %w", err)
	}

	return nil
}

func updateReplicaDeletionProtection(ctx context.Context, conn *dynamodb.Client, tableName, region string, enabled bool, timeout time.Duration) error {
	log.Printf("[DEBUG] Updating DynamoDB deletion protection to %v (%s)", enabled, region)
	input := dynamodb.UpdateTableInput{
		TableName:                 aws.String(tableName),
		DeletionProtectionEnabled: aws.Bool(enabled),
	}

	optFn := func(o *dynamodb.Options) { o.Region = region }
	_, err := conn.UpdateTable(ctx, &input, optFn)
	if err != nil {
		return fmt.Errorf("updating deletion protection: %w", err)
	}

	if _, err := waitReplicaActive(ctx, conn, tableName, region, timeout, replicaPropagationDelay); err != nil {
		return fmt.Errorf("waiting for deletion protection update: %w", err)
	}

	return nil
}

func updateReplica(ctx context.Context, conn *dynamodb.Client, d *schema.ResourceData) error {
	oRaw, nRaw := d.GetChange("replica")
	o := oRaw.(*schema.Set)
	n := nRaw.(*schema.Set)

	removeRaw := o.Difference(n).List()
	addRaw := n.Difference(o).List()

	var removeFirst []any // replicas to delete before recreating (like ForceNew without recreating table)
	var toAdd []any
	var toRemove []any

	// first pass - add replicas that don't have corresponding remove entry
	for _, a := range addRaw {
		add := true
		ma := a.(map[string]any)
		for _, r := range removeRaw {
			mr := r.(map[string]any)

			if ma["region_name"].(string) == mr["region_name"].(string) {
				add = false
				break
			}
		}

		if add {
			toAdd = append(toAdd, ma)
		}
	}

	// second pass - remove replicas that don't have corresponding add entry
	for _, r := range removeRaw {
		remove := true
		mr := r.(map[string]any)
		for _, a := range addRaw {
			ma := a.(map[string]any)

			if ma["region_name"].(string) == mr["region_name"].(string) {
				remove = false
				break
			}
		}

		if remove {
			toRemove = append(toRemove, mr)
		}
	}

	// third pass - for replicas that exist in both add and remove
	// For true updates, don't remove and add, just update
	for _, a := range addRaw {
		ma := a.(map[string]any)
		for _, r := range removeRaw {
			mr := r.(map[string]any)

			if ma["region_name"].(string) != mr["region_name"].(string) {
				continue
			}

			// like "ForceNew" for the replica - KMS change
			if ma[names.AttrKMSKeyARN].(string) != mr[names.AttrKMSKeyARN].(string) {
				toRemove = append(toRemove, mr)
				toAdd = append(toAdd, ma)
				break
			}

			// like "ForceNew" for the replica - consistency_mode change
			if v1, ok1 := ma["consistency_mode"].(string); ok1 {
				if v2, ok2 := mr["consistency_mode"].(string); ok2 && v1 != v2 {
					toRemove = append(toRemove, mr)
					toAdd = append(toAdd, ma)
					break
				}
			}

			// just update PITR
			if ma["point_in_time_recovery"].(bool) != mr["point_in_time_recovery"].(bool) {
				if err := updatePITR(ctx, conn, d.Id(), ma["point_in_time_recovery"].(bool), nil, ma["region_name"].(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return fmt.Errorf("updating replica (%s) point in time recovery: %w", ma["region_name"].(string), err)
				}
				break
			}

			// just update deletion protection
			if ma["deletion_protection_enabled"].(bool) != mr["deletion_protection_enabled"].(bool) {
				if err := updateReplicaDeletionProtection(ctx, conn, d.Id(), ma["region_name"].(string), ma["deletion_protection_enabled"].(bool), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return fmt.Errorf("updating replica (%s) deletion protection: %w", ma["region_name"].(string), err)
				}
				break
			}

			// nothing changed, assuming propagate_tags changed so do nothing here
			break
		}
	}

	globalTableWitnessRegionName := expandGlobalTableWitness(d.Get("global_table_witness"))

	if len(removeFirst) > 0 { // mini ForceNew, recreates replica but doesn't recreate the table
		if err := deleteReplicas(ctx, conn, d.Id(), removeFirst, globalTableWitnessRegionName, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("updating replicas, while deleting: %w", err)
		}
	}

	if len(toRemove) > 0 {
		if err := deleteReplicas(ctx, conn, d.Id(), toRemove, globalTableWitnessRegionName, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("updating replicas, while deleting: %w", err)
		}
	}

	if len(toAdd) > 0 {
		if err := createReplicas(ctx, conn, d.Id(), toAdd, globalTableWitnessRegionName, true, d.Timeout(schema.TimeoutCreate)); err != nil {
			return fmt.Errorf("updating replicas, while creating: %w", err)
		}
	}

	return nil
}

func updateWarmThroughput(ctx context.Context, conn *dynamodb.Client, warmList []any, tableName string, timeout time.Duration) error {
	if len(warmList) < 1 || warmList[0] == nil {
		return nil
	}

	warmMap := warmList[0].(map[string]any)

	input := &dynamodb.UpdateTableInput{
		TableName:      aws.String(tableName),
		WarmThroughput: expandWarmThroughput(warmMap),
	}

	if _, err := conn.UpdateTable(ctx, input); err != nil {
		return err
	}

	if _, err := waitTableActive(ctx, conn, tableName, timeout); err != nil {
		return fmt.Errorf("waiting for warm throughput: %w", err)
	}

	if err := waitTableWarmThroughputActive(ctx, conn, tableName, timeout); err != nil {
		return fmt.Errorf("waiting for warm throughput: %w", err)
	}

	return nil
}

func updateDiffGSI(oldGsi, newGsi []any, billingMode awstypes.BillingMode) ([]awstypes.GlobalSecondaryIndexUpdate, error) {
	// Transform slices into maps
	oldGsis := make(map[string]any)
	for _, gsidata := range oldGsi {
		m := gsidata.(map[string]any)
		oldGsis[m[names.AttrName].(string)] = m
	}
	newGsis := make(map[string]any)
	for _, gsidata := range newGsi {
		m := gsidata.(map[string]any)
		// validate throughput input early, to avoid unnecessary processing
		if err := validateGSIProvisionedThroughput(m, billingMode); err != nil {
			return nil, err
		}
		newGsis[m[names.AttrName].(string)] = m
	}

	var ops []awstypes.GlobalSecondaryIndexUpdate

	for _, data := range newGsi {
		newMap := data.(map[string]any)
		newName := newMap[names.AttrName].(string)

		if _, exists := oldGsis[newName]; !exists {
			m := data.(map[string]any)
			idxName := m[names.AttrName].(string)

			c := awstypes.CreateGlobalSecondaryIndexAction{
				IndexName:  aws.String(idxName),
				KeySchema:  expandKeySchema(m),
				Projection: expandProjection(m),
			}

			switch billingMode {
			case awstypes.BillingModeProvisioned:
				c.ProvisionedThroughput = expandProvisionedThroughput(m, billingMode)

			case awstypes.BillingModePayPerRequest:
				if v, ok := m["on_demand_throughput"].([]any); ok && len(v) > 0 && v[0] != nil {
					c.OnDemandThroughput = expandOnDemandThroughput(v[0].(map[string]any))
				}
			}

			if v, ok := m["warm_throughput"].([]any); ok && len(v) > 0 && v[0] != nil {
				c.WarmThroughput = expandWarmThroughput(v[0].(map[string]any))
			}

			ops = append(ops, awstypes.GlobalSecondaryIndexUpdate{
				Create: &c,
			})
		}
	}

	for _, data := range oldGsi {
		oldMap := data.(map[string]any)
		oldName := oldMap[names.AttrName].(string)

		newData, exists := newGsis[oldName]
		if exists {
			newMap := newData.(map[string]any)
			idxName := newMap[names.AttrName].(string)

			oldWriteCapacity, oldReadCapacity := oldMap["write_capacity"].(int), oldMap["read_capacity"].(int)
			newWriteCapacity, newReadCapacity := newMap["write_capacity"].(int), newMap["read_capacity"].(int)
			capacityChanged := (oldWriteCapacity != newWriteCapacity || oldReadCapacity != newReadCapacity)

			var oldOnDemandThroughput *awstypes.OnDemandThroughput
			var newOnDemandThroughput *awstypes.OnDemandThroughput
			if v, ok := oldMap["on_demand_throughput"].([]any); ok && len(v) > 0 && v[0] != nil {
				oldOnDemandThroughput = expandOnDemandThroughput(v[0].(map[string]any))
			}

			if v, ok := newMap["on_demand_throughput"].([]any); ok && len(v) > 0 && v[0] != nil {
				newOnDemandThroughput = expandOnDemandThroughput(v[0].(map[string]any))
			}
			var onDemandThroughputChanged bool
			if !reflect.DeepEqual(oldOnDemandThroughput, newOnDemandThroughput) {
				onDemandThroughputChanged = true
			}

			var oldWarmThroughput *awstypes.WarmThroughput
			var newWarmThroughput *awstypes.WarmThroughput
			if v, ok := oldMap["warm_throughput"].([]any); ok && len(v) > 0 && v[0] != nil {
				oldWarmThroughput = expandWarmThroughput(v[0].(map[string]any))
			}

			if v, ok := newMap["warm_throughput"].([]any); ok && len(v) > 0 && v[0] != nil {
				newWarmThroughput = expandWarmThroughput(v[0].(map[string]any))
			}

			var warmThroughputChanged bool
			if !reflect.DeepEqual(oldWarmThroughput, newWarmThroughput) {
				warmThroughputChanged = true
			}

			var warmThroughPutDecreased bool
			if warmThroughputChanged && newWarmThroughput != nil && oldWarmThroughput != nil {
				warmThroughPutDecreased = (aws.ToInt64(newWarmThroughput.ReadUnitsPerSecond) < aws.ToInt64(oldWarmThroughput.ReadUnitsPerSecond) ||
					aws.ToInt64(newWarmThroughput.WriteUnitsPerSecond) < aws.ToInt64(oldWarmThroughput.WriteUnitsPerSecond))
			}

			// pluck non_key_attributes from oldAttributes and newAttributes as reflect.DeepEquals will compare
			// ordinal of elements in its equality (which we actually don't care about)
			nonKeyAttributesChanged := checkIfNonKeyAttributesChanged(oldMap, newMap)

			recreateAttributesChanged := checkIfGSIRecreateAttributesChanged(oldMap, newMap)

			gsiNeedsRecreate := nonKeyAttributesChanged || recreateAttributesChanged || warmThroughPutDecreased

			if gsiNeedsRecreate {
				ops = append(ops, awstypes.GlobalSecondaryIndexUpdate{
					Delete: &awstypes.DeleteGlobalSecondaryIndexAction{
						IndexName: aws.String(idxName),
					},
				})

				c := awstypes.CreateGlobalSecondaryIndexAction{
					IndexName:      aws.String(idxName),
					KeySchema:      expandKeySchema(newMap),
					Projection:     expandProjection(newMap),
					WarmThroughput: newWarmThroughput,
				}
				switch billingMode {
				case awstypes.BillingModeProvisioned:
					c.ProvisionedThroughput = expandProvisionedThroughput(newMap, billingMode)

				case awstypes.BillingModePayPerRequest:
					c.OnDemandThroughput = newOnDemandThroughput
				}
				ops = append(ops, awstypes.GlobalSecondaryIndexUpdate{
					Create: &c,
				})
			} else {
				if capacityChanged && billingMode == awstypes.BillingModeProvisioned {
					update := awstypes.GlobalSecondaryIndexUpdate{
						Update: &awstypes.UpdateGlobalSecondaryIndexAction{
							IndexName:             aws.String(idxName),
							ProvisionedThroughput: expandProvisionedThroughput(newMap, billingMode),
						},
					}
					ops = append(ops, update)
				} else if onDemandThroughputChanged && billingMode == awstypes.BillingModePayPerRequest {
					update := awstypes.GlobalSecondaryIndexUpdate{
						Update: &awstypes.UpdateGlobalSecondaryIndexAction{
							IndexName:          aws.String(idxName),
							OnDemandThroughput: newOnDemandThroughput,
						},
					}
					ops = append(ops, update)
				}
				// Only update WarmThroughput if the user has set it in the new config (not omitted or empty)
				newWarmThroughputOmitted := true
				if v, ok := newMap["warm_throughput"]; ok {
					if arr, ok := v.([]any); ok {
						if len(arr) > 0 && arr[0] != nil {
							newWarmThroughputOmitted = false
						}
					}
				}
				if warmThroughputChanged && !newWarmThroughputOmitted {
					update := awstypes.GlobalSecondaryIndexUpdate{
						Update: &awstypes.UpdateGlobalSecondaryIndexAction{
							IndexName:      aws.String(idxName),
							WarmThroughput: newWarmThroughput,
						},
					}
					ops = append(ops, update)
				}
			}
		} else {
			idxName := oldName
			ops = append(ops, awstypes.GlobalSecondaryIndexUpdate{
				Delete: &awstypes.DeleteGlobalSecondaryIndexAction{
					IndexName: aws.String(idxName),
				},
			})
		}
	}
	return ops, nil
}

// checkIfNonKeyAttributesChanged returns true if non_key_attributes between old map and new map are different
func checkIfNonKeyAttributesChanged(oldMap, newMap map[string]any) bool {
	oldNonKeyAttributes, oldNkaExists := oldMap["non_key_attributes"].(*schema.Set)
	newNonKeyAttributes, newNkaExists := newMap["non_key_attributes"].(*schema.Set)

	if oldNkaExists && newNkaExists {
		return !oldNonKeyAttributes.Equal(newNonKeyAttributes)
	}

	return oldNkaExists != newNkaExists
}

func checkIfGSIRecreateAttributesChanged(oldMap, newMap map[string]any) bool {
	oldAttributes := stripGSIUpdatableAttributes(oldMap)

	newAttributes := stripGSIUpdatableAttributes(newMap)

	oHk, oOk := oldAttributes["hash_key"]
	nHk, nOk := newAttributes["hash_key"]
	if oOk && nOk && oHk == nHk && oHk != "" {
		delete(oldAttributes, "key_schema")
		delete(newAttributes, "key_schema")
	}

	return !reflect.DeepEqual(oldAttributes, newAttributes)
}

func deleteTable(ctx context.Context, conn *dynamodb.Client, tableName string) error {
	input := &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	}

	_, err := tfresource.RetryWhen(ctx, deleteTableTimeout, func(ctx context.Context) (any, error) {
		return conn.DeleteTable(ctx, input)
	}, func(err error) (bool, error) {
		// Subscriber limit exceeded: Only 10 tables can be created, updated, or deleted simultaneously
		if errs.IsAErrorMessageContains[*awstypes.LimitExceededException](err, "simultaneously") {
			return true, err
		}
		// This handles multiple scenarios in the DynamoDB API:
		// 1. Updating a table immediately before deletion may return:
		//    ResourceInUseException: Attempt to change a resource which is still in use: Table is being updated:
		// 2. Removing a table from a DynamoDB global table may return:
		//    ResourceInUseException: Attempt to change a resource which is still in use: Table is being deleted:
		if errs.IsA[*awstypes.ResourceInUseException](err) {
			return true, err
		}

		return false, err
	})

	return err
}

func deleteReplicas(ctx context.Context, conn *dynamodb.Client, tableName string, tfList []any, globalTableWitnessRegionName string, timeout time.Duration) error {
	var g tfsync.Group

	var replicaDeletes []awstypes.ReplicationGroupUpdate
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		var regionName string

		if v, ok := tfMap["region_name"].(string); ok {
			regionName = v
		}

		if regionName == "" {
			continue
		}

		if v, ok := tfMap["consistency_mode"].(string); ok {
			if awstypes.MultiRegionConsistency(v) == awstypes.MultiRegionConsistencyStrong {
				replicaDeletes = append(replicaDeletes, awstypes.ReplicationGroupUpdate{
					Delete: &awstypes.DeleteReplicationGroupMemberAction{
						RegionName: aws.String(regionName),
					},
				})
			}
		}
	}

	// We built an array of MultiRegionStrongConsistency replicas that need deletion.
	// These need to all happen concurrently
	if len(replicaDeletes) > 0 {
		var witnessDeletes []awstypes.GlobalTableWitnessGroupUpdate
		if globalTableWitnessRegionName != "" {
			witnessDeletes = append(witnessDeletes, awstypes.GlobalTableWitnessGroupUpdate{
				Delete: &awstypes.DeleteGlobalTableWitnessGroupMemberAction{
					RegionName: aws.String(globalTableWitnessRegionName),
				},
			})
		}

		input := dynamodb.UpdateTableInput{
			GlobalTableWitnessUpdates: witnessDeletes,
			ReplicaUpdates:            replicaDeletes,
			TableName:                 aws.String(tableName),
		}
		err := tfresource.Retry(ctx, updateTableTimeout, func(ctx context.Context) *tfresource.RetryError {
			_, err := conn.UpdateTable(ctx, &input)
			notFoundRetries := 0
			if err != nil {
				if tfawserr.ErrCodeEquals(err, errCodeThrottlingException) {
					return tfresource.RetryableError(err)
				}
				if errs.IsA[*awstypes.ResourceNotFoundException](err) {
					notFoundRetries++
					if notFoundRetries > 3 {
						return tfresource.NonRetryableError(err)
					}
					return tfresource.RetryableError(err)
				}
				if errs.IsAErrorMessageContains[*awstypes.LimitExceededException](err, "can be created, updated, or deleted simultaneously") {
					return tfresource.RetryableError(err)
				}
				if errs.IsA[*awstypes.ResourceInUseException](err) {
					return tfresource.RetryableError(err)
				}

				return tfresource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil && !errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return fmt.Errorf("deleting replica(s): %w", err)
		}

		for _, tfMapRaw := range tfList {
			tfMap, ok := tfMapRaw.(map[string]any)
			if !ok {
				continue
			}
			var regionName = tfMap["region_name"].(string)
			if _, err := waitReplicaDeleted(ctx, conn, tableName, regionName, timeout); err != nil {
				return fmt.Errorf("waiting for replica (%s) deletion: %w", regionName, err)
			}
		}
		return nil
	} else {
		for _, tfMapRaw := range tfList {
			tfMap, ok := tfMapRaw.(map[string]any)

			if !ok {
				continue
			}

			var regionName string

			if v, ok := tfMap["region_name"].(string); ok {
				regionName = v
			}

			if regionName == "" {
				continue
			}

			g.Go(ctx, func(ctx context.Context) error {
				input := &dynamodb.UpdateTableInput{
					TableName: aws.String(tableName),
					ReplicaUpdates: []awstypes.ReplicationGroupUpdate{
						{
							Delete: &awstypes.DeleteReplicationGroupMemberAction{
								RegionName: aws.String(regionName),
							},
						},
					},
				}

				err := tfresource.Retry(ctx, updateTableTimeout, func(ctx context.Context) *tfresource.RetryError {
					_, err := conn.UpdateTable(ctx, input)
					notFoundRetries := 0
					if err != nil {
						if tfawserr.ErrCodeEquals(err, errCodeThrottlingException) {
							return tfresource.RetryableError(err)
						}
						if errs.IsA[*awstypes.ResourceNotFoundException](err) {
							notFoundRetries++
							if notFoundRetries > 3 {
								return tfresource.NonRetryableError(err)
							}
							return tfresource.RetryableError(err)
						}
						if errs.IsAErrorMessageContains[*awstypes.LimitExceededException](err, "can be created, updated, or deleted simultaneously") {
							return tfresource.RetryableError(err)
						}
						if errs.IsA[*awstypes.ResourceInUseException](err) {
							return tfresource.RetryableError(err)
						}

						return tfresource.NonRetryableError(err)
					}
					return nil
				})

				if err != nil && !errs.IsA[*awstypes.ResourceNotFoundException](err) {
					return fmt.Errorf("deleting replica (%s): %w", regionName, err)
				}

				if _, err := waitReplicaDeleted(ctx, conn, tableName, regionName, timeout); err != nil {
					return fmt.Errorf("waiting for replica (%s) deletion: %w", regionName, err)
				}

				return nil
			})
		}
		return g.Wait(ctx)
	}
}

func replicaPITR(ctx context.Context, conn *dynamodb.Client, tableName string, region string) (bool, error) {
	// To manage replicas you need connections from the different regions. However, they
	// have to be created from the starting/main region.
	optFn := func(o *dynamodb.Options) {
		o.Region = region
	}

	input := dynamodb.DescribeContinuousBackupsInput{
		TableName: aws.String(tableName),
	}
	pitrOut, err := conn.DescribeContinuousBackups(ctx, &input, optFn)
	// When a Table is `ARCHIVED`, DescribeContinuousBackups returns `TableNotFoundException`
	if err != nil && !tfawserr.ErrCodeEquals(err, errCodeUnknownOperationException, errCodeTableNotFoundException) {
		return false, fmt.Errorf("describing Continuous Backups: %w", err)
	}

	if pitrOut == nil {
		return false, nil
	}

	enabled := false

	if pitrOut.ContinuousBackupsDescription != nil {
		pitr := pitrOut.ContinuousBackupsDescription.PointInTimeRecoveryDescription
		if pitr != nil {
			enabled = (pitr.PointInTimeRecoveryStatus == awstypes.PointInTimeRecoveryStatusEnabled)
		}
	}

	return enabled, nil
}

func addReplicaPITRs(ctx context.Context, conn *dynamodb.Client, tableName string, replicas []any) ([]any, error) {
	// This non-standard approach is needed because PITR info for a replica
	// must come from a region-specific connection.
	for i, replicaRaw := range replicas {
		replica := replicaRaw.(map[string]any)

		var enabled bool
		var err error
		if enabled, err = replicaPITR(ctx, conn, tableName, replica["region_name"].(string)); err != nil {
			return nil, err
		}
		replica["point_in_time_recovery"] = enabled
		replicas[i] = replica
	}

	return replicas, nil
}

func addReplicaMultiRegionConsistency(replicas []any, mrc awstypes.MultiRegionConsistency) []any {
	// This non-standard approach is needed because MRSC info for a replica
	// comes from the Table object
	for i, tfMapRaw := range replicas {
		tfMap := tfMapRaw.(map[string]any)
		tfMap["consistency_mode"] = string(mrc)
		replicas[i] = tfMap
	}

	return replicas
}

func enrichReplicas(ctx context.Context, conn *dynamodb.Client, arn, tableName string, tfList []any) ([]any, error) {
	// This non-standard approach is needed because PITR info for a replica
	// must come from a region-specific connection.
	for i, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		newARN, err := arnForNewRegion(arn, tfMap["region_name"].(string))
		if err != nil {
			return nil, fmt.Errorf("creating new-region ARN: %w", err)
		}
		tfMap[names.AttrARN] = newARN

		optFn := func(o *dynamodb.Options) {
			o.Region = tfMap["region_name"].(string)
		}
		table, err := findTableByName(ctx, conn, tableName, optFn)
		if err != nil {
			log.Printf("[WARN] When attempting to get replica (%s) stream information, ignoring encountered error: %s", tableName, err)
			continue
		}

		tfMap["deletion_protection_enabled"] = aws.ToBool(table.DeletionProtectionEnabled)
		if table.SSEDescription != nil {
			tfMap[names.AttrKMSKeyARN] = aws.ToString(table.SSEDescription.KMSMasterKeyArn)
		}
		tfMap[names.AttrStreamARN] = aws.ToString(table.LatestStreamArn)
		tfMap["stream_label"] = aws.ToString(table.LatestStreamLabel)

		tfList[i] = tfMap
	}

	return tfList, nil
}

func addReplicaTagPropagates(configReplicas *schema.Set, replicas []any) []any {
	if configReplicas.Len() == 0 {
		return replicas
	}

	l := configReplicas.List()

	for i, replicaRaw := range replicas {
		replica := replicaRaw.(map[string]any)

		prop := false

		for _, configReplicaRaw := range l {
			configReplica := configReplicaRaw.(map[string]any)

			if v, ok := configReplica["region_name"].(string); ok && v != replica["region_name"].(string) {
				continue
			}

			if v, ok := configReplica[names.AttrPropagateTags].(bool); ok && v {
				prop = true
				break
			}
		}
		replica[names.AttrPropagateTags] = prop
		replicas[i] = replica
	}

	return replicas
}

// clearSSEDefaultKey sets the kms_key_arn to "" if it is the default key alias/aws/dynamodb.
// Not clearing the key causes diff problems and sends the key to AWS when it should not be.
func clearSSEDefaultKey(ctx context.Context, client *conns.AWSClient, sseList []any) []any {
	if len(sseList) == 0 {
		return sseList
	}

	sse := sseList[0].(map[string]any)

	dk, err := kms.FindDefaultKeyARNForService(ctx, client.KMSClient(ctx), "dynamodb", client.Region(ctx))
	if err != nil {
		return sseList
	}

	if v, ok := sse[names.AttrKMSKeyARN].(string); ok && v == dk {
		sse[names.AttrKMSKeyARN] = ""
		return []any{sse}
	}

	return sseList
}

// clearReplicaDefaultKeys sets a replica's kms_key_arn to "" if it is the default key alias/aws/dynamodb for
// the replica's region. Not clearing the key causes diff problems and sends the key to AWS when it should not be.
func clearReplicaDefaultKeys(ctx context.Context, client *conns.AWSClient, replicas []any) []any {
	if len(replicas) == 0 {
		return replicas
	}

	for i, replicaRaw := range replicas {
		replica := replicaRaw.(map[string]any)

		if v, ok := replica[names.AttrKMSKeyARN].(string); !ok || v == "" {
			continue
		}

		if v, ok := replica["region_name"].(string); !ok || v == "" {
			continue
		}

		dk, err := kms.FindDefaultKeyARNForService(ctx, client.KMSClient(ctx), "dynamodb", replica["region_name"].(string))
		if err != nil {
			continue
		}

		if replica[names.AttrKMSKeyARN].(string) == dk {
			replica[names.AttrKMSKeyARN] = ""
		}

		replicas[i] = replica
	}

	return replicas
}

// flatteners, expanders

func flattenTableAttributeDefinitions(definitions []awstypes.AttributeDefinition) []any {
	if len(definitions) == 0 {
		return []any{}
	}

	var attributes []any

	for _, d := range definitions {
		m := map[string]any{
			names.AttrName: aws.ToString(d.AttributeName),
			names.AttrType: d.AttributeType,
		}

		attributes = append(attributes, m)
	}

	return attributes
}

func flattenTableLocalSecondaryIndex(lsi []awstypes.LocalSecondaryIndexDescription) []any {
	if len(lsi) == 0 {
		return []any{}
	}

	var output []any

	for _, l := range lsi {
		m := map[string]any{
			names.AttrName: aws.ToString(l.IndexName),
		}

		if l.Projection != nil {
			m["projection_type"] = l.Projection.ProjectionType
			m["non_key_attributes"] = l.Projection.NonKeyAttributes
		}

		for _, attribute := range l.KeySchema {
			if attribute.KeyType == awstypes.KeyTypeRange {
				m["range_key"] = aws.ToString(attribute.AttributeName)
			}
		}

		output = append(output, m)
	}

	return output
}

func flattenTableGlobalSecondaryIndex(gsi []awstypes.GlobalSecondaryIndexDescription) []any {
	if len(gsi) == 0 {
		return []any{}
	}

	var output []any

	for _, g := range gsi {
		gsi := make(map[string]any)

		gsi[names.AttrName] = aws.ToString(g.IndexName)

		if g.ProvisionedThroughput != nil {
			gsi["write_capacity"] = aws.ToInt64(g.ProvisionedThroughput.WriteCapacityUnits)
			gsi["read_capacity"] = aws.ToInt64(g.ProvisionedThroughput.ReadCapacityUnits)
		}

		gsi["key_schema"] = flattenKeySchema(g.KeySchema)

		var hashKeys []string
		var rangeKeys []string

		for _, attribute := range g.KeySchema {
			if attribute.KeyType == awstypes.KeyTypeHash {
				hashKeys = append(hashKeys, aws.ToString(attribute.AttributeName))
			}
			if attribute.KeyType == awstypes.KeyTypeRange {
				rangeKeys = append(rangeKeys, aws.ToString(attribute.AttributeName))
			}
		}

		// Set single values or lists based on count
		if len(hashKeys) == 1 {
			gsi["hash_key"] = hashKeys[0]
		}

		if len(rangeKeys) == 1 {
			gsi["range_key"] = rangeKeys[0]
		}

		if g.Projection != nil {
			gsi["projection_type"] = g.Projection.ProjectionType
			gsi["non_key_attributes"] = g.Projection.NonKeyAttributes
		}

		if g.OnDemandThroughput != nil {
			gsi["on_demand_throughput"] = flattenOnDemandThroughput(g.OnDemandThroughput)
		}

		if g.WarmThroughput != nil {
			gsi["warm_throughput"] = flattenGSIWarmThroughput(g.WarmThroughput)
		}

		output = append(output, gsi)
	}

	return output
}

func flattenKeySchema(elements []awstypes.KeySchemaElement) []map[string]any {
	result := make([]map[string]any, len(elements))

	for i, attribute := range elements {
		result[i] = map[string]any{
			"attribute_name": attribute.AttributeName,
			"key_type":       attribute.KeyType,
		}
	}

	return result
}

func flattenTableServerSideEncryption(description *awstypes.SSEDescription) []any {
	if description == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrEnabled:   description.Status == awstypes.SSEStatusEnabled,
		names.AttrKMSKeyARN: aws.ToString(description.KMSMasterKeyArn),
	}

	return []any{m}
}

func expandAttributes(cfg []any) []awstypes.AttributeDefinition {
	attributes := make([]awstypes.AttributeDefinition, len(cfg))
	for i, attribute := range cfg {
		attr := attribute.(map[string]any)
		attributes[i] = awstypes.AttributeDefinition{
			AttributeName: aws.String(attr[names.AttrName].(string)),
			AttributeType: awstypes.ScalarAttributeType(attr[names.AttrType].(string)),
		}
	}
	return attributes
}

func flattenOnDemandThroughput(apiObject *awstypes.OnDemandThroughput) []any {
	if apiObject == nil {
		return []any{}
	}

	m := map[string]any{}

	if v := apiObject.MaxReadRequestUnits; v != nil {
		m["max_read_request_units"] = aws.ToInt64(v)
	}

	if v := apiObject.MaxWriteRequestUnits; v != nil {
		m["max_write_request_units"] = aws.ToInt64(v)
	}

	return []any{m}
}

func flattenTableWarmThroughput(apiObject *awstypes.TableWarmThroughputDescription) []any {
	if apiObject == nil {
		return []any{}
	}

	// AWS may return values below the minimum when warm throughput is not actually configured
	// Also treat exact minimum values as defaults since AWS sets these automatically
	readUnits := aws.ToInt64(apiObject.ReadUnitsPerSecond)
	writeUnits := aws.ToInt64(apiObject.WriteUnitsPerSecond)

	// Return empty if values are below minimums OR exactly at minimums (AWS defaults)
	if (readUnits < 12000 && writeUnits < 4000) || (readUnits == 12000 && writeUnits == 4000) {
		return []any{}
	}

	m := map[string]any{}

	if v := apiObject.ReadUnitsPerSecond; v != nil {
		m["read_units_per_second"] = aws.ToInt64(v)
	}

	if v := apiObject.WriteUnitsPerSecond; v != nil {
		m["write_units_per_second"] = aws.ToInt64(v)
	}

	return []any{m}
}

func flattenGSIWarmThroughput(apiObject *awstypes.GlobalSecondaryIndexWarmThroughputDescription) []any {
	if apiObject == nil {
		return []any{}
	}

	// AWS may return values below the minimum when warm throughput is not actually configured
	// Also treat exact minimum values as defaults since AWS sets these automatically
	readUnits := aws.ToInt64(apiObject.ReadUnitsPerSecond)
	writeUnits := aws.ToInt64(apiObject.WriteUnitsPerSecond)

	// Return empty if values are below minimums OR exactly at minimums (AWS defaults)
	if (readUnits < 12000 && writeUnits < 4000) || (readUnits == 12000 && writeUnits == 4000) {
		return []any{}
	}

	m := map[string]any{}

	if v := apiObject.ReadUnitsPerSecond; v != nil {
		m["read_units_per_second"] = aws.ToInt64(v)
	}

	if v := apiObject.WriteUnitsPerSecond; v != nil {
		m["write_units_per_second"] = aws.ToInt64(v)
	}

	return []any{m}
}

func expandGlobalTableWitness(v any) string {
	if v == nil || len(v.([]any)) == 0 || v.([]any)[0] == nil {
		return ""
	}

	return v.([]any)[0].(map[string]any)["region_name"].(string)
}

func flattenGlobalTableWitnesses(apiObjects []awstypes.GlobalTableWitnessDescription) []any {
	if len(apiObjects) != 1 {
		return []any{}
	}

	return []any{map[string]any{
		"region_name": aws.ToString(apiObjects[0].RegionName),
	}}
}

func flattenReplicaDescription(apiObject *awstypes.ReplicaDescription) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.KMSMasterKeyId != nil {
		tfMap[names.AttrKMSKeyARN] = aws.ToString(apiObject.KMSMasterKeyId)
	}

	if apiObject.RegionName != nil {
		tfMap["region_name"] = aws.ToString(apiObject.RegionName)
	}

	return tfMap
}

func flattenReplicaDescriptions(apiObjects []awstypes.ReplicaDescription) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenReplicaDescription(&apiObject))
	}

	return tfList
}

func flattenTTL(apiObject *dynamodb.DescribeTimeToLiveOutput) []any {
	tfMap := map[string]any{
		names.AttrEnabled: false,
	}

	if apiObject == nil || apiObject.TimeToLiveDescription == nil {
		return []any{tfMap}
	}

	ttlDesc := apiObject.TimeToLiveDescription

	tfMap["attribute_name"] = aws.ToString(ttlDesc.AttributeName)
	tfMap[names.AttrEnabled] = (ttlDesc.TimeToLiveStatus == awstypes.TimeToLiveStatusEnabled)

	return []any{tfMap}
}

func flattenPITR(apiObject *dynamodb.DescribeContinuousBackupsOutput) []any {
	tfMap := map[string]any{
		names.AttrEnabled: false,
	}

	if apiObject == nil {
		return []any{tfMap}
	}

	if apiObject.ContinuousBackupsDescription != nil {
		if pitr := apiObject.ContinuousBackupsDescription.PointInTimeRecoveryDescription; pitr != nil {
			tfMap[names.AttrEnabled] = (pitr.PointInTimeRecoveryStatus == awstypes.PointInTimeRecoveryStatusEnabled)
			if pitr.PointInTimeRecoveryStatus == awstypes.PointInTimeRecoveryStatusEnabled {
				tfMap["recovery_period_in_days"] = aws.ToInt32(pitr.RecoveryPeriodInDays)
			}
		}
	}

	return []any{tfMap}
}

// TODO: Get rid of keySchemaM - the user should just explicitly define
// this in the config, we shouldn't magically be setting it like this.
// Removal will however require config change, hence BC. :/
func expandLocalSecondaryIndexes(cfg []any, keySchemaM map[string]any) []awstypes.LocalSecondaryIndex {
	indexes := make([]awstypes.LocalSecondaryIndex, len(cfg))
	for i, lsi := range cfg {
		m := lsi.(map[string]any)
		idxName := m[names.AttrName].(string)

		// TODO: See https://github.com/hashicorp/terraform-provider-aws/issues/3176
		if _, ok := m["hash_key"]; !ok {
			m["hash_key"] = keySchemaM["hash_key"]
		}

		indexes[i] = awstypes.LocalSecondaryIndex{
			IndexName:  aws.String(idxName),
			KeySchema:  expandKeySchema(m),
			Projection: expandProjection(m),
		}
	}
	return indexes
}

func expandImportTable(data map[string]any) *dynamodb.ImportTableInput {
	a := &dynamodb.ImportTableInput{
		ClientToken: aws.String(sdkid.UniqueId()),
	}

	if v, ok := data["input_compression_type"].(string); ok {
		a.InputCompressionType = awstypes.InputCompressionType(v)
	}

	if v, ok := data["input_format"].(string); ok {
		a.InputFormat = awstypes.InputFormat(v)
	}

	if v, ok := data["input_format_options"].([]any); ok && len(v) > 0 && v[0] != nil {
		a.InputFormatOptions = expandInputFormatOptions(v)
	}

	if v, ok := data["s3_bucket_source"].([]any); ok && len(v) > 0 {
		a.S3BucketSource = expandS3BucketSource(v[0].(map[string]any))
	}

	return a
}

func expandGlobalSecondaryIndex(data map[string]any, billingMode awstypes.BillingMode) *awstypes.GlobalSecondaryIndex {
	output := awstypes.GlobalSecondaryIndex{
		IndexName:  aws.String(data[names.AttrName].(string)),
		KeySchema:  expandKeySchema(data),
		Projection: expandProjection(data),
	}

	switch billingMode {
	case awstypes.BillingModeProvisioned:
		output.ProvisionedThroughput = expandProvisionedThroughput(data, billingMode)

	case awstypes.BillingModePayPerRequest:
		if v, ok := data["on_demand_throughput"].([]any); ok && len(v) > 0 && v[0] != nil {
			output.OnDemandThroughput = expandOnDemandThroughput(v[0].(map[string]any))
		}
	}

	if v, ok := data["warm_throughput"].([]any); ok && len(v) > 0 && v[0] != nil {
		output.WarmThroughput = expandWarmThroughput(v[0].(map[string]any))
	}

	return &output
}

func expandProvisionedThroughput(data map[string]any, billingMode awstypes.BillingMode) *awstypes.ProvisionedThroughput {
	return expandProvisionedThroughputUpdate("", data, billingMode, "")
}

func expandProvisionedThroughputUpdate(id string, data map[string]any, billingMode, oldBillingMode awstypes.BillingMode) *awstypes.ProvisionedThroughput {
	if billingMode == awstypes.BillingModePayPerRequest {
		return nil
	}

	return &awstypes.ProvisionedThroughput{
		ReadCapacityUnits:  aws.Int64(expandProvisionedThroughputField(id, data, "read_capacity", billingMode, oldBillingMode)),
		WriteCapacityUnits: aws.Int64(expandProvisionedThroughputField(id, data, "write_capacity", billingMode, oldBillingMode)),
	}
}

func expandProvisionedThroughputField(id string, data map[string]any, key string, billingMode, oldBillingMode awstypes.BillingMode) int64 {
	v := data[key].(int)
	if v == 0 && billingMode == awstypes.BillingModeProvisioned && oldBillingMode == awstypes.BillingModePayPerRequest {
		log.Printf("[WARN] Overriding %[1]s on DynamoDB Table (%[2]s) to %[3]d. Switching from billing mode %[4]q to %[5]q without value for %[1]s. Assuming changes are being ignored.",
			key, id, provisionedThroughputMinValue, oldBillingMode, billingMode)
		v = provisionedThroughputMinValue
	}
	return int64(v)
}

func expandProjection(data map[string]any) *awstypes.Projection {
	projection := &awstypes.Projection{
		ProjectionType: awstypes.ProjectionType(data["projection_type"].(string)),
	}

	if v, ok := data["non_key_attributes"].([]any); ok && len(v) > 0 {
		projection.NonKeyAttributes = flex.ExpandStringValueList(v)
	}

	if v, ok := data["non_key_attributes"].(*schema.Set); ok && v.Len() > 0 {
		projection.NonKeyAttributes = flex.ExpandStringValueSet(v)
	}

	return projection
}

func expandKeySchema(data map[string]any) []awstypes.KeySchemaElement {
	var keySchema []awstypes.KeySchemaElement

	hKeys, hKsok := data["key_schema"].([]any)
	if hKsok {
		for _, v := range hKeys {
			element := v.(map[string]any)
			keySchema = append(keySchema, awstypes.KeySchemaElement{
				AttributeName: aws.String(element["attribute_name"].(string)),
				KeyType:       awstypes.KeyType(element["key_type"].(string)),
			})
		}
	}

	hKey, hKok := data["hash_key"]
	if hKok {
		if (hKey != nil && hKey != "") && (len(hKeys) == 0) {
			// use hash_key
			keySchema = append(keySchema, awstypes.KeySchemaElement{
				AttributeName: aws.String(hKey.(string)),
				KeyType:       awstypes.KeyTypeHash,
			})
		}
	}

	rKey, rKok := data["range_key"]
	if rKok {
		if rKey != nil && rKey != "" {
			// use range_key
			keySchema = append(keySchema, awstypes.KeySchemaElement{
				AttributeName: aws.String(rKey.(string)),
				KeyType:       awstypes.KeyTypeRange,
			})
		}
	}

	return keySchema
}

func expandEncryptAtRestOptions(vOptions []any) *awstypes.SSESpecification {
	options := &awstypes.SSESpecification{}

	enabled := false
	if len(vOptions) > 0 {
		mOptions := vOptions[0].(map[string]any)

		enabled = mOptions[names.AttrEnabled].(bool)
		if enabled {
			if vKmsKeyArn, ok := mOptions[names.AttrKMSKeyARN].(string); ok && vKmsKeyArn != "" {
				options.KMSMasterKeyId = aws.String(vKmsKeyArn)
				options.SSEType = awstypes.SSETypeKms
			}
		}
	}
	options.Enabled = aws.Bool(enabled)

	return options
}

func expandInputFormatOptions(data []any) *awstypes.InputFormatOptions {
	if data == nil {
		return nil
	}

	m := data[0].(map[string]any)
	a := &awstypes.InputFormatOptions{}

	if v, ok := m["csv"].([]any); ok && len(v) > 0 {
		a.Csv = &awstypes.CsvOptions{}

		csv := v[0].(map[string]any)

		if s, ok := csv["delimiter"].(string); ok && s != "" {
			a.Csv.Delimiter = aws.String(s)
		}

		if s, ok := csv["header_list"].(*schema.Set); ok && s.Len() > 0 {
			a.Csv.HeaderList = flex.ExpandStringValueSet(s)
		}
	}

	return a
}

func expandOnDemandThroughput(tfMap map[string]any) *awstypes.OnDemandThroughput {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.OnDemandThroughput{}

	if v, ok := tfMap["max_read_request_units"].(int); ok && v != 0 {
		apiObject.MaxReadRequestUnits = aws.Int64(int64(v))
	}

	if v, ok := tfMap["max_write_request_units"].(int); ok && v != 0 {
		apiObject.MaxWriteRequestUnits = aws.Int64(int64(v))
	}

	return apiObject
}

func expandWarmThroughput(tfMap map[string]any) *awstypes.WarmThroughput {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.WarmThroughput{}

	if v, ok := tfMap["read_units_per_second"].(int); ok && v != 0 {
		apiObject.ReadUnitsPerSecond = aws.Int64(int64(v))
	}

	if v, ok := tfMap["write_units_per_second"].(int); ok && v != 0 {
		apiObject.WriteUnitsPerSecond = aws.Int64(int64(v))
	}

	return apiObject
}

func expandS3BucketSource(data map[string]any) *awstypes.S3BucketSource {
	if data == nil {
		return nil
	}

	a := &awstypes.S3BucketSource{}

	if s, ok := data[names.AttrBucket].(string); ok && s != "" {
		a.S3Bucket = aws.String(s)
	}

	if s, ok := data["bucket_owner"].(string); ok && s != "" {
		a.S3BucketOwner = aws.String(s)
	}

	if s, ok := data["key_prefix"].(string); ok && s != "" {
		a.S3KeyPrefix = aws.String(s)
	}

	return a
}

// validators

func validateTableAttributes(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	c := meta.(*conns.AWSClient)
	conn := c.DynamoDBClient(ctx)
	// Collect all indexed attributes
	indexedAttributes := map[string]bool{}
	if v, ok := d.GetOk("hash_key"); ok {
		indexedAttributes[v.(string)] = true
	}
	if v, ok := d.GetOk("range_key"); ok {
		indexedAttributes[v.(string)] = true
	}
	if v, ok := d.GetOk("local_secondary_index"); ok {
		indexes := v.(*schema.Set).List()
		for _, idx := range indexes {
			index := idx.(map[string]any)
			rangeKey := index["range_key"].(string)
			indexedAttributes[rangeKey] = true
		}
	}

	// schema.ResourceDiff.GetOk() has a bug when retrieving a list inside a set
	planRaw := d.GetRawPlan()
	if planRaw.IsKnown() && !planRaw.IsNull() {
		planGSI := planRaw.GetAttr("global_secondary_index")
		if planGSI.IsKnown() && !planGSI.IsNull() {
			for v := range tfcty.ValueElementValues(planGSI) {
				hashKey := v.GetAttr("hash_key")
				if hashKey.IsKnown() && !hashKey.IsNull() {
					indexedAttributes[hashKey.AsString()] = true
				}
				rangeKey := v.GetAttr("range_key")
				if rangeKey.IsKnown() && !rangeKey.IsNull() {
					indexedAttributes[rangeKey.AsString()] = true
				}
				keySchema := v.GetAttr("key_schema")
				if keySchema.IsKnown() && !keySchema.IsNull() {
					for v := range tfcty.ValueElementValues(keySchema) {
						name := v.GetAttr("attribute_name")
						indexedAttributes[name.AsString()] = true
					}
				}
			}
		}
	}

	// validate against remote as well, because we're using the remote state as a bridge between the table and gsi resources
	remoteGSIAttributes := map[string]bool{}
	name := planRaw.GetAttr(names.AttrName)
	if name.IsKnown() {
		table, err := findTableByName(ctx, conn, name.AsString())
		if err != nil && !retry.NotFound(err) {
			return err
		}

		if table != nil {
			for _, g := range table.GlobalSecondaryIndexes {
				for _, ks := range g.KeySchema {
					remoteGSIAttributes[aws.ToString(ks.AttributeName)] = true
					delete(indexedAttributes, aws.ToString(ks.AttributeName))
				}
			}
		}
	}

	// Check if all indexed attributes have an attribute definition
	attributes := d.Get("attribute").(*schema.Set).List()

	unindexedAttributes := []string{}
	for _, attr := range attributes {
		attribute := attr.(map[string]any)
		attrName := attribute[names.AttrName].(string)

		_, ok1 := indexedAttributes[attrName]
		_, ok2 := remoteGSIAttributes[attrName]

		if !ok1 && !ok2 {
			unindexedAttributes = append(unindexedAttributes, attrName)
		} else {
			delete(indexedAttributes, attrName)
		}
	}

	var errs []error

	if len(unindexedAttributes) > 0 {
		slices.Sort(unindexedAttributes)

		errs = append(errs, fmt.Errorf("all attributes must be indexed. Unused attributes: %q", unindexedAttributes))
	}

	if len(indexedAttributes) > 0 {
		missingIndexes := tfmaps.Keys(indexedAttributes)
		slices.Sort(missingIndexes)

		errs = append(errs, fmt.Errorf("all indexes must match a defined attribute. Unmatched indexes: %q", missingIndexes))
	}
	return errors.Join(errs...)
}

func validateGSIProvisionedThroughput(data map[string]any, billingMode awstypes.BillingMode) error {
	// if billing mode is PAY_PER_REQUEST, don't need to validate the throughput settings
	if billingMode == awstypes.BillingModePayPerRequest {
		return nil
	}

	writeCapacity, writeCapacitySet := data["write_capacity"].(int)
	readCapacity, readCapacitySet := data["read_capacity"].(int)

	if !writeCapacitySet || !readCapacitySet {
		return fmt.Errorf("read and write capacity should be set when billing mode is %s", awstypes.BillingModeProvisioned)
	}

	if writeCapacity < 1 {
		return fmt.Errorf("write capacity must be > 0 when billing mode is %s", awstypes.BillingModeProvisioned)
	}

	if readCapacity < 1 {
		return fmt.Errorf("read capacity must be > 0 when billing mode is %s", awstypes.BillingModeProvisioned)
	}

	return nil
}

func validateGlobalSecondaryIndexes(ctx context.Context, req schema.ValidateResourceConfigFuncRequest, resp *schema.ValidateResourceConfigFuncResponse) {
	gsisPath := cty.GetAttrPath("global_secondary_index")

	gsis, err := gsisPath.Apply(req.RawConfig)
	if err != nil {
		resp.Diagnostics = sdkdiag.AppendFromErr(resp.Diagnostics, err)
		return
	}

	if !gsis.IsKnown() || gsis.IsNull() {
		return
	}
	for i, gsiElem := range tfcty.ValueElements(gsis) {
		gsiElemPath := gsisPath.Index(i)
		validateGlobalSecondaryIndex(ctx, gsiElem, gsiElemPath, &resp.Diagnostics)
	}
}

func validateGlobalSecondaryIndex(ctx context.Context, gsi cty.Value, gsiPath cty.Path, diags *diag.Diagnostics) {
	keySchemaPath := gsiPath.GetAttr("key_schema")
	keySchema := gsi.GetAttr("key_schema")
	keySchemaIsZero := keySchema.IsKnown() && (keySchema.IsNull() || keySchema.LengthInt() == 0)

	hashKey := gsi.GetAttr("hash_key")
	hashKeyIsZero := hashKey.IsKnown() && hashKey.IsNull()

	if hashKeyIsZero && keySchemaIsZero {
		*diags = append(*diags, errs.NewExactlyOneOfChildrenError(
			gsiPath, 0, cty.GetAttrPath("key_schema"), cty.GetAttrPath("hash_key"),
		))
	} else if !hashKeyIsZero && !keySchemaIsZero {
		*diags = append(*diags, errs.NewExactlyOneOfChildrenError(
			gsiPath, 2, cty.GetAttrPath("key_schema"), cty.GetAttrPath("hash_key"), //nolint:mnd
		))
	}

	rangeKey := gsi.GetAttr("range_key")
	rangeKeyIsZero := rangeKey.IsKnown() && rangeKey.IsNull()

	if !rangeKeyIsZero && !keySchemaIsZero {
		*diags = append(*diags, errs.NewAttributeConflictsWithError(
			gsiPath.GetAttr("range_key"),
			keySchemaPath,
		))
	}

	if !keySchemaIsZero {
		validateGSIKeySchema(ctx, keySchema, keySchemaPath, diags)
	}
}

func validateGSIKeySchema(_ context.Context, keySchema cty.Value, keySchemaPath cty.Path, diags *diag.Diagnostics) {
	var hashCount, rangeCount int
	var lastKeyType awstypes.KeyType

	if !keySchema.IsKnown() || keySchema.IsNull() {
		return
	}
	for i, gsi := range tfcty.ValueElements(keySchema) {
		keyType := awstypes.KeyType(gsi.GetAttr("key_type").AsString())
		switch keyType {
		case awstypes.KeyTypeHash:
			if lastKeyType == awstypes.KeyTypeRange {
				elementPath := keySchemaPath.Index(i)
				*diags = append(*diags, errs.NewAttributeErrorDiagnostic(
					elementPath,
					"Invalid Attribute Value",
					fmt.Sprintf(`All elements of %s with "key_type" "`+string(awstypes.KeyTypeHash)+`" must be before elements with "key_type" "`+string(awstypes.KeyTypeRange)+`"`, errs.PathString(keySchemaPath)),
				))
			}
			hashCount++

		case awstypes.KeyTypeRange:
			rangeCount++
		}
		lastKeyType = keyType
	}

	if hashCount < minNumberOfHashes || hashCount > maxNumberOfHashes {
		*diags = append(*diags, errs.NewInvalidValueAttributeError(
			keySchemaPath,
			fmt.Sprintf(`The attribute %q must contain at least %d and at most %d elements with "key_type" %q, got %s`,
				errs.PathString(keySchemaPath),
				minNumberOfHashes, maxNumberOfHashes,
				awstypes.KeyTypeHash,
				strconv.Itoa(hashCount),
			),
		))
	}

	if rangeCount > maxNumberOfRanges {
		*diags = append(*diags, errs.NewInvalidValueAttributeError(
			keySchemaPath,
			fmt.Sprintf(`The attribute %q must contain at most %d elements with "key_type" %q, got %s`,
				errs.PathString(keySchemaPath),
				maxNumberOfRanges,
				awstypes.KeyTypeRange,
				strconv.Itoa(rangeCount),
			),
		))
	}
}

func validateStreamSpecification(ctx context.Context, req schema.ValidateResourceConfigFuncRequest, resp *schema.ValidateResourceConfigFuncResponse) {
	streamEnabled := req.RawConfig.GetAttr("stream_enabled")
	if !streamEnabled.IsKnown() {
		return
	}

	streamViewType := req.RawConfig.GetAttr("stream_view_type")
	if !streamViewType.IsKnown() {
		return
	}

	if streamEnabled.IsNull() {
		if !streamViewType.IsNull() && streamViewType.AsString() != "" {
			resp.Diagnostics = append(resp.Diagnostics, errs.NewAttributeAlsoRequiresError(
				cty.GetAttrPath("stream_view_type"),
				cty.GetAttrPath("stream_enabled"),
			))
		}
	} else if streamEnabled.True() {
		if streamViewType.IsNull() || streamViewType.AsString() == "" {
			resp.Diagnostics = append(resp.Diagnostics, errs.NewAttributeRequiredWhenError(
				cty.GetAttrPath("stream_view_type"),
				cty.GetAttrPath("stream_enabled"),
				"true",
			))
		}
	} else {
		if !streamViewType.IsNull() && streamViewType.AsString() != "" {
			resp.Diagnostics = append(resp.Diagnostics, errs.NewAttributeConflictsWhenWillBeError(
				cty.GetAttrPath("stream_view_type"),
				cty.GetAttrPath("stream_enabled"),
				"false",
			))
		}
	}
}

func validateProvisionedThroughputField(path cty.Path) schema.ValidateRawResourceConfigFunc {
	return func(ctx context.Context, req schema.ValidateResourceConfigFuncRequest, resp *schema.ValidateResourceConfigFuncResponse) {
		v, err := path.Apply(req.RawConfig)
		if err != nil {
			resp.Diagnostics = sdkdiag.AppendFromErr(resp.Diagnostics, err)
			return
		}

		if !v.IsKnown() || v.IsNull() {
			return
		}

		billingMode := req.RawConfig.GetAttr("billing_mode")
		if !billingMode.IsKnown() || billingMode.IsNull() {
			return
		}

		bm := awstypes.BillingMode(billingMode.AsString())
		value, _ := v.AsBigFloat().Int64()

		switch bm {
		case awstypes.BillingModeProvisioned:
			if value < provisionedThroughputMinValue {
				resp.Diagnostics = append(resp.Diagnostics, errs.NewInvalidValueAttributeCombinationError(
					path,
					fmt.Sprintf("Attribute %q must be at least %d when %q is %q.",
						errs.PathString(path),
						provisionedThroughputMinValue,
						cty.GetAttrPath("billing_mode"),
						string(awstypes.BillingModeProvisioned),
					),
				))
			}

		case awstypes.BillingModePayPerRequest:
			if value != 0 {
				resp.Diagnostics = append(resp.Diagnostics, errs.NewAttributeConflictsWhenError(
					path,
					cty.GetAttrPath("billing_mode"),
					string(awstypes.BillingModePayPerRequest),
				))
			}
		}
	}
}

func suppressTableWarmThroughputDefaults(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	configRaw := d.GetRawConfig()
	if !configRaw.IsKnown() || configRaw.IsNull() {
		return nil
	}

	// If warm throughput is explicitly configured, don't suppress any diffs
	if warmThroughput := configRaw.GetAttr("warm_throughput"); warmThroughput.IsKnown() && !warmThroughput.IsNull() && warmThroughput.LengthInt() > 0 {
		return nil
	}

	// If warm throughput is not explicitly configured, suppress AWS default values
	if !d.HasChange("warm_throughput") {
		return nil
	}

	_, new := d.GetChange("warm_throughput")
	newList, ok := new.([]any)
	if !ok || len(newList) == 0 {
		return nil
	}

	newMap, ok := newList[0].(map[string]any)
	if !ok {
		return nil
	}

	readUnits := newMap["read_units_per_second"]
	writeUnits := newMap["write_units_per_second"]

	// If AWS returns default values and no explicit config, suppress the diff
	if (readUnits == 1 && writeUnits == 1) || (readUnits == 12000 && writeUnits == 4000) {
		return d.Clear("warm_throughput")
	}

	return nil
}

func validateTTLList(ctx context.Context, req schema.ValidateResourceConfigFuncRequest, resp *schema.ValidateResourceConfigFuncResponse) {
	ttl := req.RawConfig.GetAttr("ttl")
	if !ttl.IsKnown() || ttl.IsNull() {
		return
	}

	ttlPath := cty.GetAttrPath("ttl")
	for i, ttlElem := range tfcty.ValueElements(ttl) {
		ttlElemPath := ttlPath.Index(i)
		validateTTL(ctx, ttlElem, ttlElemPath, &resp.Diagnostics)
	}
}

func validateTTL(_ context.Context, ttl cty.Value, ttlPath cty.Path, diags *diag.Diagnostics) {
	attribute := ttl.GetAttr("attribute_name")
	if !attribute.IsKnown() {
		return
	}

	enabled := ttl.GetAttr(names.AttrEnabled)
	if !enabled.IsKnown() {
		return
	}
	if enabled.IsNull() {
		return
	}

	if enabled.True() {
		if attribute.IsNull() {
			*diags = append(*diags, errs.NewAttributeRequiredWhenError(
				ttlPath.GetAttr("attribute_name"),
				ttlPath.GetAttr(names.AttrEnabled),
				"true",
			))
		} else if attribute.AsString() == "" {
			*diags = append(*diags, errs.NewInvalidValueAttributeErrorf(
				ttlPath.GetAttr("attribute_name"),
				"Attribute %q cannot have an empty value when %q is \"true\"",
				errs.PathString(ttlPath.GetAttr("attribute_name")),
				errs.PathString(ttlPath.GetAttr(names.AttrEnabled)),
			))
		}
	}

	// !! Not a validation error for attribute_name to be set when enabled is false !!
	// AWS *requires* attribute_name to be set when disabling TTL but does not return it, causing a diff.
	// The diff is handled by DiffSuppressFunc of attribute_name.
}

func customDiffGlobalSecondaryIndex(_ context.Context, diff *schema.ResourceDiff, _ any) error {
	if diff.Id() == "" {
		return nil
	}
	if !diff.HasChange("global_secondary_index") {
		return nil
	}

	stateRaw := diff.GetRawState()
	if !stateRaw.IsKnown() || stateRaw.IsNull() {
		return nil
	}
	stateGSI := stateRaw.GetAttr("global_secondary_index")
	state := collectGSI(stateGSI)

	planRaw := diff.GetRawPlan()
	if !planRaw.IsKnown() || planRaw.IsNull() {
		return nil
	}
	planGSI := planRaw.GetAttr("global_secondary_index")
	plan := collectGSI(planGSI)

	// Adding or removing GSIs
	if len(plan) != len(state) {
		return nil
	}

	// GSI name mismatch
	for name := range state {
		if _, ok := plan[name]; !ok {
			return nil
		}
	}

	for name, vState := range state {
		vPlan := plan[name]

		for attrName := range vState.Type().AttributeTypes() {
			s := vState.GetAttr(attrName)
			p := vPlan.GetAttr(attrName)
			switch attrName {
			case "hash_key":
				if p.IsNull() && !s.IsNull() && vPlan.GetAttr("key_schema").LengthInt() > 0 {
					// "key_schema" is set
					continue // change to "key_schema" will be caught by equality test
				}
				if !ctyValueLegacyEquals(s, p) {
					return nil
				}

			case "range_key":
				if p.IsNull() && !s.IsNull() && vPlan.GetAttr("key_schema").LengthInt() > 0 {
					// "key_schema" is set
					continue // change to "key_schema" will be caught by equality test
				}
				if !ctyValueLegacyEquals(s, p) {
					return nil
				}

			case "key_schema":
				// key_schema is a block nested list, so the zero-value is an empty list
				if p.LengthInt() == 0 && s.LengthInt() > 0 {
					// "hash_key" is set
					continue // change to "hash_key" will be caught by equality test
				}
				if !ctyValueLegacyEquals(s, p) {
					return nil
				}

			case "warm_throughput":
				// AWS automatically sets warm_throughput
				// values for on-demand tables, but these should not cause diffs when
				// the user hasn't explicitly configured warm_throughput.
				if p.IsNull() || (p.IsKnown() && p.LengthInt() == 0) {
					continue
				}
				if !ctyValueLegacyEquals(s, p) {
					return nil
				}

			default:
				if !ctyValueLegacyEquals(s, p) {
					return nil
				}
			}
		}
	}

	return diff.Clear("global_secondary_index")
}

func collectGSI(gsi cty.Value) map[string]cty.Value {
	result := make(map[string]cty.Value, gsi.LengthInt())
	if gsi.IsKnown() && !gsi.IsNull() {
		for v := range tfcty.ValueElementValues(gsi) {
			name := v.GetAttr(names.AttrName)
			result[name.AsString()] = v
		}
	}
	return result
}

func ctyValueLegacyEquals(lhs, rhs cty.Value) bool {
	if lhs.Equals(rhs).True() {
		return true
	}

	if !lhs.Type().Equals(rhs.Type()) {
		return false
	}

	switch {
	case lhs.Type().IsSetType():
		if rhs.IsNull() && lhs.LengthInt() == 0 {
			return true
		}
		if lhs.IsNull() && rhs.LengthInt() == 0 {
			return true
		}

	case lhs.Type().IsListType():
		if rhs.IsNull() && lhs.LengthInt() == 0 {
			return true
		}
		if lhs.IsNull() && rhs.LengthInt() == 0 {
			return true
		}

	case lhs.Type() == cty.String:
		if rhs.IsNull() && lhs.AsString() == "" {
			return true
		}
		if lhs.IsNull() && rhs.AsString() == "" {
			return true
		}

	case lhs.Type() == cty.Number:
		var zero big.Float
		if rhs.IsNull() && zero.Cmp(lhs.AsBigFloat()) == 0 {
			return true
		}
		if lhs.IsNull() && zero.Cmp(rhs.AsBigFloat()) == 0 {
			return true
		}
	}

	return false
}
