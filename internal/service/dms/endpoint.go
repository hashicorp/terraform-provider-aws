// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	dms "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dms_endpoint", name="Endpoint")
// @Tags(identifierAttribute="endpoint_arn")
func resourceEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEndpointCreate,
		ReadWithoutTimeout:   resourceEndpointRead,
		UpdateWithoutTimeout: resourceEndpointUpdate,
		DeleteWithoutTimeout: resourceEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrCertificateARN: {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"elasticsearch_settings": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint_uri": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"error_retry_duration": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							Default:      300,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"full_load_error_percentage": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							Default:      10,
							ValidateFunc: validation.IntBetween(0, 100),
						},
						"service_access_role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"use_new_mapping_type": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
							Default:  false,
						},
					},
				},
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
			"engine_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(engineName_Values(), false),
			},
			"extra_connection_attributes": {
				Type:             schema.TypeString,
				Computed:         true,
				Optional:         true,
				DiffSuppressFunc: suppressExtraConnectionAttributesDiffs,
			},
			"kafka_settings": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"broker": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"include_control_details": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"include_null_and_empty": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"include_partition_value": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"include_table_alter_operations": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"include_transaction_details": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"message_format": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.MessageFormatValueJson,
							ValidateDiagFunc: enum.Validate[awstypes.MessageFormatValue](),
						},
						"message_max_bytes": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1000000,
						},
						"no_hex_prefix": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"partition_include_schema_table": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"sasl_password": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"sasl_username": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"security_protocol": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.KafkaSecurityProtocol](),
						},
						"ssl_ca_certificate_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"ssl_client_certificate_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"ssl_client_key_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"ssl_client_key_password": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"topic": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  kafkaDefaultTopic,
						},
					},
				},
			},
			"kinesis_settings": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include_control_details": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"include_null_and_empty": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"include_partition_value": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"include_table_alter_operations": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"include_transaction_details": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"message_format": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							Default:          awstypes.MessageFormatValueJson,
							ValidateDiagFunc: enum.Validate[awstypes.MessageFormatValue](),
						},
						"partition_include_schema_table": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"service_access_role_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrStreamARN: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrKMSKeyARN: {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"mongodb_settings": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth_mechanism": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      mongoDBAuthMechanismValueDefault,
							ValidateFunc: validation.StringInSlice(mongoDBAuthMechanismValue_Values(), false),
						},
						"auth_source": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  mongoDBAuthSourceAdmin,
						},
						"auth_type": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.AuthTypeValuePassword,
							ValidateDiagFunc: enum.Validate[awstypes.AuthTypeValue](),
						},
						"docs_to_investigate": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "1000",
						},
						"extract_doc_id": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "false",
						},
						"nesting_level": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.NestingLevelValueNone,
							ValidateDiagFunc: enum.Validate[awstypes.NestingLevelValue](),
						},
					},
				},
			},
			names.AttrPassword: {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"secrets_manager_access_role_arn", "secrets_manager_arn"},
			},
			"pause_replication_tasks": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrPort: {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"secrets_manager_access_role_arn", "secrets_manager_arn"},
			},
			"postgres_settings": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"after_connect_script": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"babelfish_database_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"capture_ddls": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"database_mode": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ddl_artifacts_schema": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"execute_timeout": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"fail_tasks_on_lob_truncation": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"heartbeat_enable": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"heartbeat_frequency": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"heartbeat_schema": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"map_boolean_as_boolean": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"map_jsonb_as_clob": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"map_long_varchar_as": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"max_file_size": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"plugin_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"slot_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"redis_settings": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth_password": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"auth_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.RedisAuthTypeValue](),
						},
						"auth_user_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrPort: {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"server_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ssl_ca_certificate_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ssl_security_protocol": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.SslSecurityProtocolValueSslEncryption,
							ValidateDiagFunc: enum.Validate[awstypes.SslSecurityProtocolValue](),
						},
					},
				},
			},
			"redshift_settings": {
				Type:             schema.TypeList,
				Optional:         true,
				Computed:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_folder": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrBucketName: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"encryption_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      encryptionModeSseS3,
							ValidateFunc: validation.StringInSlice(encryptionMode_Values(), false),
						},
						"server_side_encryption_kms_key_id": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: tfkms.DiffSuppressKey,
							ValidateFunc:     tfkms.ValidateKey,
						},
						"service_access_role_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"s3_settings": {
				Description:      "This argument is deprecated and will be removed in a future version; use aws_dms_s3_endpoint instead",
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"add_column_name": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"bucket_folder": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						names.AttrBucketName: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"canned_acl_for_objects": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.CannedAclForObjectsValueNone,
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
							Default:      60,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"cdc_min_file_size": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      32000,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"cdc_path": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"compression_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      s3SettingsCompressionTypeNone,
							ValidateFunc: validation.StringInSlice(s3SettingsCompressionType_Values(), false),
						},
						"csv_delimiter": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  ",",
						},
						"csv_no_sup_value": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
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
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.DataFormatValueCsv,
							ValidateDiagFunc: enum.Validate[awstypes.DataFormatValue](),
						},
						"data_page_size": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1048576,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"date_partition_delimiter": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.DatePartitionDelimiterValueSlash,
							ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.DatePartitionDelimiterValue](),
							StateFunc: func(v interface{}) string {
								return strings.ToLower(v.(string))
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
							Default:          awstypes.DatePartitionSequenceValueYyyymmdd,
							ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.DatePartitionSequenceValue](),
							StateFunc: func(v interface{}) string {
								return strings.ToLower(v.(string))
							},
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
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.EncodingTypeValueRleDictionary,
							ValidateDiagFunc: enum.Validate[awstypes.EncodingTypeValue](),
						},
						"encryption_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      encryptionModeSseS3,
							ValidateFunc: validation.StringInSlice(encryptionMode_Values(), false),
						},
						"external_table_definition": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"glue_catalog_generation": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
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
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.ParquetVersionValueParquet10,
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
							Default:      10000,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"server_side_encryption_kms_key_id": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: tfkms.DiffSuppressKey,
							ValidateFunc:     tfkms.ValidateKey,
						},
						"service_access_role_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "",
							ValidateFunc: verify.ValidARN,
						},
						"timestamp_column_name": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
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
				},
			},
			"secrets_manager_access_role_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  verify.ValidARN,
				RequiredWith:  []string{"secrets_manager_arn"},
				ConflictsWith: []string{names.AttrUsername, names.AttrPassword, "server_name", names.AttrPort},
			},
			"secrets_manager_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  verify.ValidARN,
				RequiredWith:  []string{"secrets_manager_access_role_arn"},
				ConflictsWith: []string{names.AttrUsername, names.AttrPassword, "server_name", names.AttrPort},
			},
			"server_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"secrets_manager_access_role_arn", "secrets_manager_arn"},
			},
			"service_access_role": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ssl_mode": {
				Type:             schema.TypeString,
				Computed:         true,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.DmsSslModeValue](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrUsername: {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"secrets_manager_access_role_arn", "secrets_manager_arn"},
			},
		},

		CustomizeDiff: customdiff.All(
			requireEngineSettingsCustomizeDiff,
			validateKMSKeyEngineCustomizeDiff,
			validateS3SSEKMSKeyCustomizeDiff,
			validateRedshiftSSEKMSKeyCustomizeDiff,
			verify.SetTagsDiff,
		),
	}
}

func resourceEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	endpointID := d.Get("endpoint_id").(string)
	input := &dms.CreateEndpointInput{
		EndpointIdentifier: aws.String(endpointID),
		EndpointType:       awstypes.ReplicationEndpointTypeValue(d.Get(names.AttrEndpointType).(string)),
		EngineName:         aws.String(d.Get("engine_name").(string)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrCertificateARN); ok {
		input.CertificateArn = aws.String(v.(string))
	}

	// Send ExtraConnectionAttributes in the API request for all resource types
	// per https://github.com/hashicorp/terraform-provider-aws/issues/8009
	if v, ok := d.GetOk("extra_connection_attributes"); ok {
		input.ExtraConnectionAttributes = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ssl_mode"); ok {
		input.SslMode = awstypes.DmsSslModeValue(v.(string))
	}

	switch d.Get("engine_name").(string) {
	case engineNameAurora, engineNameMariadb, engineNameMySQL:
		if _, ok := d.GetOk("secrets_manager_arn"); ok {
			input.MySQLSettings = &awstypes.MySQLSettings{
				SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
				SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
			}
		} else {
			input.MySQLSettings = &awstypes.MySQLSettings{
				Username:     aws.String(d.Get(names.AttrUsername).(string)),
				Password:     aws.String(d.Get(names.AttrPassword).(string)),
				ServerName:   aws.String(d.Get("server_name").(string)),
				Port:         aws.Int32(int32(d.Get(names.AttrPort).(int))),
				DatabaseName: aws.String(d.Get(names.AttrDatabaseName).(string)),
			}

			// Set connection info in top-level namespace as well
			expandTopLevelConnectionInfo(d, input)
		}
	case engineNameAuroraPostgresql, engineNamePostgres:
		settings := &awstypes.PostgreSQLSettings{}
		if _, ok := d.GetOk("postgres_settings"); ok {
			settings = expandPostgreSQLSettings(d.Get("postgres_settings").([]interface{})[0].(map[string]interface{}))
		}

		if _, ok := d.GetOk("secrets_manager_arn"); ok {
			settings.SecretsManagerAccessRoleArn = aws.String(d.Get("secrets_manager_access_role_arn").(string))
			settings.SecretsManagerSecretId = aws.String(d.Get("secrets_manager_arn").(string))
			settings.DatabaseName = aws.String(d.Get(names.AttrDatabaseName).(string))
		} else {
			settings.Username = aws.String(d.Get(names.AttrUsername).(string))
			settings.Password = aws.String(d.Get(names.AttrPassword).(string))
			settings.ServerName = aws.String(d.Get("server_name").(string))
			settings.Port = aws.Int32(int32(d.Get(names.AttrPort).(int)))
			settings.DatabaseName = aws.String(d.Get(names.AttrDatabaseName).(string))

			// Set connection info in top-level namespace as well
			expandTopLevelConnectionInfo(d, input)
		}

		input.PostgreSQLSettings = settings
	case engineNameDynamoDB:
		input.DynamoDbSettings = &awstypes.DynamoDbSettings{
			ServiceAccessRoleArn: aws.String(d.Get("service_access_role").(string)),
		}
	case engineNameElasticsearch, engineNameOpenSearch:
		input.ElasticsearchSettings = &awstypes.ElasticsearchSettings{
			ServiceAccessRoleArn:    aws.String(d.Get("elasticsearch_settings.0.service_access_role_arn").(string)),
			EndpointUri:             aws.String(d.Get("elasticsearch_settings.0.endpoint_uri").(string)),
			ErrorRetryDuration:      aws.Int32(int32(d.Get("elasticsearch_settings.0.error_retry_duration").(int))),
			FullLoadErrorPercentage: aws.Int32(int32(d.Get("elasticsearch_settings.0.full_load_error_percentage").(int))),
			UseNewMappingType:       aws.Bool(d.Get("elasticsearch_settings.0.use_new_mapping_type").(bool)),
		}
	case engineNameKafka:
		input.KafkaSettings = expandKafkaSettings(d.Get("kafka_settings").([]interface{})[0].(map[string]interface{}))
	case engineNameKinesis:
		input.KinesisSettings = expandKinesisSettings(d.Get("kinesis_settings").([]interface{})[0].(map[string]interface{}))
	case engineNameMongodb:
		var settings = &awstypes.MongoDbSettings{}

		if _, ok := d.GetOk("secrets_manager_arn"); ok {
			settings.SecretsManagerAccessRoleArn = aws.String(d.Get("secrets_manager_access_role_arn").(string))
			settings.SecretsManagerSecretId = aws.String(d.Get("secrets_manager_arn").(string))
		} else {
			settings.Username = aws.String(d.Get(names.AttrUsername).(string))
			settings.Password = aws.String(d.Get(names.AttrPassword).(string))
			settings.ServerName = aws.String(d.Get("server_name").(string))
			settings.Port = aws.Int32(int32(d.Get(names.AttrPort).(int)))

			// Set connection info in top-level namespace as well
			expandTopLevelConnectionInfo(d, input)
		}

		settings.DatabaseName = aws.String(d.Get(names.AttrDatabaseName).(string))
		settings.KmsKeyId = aws.String(d.Get(names.AttrKMSKeyARN).(string))
		settings.AuthType = awstypes.AuthTypeValue(d.Get("mongodb_settings.0.auth_type").(string))
		settings.AuthMechanism = awstypes.AuthMechanismValue(d.Get("mongodb_settings.0.auth_mechanism").(string))
		settings.NestingLevel = awstypes.NestingLevelValue(d.Get("mongodb_settings.0.nesting_level").(string))
		settings.ExtractDocId = aws.String(d.Get("mongodb_settings.0.extract_doc_id").(string))
		settings.DocsToInvestigate = aws.String(d.Get("mongodb_settings.0.docs_to_investigate").(string))
		settings.AuthSource = aws.String(d.Get("mongodb_settings.0.auth_source").(string))

		input.MongoDbSettings = settings
	case engineNameOracle:
		if _, ok := d.GetOk("secrets_manager_arn"); ok {
			input.OracleSettings = &awstypes.OracleSettings{
				SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
				SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
				DatabaseName:                aws.String(d.Get(names.AttrDatabaseName).(string)),
			}
		} else {
			input.OracleSettings = &awstypes.OracleSettings{
				Username:     aws.String(d.Get(names.AttrUsername).(string)),
				Password:     aws.String(d.Get(names.AttrPassword).(string)),
				ServerName:   aws.String(d.Get("server_name").(string)),
				Port:         aws.Int32(int32(d.Get(names.AttrPort).(int))),
				DatabaseName: aws.String(d.Get(names.AttrDatabaseName).(string)),
			}

			// Set connection info in top-level namespace as well
			expandTopLevelConnectionInfo(d, input)
		}
	case engineNameRedis:
		input.RedisSettings = expandRedisSettings(d.Get("redis_settings").([]interface{})[0].(map[string]interface{}))
	case engineNameRedshift:
		var settings = &awstypes.RedshiftSettings{
			DatabaseName: aws.String(d.Get(names.AttrDatabaseName).(string)),
		}

		if _, ok := d.GetOk("secrets_manager_arn"); ok {
			settings.SecretsManagerAccessRoleArn = aws.String(d.Get("secrets_manager_access_role_arn").(string))
			settings.SecretsManagerSecretId = aws.String(d.Get("secrets_manager_arn").(string))
		} else {
			settings.Username = aws.String(d.Get(names.AttrUsername).(string))
			settings.Password = aws.String(d.Get(names.AttrPassword).(string))
			settings.ServerName = aws.String(d.Get("server_name").(string))
			settings.Port = aws.Int32(int32(d.Get(names.AttrPort).(int)))

			// Set connection info in top-level namespace as well
			expandTopLevelConnectionInfo(d, input)
		}

		if v, ok := d.GetOk("redshift_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})

			if v, ok := tfMap["bucket_folder"].(string); ok && v != "" {
				settings.BucketFolder = aws.String(v)
			}

			if v, ok := tfMap[names.AttrBucketName].(string); ok && v != "" {
				settings.BucketName = aws.String(v)
			}

			if v, ok := tfMap["encryption_mode"].(string); ok && v != "" {
				settings.EncryptionMode = awstypes.EncryptionModeValue(v)
			}

			if v, ok := tfMap["server_side_encryption_kms_key_id"].(string); ok && v != "" {
				settings.ServerSideEncryptionKmsKeyId = aws.String(v)
			}

			if v, ok := tfMap["service_access_role_arn"].(string); ok && v != "" {
				settings.ServiceAccessRoleArn = aws.String(v)
			}
		}

		input.RedshiftSettings = settings
	case engineNameSQLServer, engineNameBabelfish:
		if _, ok := d.GetOk("secrets_manager_arn"); ok {
			input.MicrosoftSQLServerSettings = &awstypes.MicrosoftSQLServerSettings{
				SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
				SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
				DatabaseName:                aws.String(d.Get(names.AttrDatabaseName).(string)),
			}
		} else {
			input.MicrosoftSQLServerSettings = &awstypes.MicrosoftSQLServerSettings{
				Username:     aws.String(d.Get(names.AttrUsername).(string)),
				Password:     aws.String(d.Get(names.AttrPassword).(string)),
				ServerName:   aws.String(d.Get("server_name").(string)),
				Port:         aws.Int32(int32(d.Get(names.AttrPort).(int))),
				DatabaseName: aws.String(d.Get(names.AttrDatabaseName).(string)),
			}

			// Set connection info in top-level namespace as well
			expandTopLevelConnectionInfo(d, input)
		}
	case engineNameSybase:
		if _, ok := d.GetOk("secrets_manager_arn"); ok {
			input.SybaseSettings = &awstypes.SybaseSettings{
				SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
				SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
				DatabaseName:                aws.String(d.Get(names.AttrDatabaseName).(string)),
			}
		} else {
			input.SybaseSettings = &awstypes.SybaseSettings{
				Username:     aws.String(d.Get(names.AttrUsername).(string)),
				Password:     aws.String(d.Get(names.AttrPassword).(string)),
				ServerName:   aws.String(d.Get("server_name").(string)),
				Port:         aws.Int32(int32(d.Get(names.AttrPort).(int))),
				DatabaseName: aws.String(d.Get(names.AttrDatabaseName).(string)),
			}

			// Set connection info in top-level namespace as well
			expandTopLevelConnectionInfo(d, input)
		}
	case engineNameDB2, engineNameDB2zOS:
		if _, ok := d.GetOk("secrets_manager_arn"); ok {
			input.IBMDb2Settings = &awstypes.IBMDb2Settings{
				SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
				SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
				DatabaseName:                aws.String(d.Get(names.AttrDatabaseName).(string)),
			}
		} else {
			input.IBMDb2Settings = &awstypes.IBMDb2Settings{
				Username:     aws.String(d.Get(names.AttrUsername).(string)),
				Password:     aws.String(d.Get(names.AttrPassword).(string)),
				ServerName:   aws.String(d.Get("server_name").(string)),
				Port:         aws.Int32(int32(d.Get(names.AttrPort).(int))),
				DatabaseName: aws.String(d.Get(names.AttrDatabaseName).(string)),
			}

			// Set connection info in top-level namespace as well
			expandTopLevelConnectionInfo(d, input)
		}
	case engineNameS3:
		input.S3Settings = expandS3Settings(d.Get("s3_settings").([]interface{})[0].(map[string]interface{}))
	default:
		expandTopLevelConnectionInfo(d, input)
	}

	_, err := tfresource.RetryWhenIsA[*awstypes.AccessDeniedFault](ctx, d.Timeout(schema.TimeoutCreate),
		func() (interface{}, error) {
			return conn.CreateEndpoint(ctx, input)
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DMS Endpoint (%s): %s", endpointID, err)
	}

	d.SetId(endpointID)

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	endpoint, err := findEndpointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DMS Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Endpoint (%s): %s", d.Id(), err)
	}

	if err := resourceEndpointSetState(d, endpoint); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func resourceEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		endpointARN := d.Get("endpoint_arn").(string)
		pauseTasks := d.Get("pause_replication_tasks").(bool)
		var tasks []awstypes.ReplicationTask

		if pauseTasks {
			var err error
			tasks, err = stopEndpointReplicationTasks(ctx, conn, endpointARN)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "stopping replication tasks before updating DMS Endpoint (%s): %s", d.Id(), err)
			}
		}

		if d.HasChangesExcept("pause_replication_tasks") {
			input := &dms.ModifyEndpointInput{
				EndpointArn: aws.String(endpointARN),
			}

			if d.HasChange(names.AttrCertificateARN) {
				input.CertificateArn = aws.String(d.Get(names.AttrCertificateARN).(string))
			}

			if d.HasChange(names.AttrEndpointType) {
				input.EndpointType = awstypes.ReplicationEndpointTypeValue(d.Get(names.AttrEndpointType).(string))
			}

			if d.HasChange("engine_name") {
				input.EngineName = aws.String(d.Get("engine_name").(string))
			}

			if d.HasChange("extra_connection_attributes") {
				input.ExtraConnectionAttributes = aws.String(d.Get("extra_connection_attributes").(string))
			}

			if d.HasChange("service_access_role") {
				input.DynamoDbSettings = &awstypes.DynamoDbSettings{
					ServiceAccessRoleArn: aws.String(d.Get("service_access_role").(string)),
				}
			}

			if d.HasChange("ssl_mode") {
				input.SslMode = awstypes.DmsSslModeValue(d.Get("ssl_mode").(string))
			}

			switch engineName := d.Get("engine_name").(string); engineName {
			case engineNameAurora, engineNameMariadb, engineNameMySQL:
				if d.HasChanges(
					names.AttrUsername, names.AttrPassword, "server_name", names.AttrPort, names.AttrDatabaseName, "secrets_manager_access_role_arn",
					"secrets_manager_arn") {
					if _, ok := d.GetOk("secrets_manager_arn"); ok {
						input.MySQLSettings = &awstypes.MySQLSettings{
							SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
							SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
						}
					} else {
						input.MySQLSettings = &awstypes.MySQLSettings{
							Username:     aws.String(d.Get(names.AttrUsername).(string)),
							Password:     aws.String(d.Get(names.AttrPassword).(string)),
							ServerName:   aws.String(d.Get("server_name").(string)),
							Port:         aws.Int32(int32(d.Get(names.AttrPort).(int))),
							DatabaseName: aws.String(d.Get(names.AttrDatabaseName).(string)),
						}
						input.EngineName = aws.String(engineName)

						// Update connection info in top-level namespace as well
						expandTopLevelConnectionInfoModify(d, input)
					}
				}
			case engineNameAuroraPostgresql, engineNamePostgres:
				if d.HasChanges(
					names.AttrUsername, names.AttrPassword, "server_name", names.AttrPort, names.AttrDatabaseName, "secrets_manager_access_role_arn",
					"secrets_manager_arn") {
					if _, ok := d.GetOk("secrets_manager_arn"); ok {
						input.PostgreSQLSettings = &awstypes.PostgreSQLSettings{
							DatabaseName:                aws.String(d.Get(names.AttrDatabaseName).(string)),
							SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
							SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
						}
					} else {
						input.PostgreSQLSettings = &awstypes.PostgreSQLSettings{
							Username:     aws.String(d.Get(names.AttrUsername).(string)),
							Password:     aws.String(d.Get(names.AttrPassword).(string)),
							ServerName:   aws.String(d.Get("server_name").(string)),
							Port:         aws.Int32(int32(d.Get(names.AttrPort).(int))),
							DatabaseName: aws.String(d.Get(names.AttrDatabaseName).(string)),
						}
						input.EngineName = aws.String(engineName) // Must be included (should be 'postgres')

						// Update connection info in top-level namespace as well
						expandTopLevelConnectionInfoModify(d, input)
					}
				}
			case engineNameDynamoDB:
				if d.HasChange("service_access_role") {
					input.DynamoDbSettings = &awstypes.DynamoDbSettings{
						ServiceAccessRoleArn: aws.String(d.Get("service_access_role").(string)),
					}
				}
			case engineNameElasticsearch, engineNameOpenSearch:
				if d.HasChanges(
					"elasticsearch_settings.0.endpoint_uri",
					"elasticsearch_settings.0.error_retry_duration",
					"elasticsearch_settings.0.full_load_error_percentage",
					"elasticsearch_settings.0.service_access_role_arn",
					"elasticsearch_settings.0.use_new_mapping_type") {
					input.ElasticsearchSettings = &awstypes.ElasticsearchSettings{
						ServiceAccessRoleArn:    aws.String(d.Get("elasticsearch_settings.0.service_access_role_arn").(string)),
						EndpointUri:             aws.String(d.Get("elasticsearch_settings.0.endpoint_uri").(string)),
						ErrorRetryDuration:      aws.Int32(int32(d.Get("elasticsearch_settings.0.error_retry_duration").(int))),
						FullLoadErrorPercentage: aws.Int32(int32(d.Get("elasticsearch_settings.0.full_load_error_percentage").(int))),
						UseNewMappingType:       aws.Bool(d.Get("elasticsearch_settings.0.use_new_mapping_type").(bool)),
					}
					input.EngineName = aws.String(engineName)
				}
			case engineNameKafka:
				if d.HasChange("kafka_settings") {
					input.KafkaSettings = expandKafkaSettings(d.Get("kafka_settings").([]interface{})[0].(map[string]interface{}))
					input.EngineName = aws.String(engineName)
				}
			case engineNameKinesis:
				if d.HasChanges("kinesis_settings") {
					input.KinesisSettings = expandKinesisSettings(d.Get("kinesis_settings").([]interface{})[0].(map[string]interface{}))
					input.EngineName = aws.String(engineName)
				}
			case engineNameMongodb:
				if d.HasChanges(
					names.AttrUsername, names.AttrPassword, "server_name", names.AttrPort, names.AttrDatabaseName, "mongodb_settings.0.auth_type",
					"mongodb_settings.0.auth_mechanism", "mongodb_settings.0.nesting_level", "mongodb_settings.0.extract_doc_id",
					"mongodb_settings.0.docs_to_investigate", "mongodb_settings.0.auth_source", "secrets_manager_access_role_arn",
					"secrets_manager_arn") {
					if _, ok := d.GetOk("secrets_manager_arn"); ok {
						input.MongoDbSettings = &awstypes.MongoDbSettings{
							SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
							SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
							DatabaseName:                aws.String(d.Get(names.AttrDatabaseName).(string)),
							KmsKeyId:                    aws.String(d.Get(names.AttrKMSKeyARN).(string)),

							AuthType:          awstypes.AuthTypeValue(d.Get("mongodb_settings.0.auth_type").(string)),
							AuthMechanism:     awstypes.AuthMechanismValue(d.Get("mongodb_settings.0.auth_mechanism").(string)),
							NestingLevel:      awstypes.NestingLevelValue(d.Get("mongodb_settings.0.nesting_level").(string)),
							ExtractDocId:      aws.String(d.Get("mongodb_settings.0.extract_doc_id").(string)),
							DocsToInvestigate: aws.String(d.Get("mongodb_settings.0.docs_to_investigate").(string)),
							AuthSource:        aws.String(d.Get("mongodb_settings.0.auth_source").(string)),
						}
					} else {
						input.MongoDbSettings = &awstypes.MongoDbSettings{
							Username:     aws.String(d.Get(names.AttrUsername).(string)),
							Password:     aws.String(d.Get(names.AttrPassword).(string)),
							ServerName:   aws.String(d.Get("server_name").(string)),
							Port:         aws.Int32(int32(d.Get(names.AttrPort).(int))),
							DatabaseName: aws.String(d.Get(names.AttrDatabaseName).(string)),
							KmsKeyId:     aws.String(d.Get(names.AttrKMSKeyARN).(string)),

							AuthType:          awstypes.AuthTypeValue(d.Get("mongodb_settings.0.auth_type").(string)),
							AuthMechanism:     awstypes.AuthMechanismValue(d.Get("mongodb_settings.0.auth_mechanism").(string)),
							NestingLevel:      awstypes.NestingLevelValue(d.Get("mongodb_settings.0.nesting_level").(string)),
							ExtractDocId:      aws.String(d.Get("mongodb_settings.0.extract_doc_id").(string)),
							DocsToInvestigate: aws.String(d.Get("mongodb_settings.0.docs_to_investigate").(string)),
							AuthSource:        aws.String(d.Get("mongodb_settings.0.auth_source").(string)),
						}
						input.EngineName = aws.String(engineName)

						// Update connection info in top-level namespace as well
						expandTopLevelConnectionInfoModify(d, input)
					}
				}
			case engineNameOracle:
				if d.HasChanges(
					names.AttrUsername, names.AttrPassword, "server_name", names.AttrPort, names.AttrDatabaseName, "secrets_manager_access_role_arn",
					"secrets_manager_arn") {
					if _, ok := d.GetOk("secrets_manager_arn"); ok {
						input.OracleSettings = &awstypes.OracleSettings{
							DatabaseName:                aws.String(d.Get(names.AttrDatabaseName).(string)),
							SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
							SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
						}
					} else {
						input.OracleSettings = &awstypes.OracleSettings{
							Username:     aws.String(d.Get(names.AttrUsername).(string)),
							Password:     aws.String(d.Get(names.AttrPassword).(string)),
							ServerName:   aws.String(d.Get("server_name").(string)),
							Port:         aws.Int32(int32(d.Get(names.AttrPort).(int))),
							DatabaseName: aws.String(d.Get(names.AttrDatabaseName).(string)),
						}
						input.EngineName = aws.String(engineName) // Must be included (should be 'oracle')

						// Update connection info in top-level namespace as well
						expandTopLevelConnectionInfoModify(d, input)
					}
				}
			case engineNameRedis:
				if d.HasChanges("redis_settings") {
					input.RedisSettings = expandRedisSettings(d.Get("redis_settings").([]interface{})[0].(map[string]interface{}))
					input.EngineName = aws.String(engineName)
				}
			case engineNameRedshift:
				if d.HasChanges(
					names.AttrUsername, names.AttrPassword, "server_name", names.AttrPort, names.AttrDatabaseName,
					"redshift_settings", "secrets_manager_access_role_arn",
					"secrets_manager_arn") {
					if _, ok := d.GetOk("secrets_manager_arn"); ok {
						input.RedshiftSettings = &awstypes.RedshiftSettings{
							DatabaseName:                aws.String(d.Get(names.AttrDatabaseName).(string)),
							SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
							SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
						}
					} else {
						input.RedshiftSettings = &awstypes.RedshiftSettings{
							Username:     aws.String(d.Get(names.AttrUsername).(string)),
							Password:     aws.String(d.Get(names.AttrPassword).(string)),
							ServerName:   aws.String(d.Get("server_name").(string)),
							Port:         aws.Int32(int32(d.Get(names.AttrPort).(int))),
							DatabaseName: aws.String(d.Get(names.AttrDatabaseName).(string)),
						}
						input.EngineName = aws.String(engineName) // Must be included (should be 'redshift')

						// Update connection info in top-level namespace as well
						expandTopLevelConnectionInfoModify(d, input)

						if v, ok := d.GetOk("redshift_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
							tfMap := v.([]interface{})[0].(map[string]interface{})

							if v, ok := tfMap["bucket_folder"].(string); ok && v != "" {
								input.RedshiftSettings.BucketFolder = aws.String(v)
							}

							if v, ok := tfMap[names.AttrBucketName].(string); ok && v != "" {
								input.RedshiftSettings.BucketName = aws.String(v)
							}

							if v, ok := tfMap["encryption_mode"].(string); ok && v != "" {
								input.RedshiftSettings.EncryptionMode = awstypes.EncryptionModeValue(v)
							}

							if v, ok := tfMap["server_side_encryption_kms_key_id"].(string); ok && v != "" {
								input.RedshiftSettings.ServerSideEncryptionKmsKeyId = aws.String(v)
							}

							if v, ok := tfMap["service_access_role_arn"].(string); ok && v != "" {
								input.RedshiftSettings.ServiceAccessRoleArn = aws.String(v)
							}
						}
					}
				}
			case engineNameSQLServer, engineNameBabelfish:
				if d.HasChanges(
					names.AttrUsername, names.AttrPassword, "server_name", names.AttrPort, names.AttrDatabaseName, "secrets_manager_access_role_arn",
					"secrets_manager_arn") {
					if _, ok := d.GetOk("secrets_manager_arn"); ok {
						input.MicrosoftSQLServerSettings = &awstypes.MicrosoftSQLServerSettings{
							DatabaseName:                aws.String(d.Get(names.AttrDatabaseName).(string)),
							SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
							SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
						}
					} else {
						input.MicrosoftSQLServerSettings = &awstypes.MicrosoftSQLServerSettings{
							Username:     aws.String(d.Get(names.AttrUsername).(string)),
							Password:     aws.String(d.Get(names.AttrPassword).(string)),
							ServerName:   aws.String(d.Get("server_name").(string)),
							Port:         aws.Int32(int32(d.Get(names.AttrPort).(int))),
							DatabaseName: aws.String(d.Get(names.AttrDatabaseName).(string)),
						}
						input.EngineName = aws.String(engineName) // Must be included (should be 'postgres')

						// Update connection info in top-level namespace as well
						expandTopLevelConnectionInfoModify(d, input)
					}
				}
			case engineNameSybase:
				if d.HasChanges(
					names.AttrUsername, names.AttrPassword, "server_name", names.AttrPort, names.AttrDatabaseName, "secrets_manager_access_role_arn",
					"secrets_manager_arn") {
					if _, ok := d.GetOk("secrets_manager_arn"); ok {
						input.SybaseSettings = &awstypes.SybaseSettings{
							DatabaseName:                aws.String(d.Get(names.AttrDatabaseName).(string)),
							SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
							SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
						}
					} else {
						input.SybaseSettings = &awstypes.SybaseSettings{
							Username:     aws.String(d.Get(names.AttrUsername).(string)),
							Password:     aws.String(d.Get(names.AttrPassword).(string)),
							ServerName:   aws.String(d.Get("server_name").(string)),
							Port:         aws.Int32(int32(d.Get(names.AttrPort).(int))),
							DatabaseName: aws.String(d.Get(names.AttrDatabaseName).(string)),
						}
						input.EngineName = aws.String(engineName) // Must be included (should be 'postgres')

						// Update connection info in top-level namespace as well
						expandTopLevelConnectionInfoModify(d, input)
					}
				}
			case engineNameDB2, engineNameDB2zOS:
				if d.HasChanges(
					names.AttrUsername, names.AttrPassword, "server_name", names.AttrPort, names.AttrDatabaseName, "secrets_manager_access_role_arn",
					"secrets_manager_arn") {
					if _, ok := d.GetOk("secrets_manager_arn"); ok {
						input.IBMDb2Settings = &awstypes.IBMDb2Settings{
							DatabaseName:                aws.String(d.Get(names.AttrDatabaseName).(string)),
							SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
							SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
						}
					} else {
						input.IBMDb2Settings = &awstypes.IBMDb2Settings{
							Username:     aws.String(d.Get(names.AttrUsername).(string)),
							Password:     aws.String(d.Get(names.AttrPassword).(string)),
							ServerName:   aws.String(d.Get("server_name").(string)),
							Port:         aws.Int32(int32(d.Get(names.AttrPort).(int))),
							DatabaseName: aws.String(d.Get(names.AttrDatabaseName).(string)),
						}
						input.EngineName = aws.String(engineName) // Must be included (should be 'db2')

						// Update connection info in top-level namespace as well
						expandTopLevelConnectionInfoModify(d, input)
					}
				}
			case engineNameS3:
				if d.HasChanges("s3_settings") {
					input.S3Settings = expandS3Settings(d.Get("s3_settings").([]interface{})[0].(map[string]interface{}))
					input.EngineName = aws.String(engineName)
				}
			default:
				if d.HasChange(names.AttrDatabaseName) {
					input.DatabaseName = aws.String(d.Get(names.AttrDatabaseName).(string))
				}

				if d.HasChange(names.AttrPassword) {
					input.Password = aws.String(d.Get(names.AttrPassword).(string))
				}

				if d.HasChange(names.AttrPort) {
					input.Port = aws.Int32(int32(d.Get(names.AttrPort).(int)))
				}

				if d.HasChange("server_name") {
					input.ServerName = aws.String(d.Get("server_name").(string))
				}

				if d.HasChange(names.AttrUsername) {
					input.Username = aws.String(d.Get(names.AttrUsername).(string))
				}
			}

			_, err := conn.ModifyEndpoint(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating DMS Endpoint (%s): %s", d.Id(), err)
			}
		}

		if pauseTasks && len(tasks) > 0 {
			if err := startEndpointReplicationTasks(ctx, conn, endpointARN, tasks); err != nil {
				return sdkdiag.AppendErrorf(diags, "starting replication tasks after updating DMS Endpoint (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return sdkdiag.AppendErrorf(diags, "deleting DMS Endpoint (%s): %s", d.Id(), err)
	}

	if _, err := waitEndpointDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DMS Endpoint (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func requireEngineSettingsCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	switch engineName := diff.Get("engine_name").(string); engineName {
	case engineNameElasticsearch, engineNameOpenSearch:
		if v, ok := diff.GetOk("elasticsearch_settings"); !ok || len(v.([]interface{})) == 0 || v.([]interface{})[0] == nil {
			return fmt.Errorf("elasticsearch_settings must be set when engine_name = %q", engineName)
		}
	case engineNameKafka:
		if v, ok := diff.GetOk("kafka_settings"); !ok || len(v.([]interface{})) == 0 || v.([]interface{})[0] == nil {
			return fmt.Errorf("kafka_settings must be set when engine_name = %q", engineName)
		}
	case engineNameKinesis:
		if v, ok := diff.GetOk("kinesis_settings"); !ok || len(v.([]interface{})) == 0 || v.([]interface{})[0] == nil {
			return fmt.Errorf("kinesis_settings must be set when engine_name = %q", engineName)
		}
	case engineNameMongodb:
		if v, ok := diff.GetOk("mongodb_settings"); !ok || len(v.([]interface{})) == 0 || v.([]interface{})[0] == nil {
			return fmt.Errorf("mongodb_settings must be set when engine_name = %q", engineName)
		}
	case engineNameRedis:
		if v, ok := diff.GetOk("redis_settings"); !ok || len(v.([]interface{})) == 0 || v.([]interface{})[0] == nil {
			return fmt.Errorf("redis_settings must be set when engine_name = %q", engineName)
		}
	case engineNameS3:
		if v, ok := diff.GetOk("s3_settings"); !ok || len(v.([]interface{})) == 0 || v.([]interface{})[0] == nil {
			return fmt.Errorf("s3_settings must be set when engine_name = %q", engineName)
		}
	}

	return nil
}

func validateKMSKeyEngineCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	if d.Get("engine_name").(string) == engineNameS3 {
		if d.Get(names.AttrKMSKeyARN) != "" {
			return fmt.Errorf("kms_key_arn must not be set when engine is %q. Use s3_settings.server_side_encryption_kms_key_id instead", engineNameS3)
		}
	}
	return nil
}

func validateS3SSEKMSKeyCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	if d.Get("engine_name").(string) == engineNameS3 {
		return validateSSEKMSKey("s3_settings", d)
	}
	return nil
}

func validateRedshiftSSEKMSKeyCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	if d.Get("engine_name").(string) == engineNameRedshift {
		return validateSSEKMSKey("redshift_settings", d)
	}
	return nil
}

func validateSSEKMSKey(settingsAttrName string, d *schema.ResourceDiff) error {
	rawConfig := d.GetRawConfig()
	settings := rawConfig.GetAttr(settingsAttrName)
	if settings.IsKnown() && !settings.IsNull() && settings.LengthInt() > 0 {
		setting := settings.Index(cty.NumberIntVal(0))
		if setting.IsKnown() && !setting.IsNull() {
			kmsKeyId := setting.GetAttr("server_side_encryption_kms_key_id")
			if !kmsKeyId.IsKnown() {
				return nil
			}
			encryptionMode := setting.GetAttr("encryption_mode")
			if encryptionMode.IsKnown() && !encryptionMode.IsNull() {
				id := ""
				if !kmsKeyId.IsNull() {
					id = kmsKeyId.AsString()
				}
				switch encryptionMode.AsString() {
				case encryptionModeSseS3:
					if id != "" {
						return fmt.Errorf("%s.server_side_encryption_kms_key_id must not be set when encryption_mode is %q", settingsAttrName, encryptionModeSseS3)
					}
				case encryptionModeSseKMS:
					if id == "" {
						return fmt.Errorf("%s.server_side_encryption_kms_key_id is required when encryption_mode is %q", settingsAttrName, encryptionModeSseKMS)
					}
				}
			}
		}
	}
	return nil
}

