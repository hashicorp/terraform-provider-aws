package dms

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceS3Endpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceS3EndpointCreate,
		Read:   resourceS3EndpointRead,
		Update: resourceS3EndpointUpdate,
		Delete: resourceS3EndpointDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"endpoint_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validEndpointID,
			},
			"endpoint_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(dms.ReplicationEndpointTypeValue_Values(), false),
			},
			"engine_display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"external_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssl_mode": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(dms.DmsSslModeValue_Values(), false),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),

			/////// S3-Specific Settings
			"add_column_name": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"bucket_folder": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"bucket_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"canned_acl_for_objects": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      dms.CannedAclForObjectsValueNone,
				ValidateFunc: validation.StringInSlice(dms.CannedAclForObjectsValue_Values(), true),
				StateFunc: func(v interface{}) string {
					return strings.ToLower(v.(string))
				},
			},
			"cdc_inserts_and_updates": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"cdc_inserts_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"cdc_max_batch_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      60,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"cdc_min_file_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      32,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"cdc_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"compression_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      dms.CompressionTypeValueNone,
				ValidateFunc: validation.StringInSlice(dms.CompressionTypeValue_Values(), true),
			},
			"csv_delimiter": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ",",
			},
			"csv_no_sup_value": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"csv_null_value": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "NULL",
			},
			"csv_row_delimiter": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "\\n",
			},
			"data_format": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      dms.DataFormatValueCsv,
				ValidateFunc: validation.StringInSlice(dms.DataFormatValue_Values(), false),
			},
			"data_page_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1048576,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"date_partition_delimiter": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      strings.ToUpper(dms.DatePartitionDelimiterValueSlash),
				ValidateFunc: validation.StringInSlice(dms.DatePartitionDelimiterValue_Values(), true),
				StateFunc: func(v interface{}) string {
					return strings.ToUpper(v.(string))
				},
			},
			"date_partition_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"date_partition_sequence": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      dms.DatePartitionSequenceValueYyyymmdd,
				ValidateFunc: validation.StringInSlice(dms.DatePartitionSequenceValue_Values(), true),
				StateFunc: func(v interface{}) string {
					return strings.ToLower(v.(string))
				},
			},
			"date_partition_timezone": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dict_page_size_limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1048576,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"enable_statistics": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"encoding_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      dms.EncodingTypeValueRleDictionary,
				ValidateFunc: validation.StringInSlice(dms.EncodingTypeValue_Values(), false),
			},
			"encryption_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      s3SettingsEncryptionModeSseS3,
				ValidateFunc: validation.StringInSlice(s3SettingsEncryptionMode_Values(), false),
			},
			"external_table_definition": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ignore_header_rows": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntInSlice([]int{0, 1}),
			},
			"include_op_for_full_load": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"max_file_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1048576,
				ValidateFunc: validation.IntBetween(1, 1048576),
			},
			"parquet_timestamp_in_millisecond": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"parquet_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      dms.ParquetVersionValueParquet10,
				ValidateFunc: validation.StringInSlice(dms.ParquetVersionValue_Values(), false),
			},
			"preserve_transactions": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"rfc_4180": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"row_group_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      10000,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"service_access_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"timestamp_column_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"use_csv_no_sup_value": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"use_task_start_time_for_full_load_timestamp": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceS3EndpointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DMSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	request := &dms.CreateEndpointInput{
		EndpointIdentifier: aws.String(d.Get("endpoint_id").(string)),
		EndpointType:       aws.String(d.Get("endpoint_type").(string)),
		EngineName:         aws.String("s3"),
		Tags:               Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("ssl_mode"); ok {
		request.SslMode = aws.String(v.(string))
	}

	request.S3Settings = &dms.S3Settings{}

	if v, ok := d.GetOk("add_column_name"); ok {
		request.S3Settings.AddColumnName = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("bucket_folder"); ok {
		request.S3Settings.BucketFolder = aws.String(v.(string))
	}

	if v, ok := d.GetOk("bucket_name"); ok {
		request.S3Settings.BucketName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("canned_acl_for_objects"); ok {
		request.S3Settings.CannedAclForObjects = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cdc_inserts_and_updates"); ok {
		request.S3Settings.CdcInsertsAndUpdates = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("cdc_inserts_only"); ok {
		request.S3Settings.CdcInsertsOnly = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("cdc_max_batch_interval"); ok {
		request.S3Settings.CdcMaxBatchInterval = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("cdc_min_file_size"); ok {
		request.S3Settings.CdcMinFileSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("cdc_path"); ok {
		request.S3Settings.CdcPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("compression_type"); ok {
		request.S3Settings.CompressionType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("csv_delimiter"); ok {
		request.S3Settings.CsvDelimiter = aws.String(v.(string))
	}

	if v, ok := d.GetOk("csv_no_sup_value"); ok {
		request.S3Settings.CsvNoSupValue = aws.String(v.(string))
	}

	if v, ok := d.GetOk("csv_null_value"); ok {
		request.S3Settings.CsvNullValue = aws.String(v.(string))
	}

	if v, ok := d.GetOk("csv_row_delimiter"); ok {
		request.S3Settings.CsvRowDelimiter = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_format"); ok {
		request.S3Settings.DataFormat = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_page_size"); ok {
		request.S3Settings.DataPageSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("date_partition_delimiter"); ok {
		request.S3Settings.DatePartitionDelimiter = aws.String(v.(string))
	}

	if v, ok := d.GetOk("date_partition_enabled"); ok {
		request.S3Settings.DatePartitionEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("date_partition_sequence"); ok {
		request.S3Settings.DatePartitionSequence = aws.String(v.(string))
	}

	if v, ok := d.GetOk("date_partition_timezone"); ok {
		request.S3Settings.DatePartitionTimezone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dict_page_size_limit"); ok {
		request.S3Settings.DictPageSizeLimit = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("enable_statistics"); !ok {
		request.S3Settings.EnableStatistics = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("encoding_type"); ok {
		request.S3Settings.EncodingType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("encryption_mode"); ok {
		request.S3Settings.EncryptionMode = aws.String(v.(string))
	}

	if v, ok := d.GetOk("external_table_definition"); ok {
		request.S3Settings.ExternalTableDefinition = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ignore_header_rows"); ok {
		request.S3Settings.IgnoreHeaderRows = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("include_op_for_full_load"); ok {
		request.S3Settings.IncludeOpForFullLoad = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("max_file_size"); ok {
		request.S3Settings.MaxFileSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("parquet_timestamp_in_millisecond"); ok {
		request.S3Settings.ParquetTimestampInMillisecond = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("parquet_version"); ok {
		request.S3Settings.ParquetVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("preserve_transactions"); ok {
		request.S3Settings.PreserveTransactions = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("rfc_4180"); ok {
		request.S3Settings.Rfc4180 = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("row_group_length"); ok {
		request.S3Settings.RowGroupLength = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("service_access_role_arn"); ok {
		request.S3Settings.ServiceAccessRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("timestamp_column_name"); ok {
		request.S3Settings.TimestampColumnName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("use_csv_no_sup_value"); ok {
		request.S3Settings.UseCsvNoSupValue = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("use_task_start_time_for_full_load_timestamp"); ok {
		request.S3Settings.UseTaskStartTimeForFullLoadTimestamp = aws.Bool(v.(bool))
	}

	log.Println("[DEBUG] DMS create endpoint:", request)

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.CreateEndpoint(request)

		if tfawserr.ErrCodeEquals(err, "AccessDeniedFault") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.CreateEndpoint(request)
	}

	if err != nil {
		return fmt.Errorf("Error creating DMS endpoint: %s", err)
	}

	d.SetId(d.Get("endpoint_id").(string))

	return resourceS3EndpointRead(d, meta)
}

func resourceS3EndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DMSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	endpoint, err := FindEndpointByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DMS Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DMS Endpoint (%s): %w", d.Id(), err)
	}

	if endpoint.S3Settings == nil {
		return fmt.Errorf("reading DMS S3 Endpoint (%s): no S3 settings returned", d.Id())
	}

	d.Set("endpoint_arn", endpoint.EndpointArn)
	d.Set("endpoint_id", endpoint.EndpointIdentifier)
	d.Set("endpoint_type", strings.ToLower(*endpoint.EndpointType)) // For some reason the AWS API only accepts lowercase type but returns it as uppercase
	d.Set("engine_display_name", endpoint.EngineDisplayName)
	d.Set("external_id", endpoint.ExternalId)
	d.Set("external_table_definition", endpoint.ExternalTableDefinition)
	d.Set("service_access_role_arn", endpoint.ServiceAccessRoleArn)
	d.Set("ssl_mode", endpoint.SslMode)
	d.Set("status", endpoint.Status)

	s3settings := endpoint.S3Settings
	d.Set("add_column_name", s3settings.AddColumnName)
	d.Set("bucket_folder", s3settings.BucketFolder)
	d.Set("bucket_name", s3settings.BucketName)
	d.Set("canned_acl_for_objects", s3settings.CannedAclForObjects)
	d.Set("cdc_inserts_and_updates", s3settings.CdcInsertsAndUpdates)
	d.Set("cdc_inserts_only", s3settings.CdcInsertsOnly)
	d.Set("cdc_max_batch_interval", s3settings.CdcMaxBatchInterval)
	d.Set("cdc_min_file_size", s3settings.CdcMinFileSize)
	d.Set("cdc_path", s3settings.CdcPath)
	d.Set("compression_type", s3settings.CompressionType)
	d.Set("csv_delimiter", s3settings.CsvDelimiter)
	d.Set("csv_no_sup_value", s3settings.CsvNoSupValue)
	d.Set("csv_null_value", s3settings.CsvNullValue)
	d.Set("csv_row_delimiter", s3settings.CsvRowDelimiter)
	d.Set("data_format", s3settings.DataFormat)
	d.Set("data_page_size", s3settings.DataPageSize)
	d.Set("date_partition_delimiter", strings.ToUpper(aws.StringValue(s3settings.DatePartitionDelimiter)))
	d.Set("date_partition_enabled", s3settings.DatePartitionEnabled)
	d.Set("date_partition_sequence", s3settings.DatePartitionSequence)
	d.Set("date_partition_timezone", s3settings.DatePartitionTimezone)
	d.Set("dict_page_size_limit", s3settings.DictPageSizeLimit)
	d.Set("enable_statistics", s3settings.EnableStatistics)
	d.Set("encoding_type", s3settings.EncodingType)
	d.Set("encryption_mode", s3settings.EncryptionMode)
	d.Set("external_table_definition", s3settings.ExternalTableDefinition)
	d.Set("ignore_header_rows", s3settings.IgnoreHeaderRows)
	d.Set("include_op_for_full_load", s3settings.IncludeOpForFullLoad)
	d.Set("max_file_size", s3settings.MaxFileSize)
	d.Set("parquet_timestamp_in_millisecond", s3settings.ParquetTimestampInMillisecond)
	d.Set("parquet_version", s3settings.ParquetVersion)
	d.Set("preserve_transactions", s3settings.PreserveTransactions)
	d.Set("rfc_4180", s3settings.Rfc4180)
	d.Set("row_group_length", s3settings.RowGroupLength)
	d.Set("service_access_role_arn", s3settings.ServiceAccessRoleArn)
	d.Set("timestamp_column_name", s3settings.TimestampColumnName)
	d.Set("use_csv_no_sup_value", s3settings.UseCsvNoSupValue)
	d.Set("use_task_start_time_for_full_load_timestamp", s3settings.UseTaskStartTimeForFullLoadTimestamp)

	tags, err := ListTags(conn, d.Get("endpoint_arn").(string))
	if err != nil {
		return fmt.Errorf("error listing tags for DMS Endpoint (%s): %w", d.Get("endpoint_arn").(string), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceS3EndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DMSConn

	request := &dms.ModifyEndpointInput{
		EndpointArn: aws.String(d.Get("endpoint_arn").(string)),
	}
	hasChanges := false

	request.EngineName = aws.String(engineNameS3)

	if d.HasChange("endpoint_type") {
		request.EndpointType = aws.String(d.Get("endpoint_type").(string))
		hasChanges = true
	}

	if d.HasChange("ssl_mode") {
		request.SslMode = aws.String(d.Get("ssl_mode").(string))
		hasChanges = true
	}

	if d.HasChangesExcept(
		"endpoint_type",
		"ssl_mode",
		"tags_all",
	) {
		request.S3Settings = &dms.S3Settings{}

		request.S3Settings.AddColumnName = aws.Bool(d.Get("add_column_name").(bool))
		request.S3Settings.BucketFolder = aws.String(d.Get("bucket_folder").(string))
		request.S3Settings.CannedAclForObjects = aws.String(d.Get("canned_acl_for_objects").(string))
		request.S3Settings.CdcInsertsAndUpdates = aws.Bool(d.Get("cdc_inserts_and_updates").(bool))
		request.S3Settings.CdcInsertsOnly = aws.Bool(d.Get("cdc_inserts_only").(bool))
		request.S3Settings.CdcMaxBatchInterval = aws.Int64(int64(d.Get("cdc_max_batch_interval").(int)))
		request.S3Settings.CdcMinFileSize = aws.Int64(int64(d.Get("cdc_min_file_size").(int)))
		request.S3Settings.CdcPath = aws.String(d.Get("cdc_path").(string))
		request.S3Settings.CompressionType = aws.String(d.Get("compression_type").(string))
		request.S3Settings.CsvDelimiter = aws.String(d.Get("csv_delimiter").(string))
		request.S3Settings.CsvNoSupValue = aws.String(d.Get("csv_no_sup_value").(string))
		request.S3Settings.CsvNullValue = aws.String(d.Get("csv_null_value").(string))
		request.S3Settings.CsvRowDelimiter = aws.String(d.Get("csv_row_delimiter").(string))
		request.S3Settings.DataFormat = aws.String(d.Get("data_format").(string))
		request.S3Settings.DataPageSize = aws.Int64(int64(d.Get("data_page_size").(int)))
		request.S3Settings.DatePartitionDelimiter = aws.String(d.Get("date_partition_delimiter").(string))
		request.S3Settings.DatePartitionEnabled = aws.Bool(d.Get("date_partition_enabled").(bool))
		request.S3Settings.DatePartitionSequence = aws.String(d.Get("date_partition_sequence").(string))
		request.S3Settings.DatePartitionTimezone = aws.String(d.Get("date_partition_timezone").(string))
		request.S3Settings.DictPageSizeLimit = aws.Int64(int64(d.Get("dict_page_size_limit").(int)))
		request.S3Settings.EnableStatistics = aws.Bool(d.Get("enable_statistics").(bool))
		request.S3Settings.EncodingType = aws.String(d.Get("encoding_type").(string))
		request.S3Settings.EncryptionMode = aws.String(d.Get("encryption_mode").(string))
		request.S3Settings.ExternalTableDefinition = aws.String(d.Get("external_table_definition").(string))
		request.S3Settings.IgnoreHeaderRows = aws.Int64(int64(d.Get("ignore_header_rows").(int)))
		request.S3Settings.IncludeOpForFullLoad = aws.Bool(d.Get("include_op_for_full_load").(bool))
		request.S3Settings.MaxFileSize = aws.Int64(int64(d.Get("max_file_size").(int)))
		request.S3Settings.ParquetTimestampInMillisecond = aws.Bool(d.Get("parquet_timestamp_in_millisecond").(bool))
		request.S3Settings.ParquetVersion = aws.String(d.Get("parquet_version").(string))
		request.S3Settings.PreserveTransactions = aws.Bool(d.Get("preserve_transactions").(bool))
		request.S3Settings.Rfc4180 = aws.Bool(d.Get("rfc_4180").(bool))
		request.S3Settings.RowGroupLength = aws.Int64(int64(d.Get("row_group_length").(int)))
		request.S3Settings.ServiceAccessRoleArn = aws.String(d.Get("service_access_role_arn").(string))
		request.S3Settings.TimestampColumnName = aws.String(d.Get("timestamp_column_name").(string))
		request.S3Settings.UseCsvNoSupValue = aws.Bool(d.Get("use_csv_no_sup_value").(bool))
		request.S3Settings.UseTaskStartTimeForFullLoadTimestamp = aws.Bool(d.Get("use_task_start_time_for_full_load_timestamp").(bool))

		request.ServiceAccessRoleArn = aws.String(d.Get("service_access_role_arn").(string))

		hasChanges = true
	}

	if d.HasChange("tags_all") {
		arn := d.Get("endpoint_arn").(string)
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating DMS Endpoint (%s) tags: %s", arn, err)
		}
	}

	if hasChanges {
		log.Println("[DEBUG] DMS update endpoint:", request)

		_, err := conn.ModifyEndpoint(request)
		if err != nil {
			return err
		}

		return resourceS3EndpointRead(d, meta)
	}

	return nil
}

func resourceS3EndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DMSConn

	log.Printf("[DEBUG] Deleting DMS Endpoint: (%s)", d.Id())
	_, err := conn.DeleteEndpoint(&dms.DeleteEndpointInput{
		EndpointArn: aws.String(d.Get("endpoint_arn").(string)),
	})

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DMS Endpoint (%s): %w", d.Id(), err)
	}

	_, err = waitEndpointDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for DMS Endpoint (%s) delete: %w", d.Id(), err)
	}

	return err
}
