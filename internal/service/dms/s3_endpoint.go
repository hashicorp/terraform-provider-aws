// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	dms "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
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

// @SDKResource("aws_dms_s3_endpoint", name="S3 Endpoint")
// @Tags(identifierAttribute="endpoint_arn")
func resourceS3Endpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceS3EndpointCreate,
		ReadWithoutTimeout:   resourceS3EndpointRead,
		UpdateWithoutTimeout: resourceS3EndpointUpdate,
		DeleteWithoutTimeout: resourceS3EndpointDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrCertificateARN: {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
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
			names.AttrEndpointType: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ReplicationEndpointTypeValue](),
			},
			"engine_display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrExternalID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyARN: {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"ssl_mode": {
				Type:             schema.TypeString,
				Computed:         true,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.DmsSslModeValue](),
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),

			/////// S3-Specific Settings
			"add_column_name": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"add_trailing_padding_character": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"bucket_folder": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrBucketName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"canned_acl_for_objects": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.CannedAclForObjectsValue](),
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
				ValidateFunc: validation.IntAtLeast(0),
			},
			"cdc_min_file_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"cdc_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"compression_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.CompressionTypeValue](),
				Default:          strings.ToUpper(string(awstypes.CompressionTypeValueNone)),
				StateFunc: func(v interface{}) string {
					return strings.ToUpper(v.(string))
				},
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
			},
			"csv_row_delimiter": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "\\n",
			},
			"data_format": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.DataFormatValue](),
			},
			"data_page_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"date_partition_delimiter": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.DatePartitionDelimiterValue](),
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
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.DatePartitionSequenceValue](),
				StateFunc: func(v interface{}) string {
					return strings.ToLower(v.(string))
				},
			},
			"date_partition_timezone": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"detach_target_on_lob_lookup_failure_parquet": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"dict_page_size_limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"enable_statistics": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"encoding_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.EncodingTypeValue](),
			},
			"encryption_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(encryptionMode_Values(), false),
			},
			names.AttrExpectedBucketOwner: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"external_table_definition": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"glue_catalog_generation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ignore_header_rows": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
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
				ValidateFunc: validation.IntBetween(1, 1048576),
			},
			"parquet_timestamp_in_millisecond": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"parquet_version": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ParquetVersionValue](),
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
				ValidateFunc: validation.IntAtLeast(0),
			},
			"server_side_encryption_kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"service_access_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
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

func resourceS3EndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	endpointID := d.Get("endpoint_id").(string)
	input := &dms.CreateEndpointInput{
		EndpointIdentifier: aws.String(endpointID),
		EndpointType:       awstypes.ReplicationEndpointTypeValue(d.Get(names.AttrEndpointType).(string)),
		EngineName:         aws.String("s3"),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrCertificateARN); ok {
		input.CertificateArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("service_access_role_arn"); ok {
		input.ServiceAccessRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ssl_mode"); ok {
		input.SslMode = awstypes.DmsSslModeValue(v.(string))
	}

	input.S3Settings = s3Settings(d, d.Get(names.AttrEndpointType).(string) == string(awstypes.ReplicationEndpointTypeValueTarget))

	input.ExtraConnectionAttributes = extraConnectionAnomalies(d)

	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.AccessDeniedFault](ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return conn.CreateEndpoint(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DMS S3 Endpoint (%s): %s", endpointID, err)
	}

	d.SetId(endpointID)
	d.Set("endpoint_arn", outputRaw.(*dms.CreateEndpointOutput).Endpoint.EndpointArn)

	// AWS bug? ssekki is ignored on create but sets on update
	if _, ok := d.GetOk("server_side_encryption_kms_key_id"); ok {
		return append(diags, resourceS3EndpointUpdate(ctx, d, meta)...)
	}

	return append(diags, resourceS3EndpointRead(ctx, d, meta)...)
}

