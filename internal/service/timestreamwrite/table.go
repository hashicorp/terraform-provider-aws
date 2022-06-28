package timestreamwrite

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTable() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTableCreate,
		ReadWithoutTimeout:   resourceTableRead,
		UpdateWithoutTimeout: resourceTableUpdate,
		DeleteWithoutTimeout: resourceTableDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"database_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
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
												"bucket_name": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"encryption_option": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(timestreamwrite.S3EncryptionOption_Values(), false),
												},
												"kms_key_id": {
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

			"table_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
				),
			},

			"tags": tftags.TagsSchema(),

			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TimestreamWriteConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	tableName := d.Get("table_name").(string)
	input := &timestreamwrite.CreateTableInput{
		DatabaseName: aws.String(d.Get("database_name").(string)),
		TableName:    aws.String(tableName),
	}

	if v, ok := d.GetOk("retention_properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{}) != nil {
		input.RetentionProperties = expandRetentionProperties(v.([]interface{}))
	}

	if v, ok := d.GetOk("magnetic_store_write_properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{}) != nil {
		input.MagneticStoreWriteProperties = expandMagneticStoreWriteProperties(v.([]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateTableWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Timestream Table (%s): %w", tableName, err))
	}

	if output == nil || output.Table == nil {
		return diag.FromErr(fmt.Errorf("error creating Timestream Table (%s): empty output", tableName))
	}

	d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(output.Table.TableName), aws.StringValue(output.Table.DatabaseName)))

	return resourceTableRead(ctx, d, meta)
}

func resourceTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TimestreamWriteConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	tableName, databaseName, err := TableParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &timestreamwrite.DescribeTableInput{
		DatabaseName: aws.String(databaseName),
		TableName:    aws.String(tableName),
	}

	output, err := conn.DescribeTableWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, timestreamwrite.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Timestream Table %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if output == nil || output.Table == nil {
		return diag.FromErr(fmt.Errorf("error reading Timestream Table (%s): empty output", d.Id()))
	}

	table := output.Table
	arn := aws.StringValue(table.Arn)

	d.Set("arn", arn)
	d.Set("database_name", table.DatabaseName)

	if err := d.Set("retention_properties", flattenRetentionProperties(table.RetentionProperties)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting retention_properties: %w", err))
	}

	if err := d.Set("magnetic_store_write_properties", flattenMagneticStoreWriteProperties(table.MagneticStoreWriteProperties)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting magnetic_store_write_properties: %w", err))
	}

	d.Set("table_name", table.TableName)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing tags for Timestream Table (%s): %w", arn, err))
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceTableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TimestreamWriteConn

	if d.HasChangesExcept("tags", "tags_all") {
		tableName, databaseName, err := TableParseID(d.Id())

		if err != nil {
			return diag.FromErr(err)
		}

		input := &timestreamwrite.UpdateTableInput{
			DatabaseName: aws.String(databaseName),
			TableName:    aws.String(tableName),
		}

		if d.HasChange("retention_properties") {
			input.RetentionProperties = expandRetentionProperties(d.Get("retention_properties").([]interface{}))
		}

		if d.HasChange("magnetic_store_write_properties") {
			input.MagneticStoreWriteProperties = expandMagneticStoreWriteProperties(d.Get("magnetic_store_write_properties").([]interface{}))
		}

		_, err = conn.UpdateTableWithContext(ctx, input)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating Timestream Table (%s): %w", d.Id(), err))
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating Timestream Table (%s) tags: %w", d.Get("arn").(string), err))
		}
	}

	return resourceTableRead(ctx, d, meta)
}

func resourceTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TimestreamWriteConn

	tableName, databaseName, err := TableParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &timestreamwrite.DeleteTableInput{
		DatabaseName: aws.String(databaseName),
		TableName:    aws.String(tableName),
	}

	_, err = conn.DeleteTableWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, timestreamwrite.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Timestream Table (%s): %w", d.Id(), err))
	}

	return nil
}

