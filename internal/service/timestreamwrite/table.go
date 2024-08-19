// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamwrite

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_timestreamwrite_table", name="Table")
// @Tags(identifierAttribute="arn")
func resourceTable() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTableCreate,
		ReadWithoutTimeout:   resourceTableRead,
		UpdateWithoutTimeout: resourceTableUpdate,
		DeleteWithoutTimeout: resourceTableDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
				),
			},
			"magnetic_store_write_properties": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_magnetic_store_writes": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"magnetic_store_rejected_data_location": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrBucketName: {
													Type:     schema.TypeString,
													Optional: true,
												},
												"encryption_option": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[types.S3EncryptionOption](),
												},
												names.AttrKMSKeyID: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"object_key_prefix": {
													Type:     schema.TypeString,
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
			"retention_properties": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"magnetic_store_retention_period_in_days": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 73000),
						},
						"memory_store_retention_period_in_hours": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 8766),
						},
					},
				},
			},
			names.AttrSchema: {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"composite_partition_key": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enforcement_in_record": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.PartitionKeyEnforcementLevel](),
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									names.AttrType: {
										Type:             schema.TypeString,
										Required:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[types.PartitionKeyType](),
									},
								},
							},
						},
					},
				},
			},
			names.AttrTableName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TimestreamWriteClient(ctx)

	databaseName := d.Get(names.AttrDatabaseName).(string)
	tableName := d.Get(names.AttrTableName).(string)
	id := tableCreateResourceID(tableName, databaseName)
	input := &timestreamwrite.CreateTableInput{
		DatabaseName: aws.String(databaseName),
		TableName:    aws.String(tableName),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("magnetic_store_write_properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{}) != nil {
		input.MagneticStoreWriteProperties = expandMagneticStoreWriteProperties(v.([]interface{}))
	}

	if v, ok := d.GetOk("retention_properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{}) != nil {
		input.RetentionProperties = expandRetentionProperties(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrSchema); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Schema = expandSchema(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.CreateTable(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Timestream Table (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceTableRead(ctx, d, meta)...)
}

func resourceTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TimestreamWriteClient(ctx)

	tableName, databaseName, err := tableParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	table, err := findTableByTwoPartKey(ctx, conn, databaseName, tableName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Timestream Table %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Timestream Table (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, table.Arn)
	d.Set(names.AttrDatabaseName, table.DatabaseName)
	if err := d.Set("magnetic_store_write_properties", flattenMagneticStoreWriteProperties(table.MagneticStoreWriteProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting magnetic_store_write_properties: %s", err)
	}
	if err := d.Set("retention_properties", flattenRetentionProperties(table.RetentionProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting retention_properties: %s", err)
	}
	if table.Schema != nil {
		if err := d.Set(names.AttrSchema, []interface{}{flattenSchema(table.Schema)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting schema: %s", err)
		}
	} else {
		d.Set(names.AttrSchema, nil)
	}
	d.Set(names.AttrTableName, table.TableName)

	return diags
}

func resourceTableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TimestreamWriteClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		tableName, databaseName, err := tableParseResourceID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &timestreamwrite.UpdateTableInput{
			DatabaseName: aws.String(databaseName),
			TableName:    aws.String(tableName),
		}

		if d.HasChange("magnetic_store_write_properties") {
			input.MagneticStoreWriteProperties = expandMagneticStoreWriteProperties(d.Get("magnetic_store_write_properties").([]interface{}))
		}

		if d.HasChange("retention_properties") {
			input.RetentionProperties = expandRetentionProperties(d.Get("retention_properties").([]interface{}))
		}

		if d.HasChange(names.AttrSchema) {
			if v, ok := d.GetOk(names.AttrSchema); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.Schema = expandSchema(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		_, err = conn.UpdateTable(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Timestream Table (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTableRead(ctx, d, meta)...)
}

func resourceTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TimestreamWriteClient(ctx)

	tableName, databaseName, err := tableParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting Timestream Table: %s", d.Id())
	_, err = conn.DeleteTable(ctx, &timestreamwrite.DeleteTableInput{
		DatabaseName: aws.String(databaseName),
		TableName:    aws.String(tableName),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Timestream Table (%s): %s", d.Id(), err)
	}

	return diags
}

func findTableByTwoPartKey(ctx context.Context, conn *timestreamwrite.Client, databaseName, tableName string) (*types.Table, error) {
	input := &timestreamwrite.DescribeTableInput{
		DatabaseName: aws.String(databaseName),
		TableName:    aws.String(tableName),
	}

	output, err := conn.DescribeTable(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Table == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Table, nil
}

func expandRetentionProperties(tfList []interface{}) *types.RetentionProperties {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &types.RetentionProperties{}

	if v, ok := tfMap["magnetic_store_retention_period_in_days"].(int); ok {
		apiObject.MagneticStoreRetentionPeriodInDays = aws.Int64(int64(v))
	}

	if v, ok := tfMap["memory_store_retention_period_in_hours"].(int); ok {
		apiObject.MemoryStoreRetentionPeriodInHours = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenRetentionProperties(apiObject *types.RetentionProperties) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"magnetic_store_retention_period_in_days": aws.ToInt64(apiObject.MagneticStoreRetentionPeriodInDays),
		"memory_store_retention_period_in_hours":  aws.ToInt64(apiObject.MemoryStoreRetentionPeriodInHours),
	}

	return []interface{}{tfMap}
}

func expandMagneticStoreWriteProperties(tfList []interface{}) *types.MagneticStoreWriteProperties {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &types.MagneticStoreWriteProperties{
		EnableMagneticStoreWrites: aws.Bool(tfMap["enable_magnetic_store_writes"].(bool)),
	}

	if v, ok := tfMap["magnetic_store_rejected_data_location"].([]interface{}); ok && len(v) > 0 {
		apiObject.MagneticStoreRejectedDataLocation = expandMagneticStoreRejectedDataLocation(v)
	}

	return apiObject
}

func flattenMagneticStoreWriteProperties(apiObject *types.MagneticStoreWriteProperties) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"enable_magnetic_store_writes":          aws.ToBool(apiObject.EnableMagneticStoreWrites),
		"magnetic_store_rejected_data_location": flattenMagneticStoreRejectedDataLocation(apiObject.MagneticStoreRejectedDataLocation),
	}

	return []interface{}{tfMap}
}

func expandMagneticStoreRejectedDataLocation(tfList []interface{}) *types.MagneticStoreRejectedDataLocation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &types.MagneticStoreRejectedDataLocation{}

	if v, ok := tfMap["s3_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.S3Configuration = expandS3Configuration(v)
	}

	return apiObject
}

func flattenMagneticStoreRejectedDataLocation(apiObject *types.MagneticStoreRejectedDataLocation) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"s3_configuration": flattenS3Configuration(apiObject.S3Configuration),
	}

	return []interface{}{tfMap}
}

func expandS3Configuration(tfList []interface{}) *types.S3Configuration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &types.S3Configuration{}

	if v, ok := tfMap[names.AttrBucketName].(string); ok && v != "" {
		apiObject.BucketName = aws.String(v)
	}

	if v, ok := tfMap["encryption_option"].(string); ok && v != "" {
		apiObject.EncryptionOption = types.S3EncryptionOption(v)
	}

	if v, ok := tfMap[names.AttrKMSKeyID].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["object_key_prefix"].(string); ok && v != "" {
		apiObject.ObjectKeyPrefix = aws.String(v)
	}

	return apiObject
}

func flattenS3Configuration(apiObject *types.S3Configuration) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		names.AttrBucketName: aws.ToString(apiObject.BucketName),
		"encryption_option":  string(apiObject.EncryptionOption),
		names.AttrKMSKeyID:   aws.ToString(apiObject.KmsKeyId),
		"object_key_prefix":  aws.ToString(apiObject.ObjectKeyPrefix),
	}

	return []interface{}{tfMap}
}

func expandSchema(tfMap map[string]interface{}) *types.Schema {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Schema{}

	if v, ok := tfMap["composite_partition_key"].([]interface{}); ok && len(v) > 0 {
		apiObject.CompositePartitionKey = expandPartitionKeys(v)
	}

	return apiObject
}

func expandPartitionKey(tfMap map[string]interface{}) *types.PartitionKey {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PartitionKey{}

	if v, ok := tfMap["enforcement_in_record"].(string); ok && v != "" {
		apiObject.EnforcementInRecord = types.PartitionKeyEnforcementLevel(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = types.PartitionKeyType(v)
	}

	return apiObject
}

func expandPartitionKeys(tfList []interface{}) []types.PartitionKey {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.PartitionKey

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandPartitionKey(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenSchema(apiObject *types.Schema) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CompositePartitionKey; v != nil {
		tfMap["composite_partition_key"] = flattenPartitionKeys(v)
	}

	return tfMap
}

func flattenPartitionKey(apiObject *types.PartitionKey) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"enforcement_in_record": apiObject.EnforcementInRecord,
		names.AttrType:          apiObject.Type,
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	return tfMap
}

func flattenPartitionKeys(apiObjects []types.PartitionKey) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPartitionKey(&apiObject))
	}

	return tfList
}

const tableIDSeparator = ":"

func tableCreateResourceID(tableName, databaseName string) string {
	parts := []string{tableName, databaseName}
	id := strings.Join(parts, tableIDSeparator)

	return id
}

func tableParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, tableIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected table_name%[2]sdatabase_name", id, tableIDSeparator)
	}

	return parts[0], parts[1], nil
}