func resourceEndpointSetState(d *schema.ResourceData, endpoint *awstypes.Endpoint) error {
	d.SetId(aws.ToString(endpoint.EndpointIdentifier))

	d.Set(names.AttrCertificateARN, endpoint.CertificateArn)
	d.Set("endpoint_arn", endpoint.EndpointArn)
	d.Set("endpoint_id", endpoint.EndpointIdentifier)
	// For some reason the AWS API only accepts lowercase type but returns it as uppercase
	d.Set(names.AttrEndpointType, strings.ToLower(string(endpoint.EndpointType)))
	d.Set("engine_name", endpoint.EngineName)
	d.Set("extra_connection_attributes", endpoint.ExtraConnectionAttributes)

	switch aws.ToString(endpoint.EngineName) {
	case engineNameAurora, engineNameMariadb, engineNameMySQL:
		if endpoint.MySQLSettings != nil {
			d.Set(names.AttrUsername, endpoint.MySQLSettings.Username)
			d.Set("server_name", endpoint.MySQLSettings.ServerName)
			d.Set(names.AttrPort, endpoint.MySQLSettings.Port)
			d.Set(names.AttrDatabaseName, endpoint.MySQLSettings.DatabaseName)
			d.Set("secrets_manager_access_role_arn", endpoint.MySQLSettings.SecretsManagerAccessRoleArn)
			d.Set("secrets_manager_arn", endpoint.MySQLSettings.SecretsManagerSecretId)
		} else {
			flattenTopLevelConnectionInfo(d, endpoint)
		}
	case engineNameAuroraPostgresql, engineNamePostgres:
		if endpoint.PostgreSQLSettings != nil {
			d.Set(names.AttrUsername, endpoint.PostgreSQLSettings.Username)
			d.Set("server_name", endpoint.PostgreSQLSettings.ServerName)
			d.Set(names.AttrPort, endpoint.PostgreSQLSettings.Port)
			d.Set(names.AttrDatabaseName, endpoint.PostgreSQLSettings.DatabaseName)
			d.Set("secrets_manager_access_role_arn", endpoint.PostgreSQLSettings.SecretsManagerAccessRoleArn)
			d.Set("secrets_manager_arn", endpoint.PostgreSQLSettings.SecretsManagerSecretId)
		} else {
			flattenTopLevelConnectionInfo(d, endpoint)
		}
		if err := d.Set("postgres_settings", flattenPostgreSQLSettings(endpoint.PostgreSQLSettings)); err != nil {
			return fmt.Errorf("setting postgres_settings: %w", err)
		}
	case engineNameDynamoDB:
		if endpoint.DynamoDbSettings != nil {
			d.Set("service_access_role", endpoint.DynamoDbSettings.ServiceAccessRoleArn)
		} else {
			d.Set("service_access_role", "")
		}
	case engineNameElasticsearch, engineNameOpenSearch:
		if err := d.Set("elasticsearch_settings", flattenOpenSearchSettings(endpoint.ElasticsearchSettings)); err != nil {
			return fmt.Errorf("setting elasticsearch_settings: %w", err)
		}
	case engineNameKafka:
		if endpoint.KafkaSettings != nil {
			// SASL password isn't returned in API. Propagate state value.
			tfMap := flattenKafkaSettings(endpoint.KafkaSettings)
			tfMap["sasl_password"] = d.Get("kafka_settings.0.sasl_password").(string)

			if err := d.Set("kafka_settings", []interface{}{tfMap}); err != nil {
				return fmt.Errorf("setting kafka_settings: %w", err)
			}
		} else {
			d.Set("kafka_settings", nil)
		}
	case engineNameKinesis:
		if err := d.Set("kinesis_settings", []interface{}{flattenKinesisSettings(endpoint.KinesisSettings)}); err != nil {
			return fmt.Errorf("setting kinesis_settings: %w", err)
		}
	case engineNameMongodb:
		if endpoint.MongoDbSettings != nil {
			d.Set(names.AttrUsername, endpoint.MongoDbSettings.Username)
			d.Set("server_name", endpoint.MongoDbSettings.ServerName)
			d.Set(names.AttrPort, endpoint.MongoDbSettings.Port)
			d.Set(names.AttrDatabaseName, endpoint.MongoDbSettings.DatabaseName)
			d.Set("secrets_manager_access_role_arn", endpoint.MongoDbSettings.SecretsManagerAccessRoleArn)
			d.Set("secrets_manager_arn", endpoint.MongoDbSettings.SecretsManagerSecretId)
		} else {
			flattenTopLevelConnectionInfo(d, endpoint)
		}
		if err := d.Set("mongodb_settings", flattenMongoDBSettings(endpoint.MongoDbSettings)); err != nil {
			return fmt.Errorf("setting mongodb_settings: %w", err)
		}
	case engineNameOracle:
		if endpoint.OracleSettings != nil {
			d.Set(names.AttrUsername, endpoint.OracleSettings.Username)
			d.Set("server_name", endpoint.OracleSettings.ServerName)
			d.Set(names.AttrPort, endpoint.OracleSettings.Port)
			d.Set(names.AttrDatabaseName, endpoint.OracleSettings.DatabaseName)
			d.Set("secrets_manager_access_role_arn", endpoint.OracleSettings.SecretsManagerAccessRoleArn)
			d.Set("secrets_manager_arn", endpoint.OracleSettings.SecretsManagerSecretId)
		} else {
			flattenTopLevelConnectionInfo(d, endpoint)
		}
	case engineNameRedis:
		// Auth password isn't returned in API. Propagate state value.
		tfMap := flattenRedisSettings(endpoint.RedisSettings)
		tfMap["auth_password"] = d.Get("redis_settings.0.auth_password").(string)

		if err := d.Set("redis_settings", []interface{}{tfMap}); err != nil {
			return fmt.Errorf("setting redis_settings: %w", err)
		}
	case engineNameRedshift:
		if endpoint.RedshiftSettings != nil {
			d.Set(names.AttrUsername, endpoint.RedshiftSettings.Username)
			d.Set("server_name", endpoint.RedshiftSettings.ServerName)
			d.Set(names.AttrPort, endpoint.RedshiftSettings.Port)
			d.Set(names.AttrDatabaseName, endpoint.RedshiftSettings.DatabaseName)
			d.Set("secrets_manager_access_role_arn", endpoint.RedshiftSettings.SecretsManagerAccessRoleArn)
			d.Set("secrets_manager_arn", endpoint.RedshiftSettings.SecretsManagerSecretId)
		} else {
			flattenTopLevelConnectionInfo(d, endpoint)
		}
		if err := d.Set("redshift_settings", flattenRedshiftSettings(endpoint.RedshiftSettings)); err != nil {
			return fmt.Errorf("setting redshift_settings: %w", err)
		}
	case engineNameSQLServer, engineNameBabelfish:
		if endpoint.MicrosoftSQLServerSettings != nil {
			d.Set(names.AttrUsername, endpoint.MicrosoftSQLServerSettings.Username)
			d.Set("server_name", endpoint.MicrosoftSQLServerSettings.ServerName)
			d.Set(names.AttrPort, endpoint.MicrosoftSQLServerSettings.Port)
			d.Set(names.AttrDatabaseName, endpoint.MicrosoftSQLServerSettings.DatabaseName)
			d.Set("secrets_manager_access_role_arn", endpoint.MicrosoftSQLServerSettings.SecretsManagerAccessRoleArn)
			d.Set("secrets_manager_arn", endpoint.MicrosoftSQLServerSettings.SecretsManagerSecretId)
		} else {
			flattenTopLevelConnectionInfo(d, endpoint)
		}
	case engineNameSybase:
		if endpoint.SybaseSettings != nil {
			d.Set(names.AttrUsername, endpoint.SybaseSettings.Username)
			d.Set("server_name", endpoint.SybaseSettings.ServerName)
			d.Set(names.AttrPort, endpoint.SybaseSettings.Port)
			d.Set(names.AttrDatabaseName, endpoint.SybaseSettings.DatabaseName)
			d.Set("secrets_manager_access_role_arn", endpoint.SybaseSettings.SecretsManagerAccessRoleArn)
			d.Set("secrets_manager_arn", endpoint.SybaseSettings.SecretsManagerSecretId)
		} else {
			flattenTopLevelConnectionInfo(d, endpoint)
		}
	case engineNameDB2, engineNameDB2zOS:
		if endpoint.IBMDb2Settings != nil {
			d.Set(names.AttrUsername, endpoint.IBMDb2Settings.Username)
			d.Set("server_name", endpoint.IBMDb2Settings.ServerName)
			d.Set(names.AttrPort, endpoint.IBMDb2Settings.Port)
			d.Set(names.AttrDatabaseName, endpoint.IBMDb2Settings.DatabaseName)
			d.Set("secrets_manager_access_role_arn", endpoint.IBMDb2Settings.SecretsManagerAccessRoleArn)
			d.Set("secrets_manager_arn", endpoint.IBMDb2Settings.SecretsManagerSecretId)
		} else {
			flattenTopLevelConnectionInfo(d, endpoint)
		}
	case engineNameS3:
		if err := d.Set("s3_settings", flattenS3Settings(endpoint.S3Settings)); err != nil {
			return fmt.Errorf("setting s3_settings for DMS: %s", err)
		}
	default:
		d.Set(names.AttrDatabaseName, endpoint.DatabaseName)
		d.Set(names.AttrPort, endpoint.Port)
		d.Set("server_name", endpoint.ServerName)
		d.Set(names.AttrUsername, endpoint.Username)
	}

	d.Set(names.AttrKMSKeyARN, endpoint.KmsKeyId)
	d.Set("ssl_mode", endpoint.SslMode)

	return nil
}