func expandRetentionProperties(l []interface{}) *timestreamwrite.RetentionProperties {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	rp := &timestreamwrite.RetentionProperties{}

	if v, ok := tfMap["magnetic_store_retention_period_in_days"].(int); ok {
		rp.MagneticStoreRetentionPeriodInDays = aws.Int64(int64(v))
	}

	if v, ok := tfMap["memory_store_retention_period_in_hours"].(int); ok {
		rp.MemoryStoreRetentionPeriodInHours = aws.Int64(int64(v))
	}

	return rp
}

func flattenRetentionProperties(rp *timestreamwrite.RetentionProperties) []interface{} {
	if rp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"magnetic_store_retention_period_in_days": aws.Int64Value(rp.MagneticStoreRetentionPeriodInDays),
		"memory_store_retention_period_in_hours":  aws.Int64Value(rp.MemoryStoreRetentionPeriodInHours),
	}

	return []interface{}{m}
}

func expandMagneticStoreWriteProperties(l []interface{}) *timestreamwrite.MagneticStoreWriteProperties {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	rp := &timestreamwrite.MagneticStoreWriteProperties{
		EnableMagneticStoreWrites: aws.Bool(tfMap["enable_magnetic_store_writes"].(bool)),
	}

	if v, ok := tfMap["magnetic_store_rejected_data_location"].([]interface{}); ok && len(v) > 0 {
		rp.MagneticStoreRejectedDataLocation = expandMagneticStoreRejectedDataLocation(v)
	}

	return rp
}

func flattenMagneticStoreWriteProperties(rp *timestreamwrite.MagneticStoreWriteProperties) []interface{} {
	if rp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"enable_magnetic_store_writes":          aws.BoolValue(rp.EnableMagneticStoreWrites),
		"magnetic_store_rejected_data_location": flattenMagneticStoreRejectedDataLocation(rp.MagneticStoreRejectedDataLocation),
	}

	return []interface{}{m}
}

func expandMagneticStoreRejectedDataLocation(l []interface{}) *timestreamwrite.MagneticStoreRejectedDataLocation {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	rp := &timestreamwrite.MagneticStoreRejectedDataLocation{}

	if v, ok := tfMap["s3_configuration"].([]interface{}); ok && len(v) > 0 {
		rp.S3Configuration = expandS3Configuration(v)
	}

	return rp
}

func flattenMagneticStoreRejectedDataLocation(rp *timestreamwrite.MagneticStoreRejectedDataLocation) []interface{} {
	if rp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"s3_configuration": flattenS3Configuration(rp.S3Configuration),
	}

	return []interface{}{m}
}

func expandS3Configuration(l []interface{}) *timestreamwrite.S3Configuration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	rp := &timestreamwrite.S3Configuration{}

	if v, ok := tfMap["bucket_name"].(string); ok && v != "" {
		rp.BucketName = aws.String(v)
	}

	if v, ok := tfMap["object_key_prefix"].(string); ok && v != "" {
		rp.ObjectKeyPrefix = aws.String(v)
	}

	if v, ok := tfMap["kms_key_id"].(string); ok && v != "" {
		rp.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["encryption_option"].(string); ok && v != "" {
		rp.EncryptionOption = aws.String(v)
	}

	return rp
}

func flattenS3Configuration(rp *timestreamwrite.S3Configuration) []interface{} {
	if rp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"bucket_name":       aws.StringValue(rp.BucketName),
		"object_key_prefix": aws.StringValue(rp.ObjectKeyPrefix),
		"kms_key_id":        aws.StringValue(rp.KmsKeyId),
		"encryption_option": aws.StringValue(rp.EncryptionOption),
	}

	return []interface{}{m}
}

func TableParseID(id string) (string, string, error) {
	idParts := strings.SplitN(id, ":", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected table_name:database_name", id)
	}
	return idParts[0], idParts[1], nil
}
