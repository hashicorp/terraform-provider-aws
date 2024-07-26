// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/service/kms"
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
			func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
				return validStreamSpec(diff)
			},
			func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
				return validateTableAttributes(diff)
			},
			func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
				if diff.Id() != "" && diff.HasChange("server_side_encryption") {
					o, n := diff.GetChange("server_side_encryption")
					if isTableOptionDisabled(o) && isTableOptionDisabled(n) {
						return diff.Clear("server_side_encryption")
					}
				}
				return nil
			},
			func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
				if diff.Id() != "" && diff.HasChange("point_in_time_recovery") {
					o, n := diff.GetChange("point_in_time_recovery")
					if isTableOptionDisabled(o) && isTableOptionDisabled(n) {
						return diff.Clear("point_in_time_recovery")
					}
				}
				return nil
			},
			func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
				if diff.Id() != "" && diff.HasChange("stream_enabled") {
					if err := diff.SetNewComputed(names.AttrStreamARN); err != nil {
						return fmt.Errorf("setting stream_arn to computed: %s", err)
					}
				}
				return nil
			},
			func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
				if v := diff.Get("restore_source_name"); v != "" {
					return nil
				}

				var errs []error
				if err := validateProvisionedThroughputField(diff, "read_capacity"); err != nil {
					errs = append(errs, err)
				}
				if err := validateProvisionedThroughputField(diff, "write_capacity"); err != nil {
					errs = append(errs, err)
				}
				return errors.Join(errs...)
			},
			customdiff.ForceNewIfChange("restore_source_name", func(_ context.Context, old, new, meta interface{}) bool {
				// If they differ force new unless new is cleared
				// https://github.com/hashicorp/terraform-provider-aws/issues/25214
				return old.(string) != new.(string) && new.(string) != ""
			}),
			validateTTLCustomDiff,
			verify.SetTagsDiff,
		),

		SchemaVersion: 1,
		MigrateState:  resourceTableMigrateState,

		Schema: map[string]*schema.Schema{
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
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hash_key": {
							Type:     schema.TypeString,
							Required: true,
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
						"projection_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ProjectionType](),
						},
						"range_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"read_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"write_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
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
					},
				},
			},
			"range_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"read_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
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
			"import_table": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"restore_source_name"},
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
			"restore_date_time": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidUTCTimestamp,
			},
			"restore_source_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"import_table"},
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
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				StateFunc: func(v interface{}) string {
					value := v.(string)
					return strings.ToUpper(value)
				},
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
			"write_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
		},
	}
}

func resourceTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	tableName := d.Get(names.AttrName).(string)
	keySchemaMap := map[string]interface{}{
		"hash_key": d.Get("hash_key").(string),
	}
	if v, ok := d.GetOk("range_key"); ok {
		keySchemaMap["range_key"] = v.(string)
	}

	if v, ok := d.GetOk("restore_source_name"); ok {
		input := &dynamodb.RestoreTableToPointInTimeInput{
			SourceTableName: aws.String(v.(string)),
			TargetTableName: aws.String(tableName),
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
				capacityMap := map[string]interface{}{
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
				gsi := gsiObject.(map[string]interface{})
				if err := validateGSIProvisionedThroughput(gsi, billingModeOverride); err != nil {
					return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, d.Get(names.AttrName).(string), err)
				}

				gsiObject := expandGlobalSecondaryIndex(gsi, billingModeOverride)
				globalSecondaryIndexes = append(globalSecondaryIndexes, *gsiObject)
			}
			input.GlobalSecondaryIndexOverride = globalSecondaryIndexes
		}

		if v, ok := d.GetOk("server_side_encryption"); ok {
			input.SSESpecificationOverride = expandEncryptAtRestOptions(v.([]interface{}))
		}

		_, err := tfresource.RetryWhen(ctx, createTableTimeout, func() (interface{}, error) {
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
	} else if vit, ok := d.GetOk("import_table"); ok && len(vit.([]interface{})) > 0 && vit.([]interface{})[0] != nil {
		input := expandImportTable(vit.([]interface{})[0].(map[string]interface{}))

		tcp := &awstypes.TableCreationParameters{
			TableName:   aws.String(tableName),
			BillingMode: awstypes.BillingMode(d.Get("billing_mode").(string)),
			KeySchema:   expandKeySchema(keySchemaMap),
		}

		capacityMap := map[string]interface{}{
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
			tcp.SSESpecification = expandEncryptAtRestOptions(v.([]interface{}))
		}

		if v, ok := d.GetOk("global_secondary_index"); ok {
			globalSecondaryIndexes := []awstypes.GlobalSecondaryIndex{}
			gsiSet := v.(*schema.Set)

			for _, gsiObject := range gsiSet.List() {
				gsi := gsiObject.(map[string]interface{})
				if err := validateGSIProvisionedThroughput(gsi, billingMode); err != nil {
					return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, d.Get(names.AttrName).(string), err)
				}

				gsiObject := expandGlobalSecondaryIndex(gsi, billingMode)
				globalSecondaryIndexes = append(globalSecondaryIndexes, *gsiObject)
			}
			tcp.GlobalSecondaryIndexes = globalSecondaryIndexes
		}

		input.TableCreationParameters = tcp

		importTableOutput, err := tfresource.RetryWhen(ctx, createTableTimeout, func() (interface{}, error) {
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

		capacityMap := map[string]interface{}{
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
				gsi := gsiObject.(map[string]interface{})
				if err := validateGSIProvisionedThroughput(gsi, billingMode); err != nil {
					return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, tableName, err)
				}

				gsiObject := expandGlobalSecondaryIndex(gsi, billingMode)
				globalSecondaryIndexes = append(globalSecondaryIndexes, *gsiObject)
			}
			input.GlobalSecondaryIndexes = globalSecondaryIndexes
		}

		if v, ok := d.GetOk("stream_enabled"); ok {
			input.StreamSpecification = &awstypes.StreamSpecification{
				StreamEnabled:  aws.Bool(v.(bool)),
				StreamViewType: awstypes.StreamViewType(d.Get("stream_view_type").(string)),
			}
		}

		if v, ok := d.GetOk("server_side_encryption"); ok {
			input.SSESpecification = expandEncryptAtRestOptions(v.([]interface{}))
		}

		if v, ok := d.GetOk("table_class"); ok {
			input.TableClass = awstypes.TableClass(v.(string))
		}

		_, err := tfresource.RetryWhen(ctx, createTableTimeout, func() (interface{}, error) {
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

	if v, ok := d.GetOk("global_secondary_index"); ok {
		gsiSet := v.(*schema.Set)

		for _, gsiObject := range gsiSet.List() {
			gsi := gsiObject.(map[string]interface{})

			if _, err := waitGSIActive(ctx, conn, d.Id(), gsi[names.AttrName].(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForCreation, resNameTable, d.Id(), fmt.Errorf("GSI (%s): %w", gsi[names.AttrName].(string), err))
			}
		}
	}

	if d.Get("ttl.0.enabled").(bool) {
		if err := updateTimeToLive(ctx, conn, d.Id(), d.Get("ttl").([]interface{}), d.Timeout(schema.TimeoutCreate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, d.Id(), fmt.Errorf("enabling TTL: %w", err))
		}
	}

	if d.Get("point_in_time_recovery.0.enabled").(bool) {
		if err := updatePITR(ctx, conn, d.Id(), true, meta.(*conns.AWSClient).Region, d.Timeout(schema.TimeoutCreate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, d.Id(), fmt.Errorf("enabling point in time recovery: %w", err))
		}
	}

	if v := d.Get("replica").(*schema.Set); v.Len() > 0 {
		if err := createReplicas(ctx, conn, d.Id(), v.List(), true, d.Timeout(schema.TimeoutCreate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, d.Id(), fmt.Errorf("replicas: %w", err))
		}

		if err := updateReplicaTags(ctx, conn, aws.ToString(output.TableArn), v.List(), KeyValueTags(ctx, getTagsIn(ctx))); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTable, d.Id(), fmt.Errorf("replica tags: %w", err))
		}
	}

	return append(diags, resourceTableRead(ctx, d, meta)...)
}

func resourceTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	table, err := findTableByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
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
	sse = clearSSEDefaultKey(ctx, meta.(*conns.AWSClient), sse)

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
	replicas = clearReplicaDefaultKeys(ctx, meta.(*conns.AWSClient), replicas)

	if err := d.Set("replica", replicas); err != nil {
		return create.AppendDiagSettingError(diags, names.DynamoDB, resNameTable, d.Id(), "replica", err)
	}

	if table.TableClassSummary != nil {
		d.Set("table_class", table.TableClassSummary.TableClass)
	} else {
		d.Set("table_class", awstypes.TableClassStandard)
	}

	pitrOut, err := conn.DescribeContinuousBackups(ctx, &dynamodb.DescribeContinuousBackupsInput{
		TableName: aws.String(d.Id()),
	})

	// When a Table is `ARCHIVED`, DescribeContinuousBackups returns `TableNotFoundException`
	if err != nil && !tfawserr.ErrCodeEquals(err, errCodeUnknownOperationException, errCodeTableNotFoundException) {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, resNameTable, d.Id(), fmt.Errorf("continuous backups: %w", err))
	}

	if err := d.Set("point_in_time_recovery", flattenPITR(pitrOut)); err != nil {
		return create.AppendDiagSettingError(diags, names.DynamoDB, resNameTable, d.Id(), "point_in_time_recovery", err)
	}

	ttlOut, err := conn.DescribeTimeToLive(ctx, &dynamodb.DescribeTimeToLiveInput{
		TableName: aws.String(d.Id()),
	})

	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, resNameTable, d.Id(), fmt.Errorf("TTL: %w", err))
	}

	if err := d.Set("ttl", flattenTTL(ttlOut)); err != nil {
		return create.AppendDiagSettingError(diags, names.DynamoDB, resNameTable, d.Id(), "ttl", err)
	}

	return diags
}

func resourceTableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		_, err := conn.UpdateTable(ctx, &dynamodb.UpdateTableInput{
			TableClass: awstypes.TableClass(d.Get("table_class").(string)),
			TableName:  aws.String(d.Id()),
		})
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

		capacityMap := map[string]interface{}{
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
			if gsiUpdate.Update == nil {
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

		for _, gsiUpdate := range gsiUpdates {
			if gsiUpdate.Update == nil {
				continue
			}

			idxName := aws.ToString(gsiUpdate.Update.IndexName)

			if _, err := waitGSIActive(ctx, conn, d.Id(), idxName, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForUpdate, resNameTable, d.Id(), fmt.Errorf("GSI (%s): %w", idxName, err))
			}
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
	}

	if d.HasChange("server_side_encryption") {
		if replicas := d.Get("replica").(*schema.Set); replicas.Len() > 0 {
			log.Printf("[DEBUG] Using SSE update on replicas")
			var replicaInputs []awstypes.ReplicationGroupUpdate
			var replicaRegions []string
			for _, replica := range replicas.List() {
				tfMap, ok := replica.(map[string]interface{})
				if !ok {
					continue
				}
				var regionName string
				var KMSMasterKeyId string
				if v, ok := tfMap["region_name"].(string); ok {
					regionName = v
					replicaRegions = append(replicaRegions, v)
				}
				if v, ok := tfMap[names.AttrKMSKeyARN].(string); ok && v != "" {
					KMSMasterKeyId = v
				}
				var input = &awstypes.UpdateReplicationGroupMemberAction{
					RegionName:     aws.String(regionName),
					KMSMasterKeyId: aws.String(KMSMasterKeyId),
				}
				var update = awstypes.ReplicationGroupUpdate{Update: input}
				replicaInputs = append(replicaInputs, update)
			}
			var input = &awstypes.UpdateReplicationGroupMemberAction{
				KMSMasterKeyId: expandEncryptAtRestOptions(d.Get("server_side_encryption").([]interface{})).KMSMasterKeyId,
				RegionName:     aws.String(meta.(*conns.AWSClient).Region),
			}
			var update = awstypes.ReplicationGroupUpdate{Update: input}
			replicaInputs = append(replicaInputs, update)
			_, err := conn.UpdateTable(ctx, &dynamodb.UpdateTableInput{
				TableName:      aws.String(d.Id()),
				ReplicaUpdates: replicaInputs,
			})
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating DynamoDB Table (%s) SSE: %s", d.Id(), err)
			}
			for _, region := range replicaRegions {
				if _, err := waitReplicaSSEUpdated(ctx, conn, region, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for DynamoDB Table (%s) replica SSE update in region %q: %s", d.Id(), region, err)
				}
			}
			if _, err := waitSSEUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for DynamoDB Table (%s) SSE update: %s", d.Id(), err)
			}
		} else {
			log.Printf("[DEBUG] Using normal update for SSE")
			_, err := conn.UpdateTable(ctx, &dynamodb.UpdateTableInput{
				TableName:        aws.String(d.Id()),
				SSESpecification: expandEncryptAtRestOptions(d.Get("server_side_encryption").([]interface{})),
			})
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating DynamoDB Table (%s) SSE: %s", d.Id(), err)
			}

			if _, err := waitSSEUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for DynamoDB Table (%s) SSE update: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("ttl") {
		if err := updateTimeToLive(ctx, conn, d.Id(), d.Get("ttl").([]interface{}), d.Timeout(schema.TimeoutUpdate)); err != nil {
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
		if err := updatePITR(ctx, conn, d.Id(), d.Get("point_in_time_recovery.0.enabled").(bool), meta.(*conns.AWSClient).Region, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTable, d.Id(), err)
		}
	}

	return append(diags, resourceTableRead(ctx, d, meta)...)
}

func resourceTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	if replicas := d.Get("replica").(*schema.Set).List(); len(replicas) > 0 {
		log.Printf("[DEBUG] Deleting DynamoDB Table replicas: %s", d.Id())
		if err := deleteReplicas(ctx, conn, d.Id(), replicas, d.Timeout(schema.TimeoutDelete)); err != nil {
			// ValidationException: Replica specified in the Replica Update or Replica Delete action of the request was not found.
			if !tfawserr.ErrMessageContains(err, errCodeValidationException, "request was not found") {
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

func isTableOptionDisabled(v interface{}) bool {
	options := v.([]interface{})
	if len(options) == 0 {
		return true
	}
	e := options[0].(map[string]interface{})[names.AttrEnabled]
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
		return fmt.Errorf("cycling stream enabled: %s", err)
	}

	if _, err := waitTableActive(ctx, conn, id, timeout); err != nil {
		return fmt.Errorf("waiting for stream cycle: %s", err)
	}

	input.StreamSpecification = &awstypes.StreamSpecification{
		StreamEnabled:  aws.Bool(true),
		StreamViewType: streamViewType,
	}

	_, err = conn.UpdateTable(ctx, input)

	if err != nil {
		return fmt.Errorf("cycling stream enabled: %s", err)
	}

	if _, err := waitTableActive(ctx, conn, id, timeout); err != nil {
		return fmt.Errorf("waiting for stream cycle: %s", err)
	}

	return nil
}

func createReplicas(ctx context.Context, conn *dynamodb.Client, tableName string, tfList []interface{}, create bool, timeout time.Duration) error {
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

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

		err := retry.RetryContext(ctx, max(replicaUpdateTimeout, timeout), func() *retry.RetryError {
			_, err := conn.UpdateTable(ctx, input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, errCodeThrottlingException) {
					return retry.RetryableError(err)
				}
				if errs.IsAErrorMessageContains[*awstypes.LimitExceededException](err, "can be created, updated, or deleted simultaneously") {
					return retry.RetryableError(err)
				}
				if tfawserr.ErrMessageContains(err, errCodeValidationException, "Replica specified in the Replica Update or Replica Delete action of the request was not found") {
					return retry.RetryableError(err)
				}
				if errs.IsA[*awstypes.ResourceInUseException](err) {
					return retry.RetryableError(err)
				}

				return retry.NonRetryableError(err)
			}
			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateTable(ctx, input)
		}

		// An update that doesn't (makes no changes) returns ValidationException
		// (same region_name and kms_key_arn as currently) throws unhelpfully worded exception:
		// ValidationException: One or more parameter values were invalid: KMSMasterKeyId must be specified for each replica.

		if create && tfawserr.ErrMessageContains(err, errCodeValidationException, "already exist") {
			return createReplicas(ctx, conn, tableName, tfList, false, timeout)
		}

		if err != nil && !tfawserr.ErrMessageContains(err, errCodeValidationException, "no actions specified") {
			return fmt.Errorf("creating replica (%s): %w", tfMap["region_name"].(string), err)
		}

		if _, err := waitReplicaActive(ctx, conn, tableName, tfMap["region_name"].(string), timeout); err != nil {
			return fmt.Errorf("waiting for replica (%s) creation: %w", tfMap["region_name"].(string), err)
		}

		// pitr
		if err = updatePITR(ctx, conn, tableName, tfMap["point_in_time_recovery"].(bool), tfMap["region_name"].(string), timeout); err != nil {
			return fmt.Errorf("updating replica (%s) point in time recovery: %w", tfMap["region_name"].(string), err)
		}
	}

	return nil
}

func updateReplicaTags(ctx context.Context, conn *dynamodb.Client, rn string, replicas []interface{}, newTags interface{}) error {
	for _, tfMapRaw := range replicas {
		tfMap, ok := tfMapRaw.(map[string]interface{})

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

func updateTimeToLive(ctx context.Context, conn *dynamodb.Client, tableName string, ttlList []interface{}, timeout time.Duration) error {
	ttlMap := ttlList[0].(map[string]interface{})

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

func updatePITR(ctx context.Context, conn *dynamodb.Client, tableName string, enabled bool, region string, timeout time.Duration) error {
	// pitr must be modified from region where the main/replica resides
	log.Printf("[DEBUG] Updating DynamoDB point in time recovery status to %v (%s)", enabled, region)
	input := &dynamodb.UpdateContinuousBackupsInput{
		TableName: aws.String(tableName),
		PointInTimeRecoverySpecification: &awstypes.PointInTimeRecoverySpecification{
			PointInTimeRecoveryEnabled: aws.Bool(enabled),
		},
	}

	optFn := func(o *dynamodb.Options) {
		o.Region = region
	}
	err := retry.RetryContext(ctx, updateTableContinuousBackupsTimeout, func() *retry.RetryError {
		_, err := conn.UpdateContinuousBackups(ctx, input, optFn)
		if err != nil {
			// Backups are still being enabled for this newly created table
			if errs.IsAErrorMessageContains[*awstypes.ContinuousBackupsUnavailableException](err, "Backups are being enabled") {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.UpdateContinuousBackups(ctx, input, optFn)
	}

	if err != nil {
		return fmt.Errorf("updating PITR: %w", err)
	}

	if _, err := waitPITRUpdated(ctx, conn, tableName, enabled, timeout, optFn); err != nil {
		return fmt.Errorf("waiting for PITR update: %w", err)
	}

	return nil
}

func updateReplica(ctx context.Context, conn *dynamodb.Client, d *schema.ResourceData) error {
	oRaw, nRaw := d.GetChange("replica")
	o := oRaw.(*schema.Set)
	n := nRaw.(*schema.Set)

	removeRaw := o.Difference(n).List()
	addRaw := n.Difference(o).List()

	var removeFirst []interface{} // replicas to delete before recreating (like ForceNew without recreating table)
	var toAdd []interface{}
	var toRemove []interface{}

	// first pass - add replicas that don't have corresponding remove entry
	for _, a := range addRaw {
		add := true
		ma := a.(map[string]interface{})
		for _, r := range removeRaw {
			mr := r.(map[string]interface{})

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
		mr := r.(map[string]interface{})
		for _, a := range addRaw {
			ma := a.(map[string]interface{})

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
		ma := a.(map[string]interface{})
		for _, r := range removeRaw {
			mr := r.(map[string]interface{})

			if ma["region_name"].(string) != mr["region_name"].(string) {
				continue
			}

			// like "ForceNew" for the replica - KMS change
			if ma[names.AttrKMSKeyARN].(string) != mr[names.AttrKMSKeyARN].(string) {
				toRemove = append(toRemove, mr)
				toAdd = append(toAdd, ma)
				break
			}

			// just update PITR
			if ma["point_in_time_recovery"].(bool) != mr["point_in_time_recovery"].(bool) {
				if err := updatePITR(ctx, conn, d.Id(), ma["point_in_time_recovery"].(bool), ma["region_name"].(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return fmt.Errorf("updating replica (%s) point in time recovery: %w", ma["region_name"].(string), err)
				}
				break
			}

			// nothing changed, assuming propagate_tags changed so do nothing here
			break
		}
	}

	if len(removeFirst) > 0 { // mini ForceNew, recreates replica but doesn't recreate the table
		if err := deleteReplicas(ctx, conn, d.Id(), removeFirst, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("updating replicas, while deleting: %w", err)
		}
	}

	if len(toRemove) > 0 {
		if err := deleteReplicas(ctx, conn, d.Id(), toRemove, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("updating replicas, while deleting: %w", err)
		}
	}

	if len(toAdd) > 0 {
		if err := createReplicas(ctx, conn, d.Id(), toAdd, true, d.Timeout(schema.TimeoutCreate)); err != nil {
			return fmt.Errorf("updating replicas, while creating: %w", err)
		}
	}

	return nil
}

func updateDiffGSI(oldGsi, newGsi []interface{}, billingMode awstypes.BillingMode) ([]awstypes.GlobalSecondaryIndexUpdate, error) {
	// Transform slices into maps
	oldGsis := make(map[string]interface{})
	for _, gsidata := range oldGsi {
		m := gsidata.(map[string]interface{})
		oldGsis[m[names.AttrName].(string)] = m
	}
	newGsis := make(map[string]interface{})
	for _, gsidata := range newGsi {
		m := gsidata.(map[string]interface{})
		// validate throughput input early, to avoid unnecessary processing
		if err := validateGSIProvisionedThroughput(m, billingMode); err != nil {
			return nil, err
		}
		newGsis[m[names.AttrName].(string)] = m
	}

	var ops []awstypes.GlobalSecondaryIndexUpdate

	for _, data := range newGsi {
		newMap := data.(map[string]interface{})
		newName := newMap[names.AttrName].(string)

		if _, exists := oldGsis[newName]; !exists {
			m := data.(map[string]interface{})
			idxName := m[names.AttrName].(string)

			ops = append(ops, awstypes.GlobalSecondaryIndexUpdate{
				Create: &awstypes.CreateGlobalSecondaryIndexAction{
					IndexName:             aws.String(idxName),
					KeySchema:             expandKeySchema(m),
					ProvisionedThroughput: expandProvisionedThroughput(m, billingMode),
					Projection:            expandProjection(m),
				},
			})
		}
	}

	for _, data := range oldGsi {
		oldMap := data.(map[string]interface{})
		oldName := oldMap[names.AttrName].(string)

		newData, exists := newGsis[oldName]
		if exists {
			newMap := newData.(map[string]interface{})
			idxName := newMap[names.AttrName].(string)

			oldWriteCapacity, oldReadCapacity := oldMap["write_capacity"].(int), oldMap["read_capacity"].(int)
			newWriteCapacity, newReadCapacity := newMap["write_capacity"].(int), newMap["read_capacity"].(int)
			capacityChanged := (oldWriteCapacity != newWriteCapacity || oldReadCapacity != newReadCapacity)

			// pluck non_key_attributes from oldAttributes and newAttributes as reflect.DeepEquals will compare
			// ordinal of elements in its equality (which we actually don't care about)
			nonKeyAttributesChanged := checkIfNonKeyAttributesChanged(oldMap, newMap)

			oldAttributes, err := stripCapacityAttributes(oldMap)
			if err != nil {
				return ops, err
			}
			oldAttributes, err = stripNonKeyAttributes(oldAttributes)
			if err != nil {
				return ops, err
			}
			newAttributes, err := stripCapacityAttributes(newMap)
			if err != nil {
				return ops, err
			}
			newAttributes, err = stripNonKeyAttributes(newAttributes)
			if err != nil {
				return ops, err
			}
			otherAttributesChanged := nonKeyAttributesChanged || !reflect.DeepEqual(oldAttributes, newAttributes)

			if capacityChanged && !otherAttributesChanged {
				update := awstypes.GlobalSecondaryIndexUpdate{
					Update: &awstypes.UpdateGlobalSecondaryIndexAction{
						IndexName:             aws.String(idxName),
						ProvisionedThroughput: expandProvisionedThroughput(newMap, billingMode),
					},
				}
				ops = append(ops, update)
			} else if otherAttributesChanged {
				// Other attributes cannot be updated
				ops = append(ops, awstypes.GlobalSecondaryIndexUpdate{
					Delete: &awstypes.DeleteGlobalSecondaryIndexAction{
						IndexName: aws.String(idxName),
					},
				})

				ops = append(ops, awstypes.GlobalSecondaryIndexUpdate{
					Create: &awstypes.CreateGlobalSecondaryIndexAction{
						IndexName:             aws.String(idxName),
						KeySchema:             expandKeySchema(newMap),
						ProvisionedThroughput: expandProvisionedThroughput(newMap, billingMode),
						Projection:            expandProjection(newMap),
					},
				})
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

func deleteTable(ctx context.Context, conn *dynamodb.Client, tableName string) error {
	input := &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	}

	_, err := tfresource.RetryWhen(ctx, deleteTableTimeout, func() (interface{}, error) {
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

func deleteReplicas(ctx context.Context, conn *dynamodb.Client, tableName string, tfList []interface{}, timeout time.Duration) error {
	var g multierror.Group

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

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

		g.Go(func() error {
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

			err := retry.RetryContext(ctx, updateTableTimeout, func() *retry.RetryError {
				_, err := conn.UpdateTable(ctx, input)
				notFoundRetries := 0
				if err != nil {
					if tfawserr.ErrCodeEquals(err, errCodeThrottlingException) {
						return retry.RetryableError(err)
					}
					if errs.IsA[*awstypes.ResourceNotFoundException](err) {
						notFoundRetries++
						if notFoundRetries > 3 {
							return retry.NonRetryableError(err)
						}
						return retry.RetryableError(err)
					}
					if errs.IsAErrorMessageContains[*awstypes.LimitExceededException](err, "can be created, updated, or deleted simultaneously") {
						return retry.RetryableError(err)
					}
					if errs.IsA[*awstypes.ResourceInUseException](err) {
						return retry.RetryableError(err)
					}

					return retry.NonRetryableError(err)
				}
				return nil
			})

			if tfresource.TimedOut(err) {
				_, err = conn.UpdateTable(ctx, input)
			}

			if err != nil && !errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return fmt.Errorf("deleting replica (%s): %w", regionName, err)
			}

			if _, err := waitReplicaDeleted(ctx, conn, tableName, regionName, timeout); err != nil {
				return fmt.Errorf("waiting for replica (%s) deletion: %w", regionName, err)
			}

			return nil
		})
	}

	return g.Wait().ErrorOrNil()
}

func replicaPITR(ctx context.Context, conn *dynamodb.Client, tableName string, region string) (bool, error) {
	// To manage replicas you need connections from the different regions. However, they
	// have to be created from the starting/main region.
	optFn := func(o *dynamodb.Options) {
		o.Region = region
	}

	pitrOut, err := conn.DescribeContinuousBackups(ctx, &dynamodb.DescribeContinuousBackupsInput{
		TableName: aws.String(tableName),
	}, optFn)
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

func addReplicaPITRs(ctx context.Context, conn *dynamodb.Client, tableName string, replicas []interface{}) ([]interface{}, error) {
	// This non-standard approach is needed because PITR info for a replica
	// must come from a region-specific connection.
	for i, replicaRaw := range replicas {
		replica := replicaRaw.(map[string]interface{})

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

func enrichReplicas(ctx context.Context, conn *dynamodb.Client, arn, tableName string, tfList []interface{}) ([]interface{}, error) {
	// This non-standard approach is needed because PITR info for a replica
	// must come from a region-specific connection.
	for i, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		newARN, err := arnForNewRegion(arn, tfMap["region_name"].(string))
		if err != nil {
			return nil, fmt.Errorf("creating new-region ARN: %s", err)
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

		tfMap[names.AttrStreamARN] = aws.ToString(table.LatestStreamArn)
		tfMap["stream_label"] = aws.ToString(table.LatestStreamLabel)

		if table.SSEDescription != nil {
			tfMap[names.AttrKMSKeyARN] = aws.ToString(table.SSEDescription.KMSMasterKeyArn)
		}

		tfList[i] = tfMap
	}

	return tfList, nil
}

func addReplicaTagPropagates(configReplicas *schema.Set, replicas []interface{}) []interface{} {
	if configReplicas.Len() == 0 {
		return replicas
	}

	l := configReplicas.List()

	for i, replicaRaw := range replicas {
		replica := replicaRaw.(map[string]interface{})

		prop := false

		for _, configReplicaRaw := range l {
			configReplica := configReplicaRaw.(map[string]interface{})

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
func clearSSEDefaultKey(ctx context.Context, client *conns.AWSClient, sseList []interface{}) []interface{} {
	if len(sseList) == 0 {
		return sseList
	}

	sse := sseList[0].(map[string]interface{})

	dk, err := kms.FindDefaultKeyARNForService(ctx, client.KMSClient(ctx), "dynamodb", client.Region)
	if err != nil {
		return sseList
	}

	if sse[names.AttrKMSKeyARN].(string) == dk {
		sse[names.AttrKMSKeyARN] = ""
		return []interface{}{sse}
	}

	return sseList
}

// clearReplicaDefaultKeys sets a replica's kms_key_arn to "" if it is the default key alias/aws/dynamodb for
// the replica's region. Not clearing the key causes diff problems and sends the key to AWS when it should not be.
func clearReplicaDefaultKeys(ctx context.Context, client *conns.AWSClient, replicas []interface{}) []interface{} {
	if len(replicas) == 0 {
		return replicas
	}

	for i, replicaRaw := range replicas {
		replica := replicaRaw.(map[string]interface{})

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

func flattenTableAttributeDefinitions(definitions []awstypes.AttributeDefinition) []interface{} {
	if len(definitions) == 0 {
		return []interface{}{}
	}

	var attributes []interface{}

	for _, d := range definitions {
		m := map[string]interface{}{
			names.AttrName: aws.ToString(d.AttributeName),
			names.AttrType: d.AttributeType,
		}

		attributes = append(attributes, m)
	}

	return attributes
}

func flattenTableLocalSecondaryIndex(lsi []awstypes.LocalSecondaryIndexDescription) []interface{} {
	if len(lsi) == 0 {
		return []interface{}{}
	}

	var output []interface{}

	for _, l := range lsi {
		m := map[string]interface{}{
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

func flattenTableGlobalSecondaryIndex(gsi []awstypes.GlobalSecondaryIndexDescription) []interface{} {
	if len(gsi) == 0 {
		return []interface{}{}
	}

	var output []interface{}

	for _, g := range gsi {
		gsi := make(map[string]interface{})

		if g.ProvisionedThroughput != nil {
			gsi["write_capacity"] = aws.ToInt64(g.ProvisionedThroughput.WriteCapacityUnits)
			gsi["read_capacity"] = aws.ToInt64(g.ProvisionedThroughput.ReadCapacityUnits)
			gsi[names.AttrName] = aws.ToString(g.IndexName)
		}

		for _, attribute := range g.KeySchema {
			if attribute.KeyType == awstypes.KeyTypeHash {
				gsi["hash_key"] = aws.ToString(attribute.AttributeName)
			}

			if attribute.KeyType == awstypes.KeyTypeRange {
				gsi["range_key"] = aws.ToString(attribute.AttributeName)
			}
		}

		if g.Projection != nil {
			gsi["projection_type"] = g.Projection.ProjectionType
			gsi["non_key_attributes"] = g.Projection.NonKeyAttributes
		}

		output = append(output, gsi)
	}

	return output
}

func flattenTableServerSideEncryption(description *awstypes.SSEDescription) []interface{} {
	if description == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrEnabled:   description.Status == awstypes.SSEStatusEnabled,
		names.AttrKMSKeyARN: aws.ToString(description.KMSMasterKeyArn),
	}

	return []interface{}{m}
}

func expandAttributes(cfg []interface{}) []awstypes.AttributeDefinition {
	attributes := make([]awstypes.AttributeDefinition, len(cfg))
	for i, attribute := range cfg {
		attr := attribute.(map[string]interface{})
		attributes[i] = awstypes.AttributeDefinition{
			AttributeName: aws.String(attr[names.AttrName].(string)),
			AttributeType: awstypes.ScalarAttributeType(attr[names.AttrType].(string)),
		}
	}
	return attributes
}

func flattenReplicaDescription(apiObject *awstypes.ReplicaDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.KMSMasterKeyId != nil {
		tfMap[names.AttrKMSKeyARN] = aws.ToString(apiObject.KMSMasterKeyId)
	}

	if apiObject.RegionName != nil {
		tfMap["region_name"] = aws.ToString(apiObject.RegionName)
	}

	return tfMap
}

func flattenReplicaDescriptions(apiObjects []awstypes.ReplicaDescription) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenReplicaDescription(&apiObject))
	}

	return tfList
}

func flattenTTL(ttlOutput *dynamodb.DescribeTimeToLiveOutput) []interface{} {
	m := map[string]interface{}{
		names.AttrEnabled: false,
	}

	if ttlOutput == nil || ttlOutput.TimeToLiveDescription == nil {
		return []interface{}{m}
	}

	ttlDesc := ttlOutput.TimeToLiveDescription

	m["attribute_name"] = aws.ToString(ttlDesc.AttributeName)
	m[names.AttrEnabled] = (ttlDesc.TimeToLiveStatus == awstypes.TimeToLiveStatusEnabled)

	return []interface{}{m}
}

func flattenPITR(pitrDesc *dynamodb.DescribeContinuousBackupsOutput) []interface{} {
	m := map[string]interface{}{
		names.AttrEnabled: false,
	}

	if pitrDesc == nil {
		return []interface{}{m}
	}

	if pitrDesc.ContinuousBackupsDescription != nil {
		pitr := pitrDesc.ContinuousBackupsDescription.PointInTimeRecoveryDescription
		if pitr != nil {
			m[names.AttrEnabled] = (pitr.PointInTimeRecoveryStatus == awstypes.PointInTimeRecoveryStatusEnabled)
		}
	}

	return []interface{}{m}
}

// TODO: Get rid of keySchemaM - the user should just explicitly define
// this in the config, we shouldn't magically be setting it like this.
// Removal will however require config change, hence BC. :/
func expandLocalSecondaryIndexes(cfg []interface{}, keySchemaM map[string]interface{}) []awstypes.LocalSecondaryIndex {
	indexes := make([]awstypes.LocalSecondaryIndex, len(cfg))
	for i, lsi := range cfg {
		m := lsi.(map[string]interface{})
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

func expandImportTable(data map[string]interface{}) *dynamodb.ImportTableInput {
	a := &dynamodb.ImportTableInput{
		ClientToken: aws.String(id.UniqueId()),
	}

	if v, ok := data["input_compression_type"].(string); ok {
		a.InputCompressionType = awstypes.InputCompressionType(v)
	}

	if v, ok := data["input_format"].(string); ok {
		a.InputFormat = awstypes.InputFormat(v)
	}

	if v, ok := data["input_format_options"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.InputFormatOptions = expandInputFormatOptions(v)
	}

	if v, ok := data["s3_bucket_source"].([]interface{}); ok && len(v) > 0 {
		a.S3BucketSource = expandS3BucketSource(v[0].(map[string]interface{}))
	}

	return a
}

func expandGlobalSecondaryIndex(data map[string]interface{}, billingMode awstypes.BillingMode) *awstypes.GlobalSecondaryIndex {
	return &awstypes.GlobalSecondaryIndex{
		IndexName:             aws.String(data[names.AttrName].(string)),
		KeySchema:             expandKeySchema(data),
		Projection:            expandProjection(data),
		ProvisionedThroughput: expandProvisionedThroughput(data, billingMode),
	}
}

func expandProvisionedThroughput(data map[string]interface{}, billingMode awstypes.BillingMode) *awstypes.ProvisionedThroughput {
	return expandProvisionedThroughputUpdate("", data, billingMode, "")
}

func expandProvisionedThroughputUpdate(id string, data map[string]interface{}, billingMode, oldBillingMode awstypes.BillingMode) *awstypes.ProvisionedThroughput {
	if billingMode == awstypes.BillingModePayPerRequest {
		return nil
	}

	return &awstypes.ProvisionedThroughput{
		ReadCapacityUnits:  aws.Int64(expandProvisionedThroughputField(id, data, "read_capacity", billingMode, oldBillingMode)),
		WriteCapacityUnits: aws.Int64(expandProvisionedThroughputField(id, data, "write_capacity", billingMode, oldBillingMode)),
	}
}

func expandProvisionedThroughputField(id string, data map[string]interface{}, key string, billingMode, oldBillingMode awstypes.BillingMode) int64 {
	v := data[key].(int)
	if v == 0 && billingMode == awstypes.BillingModeProvisioned && oldBillingMode == awstypes.BillingModePayPerRequest {
		log.Printf("[WARN] Overriding %[1]s on DynamoDB Table (%[2]s) to %[3]d. Switching from billing mode %[4]q to %[5]q without value for %[1]s. Assuming changes are being ignored.",
			key, id, provisionedThroughputMinValue, oldBillingMode, billingMode)
		v = provisionedThroughputMinValue
	}
	return int64(v)
}

func expandProjection(data map[string]interface{}) *awstypes.Projection {
	projection := &awstypes.Projection{
		ProjectionType: awstypes.ProjectionType(data["projection_type"].(string)),
	}

	if v, ok := data["non_key_attributes"].([]interface{}); ok && len(v) > 0 {
		projection.NonKeyAttributes = flex.ExpandStringValueList(v)
	}

	if v, ok := data["non_key_attributes"].(*schema.Set); ok && v.Len() > 0 {
		projection.NonKeyAttributes = flex.ExpandStringValueSet(v)
	}

	return projection
}

func expandKeySchema(data map[string]interface{}) []awstypes.KeySchemaElement {
	keySchema := []awstypes.KeySchemaElement{}

	if v, ok := data["hash_key"]; ok && v != nil && v != "" {
		keySchema = append(keySchema, awstypes.KeySchemaElement{
			AttributeName: aws.String(v.(string)),
			KeyType:       awstypes.KeyTypeHash,
		})
	}

	if v, ok := data["range_key"]; ok && v != nil && v != "" {
		keySchema = append(keySchema, awstypes.KeySchemaElement{
			AttributeName: aws.String(v.(string)),
			KeyType:       awstypes.KeyTypeRange,
		})
	}

	return keySchema
}

func expandEncryptAtRestOptions(vOptions []interface{}) *awstypes.SSESpecification {
	options := &awstypes.SSESpecification{}

	enabled := false
	if len(vOptions) > 0 {
		mOptions := vOptions[0].(map[string]interface{})

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

func expandInputFormatOptions(data []interface{}) *awstypes.InputFormatOptions {
	if data == nil {
		return nil
	}

	m := data[0].(map[string]interface{})
	a := &awstypes.InputFormatOptions{}

	if v, ok := m["csv"].([]interface{}); ok && len(v) > 0 {
		a.Csv = &awstypes.CsvOptions{}

		csv := v[0].(map[string]interface{})

		if s, ok := csv["delimiter"].(string); ok && s != "" {
			a.Csv.Delimiter = aws.String(s)
		}

		if s, ok := csv["header_list"].(*schema.Set); ok && s.Len() > 0 {
			a.Csv.HeaderList = flex.ExpandStringValueSet(s)
		}
	}

	return a
}

func expandS3BucketSource(data map[string]interface{}) *awstypes.S3BucketSource {
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

func validateTableAttributes(d *schema.ResourceDiff) error {
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
			index := idx.(map[string]interface{})
			rangeKey := index["range_key"].(string)
			indexedAttributes[rangeKey] = true
		}
	}
	if v, ok := d.GetOk("global_secondary_index"); ok {
		indexes := v.(*schema.Set).List()
		for _, idx := range indexes {
			index := idx.(map[string]interface{})

			hashKey := index["hash_key"].(string)
			indexedAttributes[hashKey] = true

			if rk, ok := index["range_key"].(string); ok && rk != "" {
				indexedAttributes[rk] = true
			}
		}
	}

	// Check if all indexed attributes have an attribute definition
	attributes := d.Get("attribute").(*schema.Set).List()
	unindexedAttributes := []string{}
	for _, attr := range attributes {
		attribute := attr.(map[string]interface{})
		attrName := attribute[names.AttrName].(string)

		if _, ok := indexedAttributes[attrName]; !ok {
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

func validateGSIProvisionedThroughput(data map[string]interface{}, billingMode awstypes.BillingMode) error {
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

func validateProvisionedThroughputField(diff *schema.ResourceDiff, key string) error {
	oldBillingMode, newBillingMode := diff.GetChange("billing_mode")
	v := diff.Get(key).(int)
	if oldBillingMode, newBillingMode := awstypes.BillingMode(oldBillingMode.(string)), awstypes.BillingMode(newBillingMode.(string)); newBillingMode == awstypes.BillingModeProvisioned {
		if v < provisionedThroughputMinValue {
			// Assuming the field is ignored, likely due to autoscaling
			if oldBillingMode == awstypes.BillingModePayPerRequest {
				return nil
			}
			return fmt.Errorf("%s must be at least 1 when billing_mode is %q", key, newBillingMode)
		}
	} else if newBillingMode == awstypes.BillingModePayPerRequest && oldBillingMode != awstypes.BillingModeProvisioned {
		if v != 0 {
			return fmt.Errorf("%s can not be set when billing_mode is %q", key, awstypes.BillingModePayPerRequest)
		}
	}
	return nil
}

func validateTTLCustomDiff(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	var diags diag.Diagnostics

	configRaw := d.GetRawConfig()
	if !configRaw.IsKnown() || configRaw.IsNull() {
		return nil
	}

	ttlPath := cty.GetAttrPath("ttl")
	ttl := configRaw.GetAttr("ttl")
	if ttl.IsKnown() && !ttl.IsNull() {
		if ttl.LengthInt() == 1 {
			idx := cty.NumberIntVal(0)
			ttl := ttl.Index(idx)
			ttlPath := ttlPath.Index(idx)
			ttlPlantimeValidate(ttlPath, ttl, &diags)
		}
	}

	return sdkdiag.DiagnosticsError(diags)
}

func ttlPlantimeValidate(ttlPath cty.Path, ttl cty.Value, diags *diag.Diagnostics) {
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
				"Attribute %q cannot have an empty value",
				errs.PathString(ttlPath.GetAttr("attribute_name")),
			))
		}
	} else {
		if !(attribute.IsNull() || attribute.AsString() == "") {
			*diags = append(*diags, errs.NewAttributeConflictsWhenError(
				ttlPath.GetAttr("attribute_name"),
				ttlPath.GetAttr(names.AttrEnabled),
				"false",
			))
		}
	}
}