func steadyEndpointReplicationTasks(ctx context.Context, conn *dms.Client, arn string) error {
	tasks, err := findReplicationTasksByEndpointARN(ctx, conn, arn)

	if err != nil {
		return err
	}

	for _, task := range tasks {
		rtID := aws.ToString(task.ReplicationTaskIdentifier)
		switch aws.ToString(task.Status) {
		case replicationTaskStatusRunning, replicationTaskStatusFailed, replicationTaskStatusReady, replicationTaskStatusStopped:
			continue
		case replicationTaskStatusCreating, replicationTaskStatusDeleting, replicationTaskStatusModifying, replicationTaskStatusStopping, replicationTaskStatusStarting:
			if _, err := waitReplicationTaskSteady(ctx, conn, rtID); err != nil {
				return err
			}
		}
	}

	return nil
}

func stopEndpointReplicationTasks(ctx context.Context, conn *dms.Client, arn string) ([]awstypes.ReplicationTask, error) {
	if err := steadyEndpointReplicationTasks(ctx, conn, arn); err != nil {
		return nil, err
	}

	tasks, err := findReplicationTasksByEndpointARN(ctx, conn, arn)

	if err != nil {
		return nil, err
	}

	var stoppedTasks []awstypes.ReplicationTask
	for _, task := range tasks {
		rtID := aws.ToString(task.ReplicationTaskIdentifier)
		switch aws.ToString(task.Status) {
		case replicationTaskStatusRunning:
			err := stopReplicationTask(ctx, conn, rtID)

			if err != nil {
				return stoppedTasks, err
			}
			stoppedTasks = append(stoppedTasks, task)
		default:
			continue
		}
	}

	return stoppedTasks, nil
}

func startEndpointReplicationTasks(ctx context.Context, conn *dms.Client, arn string, tasks []awstypes.ReplicationTask) error {
	const maxConnTestWaitTime = 120 * time.Second

	if len(tasks) == 0 {
		return nil
	}

	if err := steadyEndpointReplicationTasks(ctx, conn, arn); err != nil {
		return err
	}

	for _, task := range tasks {
		_, err := conn.TestConnection(ctx, &dms.TestConnectionInput{
			EndpointArn:            aws.String(arn),
			ReplicationInstanceArn: task.ReplicationInstanceArn,
		})

		if errs.IsAErrorMessageContains[*awstypes.InvalidResourceStateFault](err, "already being tested") {
			continue
		}

		if err != nil {
			return fmt.Errorf("testing connection: %w", err)
		}

		waiter := dms.NewTestConnectionSucceedsWaiter(conn)

		err = waiter.Wait(ctx, &dms.DescribeConnectionsInput{
			Filters: []awstypes.Filter{
				{
					Name:   aws.String("endpoint-arn"),
					Values: []string{arn},
				},
			},
		}, maxConnTestWaitTime)

		if err != nil {
			return fmt.Errorf("waiting until test connection succeeds: %w", err)
		}

		if err := startReplicationTask(ctx, conn, aws.ToString(task.ReplicationTaskIdentifier)); err != nil {
			return fmt.Errorf("starting replication task: %w", err)
		}
	}

	return nil
}

func findReplicationTasksByEndpointARN(ctx context.Context, conn *dms.Client, arn string) ([]awstypes.ReplicationTask, error) {
	input := &dms.DescribeReplicationTasksInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("endpoint-arn"),
				Values: []string{arn},
			},
		},
	}

	return findReplicationTasks(ctx, conn, input)
}