func resourceS3EndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	endpoint, err := findEndpointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DMS Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err == nil && endpoint.S3Settings == nil {
		err = tfresource.NewEmptyResultError(nil)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DMS S3 Endpoint (%s): %s", d.Id(), err)
	}

	d.Set("endpoint_arn", endpoint.EndpointArn)
	d.Set(names.AttrCertificateARN, endpoint.CertificateArn)
	d.Set("endpoint_id", endpoint.EndpointIdentifier)
	d.Set(names.AttrEndpointType, strings.ToLower(string(endpoint.EndpointType))) // For some reason the AWS API only accepts lowercase type but returns it as uppercase
	d.Set("engine_display_name", endpoint.EngineDisplayName)
	d.Set(names.AttrExternalID, endpoint.ExternalId)
	// d.Set("external_table_definition", endpoint.ExternalTableDefinition) // set from s3 settings
	d.Set(names.AttrKMSKeyARN, endpoint.KmsKeyId)
	// d.Set("service_access_role_arn", endpoint.ServiceAccessRoleArn) // set from s3 settings
	d.Set("ssl_mode", endpoint.SslMode)
	d.Set(names.AttrStatus, endpoint.Status)

	setDetachTargetOnLobLookupFailureParquet(d, aws.ToString(endpoint.ExtraConnectionAttributes))

	s3settings := endpoint.S3Settings
	d.Set("add_column_name", s3settings.AddColumnName)
	d.Set("bucket_folder", s3settings.BucketFolder)
	d.Set(names.AttrBucketName, s3settings.BucketName)
	d.Set("canned_acl_for_objects", s3settings.CannedAclForObjects)
	d.Set("cdc_inserts_and_updates", s3settings.CdcInsertsAndUpdates)
	d.Set("cdc_inserts_only", s3settings.CdcInsertsOnly)
	d.Set("cdc_max_batch_interval", s3settings.CdcMaxBatchInterval)
	d.Set("cdc_min_file_size", s3settings.CdcMinFileSize)
	d.Set("cdc_path", s3settings.CdcPath)
	d.Set("csv_delimiter", s3settings.CsvDelimiter)
	d.Set("csv_null_value", s3settings.CsvNullValue)
	d.Set("csv_row_delimiter", s3settings.CsvRowDelimiter)
	d.Set("data_page_size", s3settings.DataPageSize)
	d.Set("dict_page_size_limit", s3settings.DictPageSizeLimit)
	d.Set("enable_statistics", s3settings.EnableStatistics)
	d.Set("encoding_type", s3settings.EncodingType)
	d.Set(names.AttrExpectedBucketOwner, s3settings.ExpectedBucketOwner)
	d.Set("ignore_header_rows", s3settings.IgnoreHeaderRows)
	d.Set("include_op_for_full_load", s3settings.IncludeOpForFullLoad)
	d.Set("max_file_size", s3settings.MaxFileSize)
	d.Set("rfc_4180", s3settings.Rfc4180)
	d.Set("row_group_length", s3settings.RowGroupLength)
	d.Set("service_access_role_arn", s3settings.ServiceAccessRoleArn)
	d.Set("timestamp_column_name", s3settings.TimestampColumnName)
	d.Set("use_task_start_time_for_full_load_timestamp", s3settings.UseTaskStartTimeForFullLoadTimestamp)

	if d.Get(names.AttrEndpointType).(string) == string(awstypes.ReplicationEndpointTypeValueTarget) {
		d.Set("add_trailing_padding_character", s3settings.AddTrailingPaddingCharacter)
		d.Set("compression_type", s3settings.CompressionType)
		d.Set("csv_no_sup_value", s3settings.CsvNoSupValue)
		d.Set("data_format", s3settings.DataFormat)
		d.Set("date_partition_delimiter", strings.ToUpper(string(s3settings.DatePartitionDelimiter)))
		d.Set("date_partition_enabled", s3settings.DatePartitionEnabled)
		d.Set("date_partition_sequence", s3settings.DatePartitionSequence)
		d.Set("date_partition_timezone", s3settings.DatePartitionTimezone)
		d.Set("encryption_mode", s3settings.EncryptionMode)
		d.Set("glue_catalog_generation", s3settings.GlueCatalogGeneration)
		d.Set("parquet_timestamp_in_millisecond", s3settings.ParquetTimestampInMillisecond)
		d.Set("parquet_version", s3settings.ParquetVersion)
		d.Set("preserve_transactions", s3settings.PreserveTransactions)
		d.Set("server_side_encryption_kms_key_id", s3settings.ServerSideEncryptionKmsKeyId)
		d.Set("use_csv_no_sup_value", s3settings.UseCsvNoSupValue)
	}

	p, err := structure.NormalizeJsonString(aws.ToString(s3settings.ExternalTableDefinition))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("external_table_definition", p)

	return diags
}