func flattenOpenSearchSettings(settings *awstypes.ElasticsearchSettings) []map[string]interface{} {
	if settings == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"endpoint_uri":               aws.ToString(settings.EndpointUri),
		"error_retry_duration":       aws.ToInt32(settings.ErrorRetryDuration),
		"full_load_error_percentage": aws.ToInt32(settings.FullLoadErrorPercentage),
		"service_access_role_arn":    aws.ToString(settings.ServiceAccessRoleArn),
		"use_new_mapping_type":       aws.ToBool(settings.UseNewMappingType),
	}

	return []map[string]interface{}{m}
}

func expandKafkaSettings(tfMap map[string]interface{}) *awstypes.KafkaSettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.KafkaSettings{}

	if v, ok := tfMap["broker"].(string); ok && v != "" {
		apiObject.Broker = aws.String(v)
	}

	if v, ok := tfMap["include_control_details"].(bool); ok {
		apiObject.IncludeControlDetails = aws.Bool(v)
	}

	if v, ok := tfMap["include_null_and_empty"].(bool); ok {
		apiObject.IncludeNullAndEmpty = aws.Bool(v)
	}

	if v, ok := tfMap["include_partition_value"].(bool); ok {
		apiObject.IncludePartitionValue = aws.Bool(v)
	}

	if v, ok := tfMap["include_table_alter_operations"].(bool); ok {
		apiObject.IncludeTableAlterOperations = aws.Bool(v)
	}

	if v, ok := tfMap["include_transaction_details"].(bool); ok {
		apiObject.IncludeTransactionDetails = aws.Bool(v)
	}

	if v, ok := tfMap["message_format"].(string); ok && v != "" {
		apiObject.MessageFormat = awstypes.MessageFormatValue(v)
	}

	if v, ok := tfMap["message_max_bytes"].(int); ok && v != 0 {
		apiObject.MessageMaxBytes = aws.Int32(int32(v))
	}

	if v, ok := tfMap["no_hex_prefix"].(bool); ok {
		apiObject.NoHexPrefix = aws.Bool(v)
	}

	if v, ok := tfMap["partition_include_schema_table"].(bool); ok {
		apiObject.PartitionIncludeSchemaTable = aws.Bool(v)
	}

	if v, ok := tfMap["sasl_password"].(string); ok && v != "" {
		apiObject.SaslPassword = aws.String(v)
	}

	if v, ok := tfMap["sasl_username"].(string); ok && v != "" {
		apiObject.SaslUsername = aws.String(v)
	}

	if v, ok := tfMap["security_protocol"].(string); ok && v != "" {
		apiObject.SecurityProtocol = awstypes.KafkaSecurityProtocol(v)
	}

	if v, ok := tfMap["ssl_ca_certificate_arn"].(string); ok && v != "" {
		apiObject.SslCaCertificateArn = aws.String(v)
	}

	if v, ok := tfMap["ssl_client_certificate_arn"].(string); ok && v != "" {
		apiObject.SslClientCertificateArn = aws.String(v)
	}

	if v, ok := tfMap["ssl_client_key_arn"].(string); ok && v != "" {
		apiObject.SslClientKeyArn = aws.String(v)
	}

	if v, ok := tfMap["ssl_client_key_password"].(string); ok && v != "" {
		apiObject.SslClientKeyPassword = aws.String(v)
	}

	if v, ok := tfMap["topic"].(string); ok && v != "" {
		apiObject.Topic = aws.String(v)
	}

	return apiObject
}

func flattenKafkaSettings(apiObject *awstypes.KafkaSettings) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Broker; v != nil {
		tfMap["broker"] = aws.ToString(v)
	}

	if v := apiObject.IncludeControlDetails; v != nil {
		tfMap["include_control_details"] = aws.ToBool(v)
	}

	if v := apiObject.IncludeNullAndEmpty; v != nil {
		tfMap["include_null_and_empty"] = aws.ToBool(v)
	}

	if v := apiObject.IncludePartitionValue; v != nil {
		tfMap["include_partition_value"] = aws.ToBool(v)
	}

	if v := apiObject.IncludeTableAlterOperations; v != nil {
		tfMap["include_table_alter_operations"] = aws.ToBool(v)
	}

	if v := apiObject.IncludeTransactionDetails; v != nil {
		tfMap["include_transaction_details"] = aws.ToBool(v)
	}

	tfMap["message_format"] = string(apiObject.MessageFormat)

	if v := apiObject.MessageMaxBytes; v != nil {
		tfMap["message_max_bytes"] = aws.ToInt32(v)
	}

	if v := apiObject.NoHexPrefix; v != nil {
		tfMap["no_hex_prefix"] = aws.ToBool(v)
	}

	if v := apiObject.PartitionIncludeSchemaTable; v != nil {
		tfMap["partition_include_schema_table"] = aws.ToBool(v)
	}

	if v := apiObject.SaslPassword; v != nil {
		tfMap["sasl_password"] = aws.ToString(v)
	}

	if v := apiObject.SaslUsername; v != nil {
		tfMap["sasl_username"] = aws.ToString(v)
	}

	tfMap["security_protocol"] = string(apiObject.SecurityProtocol)

	if v := apiObject.SslCaCertificateArn; v != nil {
		tfMap["ssl_ca_certificate_arn"] = aws.ToString(v)
	}

	if v := apiObject.SslClientCertificateArn; v != nil {
		tfMap["ssl_client_certificate_arn"] = aws.ToString(v)
	}

	if v := apiObject.SslClientKeyArn; v != nil {
		tfMap["ssl_client_key_arn"] = aws.ToString(v)
	}

	if v := apiObject.SslClientKeyPassword; v != nil {
		tfMap["ssl_client_key_password"] = aws.ToString(v)
	}

	if v := apiObject.Topic; v != nil {
		tfMap["topic"] = aws.ToString(v)
	}

	return tfMap
}

func expandKinesisSettings(tfMap map[string]interface{}) *awstypes.KinesisSettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.KinesisSettings{}

	if v, ok := tfMap["include_control_details"].(bool); ok {
		apiObject.IncludeControlDetails = aws.Bool(v)
	}

	if v, ok := tfMap["include_null_and_empty"].(bool); ok {
		apiObject.IncludeNullAndEmpty = aws.Bool(v)
	}

	if v, ok := tfMap["include_partition_value"].(bool); ok {
		apiObject.IncludePartitionValue = aws.Bool(v)
	}

	if v, ok := tfMap["include_table_alter_operations"].(bool); ok {
		apiObject.IncludeTableAlterOperations = aws.Bool(v)
	}

	if v, ok := tfMap["include_transaction_details"].(bool); ok {
		apiObject.IncludeTransactionDetails = aws.Bool(v)
	}

	if v, ok := tfMap["message_format"].(string); ok && v != "" {
		apiObject.MessageFormat = awstypes.MessageFormatValue(v)
	}

	if v, ok := tfMap["partition_include_schema_table"].(bool); ok {
		apiObject.PartitionIncludeSchemaTable = aws.Bool(v)
	}

	if v, ok := tfMap["service_access_role_arn"].(string); ok && v != "" {
		apiObject.ServiceAccessRoleArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrStreamARN].(string); ok && v != "" {
		apiObject.StreamArn = aws.String(v)
	}

	return apiObject
}

func flattenKinesisSettings(apiObject *awstypes.KinesisSettings) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.IncludeControlDetails; v != nil {
		tfMap["include_control_details"] = aws.ToBool(v)
	}

	if v := apiObject.IncludeNullAndEmpty; v != nil {
		tfMap["include_null_and_empty"] = aws.ToBool(v)
	}

	if v := apiObject.IncludePartitionValue; v != nil {
		tfMap["include_partition_value"] = aws.ToBool(v)
	}

	if v := apiObject.IncludeTableAlterOperations; v != nil {
		tfMap["include_table_alter_operations"] = aws.ToBool(v)
	}

	if v := apiObject.IncludeTransactionDetails; v != nil {
		tfMap["include_transaction_details"] = aws.ToBool(v)
	}

	tfMap["message_format"] = string(apiObject.MessageFormat)

	if v := apiObject.PartitionIncludeSchemaTable; v != nil {
		tfMap["partition_include_schema_table"] = aws.ToBool(v)
	}

	if v := apiObject.ServiceAccessRoleArn; v != nil {
		tfMap["service_access_role_arn"] = aws.ToString(v)
	}

	if v := apiObject.StreamArn; v != nil {
		tfMap[names.AttrStreamARN] = aws.ToString(v)
	}

	return tfMap
}

func flattenMongoDBSettings(settings *awstypes.MongoDbSettings) []map[string]interface{} {
	if settings == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"auth_type":           string(settings.AuthType),
		"auth_mechanism":      string(settings.AuthMechanism),
		"nesting_level":       string(settings.NestingLevel),
		"extract_doc_id":      aws.ToString(settings.ExtractDocId),
		"docs_to_investigate": aws.ToString(settings.DocsToInvestigate),
		"auth_source":         aws.ToString(settings.AuthSource),
	}

	return []map[string]interface{}{m}
}

func expandRedisSettings(tfMap map[string]interface{}) *awstypes.RedisSettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.RedisSettings{}

	if v, ok := tfMap["auth_password"].(string); ok && v != "" {
		apiObject.AuthPassword = aws.String(v)
	}
	if v, ok := tfMap["auth_type"].(string); ok && v != "" {
		apiObject.AuthType = awstypes.RedisAuthTypeValue(v)
	}
	if v, ok := tfMap["auth_user_name"].(string); ok && v != "" {
		apiObject.AuthUserName = aws.String(v)
	}
	if v, ok := tfMap[names.AttrPort].(int); ok {
		apiObject.Port = int32(v)
	}
	if v, ok := tfMap["server_name"].(string); ok && v != "" {
		apiObject.ServerName = aws.String(v)
	}
	if v, ok := tfMap["ssl_ca_certificate_arn"].(string); ok && v != "" {
		apiObject.SslCaCertificateArn = aws.String(v)
	}
	if v, ok := tfMap["ssl_security_protocol"].(string); ok && v != "" {
		apiObject.SslSecurityProtocol = awstypes.SslSecurityProtocolValue(v)
	}

	return apiObject
}

func flattenRedisSettings(apiObject *awstypes.RedisSettings) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AuthPassword; v != nil {
		tfMap["auth_password"] = aws.ToString(v)
	}
	tfMap["auth_type"] = string(apiObject.AuthType)
	if v := apiObject.AuthUserName; v != nil {
		tfMap["auth_user_name"] = aws.ToString(v)
	}
	tfMap[names.AttrPort] = apiObject.Port
	if v := apiObject.ServerName; v != nil {
		tfMap["server_name"] = aws.ToString(v)
	}
	if v := apiObject.SslCaCertificateArn; v != nil {
		tfMap["ssl_ca_certificate_arn"] = aws.ToString(v)
	}
	tfMap["ssl_security_protocol"] = string(apiObject.SslSecurityProtocol)
	return tfMap
}

func flattenRedshiftSettings(settings *awstypes.RedshiftSettings) []map[string]interface{} {
	if settings == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"bucket_folder":                     aws.ToString(settings.BucketFolder),
		names.AttrBucketName:                aws.ToString(settings.BucketName),
		"encryption_mode":                   string(settings.EncryptionMode),
		"server_side_encryption_kms_key_id": aws.ToString(settings.ServerSideEncryptionKmsKeyId),
		"service_access_role_arn":           aws.ToString(settings.ServiceAccessRoleArn),
	}

	return []map[string]interface{}{m}
}

func expandPostgreSQLSettings(tfMap map[string]interface{}) *awstypes.PostgreSQLSettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.PostgreSQLSettings{}

	if v, ok := tfMap["after_connect_script"].(string); ok && v != "" {
		apiObject.AfterConnectScript = aws.String(v)
	}
	if v, ok := tfMap["babelfish_database_name"].(string); ok && v != "" {
		apiObject.BabelfishDatabaseName = aws.String(v)
	}
	if v, ok := tfMap["capture_ddls"].(bool); ok {
		apiObject.CaptureDdls = aws.Bool(v)
	}
	if v, ok := tfMap["database_mode"].(string); ok && v != "" {
		apiObject.DatabaseMode = awstypes.DatabaseMode(v)
	}
	if v, ok := tfMap["ddl_artifacts_schema"].(string); ok && v != "" {
		apiObject.DdlArtifactsSchema = aws.String(v)
	}
	if v, ok := tfMap["execute_timeout"].(int); ok {
		apiObject.ExecuteTimeout = aws.Int32(int32(v))
	}
	if v, ok := tfMap["fail_tasks_on_lob_truncation"].(bool); ok {
		apiObject.FailTasksOnLobTruncation = aws.Bool(v)
	}
	if v, ok := tfMap["heartbeat_enable"].(bool); ok {
		apiObject.HeartbeatEnable = aws.Bool(v)
	}
	if v, ok := tfMap["heartbeat_frequency"].(int); ok {
		apiObject.HeartbeatFrequency = aws.Int32(int32(v))
	}
	if v, ok := tfMap["heartbeat_schema"].(string); ok && v != "" {
		apiObject.HeartbeatSchema = aws.String(v)
	}
	if v, ok := tfMap["map_boolean_as_boolean"].(bool); ok {
		apiObject.MapBooleanAsBoolean = aws.Bool(v)
	}
	if v, ok := tfMap["map_jsonb_as_clob"].(bool); ok {
		apiObject.MapJsonbAsClob = aws.Bool(v)
	}
	if v, ok := tfMap["map_long_varchar_as"].(string); ok && v != "" {
		apiObject.MapLongVarcharAs = awstypes.LongVarcharMappingType(v)
	}
	if v, ok := tfMap["max_file_size"].(int); ok {
		apiObject.MaxFileSize = aws.Int32(int32(v))
	}
	if v, ok := tfMap["plugin_name"].(string); ok && v != "" {
		apiObject.PluginName = awstypes.PluginNameValue(v)
	}
	if v, ok := tfMap["slot_name"].(string); ok && v != "" {
		apiObject.SlotName = aws.String(v)
	}

	return apiObject
}

func flattenPostgreSQLSettings(apiObject *awstypes.PostgreSQLSettings) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AfterConnectScript; v != nil {
		tfMap["after_connect_script"] = aws.ToString(v)
	}
	if v := apiObject.BabelfishDatabaseName; v != nil {
		tfMap["babelfish_database_name"] = aws.ToString(v)
	}
	if v := apiObject.CaptureDdls; v != nil {
		tfMap["capture_ddls"] = aws.ToBool(v)
	}
	tfMap["database_mode"] = string(apiObject.DatabaseMode)
	if v := apiObject.DdlArtifactsSchema; v != nil {
		tfMap["ddl_artifacts_schema"] = aws.ToString(v)
	}
	if v := apiObject.ExecuteTimeout; v != nil {
		tfMap["execute_timeout"] = aws.ToInt32(v)
	}
	if v := apiObject.FailTasksOnLobTruncation; v != nil {
		tfMap["fail_tasks_on_lob_truncation"] = aws.ToBool(v)
	}
	if v := apiObject.HeartbeatEnable; v != nil {
		tfMap["heartbeat_enable"] = aws.ToBool(v)
	}
	if v := apiObject.HeartbeatFrequency; v != nil {
		tfMap["heartbeat_frequency"] = aws.ToInt32(v)
	}
	if v := apiObject.HeartbeatSchema; v != nil {
		tfMap["heartbeat_schema"] = aws.ToString(v)
	}
	if v := apiObject.MapBooleanAsBoolean; v != nil {
		tfMap["map_boolean_as_boolean"] = aws.ToBool(v)
	}
	if v := apiObject.MapJsonbAsClob; v != nil {
		tfMap["map_jsonb_as_clob"] = aws.ToBool(v)
	}
	tfMap["map_long_varchar_as"] = string(apiObject.MapLongVarcharAs)
	if v := apiObject.MaxFileSize; v != nil {
		tfMap["max_file_size"] = aws.ToInt32(v)
	}
	tfMap["plugin_name"] = string(apiObject.PluginName)
	if v := apiObject.SlotName; v != nil {
		tfMap["slot_name"] = aws.ToString(v)
	}

	return []map[string]interface{}{tfMap}
}

func expandS3Settings(tfMap map[string]interface{}) *awstypes.S3Settings {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.S3Settings{}

	if v, ok := tfMap["add_column_name"].(bool); ok {
		apiObject.AddColumnName = aws.Bool(v)
	}
	if v, ok := tfMap["bucket_folder"].(string); ok {
		apiObject.BucketFolder = aws.String(v)
	}
	if v, ok := tfMap[names.AttrBucketName].(string); ok {
		apiObject.BucketName = aws.String(v)
	}
	if v, ok := tfMap["canned_acl_for_objects"].(string); ok {
		apiObject.CannedAclForObjects = awstypes.CannedAclForObjectsValue(v)
	}
	if v, ok := tfMap["cdc_inserts_and_updates"].(bool); ok {
		apiObject.CdcInsertsAndUpdates = aws.Bool(v)
	}
	if v, ok := tfMap["cdc_inserts_only"].(bool); ok {
		apiObject.CdcInsertsOnly = aws.Bool(v)
	}
	if v, ok := tfMap["cdc_max_batch_interval"].(int); ok {
		apiObject.CdcMaxBatchInterval = aws.Int32(int32(v))
	}
	if v, ok := tfMap["cdc_min_file_size"].(int); ok {
		apiObject.CdcMinFileSize = aws.Int32(int32(v))
	}
	if v, ok := tfMap["cdc_path"].(string); ok {
		apiObject.CdcPath = aws.String(v)
	}
	if v, ok := tfMap["compression_type"].(string); ok {
		apiObject.CompressionType = awstypes.CompressionTypeValue(v)
	}
	if v, ok := tfMap["csv_delimiter"].(string); ok {
		apiObject.CsvDelimiter = aws.String(v)
	}
	if v, ok := tfMap["csv_no_sup_value"].(string); ok {
		apiObject.CsvNoSupValue = aws.String(v)
	}
	if v, ok := tfMap["csv_null_value"].(string); ok {
		apiObject.CsvNullValue = aws.String(v)
	}
	if v, ok := tfMap["csv_row_delimiter"].(string); ok {
		apiObject.CsvRowDelimiter = aws.String(v)
	}
	if v, ok := tfMap["data_format"].(string); ok {
		apiObject.DataFormat = awstypes.DataFormatValue(v)
	}
	if v, ok := tfMap["data_page_size"].(int); ok {
		apiObject.DataPageSize = aws.Int32(int32(v))
	}
	if v, ok := tfMap["date_partition_delimiter"].(string); ok {
		apiObject.DatePartitionDelimiter = awstypes.DatePartitionDelimiterValue(v)
	}
	if v, ok := tfMap["date_partition_enabled"].(bool); ok {
		apiObject.DatePartitionEnabled = aws.Bool(v)
	}
	if v, ok := tfMap["date_partition_sequence"].(string); ok {
		apiObject.DatePartitionSequence = awstypes.DatePartitionSequenceValue(v)
	}
	if v, ok := tfMap["dict_page_size_limit"].(int); ok {
		apiObject.DictPageSizeLimit = aws.Int32(int32(v))
	}
	if v, ok := tfMap["enable_statistics"].(bool); ok {
		apiObject.EnableStatistics = aws.Bool(v)
	}
	if v, ok := tfMap["encoding_type"].(string); ok {
		apiObject.EncodingType = awstypes.EncodingTypeValue(v)
	}
	if v, ok := tfMap["encryption_mode"].(string); ok {
		apiObject.EncryptionMode = awstypes.EncryptionModeValue(v)
	}
	if v, ok := tfMap["external_table_definition"].(string); ok {
		apiObject.ExternalTableDefinition = aws.String(v)
	}
	if v, ok := tfMap["glue_catalog_generation"].(bool); ok {
		apiObject.GlueCatalogGeneration = aws.Bool(v)
	}
	if v, ok := tfMap["ignore_header_rows"].(int); ok {
		apiObject.IgnoreHeaderRows = aws.Int32(int32(v))
	}
	if v, ok := tfMap["include_op_for_full_load"].(bool); ok {
		apiObject.IncludeOpForFullLoad = aws.Bool(v)
	}
	if v, ok := tfMap["max_file_size"].(int); ok {
		apiObject.MaxFileSize = aws.Int32(int32(v))
	}
	if v, ok := tfMap["parquet_timestamp_in_millisecond"].(bool); ok {
		apiObject.ParquetTimestampInMillisecond = aws.Bool(v)
	}
	if v, ok := tfMap["parquet_version"].(string); ok {
		apiObject.ParquetVersion = awstypes.ParquetVersionValue(v)
	}
	if v, ok := tfMap["preserve_transactions"].(bool); ok {
		apiObject.PreserveTransactions = aws.Bool(v)
	}
	if v, ok := tfMap["rfc_4180"].(bool); ok {
		apiObject.Rfc4180 = aws.Bool(v)
	}
	if v, ok := tfMap["row_group_length"].(int); ok {
		apiObject.RowGroupLength = aws.Int32(int32(v))
	}
	if v, ok := tfMap["server_side_encryption_kms_key_id"].(string); ok {
		apiObject.ServerSideEncryptionKmsKeyId = aws.String(v)
	}
	if v, ok := tfMap["service_access_role_arn"].(string); ok {
		apiObject.ServiceAccessRoleArn = aws.String(v)
	}
	if v, ok := tfMap["timestamp_column_name"].(string); ok {
		apiObject.TimestampColumnName = aws.String(v)
	}
	if v, ok := tfMap["use_csv_no_sup_value"].(bool); ok {
		apiObject.UseCsvNoSupValue = aws.Bool(v)
	}
	if v, ok := tfMap["use_task_start_time_for_full_load_timestamp"].(bool); ok {
		apiObject.UseTaskStartTimeForFullLoadTimestamp = aws.Bool(v)
	}

	return apiObject
}