func resourceS3EndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &dms.ModifyEndpointInput{
			EndpointArn: aws.String(d.Get("endpoint_arn").(string)),
		}

		if d.HasChange(names.AttrCertificateARN) {
			input.CertificateArn = aws.String(d.Get(names.AttrCertificateARN).(string))
		}

		if d.HasChange(names.AttrEndpointType) {
			input.EndpointType = awstypes.ReplicationEndpointTypeValue(d.Get(names.AttrEndpointType).(string))
		}

		input.EngineName = aws.String(engineNameS3)

		if d.HasChange("ssl_mode") {
			input.SslMode = awstypes.DmsSslModeValue(d.Get("ssl_mode").(string))
		}

		if d.HasChangesExcept(
			names.AttrCertificateARN,
			names.AttrEndpointType,
			"ssl_mode",
		) {
			input.S3Settings = s3Settings(d, d.Get(names.AttrEndpointType).(string) == string(awstypes.ReplicationEndpointTypeValueTarget))
			input.ServiceAccessRoleArn = aws.String(d.Get("service_access_role_arn").(string))

			input.ExtraConnectionAttributes = extraConnectionAnomalies(d)
		}

		_, err := tfresource.RetryWhenIsA[*awstypes.AccessDeniedFault](ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
			return conn.ModifyEndpoint(ctx, input)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DMS S3 Endpoint (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceS3EndpointRead(ctx, d, meta)...)
}