func flattenS3Settings(apiObject *awstypes.S3Settings) []map[string]interface{} {
	if apiObject == nil {
		return []map[string]interface{}{}
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AddColumnName; v != nil {
		tfMap["add_column_name"] = aws.ToBool(v)
	}
	if v := apiObject.BucketFolder; v != nil {
		tfMap["bucket_folder"] = aws.ToString(v)
	}
	if v := apiObject.BucketName; v != nil {
		tfMap[names.AttrBucketName] = aws.ToString(v)
	}
	tfMap["canned_acl_for_objects"] = string(apiObject.CannedAclForObjects)
	if v := apiObject.CdcInsertsAndUpdates; v != nil {
		tfMap["cdc_inserts_and_updates"] = aws.ToBool(v)
	}
	if v := apiObject.CdcInsertsOnly; v != nil {
		tfMap["cdc_inserts_only"] = aws.ToBool(v)
	}
	if v := apiObject.CdcMaxBatchInterval; v != nil {
		tfMap["cdc_max_batch_interval"] = aws.ToInt32(v)
	}
	if v := apiObject.CdcMinFileSize; v != nil {
		tfMap["cdc_min_file_size"] = aws.ToInt32(v)
	}
	if v := apiObject.CdcPath; v != nil {
		tfMap["cdc_path"] = aws.ToString(v)
	}
	tfMap["compression_type"] = string(apiObject.CompressionType)
	if v := apiObject.CsvDelimiter; v != nil {
		tfMap["csv_delimiter"] = aws.ToString(v)
	}
	if v := apiObject.CsvNoSupValue; v != nil {
		tfMap["csv_no_sup_value"] = aws.ToString(v)
	}
	if v := apiObject.CsvNullValue; v != nil {
		tfMap["csv_null_value"] = aws.ToString(v)
	}
	if v := apiObject.CsvRowDelimiter; v != nil {
		tfMap["csv_row_delimiter"] = aws.ToString(v)
	}
	tfMap["data_format"] = string(apiObject.DataFormat)
	if v := apiObject.DataPageSize; v != nil {
		tfMap["data_page_size"] = aws.ToInt32(v)
	}
	tfMap["date_partition_delimiter"] = string(apiObject.DatePartitionDelimiter)
	if v := apiObject.DatePartitionEnabled; v != nil {
		tfMap["date_partition_enabled"] = aws.ToBool(v)
	}
	tfMap["date_partition_sequence"] = string(apiObject.DatePartitionSequence)
	if v := apiObject.DictPageSizeLimit; v != nil {
		tfMap["dict_page_size_limit"] = aws.ToInt32(v)
	}
	if v := apiObject.EnableStatistics; v != nil {
		tfMap["enable_statistics"] = aws.ToBool(v)
	}
	tfMap["encoding_type"] = string(apiObject.EncodingType)
	tfMap["encryption_mode"] = string(apiObject.EncryptionMode)
	if v := apiObject.ExternalTableDefinition; v != nil {
		tfMap["external_table_definition"] = aws.ToString(v)
	}
	if v := apiObject.GlueCatalogGeneration; v != nil {
		tfMap["glue_catalog_generation"] = aws.ToBool(v)
	}
	if v := apiObject.IgnoreHeaderRows; v != nil {
		tfMap["ignore_header_rows"] = aws.ToInt32(v)
	}
	if v := apiObject.IncludeOpForFullLoad; v != nil {
		tfMap["include_op_for_full_load"] = aws.ToBool(v)
	}
	if v := apiObject.MaxFileSize; v != nil {
		tfMap["max_file_size"] = aws.ToInt32(v)
	}
	if v := apiObject.ParquetTimestampInMillisecond; v != nil {
		tfMap["parquet_timestamp_in_millisecond"] = aws.ToBool(v)
	}
	tfMap["parquet_version"] = string(apiObject.ParquetVersion)
	if v := apiObject.Rfc4180; v != nil {
		tfMap["rfc_4180"] = aws.ToBool(v)
	}
	if v := apiObject.RowGroupLength; v != nil {
		tfMap["row_group_length"] = aws.ToInt32(v)
	}
	if v := apiObject.ServerSideEncryptionKmsKeyId; v != nil {
		tfMap["server_side_encryption_kms_key_id"] = aws.ToString(v)
	}
	if v := apiObject.ServiceAccessRoleArn; v != nil {
		tfMap["service_access_role_arn"] = aws.ToString(v)
	}
	if v := apiObject.TimestampColumnName; v != nil {
		tfMap["timestamp_column_name"] = aws.ToString(v)
	}
	if v := apiObject.UseCsvNoSupValue; v != nil {
		tfMap["use_csv_no_sup_value"] = aws.ToBool(v)
	}
	if v := apiObject.UseTaskStartTimeForFullLoadTimestamp; v != nil {
		tfMap["use_task_start_time_for_full_load_timestamp"] = aws.ToBool(v)
	}

	return []map[string]interface{}{tfMap}
}

func suppressExtraConnectionAttributesDiffs(_, old, new string, d *schema.ResourceData) bool {
	if d.Id() != "" {
		o := extraConnectionAttributesToSet(old)
		n := extraConnectionAttributesToSet(new)

		var config *schema.Set
		// when the engine is "s3" or "mongodb", the extra_connection_attributes
		// can consist of a subset of the attributes configured in the {engine}_settings block;
		// fields such as service_access_role_arn (in the case of "s3") are not returned from the API in
		// extra_connection_attributes thus we take the Set difference to ensure
		// the returned attributes were set in the {engine}_settings block or originally
		// in the extra_connection_attributes field
		if v, ok := d.GetOk("mongodb_settings"); ok {
			config = engineSettingsToSet(v.([]interface{}))
		} else if v, ok := d.GetOk("s3_settings"); ok {
			config = engineSettingsToSet(v.([]interface{}))
		}

		if o != nil && config != nil {
			diff := o.Difference(config)
			diff2 := n.Difference(config)

			return (diff.Len() == 0 && diff2.Len() == 0) || diff.Equal(n)
		}
	}
	return false
}

// extraConnectionAttributesToSet accepts an extra_connection_attributes
// string in the form of "key=value;key2=value2;" and returns
// the Set representation, with each element being the key/value pair
func extraConnectionAttributesToSet(extra string) *schema.Set {
	if extra == "" {
		return nil
	}

	s := &schema.Set{F: schema.HashString}

	parts := strings.Split(extra, ";")
	for _, part := range parts {
		kvParts := strings.Split(part, "=")
		if len(kvParts) != 2 {
			continue
		}

		k, v := kvParts[0], kvParts[1]
		// normalize key, from camelCase to snake_case,
		// and value where hyphens maybe used in a config
		// but the API returns with underscores
		matchAllCap := regexache.MustCompile("([a-z])([A-Z])")
		key := matchAllCap.ReplaceAllString(k, "${1}_${2}")
		normalizedVal := strings.Replace(strings.ToLower(v), "-", "_", -1)

		s.Add(fmt.Sprintf("%s=%s", strings.ToLower(key), normalizedVal))
	}

	return s
}

// engineSettingsToSet accepts the {engine}_settings block as a list
// and returns the Set representation, with each element being the key/value pair
func engineSettingsToSet(l []interface{}) *schema.Set {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	s := &schema.Set{F: schema.HashString}

	for k, v := range tfMap {
		switch t := v.(type) {
		case string:
			// normalize value for changes in case or where hyphens
			// maybe used in a config but the API returns with underscores
			normalizedVal := strings.Replace(strings.ToLower(t), "-", "_", -1)
			s.Add(fmt.Sprintf("%s=%v", k, normalizedVal))
		default:
			s.Add(fmt.Sprintf("%s=%v", k, t))
		}
	}

	return s
}

func expandTopLevelConnectionInfo(d *schema.ResourceData, input *dms.CreateEndpointInput) {
	input.Username = aws.String(d.Get(names.AttrUsername).(string))
	input.Password = aws.String(d.Get(names.AttrPassword).(string))
	input.ServerName = aws.String(d.Get("server_name").(string))
	input.Port = aws.Int32(int32(d.Get(names.AttrPort).(int)))

	if v, ok := d.GetOk(names.AttrDatabaseName); ok {
		input.DatabaseName = aws.String(v.(string))
	}
}

func expandTopLevelConnectionInfoModify(d *schema.ResourceData, input *dms.ModifyEndpointInput) {
	input.Username = aws.String(d.Get(names.AttrUsername).(string))
	input.Password = aws.String(d.Get(names.AttrPassword).(string))
	input.ServerName = aws.String(d.Get("server_name").(string))
	input.Port = aws.Int32(int32(d.Get(names.AttrPort).(int)))

	if v, ok := d.GetOk(names.AttrDatabaseName); ok {
		input.DatabaseName = aws.String(v.(string))
	}
}

func flattenTopLevelConnectionInfo(d *schema.ResourceData, endpoint *awstypes.Endpoint) {
	d.Set(names.AttrUsername, endpoint.Username)
	d.Set("server_name", endpoint.ServerName)
	d.Set(names.AttrPort, endpoint.Port)
	d.Set(names.AttrDatabaseName, endpoint.DatabaseName)
}

func findEndpointByID(ctx context.Context, conn *dms.Client, id string) (*awstypes.Endpoint, error) {
	input := &dms.DescribeEndpointsInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("endpoint-id"),
				Values: []string{id},
			},
		},
	}

	return findEndpoint(ctx, conn, input)
}

func findEndpoint(ctx context.Context, conn *dms.Client, input *dms.DescribeEndpointsInput) (*awstypes.Endpoint, error) {
	output, err := findEndpoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEndpoints(ctx context.Context, conn *dms.Client, input *dms.DescribeEndpointsInput) ([]awstypes.Endpoint, error) {
	var output []awstypes.Endpoint

	pages := dms.NewDescribeEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Endpoints...)
	}

	return output, nil
}

func statusEndpoint(ctx context.Context, conn *dms.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findEndpointByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitEndpointDeleted(ctx context.Context, conn *dms.Client, id string, timeout time.Duration) (*awstypes.Endpoint, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{endpointStatusDeleting},
		Target:  []string{},
		Refresh: statusEndpoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Endpoint); ok {
		return output, err
	}

	return nil, err
}