func resourceS3EndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	log.Printf("[DEBUG] Deleting DMS Endpoint: (%s)", d.Id())
	_, err := conn.DeleteEndpoint(ctx, &dms.DeleteEndpointInput{
		EndpointArn: aws.String(d.Get("endpoint_arn").(string)),
	})

	if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating DMS S3 Endpoint (%s): %s", d.Id(), err)
	}

	if _, err := waitEndpointDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DMS S3 Endpoint (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func s3Settings(d *schema.ResourceData, target bool) *awstypes.S3Settings {
	s3s := &awstypes.S3Settings{}

	if v, ok := d.Get("add_column_name").(bool); ok { // likely only useful for target
		s3s.AddColumnName = aws.Bool(v)
	}

	if v, ok := d.GetOk("add_trailing_padding_character"); ok && target { // target
		s3s.AddTrailingPaddingCharacter = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("bucket_folder"); ok {
		s3s.BucketFolder = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrBucketName); ok {
		s3s.BucketName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("canned_acl_for_objects"); ok { // likely only useful for target
		s3s.CannedAclForObjects = awstypes.CannedAclForObjectsValue(v.(string))
	}

	if v, ok := d.Get("cdc_inserts_and_updates").(bool); ok { // likely only useful for target
		s3s.CdcInsertsAndUpdates = aws.Bool(v)
	}

	if v, ok := d.Get("cdc_inserts_only").(bool); ok { // likely only useful for target
		s3s.CdcInsertsOnly = aws.Bool(v)
	}

	if v, ok := d.GetOk("cdc_max_batch_interval"); ok { // likely only useful for target
		s3s.CdcMaxBatchInterval = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("cdc_min_file_size"); ok { // likely only useful for target
		s3s.CdcMinFileSize = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("cdc_path"); ok {
		s3s.CdcPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("compression_type"); ok && target { // likely only useful for target
		s3s.CompressionType = awstypes.CompressionTypeValue(v.(string))
	}

	if v, ok := d.GetOk("csv_delimiter"); ok {
		s3s.CsvDelimiter = aws.String(v.(string))
	}

	if v, ok := d.GetOk("csv_no_sup_value"); ok && target { // target
		s3s.CsvNoSupValue = aws.String(v.(string))
	}

	if v, ok := d.GetOk("csv_null_value"); ok { // likely only useful for target
		s3s.CsvNullValue = aws.String(v.(string))
	}

	if v, ok := d.GetOk("csv_row_delimiter"); ok {
		s3s.CsvRowDelimiter = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_format"); ok && target { // target
		s3s.DataFormat = awstypes.DataFormatValue(v.(string))
	}

	if v, ok := d.GetOk("data_page_size"); ok { // likely only useful for target
		s3s.DataPageSize = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("date_partition_delimiter"); ok && target { // target
		s3s.DatePartitionDelimiter = awstypes.DatePartitionDelimiterValue(v.(string))
	}

	if v, ok := d.Get("date_partition_enabled").(bool); ok && target { // likely only useful for target
		s3s.DatePartitionEnabled = aws.Bool(v)
	}

	if v, ok := d.GetOk("date_partition_sequence"); ok && target { // target
		s3s.DatePartitionSequence = awstypes.DatePartitionSequenceValue(v.(string))
	}

	if v, ok := d.GetOk("date_partition_timezone"); ok && target { // target
		s3s.DatePartitionTimezone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dict_page_size_limit"); ok { // likely only useful for target
		s3s.DictPageSizeLimit = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.Get("enable_statistics").(bool); ok { // likely only useful for target
		s3s.EnableStatistics = aws.Bool(v)
	}

	if v, ok := d.GetOk("encoding_type"); ok { // likely only useful for target
		s3s.EncodingType = awstypes.EncodingTypeValue(v.(string))
	}

	if v, ok := d.GetOk("encryption_mode"); ok && target { // target
		s3s.EncryptionMode = awstypes.EncryptionModeValue(v.(string))
	}

	if v, ok := d.GetOk(names.AttrExpectedBucketOwner); ok { // likely only useful for target
		s3s.ExpectedBucketOwner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("external_table_definition"); ok {
		s3s.ExternalTableDefinition = aws.String(v.(string))
	}

	if v, ok := d.Get("glue_catalog_generation").(bool); ok { // target
		s3s.GlueCatalogGeneration = aws.Bool(v)
	}

	if v, ok := d.GetOk("ignore_header_rows"); ok {
		s3s.IgnoreHeaderRows = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.Get("include_op_for_full_load").(bool); ok { // likely only useful for target
		s3s.IncludeOpForFullLoad = aws.Bool(v)
	}

	if v, ok := d.GetOk("max_file_size"); ok { // likely only useful for target
		s3s.MaxFileSize = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.Get("parquet_timestamp_in_millisecond").(bool); ok && target { // target
		s3s.ParquetTimestampInMillisecond = aws.Bool(v)
	}

	if v, ok := d.GetOk("parquet_version"); ok && target { // target
		s3s.ParquetVersion = awstypes.ParquetVersionValue(v.(string))
	}

	if v, ok := d.Get("preserve_transactions").(bool); ok && target { // target
		s3s.PreserveTransactions = aws.Bool(v)
	}

	if v, ok := d.Get("rfc_4180").(bool); ok {
		s3s.Rfc4180 = aws.Bool(v)
	}

	if v, ok := d.GetOk("row_group_length"); ok { // likely only useful for target
		s3s.RowGroupLength = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("server_side_encryption_kms_key_id"); ok && target { // target
		s3s.ServerSideEncryptionKmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("service_access_role_arn"); ok {
		s3s.ServiceAccessRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("timestamp_column_name"); ok { // likely only useful for target
		s3s.TimestampColumnName = aws.String(v.(string))
	}

	if v, ok := d.Get("use_csv_no_sup_value").(bool); ok && target { // target
		s3s.UseCsvNoSupValue = aws.Bool(v)
	}

	if v, ok := d.Get("use_task_start_time_for_full_load_timestamp").(bool); ok { // likely only useful for target
		s3s.UseTaskStartTimeForFullLoadTimestamp = aws.Bool(v)
	}

	return s3s
}

func extraConnectionAnomalies(d *schema.ResourceData) *string {
	// not all attributes work in the data structures and must be passed via ex conn attr

	var anoms []string

	if v, ok := d.GetOk("cdc_path"); ok {
		anoms = append(anoms, fmt.Sprintf("%s=%s", "CdcPath", v.(string)))
	}

	if v, ok := d.GetOk("detach_target_on_lob_lookup_failure_parquet"); ok {
		anoms = append(anoms, fmt.Sprintf("%s=%t", "detachTargetOnLobLookupFailureParquet", v.(bool)))
	}

	if len(anoms) == 0 {
		return nil
	}

	return aws.String(strings.Join(anoms, ";"))
}

func setDetachTargetOnLobLookupFailureParquet(d *schema.ResourceData, eca string) {
	if strings.Contains(eca, "detachTargetOnLobLookupFailureParquet=false") {
		d.Set("detach_target_on_lob_lookup_failure_parquet", false)
	}

	if strings.Contains(eca, "detachTargetOnLobLookupFailureParquet=true") {
		d.Set("detach_target_on_lob_lookup_failure_parquet", true)
	}
}
