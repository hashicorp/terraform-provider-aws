package dms

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceEndpointCreate,
		Read:   resourceEndpointRead,
		Update: resourceEndpointUpdate,
		Delete: resourceEndpointDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"certificate_arn": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"database_name": {
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
							// API returns this error with ModifyEndpoint:
							// InvalidParameterCombinationException: OpenSearch endpoint cant be modified.
							ForceNew: true,
						},
						"error_retry_duration": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      300,
							ValidateFunc: validation.IntAtLeast(0),
							// API returns this error with ModifyEndpoint:
							// InvalidParameterCombinationException: OpenSearch endpoint cant be modified.
							ForceNew: true,
						},
						"full_load_error_percentage": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntBetween(0, 100),
							// API returns this error with ModifyEndpoint:
							// InvalidParameterCombinationException: OpenSearch endpoint cant be modified.
							ForceNew: true,
						},
						"service_access_role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
							// API returns this error with ModifyEndpoint:
							// InvalidParameterCombinationException: OpenSearch endpoint cant be modified.
							ForceNew: true,
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
			"endpoint_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(dms.ReplicationEndpointTypeValue_Values(), false),
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
							Type:         schema.TypeString,
							Optional:     true,
							Default:      dms.MessageFormatValueJson,
							ValidateFunc: validation.StringInSlice(dms.MessageFormatValue_Values(), false),
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
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(dms.KafkaSecurityProtocol_Values(), false),
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
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      dms.MessageFormatValueJson,
							ValidateFunc: validation.StringInSlice(dms.MessageFormatValue_Values(), false),
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
						"stream_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"kms_key_arn": {
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
							Type:         schema.TypeString,
							Optional:     true,
							Default:      dms.AuthTypeValuePassword,
							ValidateFunc: validation.StringInSlice(dms.AuthTypeValue_Values(), false),
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
							Type:         schema.TypeString,
							Optional:     true,
							Default:      dms.NestingLevelValueNone,
							ValidateFunc: validation.StringInSlice(dms.NestingLevelValue_Values(), false),
						},
					},
				},
			},
			"password": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"secrets_manager_access_role_arn", "secrets_manager_arn"},
			},
			"port": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"secrets_manager_access_role_arn", "secrets_manager_arn"},
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
						"bucket_name": {
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
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
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
						"bucket_name": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
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
							Default:      dms.DatePartitionDelimiterValueSlash,
							ValidateFunc: validation.StringInSlice(dms.DatePartitionDelimiterValue_Values(), true),
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
							Type:         schema.TypeString,
							Optional:     true,
							Default:      dms.DatePartitionSequenceValueYyyymmdd,
							ValidateFunc: validation.StringInSlice(dms.DatePartitionSequenceValue_Values(), true),
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
							Type:         schema.TypeString,
							Optional:     true,
							Default:      dms.EncodingTypeValueRleDictionary,
							ValidateFunc: validation.StringInSlice(dms.EncodingTypeValue_Values(), false),
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
						"ignore_headers_row": {
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
						"server_side_encryption_kms_key_id": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
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
					},
				},
			},
			"secrets_manager_access_role_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  verify.ValidARN,
				RequiredWith:  []string{"secrets_manager_arn"},
				ConflictsWith: []string{"username", "password", "server_name", "port"},
			},
			"secrets_manager_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  verify.ValidARN,
				RequiredWith:  []string{"secrets_manager_access_role_arn"},
				ConflictsWith: []string{"username", "password", "server_name", "port"},
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
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(dms.DmsSslModeValue_Values(), false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"username": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"secrets_manager_access_role_arn", "secrets_manager_arn"},
			},
		},

		CustomizeDiff: customdiff.All(
			resourceEndpointCustomizeDiff,
			verify.SetTagsDiff,
		),
	}
}

func resourceEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DMSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	endpointID := d.Get("endpoint_id").(string)
	input := &dms.CreateEndpointInput{
		EndpointIdentifier: aws.String(endpointID),
		EndpointType:       aws.String(d.Get("endpoint_type").(string)),
		EngineName:         aws.String(d.Get("engine_name").(string)),
		Tags:               Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("certificate_arn"); ok {
		input.CertificateArn = aws.String(v.(string))
	}

	// Send ExtraConnectionAttributes in the API request for all resource types
	// per https://github.com/hashicorp/terraform-provider-aws/issues/8009
	if v, ok := d.GetOk("extra_connection_attributes"); ok {
		input.ExtraConnectionAttributes = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_arn"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ssl_mode"); ok {
		input.SslMode = aws.String(v.(string))
	}

	switch d.Get("engine_name").(string) {
	case engineNameAurora, engineNameMariadb, engineNameMySQL:
		if _, ok := d.GetOk("secrets_manager_arn"); ok {
			input.MySQLSettings = &dms.MySQLSettings{
				SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
				SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
			}
		} else {
			input.MySQLSettings = &dms.MySQLSettings{
				Username:     aws.String(d.Get("username").(string)),
				Password:     aws.String(d.Get("password").(string)),
				ServerName:   aws.String(d.Get("server_name").(string)),
				Port:         aws.Int64(int64(d.Get("port").(int))),
				DatabaseName: aws.String(d.Get("database_name").(string)),
			}

			// Set connection info in top-level namespace as well
			expandTopLevelConnectionInfo(d, input)
		}
	case engineNameAuroraPostgresql, engineNamePostgres:
		if _, ok := d.GetOk("secrets_manager_arn"); ok {
			input.PostgreSQLSettings = &dms.PostgreSQLSettings{
				SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
				SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
				DatabaseName:                aws.String(d.Get("database_name").(string)),
			}
		} else {
			input.PostgreSQLSettings = &dms.PostgreSQLSettings{
				Username:     aws.String(d.Get("username").(string)),
				Password:     aws.String(d.Get("password").(string)),
				ServerName:   aws.String(d.Get("server_name").(string)),
				Port:         aws.Int64(int64(d.Get("port").(int))),
				DatabaseName: aws.String(d.Get("database_name").(string)),
			}

			// Set connection info in top-level namespace as well
			expandTopLevelConnectionInfo(d, input)
		}
	case engineNameDynamoDB:
		input.DynamoDbSettings = &dms.DynamoDbSettings{
			ServiceAccessRoleArn: aws.String(d.Get("service_access_role").(string)),
		}
	case engineNameElasticsearch, engineNameOpenSearch:
		input.ElasticsearchSettings = &dms.ElasticsearchSettings{
			ServiceAccessRoleArn:    aws.String(d.Get("elasticsearch_settings.0.service_access_role_arn").(string)),
			EndpointUri:             aws.String(d.Get("elasticsearch_settings.0.endpoint_uri").(string)),
			ErrorRetryDuration:      aws.Int64(int64(d.Get("elasticsearch_settings.0.error_retry_duration").(int))),
			FullLoadErrorPercentage: aws.Int64(int64(d.Get("elasticsearch_settings.0.full_load_error_percentage").(int))),
		}
	case engineNameKafka:
		input.KafkaSettings = expandKafkaSettings(d.Get("kafka_settings").([]interface{})[0].(map[string]interface{}))
	case engineNameKinesis:
		input.KinesisSettings = expandKinesisSettings(d.Get("kinesis_settings").([]interface{})[0].(map[string]interface{}))
	case engineNameMongodb:
		var settings = &dms.MongoDbSettings{}

		if _, ok := d.GetOk("secrets_manager_arn"); ok {
			settings.SecretsManagerAccessRoleArn = aws.String(d.Get("secrets_manager_access_role_arn").(string))
			settings.SecretsManagerSecretId = aws.String(d.Get("secrets_manager_arn").(string))
		} else {
			settings.Username = aws.String(d.Get("username").(string))
			settings.Password = aws.String(d.Get("password").(string))
			settings.ServerName = aws.String(d.Get("server_name").(string))
			settings.Port = aws.Int64(int64(d.Get("port").(int)))

			// Set connection info in top-level namespace as well
			expandTopLevelConnectionInfo(d, input)
		}

		settings.DatabaseName = aws.String(d.Get("database_name").(string))
		settings.KmsKeyId = aws.String(d.Get("kms_key_arn").(string))
		settings.AuthType = aws.String(d.Get("mongodb_settings.0.auth_type").(string))
		settings.AuthMechanism = aws.String(d.Get("mongodb_settings.0.auth_mechanism").(string))
		settings.NestingLevel = aws.String(d.Get("mongodb_settings.0.nesting_level").(string))
		settings.ExtractDocId = aws.String(d.Get("mongodb_settings.0.extract_doc_id").(string))
		settings.DocsToInvestigate = aws.String(d.Get("mongodb_settings.0.docs_to_investigate").(string))
		settings.AuthSource = aws.String(d.Get("mongodb_settings.0.auth_source").(string))

		input.MongoDbSettings = settings
	case engineNameOracle:
		if _, ok := d.GetOk("secrets_manager_arn"); ok {
			input.OracleSettings = &dms.OracleSettings{
				SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
				SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
				DatabaseName:                aws.String(d.Get("database_name").(string)),
			}
		} else {
			input.OracleSettings = &dms.OracleSettings{
				Username:     aws.String(d.Get("username").(string)),
				Password:     aws.String(d.Get("password").(string)),
				ServerName:   aws.String(d.Get("server_name").(string)),
				Port:         aws.Int64(int64(d.Get("port").(int))),
				DatabaseName: aws.String(d.Get("database_name").(string)),
			}

			// Set connection info in top-level namespace as well
			expandTopLevelConnectionInfo(d, input)
		}
	case engineNameRedshift:
		var settings = &dms.RedshiftSettings{
			DatabaseName: aws.String(d.Get("database_name").(string)),
		}

		if _, ok := d.GetOk("secrets_manager_arn"); ok {
			settings.SecretsManagerAccessRoleArn = aws.String(d.Get("secrets_manager_access_role_arn").(string))
			settings.SecretsManagerSecretId = aws.String(d.Get("secrets_manager_arn").(string))
		} else {
			settings.Username = aws.String(d.Get("username").(string))
			settings.Password = aws.String(d.Get("password").(string))
			settings.ServerName = aws.String(d.Get("server_name").(string))
			settings.Port = aws.Int64(int64(d.Get("port").(int)))

			// Set connection info in top-level namespace as well
			expandTopLevelConnectionInfo(d, input)
		}

		if v, ok := d.GetOk("redshift_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})

			if v, ok := tfMap["bucket_folder"].(string); ok && v != "" {
				settings.BucketFolder = aws.String(v)
			}

			if v, ok := tfMap["bucket_name"].(string); ok && v != "" {
				settings.BucketName = aws.String(v)
			}

			if v, ok := tfMap["encryption_mode"].(string); ok && v != "" {
				settings.EncryptionMode = aws.String(v)
			}

			if v, ok := tfMap["server_side_encryption_kms_key_id"].(string); ok && v != "" {
				settings.ServerSideEncryptionKmsKeyId = aws.String(v)
			}

			if v, ok := tfMap["service_access_role_arn"].(string); ok && v != "" {
				settings.ServiceAccessRoleArn = aws.String(v)
			}
		}

		input.RedshiftSettings = settings
	case engineNameSQLServer:
		if _, ok := d.GetOk("secrets_manager_arn"); ok {
			input.MicrosoftSQLServerSettings = &dms.MicrosoftSQLServerSettings{
				SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
				SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
				DatabaseName:                aws.String(d.Get("database_name").(string)),
			}
		} else {
			input.MicrosoftSQLServerSettings = &dms.MicrosoftSQLServerSettings{
				Username:     aws.String(d.Get("username").(string)),
				Password:     aws.String(d.Get("password").(string)),
				ServerName:   aws.String(d.Get("server_name").(string)),
				Port:         aws.Int64(int64(d.Get("port").(int))),
				DatabaseName: aws.String(d.Get("database_name").(string)),
			}

			// Set connection info in top-level namespace as well
			expandTopLevelConnectionInfo(d, input)
		}
	case engineNameS3:
		input.S3Settings = expandS3Settings(d.Get("s3_settings").([]interface{})[0].(map[string]interface{}))
	default:
		expandTopLevelConnectionInfo(d, input)
	}

	log.Printf("[DEBUG] Creating DMS Endpoint: %s", input)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(d.Timeout(schema.TimeoutCreate),
		func() (interface{}, error) {
			return conn.CreateEndpoint(input)
		},
		dms.ErrCodeAccessDeniedFault)

	if err != nil {
		return fmt.Errorf("creating DMS Endpoint (%s): %w", endpointID, err)
	}

	d.SetId(endpointID)

	return resourceEndpointRead(d, meta)
}

func resourceEndpointRead(d *schema.ResourceData, meta interface{}) error {
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
		return fmt.Errorf("reading DMS Endpoint (%s): %w", d.Id(), err)
	}

	err = resourceEndpointSetState(d, endpoint)

	if err != nil {
		return err
	}

	tags, err := ListTags(conn, d.Get("endpoint_arn").(string))

	if err != nil {
		return fmt.Errorf("listing tags for DMS Endpoint (%s): %w", d.Get("endpoint_arn").(string), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DMSConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &dms.ModifyEndpointInput{
			EndpointArn: aws.String(d.Get("endpoint_arn").(string)),
		}

		if d.HasChange("certificate_arn") {
			input.CertificateArn = aws.String(d.Get("certificate_arn").(string))
		}

		if d.HasChange("endpoint_type") {
			input.EndpointType = aws.String(d.Get("endpoint_type").(string))
		}

		if d.HasChange("engine_name") {
			input.EngineName = aws.String(d.Get("engine_name").(string))
		}

		if d.HasChange("extra_connection_attributes") {
			input.ExtraConnectionAttributes = aws.String(d.Get("extra_connection_attributes").(string))
		}

		if d.HasChange("service_access_role") {
			input.DynamoDbSettings = &dms.DynamoDbSettings{
				ServiceAccessRoleArn: aws.String(d.Get("service_access_role").(string)),
			}
		}

		if d.HasChange("ssl_mode") {
			input.SslMode = aws.String(d.Get("ssl_mode").(string))
		}

		switch engineName := d.Get("engine_name").(string); engineName {
		case engineNameAurora, engineNameMariadb, engineNameMySQL:
			if d.HasChanges(
				"username", "password", "server_name", "port", "database_name", "secrets_manager_access_role_arn",
				"secrets_manager_arn") {
				if _, ok := d.GetOk("secrets_manager_arn"); ok {
					input.MySQLSettings = &dms.MySQLSettings{
						SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
						SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
					}
				} else {
					input.MySQLSettings = &dms.MySQLSettings{
						Username:     aws.String(d.Get("username").(string)),
						Password:     aws.String(d.Get("password").(string)),
						ServerName:   aws.String(d.Get("server_name").(string)),
						Port:         aws.Int64(int64(d.Get("port").(int))),
						DatabaseName: aws.String(d.Get("database_name").(string)),
					}
					input.EngineName = aws.String(engineName)

					// Update connection info in top-level namespace as well
					expandTopLevelConnectionInfoModify(d, input)
				}
			}
		case engineNameAuroraPostgresql, engineNamePostgres:
			if d.HasChanges(
				"username", "password", "server_name", "port", "database_name", "secrets_manager_access_role_arn",
				"secrets_manager_arn") {
				if _, ok := d.GetOk("secrets_manager_arn"); ok {
					input.PostgreSQLSettings = &dms.PostgreSQLSettings{
						DatabaseName:                aws.String(d.Get("database_name").(string)),
						SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
						SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
					}
				} else {
					input.PostgreSQLSettings = &dms.PostgreSQLSettings{
						Username:     aws.String(d.Get("username").(string)),
						Password:     aws.String(d.Get("password").(string)),
						ServerName:   aws.String(d.Get("server_name").(string)),
						Port:         aws.Int64(int64(d.Get("port").(int))),
						DatabaseName: aws.String(d.Get("database_name").(string)),
					}
					input.EngineName = aws.String(engineName) // Must be included (should be 'postgres')

					// Update connection info in top-level namespace as well
					expandTopLevelConnectionInfoModify(d, input)
				}
			}
		case engineNameDynamoDB:
			if d.HasChange("service_access_role") {
				input.DynamoDbSettings = &dms.DynamoDbSettings{
					ServiceAccessRoleArn: aws.String(d.Get("service_access_role").(string)),
				}
			}
		case engineNameElasticsearch, engineNameOpenSearch:
			if d.HasChanges(
				"elasticsearch_settings.0.endpoint_uri",
				"elasticsearch_settings.0.error_retry_duration",
				"elasticsearch_settings.0.full_load_error_percentage",
				"elasticsearch_settings.0.service_access_role_arn") {
				input.ElasticsearchSettings = &dms.ElasticsearchSettings{
					ServiceAccessRoleArn:    aws.String(d.Get("elasticsearch_settings.0.service_access_role_arn").(string)),
					EndpointUri:             aws.String(d.Get("elasticsearch_settings.0.endpoint_uri").(string)),
					ErrorRetryDuration:      aws.Int64(int64(d.Get("elasticsearch_settings.0.error_retry_duration").(int))),
					FullLoadErrorPercentage: aws.Int64(int64(d.Get("elasticsearch_settings.0.full_load_error_percentage").(int))),
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
				"username", "password", "server_name", "port", "database_name", "mongodb_settings.0.auth_type",
				"mongodb_settings.0.auth_mechanism", "mongodb_settings.0.nesting_level", "mongodb_settings.0.extract_doc_id",
				"mongodb_settings.0.docs_to_investigate", "mongodb_settings.0.auth_source", "secrets_manager_access_role_arn",
				"secrets_manager_arn") {
				if _, ok := d.GetOk("secrets_manager_arn"); ok {
					input.MongoDbSettings = &dms.MongoDbSettings{
						SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
						SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
						DatabaseName:                aws.String(d.Get("database_name").(string)),
						KmsKeyId:                    aws.String(d.Get("kms_key_arn").(string)),

						AuthType:          aws.String(d.Get("mongodb_settings.0.auth_type").(string)),
						AuthMechanism:     aws.String(d.Get("mongodb_settings.0.auth_mechanism").(string)),
						NestingLevel:      aws.String(d.Get("mongodb_settings.0.nesting_level").(string)),
						ExtractDocId:      aws.String(d.Get("mongodb_settings.0.extract_doc_id").(string)),
						DocsToInvestigate: aws.String(d.Get("mongodb_settings.0.docs_to_investigate").(string)),
						AuthSource:        aws.String(d.Get("mongodb_settings.0.auth_source").(string)),
					}
				} else {
					input.MongoDbSettings = &dms.MongoDbSettings{
						Username:     aws.String(d.Get("username").(string)),
						Password:     aws.String(d.Get("password").(string)),
						ServerName:   aws.String(d.Get("server_name").(string)),
						Port:         aws.Int64(int64(d.Get("port").(int))),
						DatabaseName: aws.String(d.Get("database_name").(string)),
						KmsKeyId:     aws.String(d.Get("kms_key_arn").(string)),

						AuthType:          aws.String(d.Get("mongodb_settings.0.auth_type").(string)),
						AuthMechanism:     aws.String(d.Get("mongodb_settings.0.auth_mechanism").(string)),
						NestingLevel:      aws.String(d.Get("mongodb_settings.0.nesting_level").(string)),
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
				"username", "password", "server_name", "port", "database_name", "secrets_manager_access_role_arn",
				"secrets_manager_arn") {
				if _, ok := d.GetOk("secrets_manager_arn"); ok {
					input.OracleSettings = &dms.OracleSettings{
						DatabaseName:                aws.String(d.Get("database_name").(string)),
						SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
						SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
					}
				} else {
					input.OracleSettings = &dms.OracleSettings{
						Username:     aws.String(d.Get("username").(string)),
						Password:     aws.String(d.Get("password").(string)),
						ServerName:   aws.String(d.Get("server_name").(string)),
						Port:         aws.Int64(int64(d.Get("port").(int))),
						DatabaseName: aws.String(d.Get("database_name").(string)),
					}
					input.EngineName = aws.String(engineName) // Must be included (should be 'oracle')

					// Update connection info in top-level namespace as well
					expandTopLevelConnectionInfoModify(d, input)
				}
			}
		case engineNameRedshift:
			if d.HasChanges(
				"username", "password", "server_name", "port", "database_name",
				"redshift_settings", "secrets_manager_access_role_arn",
				"secrets_manager_arn") {
				if _, ok := d.GetOk("secrets_manager_arn"); ok {
					input.RedshiftSettings = &dms.RedshiftSettings{
						DatabaseName:                aws.String(d.Get("database_name").(string)),
						SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
						SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
					}
				} else {
					input.RedshiftSettings = &dms.RedshiftSettings{
						Username:     aws.String(d.Get("username").(string)),
						Password:     aws.String(d.Get("password").(string)),
						ServerName:   aws.String(d.Get("server_name").(string)),
						Port:         aws.Int64(int64(d.Get("port").(int))),
						DatabaseName: aws.String(d.Get("database_name").(string)),
					}
					input.EngineName = aws.String(engineName) // Must be included (should be 'redshift')

					// Update connection info in top-level namespace as well
					expandTopLevelConnectionInfoModify(d, input)

					if v, ok := d.GetOk("redshift_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
						tfMap := v.([]interface{})[0].(map[string]interface{})

						if v, ok := tfMap["bucket_folder"].(string); ok && v != "" {
							input.RedshiftSettings.BucketFolder = aws.String(v)
						}

						if v, ok := tfMap["bucket_name"].(string); ok && v != "" {
							input.RedshiftSettings.BucketName = aws.String(v)
						}

						if v, ok := tfMap["encryption_mode"].(string); ok && v != "" {
							input.RedshiftSettings.EncryptionMode = aws.String(v)
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
		case engineNameSQLServer:
			if d.HasChanges(
				"username", "password", "server_name", "port", "database_name", "secrets_manager_access_role_arn",
				"secrets_manager_arn") {
				if _, ok := d.GetOk("secrets_manager_arn"); ok {
					input.MicrosoftSQLServerSettings = &dms.MicrosoftSQLServerSettings{
						DatabaseName:                aws.String(d.Get("database_name").(string)),
						SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
						SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
					}
				} else {
					input.MicrosoftSQLServerSettings = &dms.MicrosoftSQLServerSettings{
						Username:     aws.String(d.Get("username").(string)),
						Password:     aws.String(d.Get("password").(string)),
						ServerName:   aws.String(d.Get("server_name").(string)),
						Port:         aws.Int64(int64(d.Get("port").(int))),
						DatabaseName: aws.String(d.Get("database_name").(string)),
					}
					input.EngineName = aws.String(engineName) // Must be included (should be 'postgres')

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
			if d.HasChange("database_name") {
				input.DatabaseName = aws.String(d.Get("database_name").(string))
			}

			if d.HasChange("password") {
				input.Password = aws.String(d.Get("password").(string))
			}

			if d.HasChange("port") {
				input.Port = aws.Int64(int64(d.Get("port").(int)))
			}

			if d.HasChange("server_name") {
				input.ServerName = aws.String(d.Get("server_name").(string))
			}

			if d.HasChange("username") {
				input.Username = aws.String(d.Get("username").(string))
			}
		}

		log.Printf("[DEBUG] Modifying DMS Endpoint: %s", input)
		_, err := conn.ModifyEndpoint(input)

		if err != nil {
			return fmt.Errorf("updating DMS Endpoint (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		arn := d.Get("endpoint_arn").(string)
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("updating DMS Endpoint (%s) tags: %w", arn, err)
		}
	}

	return resourceEndpointRead(d, meta)
}

func resourceEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DMSConn

	log.Printf("[DEBUG] Deleting DMS Endpoint: (%s)", d.Id())
	_, err := conn.DeleteEndpoint(&dms.DeleteEndpointInput{
		EndpointArn: aws.String(d.Get("endpoint_arn").(string)),
	})

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting DMS Endpoint (%s): %w", d.Id(), err)
	}

	if _, err = waitEndpointDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("waiting for DMS Endpoint (%s) delete: %w", d.Id(), err)
	}

	return err
}

func resourceEndpointCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
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
	case engineNameS3:
		if v, ok := diff.GetOk("s3_settings"); !ok || len(v.([]interface{})) == 0 || v.([]interface{})[0] == nil {
			return fmt.Errorf("s3_settings must be set when engine_name = %q", engineName)
		}
	}

	return nil
}

func resourceEndpointSetState(d *schema.ResourceData, endpoint *dms.Endpoint) error {
	d.SetId(aws.StringValue(endpoint.EndpointIdentifier))

	d.Set("certificate_arn", endpoint.CertificateArn)
	d.Set("endpoint_arn", endpoint.EndpointArn)
	d.Set("endpoint_id", endpoint.EndpointIdentifier)
	// For some reason the AWS API only accepts lowercase type but returns it as uppercase
	d.Set("endpoint_type", strings.ToLower(*endpoint.EndpointType))
	d.Set("engine_name", endpoint.EngineName)
	d.Set("extra_connection_attributes", endpoint.ExtraConnectionAttributes)

	switch aws.StringValue(endpoint.EngineName) {
	case engineNameAurora, engineNameMariadb, engineNameMySQL:
		if endpoint.MySQLSettings != nil {
			d.Set("username", endpoint.MySQLSettings.Username)
			d.Set("server_name", endpoint.MySQLSettings.ServerName)
			d.Set("port", endpoint.MySQLSettings.Port)
			d.Set("database_name", endpoint.MySQLSettings.DatabaseName)
			d.Set("secrets_manager_access_role_arn", endpoint.MySQLSettings.SecretsManagerAccessRoleArn)
			d.Set("secrets_manager_arn", endpoint.MySQLSettings.SecretsManagerSecretId)
		} else {
			flattenTopLevelConnectionInfo(d, endpoint)
		}
	case engineNameAuroraPostgresql, engineNamePostgres:
		if endpoint.PostgreSQLSettings != nil {
			d.Set("username", endpoint.PostgreSQLSettings.Username)
			d.Set("server_name", endpoint.PostgreSQLSettings.ServerName)
			d.Set("port", endpoint.PostgreSQLSettings.Port)
			d.Set("database_name", endpoint.PostgreSQLSettings.DatabaseName)
			d.Set("secrets_manager_access_role_arn", endpoint.PostgreSQLSettings.SecretsManagerAccessRoleArn)
			d.Set("secrets_manager_arn", endpoint.PostgreSQLSettings.SecretsManagerSecretId)
		} else {
			flattenTopLevelConnectionInfo(d, endpoint)
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
			d.Set("username", endpoint.MongoDbSettings.Username)
			d.Set("server_name", endpoint.MongoDbSettings.ServerName)
			d.Set("port", endpoint.MongoDbSettings.Port)
			d.Set("database_name", endpoint.MongoDbSettings.DatabaseName)
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
			d.Set("username", endpoint.OracleSettings.Username)
			d.Set("server_name", endpoint.OracleSettings.ServerName)
			d.Set("port", endpoint.OracleSettings.Port)
			d.Set("database_name", endpoint.OracleSettings.DatabaseName)
			d.Set("secrets_manager_access_role_arn", endpoint.OracleSettings.SecretsManagerAccessRoleArn)
			d.Set("secrets_manager_arn", endpoint.OracleSettings.SecretsManagerSecretId)
		} else {
			flattenTopLevelConnectionInfo(d, endpoint)
		}
	case engineNameRedshift:
		if endpoint.RedshiftSettings != nil {
			d.Set("username", endpoint.RedshiftSettings.Username)
			d.Set("server_name", endpoint.RedshiftSettings.ServerName)
			d.Set("port", endpoint.RedshiftSettings.Port)
			d.Set("database_name", endpoint.RedshiftSettings.DatabaseName)
			d.Set("secrets_manager_access_role_arn", endpoint.RedshiftSettings.SecretsManagerAccessRoleArn)
			d.Set("secrets_manager_arn", endpoint.RedshiftSettings.SecretsManagerSecretId)
		} else {
			flattenTopLevelConnectionInfo(d, endpoint)
		}
		if err := d.Set("redshift_settings", flattenRedshiftSettings(endpoint.RedshiftSettings)); err != nil {
			return fmt.Errorf("setting redshift_settings: %w", err)
		}
	case engineNameSQLServer:
		if endpoint.MicrosoftSQLServerSettings != nil {
			d.Set("username", endpoint.MicrosoftSQLServerSettings.Username)
			d.Set("server_name", endpoint.MicrosoftSQLServerSettings.ServerName)
			d.Set("port", endpoint.MicrosoftSQLServerSettings.Port)
			d.Set("database_name", endpoint.MicrosoftSQLServerSettings.DatabaseName)
			d.Set("secrets_manager_access_role_arn", endpoint.MicrosoftSQLServerSettings.SecretsManagerAccessRoleArn)
			d.Set("secrets_manager_arn", endpoint.MicrosoftSQLServerSettings.SecretsManagerSecretId)
		} else {
			flattenTopLevelConnectionInfo(d, endpoint)
		}
	case engineNameS3:
		if err := d.Set("s3_settings", flattenS3Settings(endpoint.S3Settings)); err != nil {
			return fmt.Errorf("Error setting s3_settings for DMS: %s", err)
		}
	default:
		d.Set("database_name", endpoint.DatabaseName)
		d.Set("extra_connection_attributes", endpoint.ExtraConnectionAttributes)
		d.Set("port", endpoint.Port)
		d.Set("server_name", endpoint.ServerName)
		d.Set("username", endpoint.Username)
	}

	d.Set("kms_key_arn", endpoint.KmsKeyId)
	d.Set("ssl_mode", endpoint.SslMode)

	return nil
}

func flattenOpenSearchSettings(settings *dms.ElasticsearchSettings) []map[string]interface{} {
	if settings == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"endpoint_uri":               aws.StringValue(settings.EndpointUri),
		"error_retry_duration":       aws.Int64Value(settings.ErrorRetryDuration),
		"full_load_error_percentage": aws.Int64Value(settings.FullLoadErrorPercentage),
		"service_access_role_arn":    aws.StringValue(settings.ServiceAccessRoleArn),
	}

	return []map[string]interface{}{m}
}

func expandKafkaSettings(tfMap map[string]interface{}) *dms.KafkaSettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &dms.KafkaSettings{}

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
		apiObject.MessageFormat = aws.String(v)
	}

	if v, ok := tfMap["message_max_bytes"].(int); ok && v != 0 {
		apiObject.MessageMaxBytes = aws.Int64(int64(v))
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
		apiObject.SecurityProtocol = aws.String(v)
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

func flattenKafkaSettings(apiObject *dms.KafkaSettings) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Broker; v != nil {
		tfMap["broker"] = aws.StringValue(v)
	}

	if v := apiObject.IncludeControlDetails; v != nil {
		tfMap["include_control_details"] = aws.BoolValue(v)
	}

	if v := apiObject.IncludeNullAndEmpty; v != nil {
		tfMap["include_null_and_empty"] = aws.BoolValue(v)
	}

	if v := apiObject.IncludePartitionValue; v != nil {
		tfMap["include_partition_value"] = aws.BoolValue(v)
	}

	if v := apiObject.IncludeTableAlterOperations; v != nil {
		tfMap["include_table_alter_operations"] = aws.BoolValue(v)
	}

	if v := apiObject.IncludeTransactionDetails; v != nil {
		tfMap["include_transaction_details"] = aws.BoolValue(v)
	}

	if v := apiObject.MessageFormat; v != nil {
		tfMap["message_format"] = aws.StringValue(v)
	}

	if v := apiObject.MessageMaxBytes; v != nil {
		tfMap["message_max_bytes"] = aws.Int64Value(v)
	}

	if v := apiObject.NoHexPrefix; v != nil {
		tfMap["no_hex_prefix"] = aws.BoolValue(v)
	}

	if v := apiObject.PartitionIncludeSchemaTable; v != nil {
		tfMap["partition_include_schema_table"] = aws.BoolValue(v)
	}

	if v := apiObject.SaslPassword; v != nil {
		tfMap["sasl_password"] = aws.StringValue(v)
	}

	if v := apiObject.SaslUsername; v != nil {
		tfMap["sasl_username"] = aws.StringValue(v)
	}

	if v := apiObject.SecurityProtocol; v != nil {
		tfMap["security_protocol"] = aws.StringValue(v)
	}

	if v := apiObject.SslCaCertificateArn; v != nil {
		tfMap["ssl_ca_certificate_arn"] = aws.StringValue(v)
	}

	if v := apiObject.SslClientCertificateArn; v != nil {
		tfMap["ssl_client_certificate_arn"] = aws.StringValue(v)
	}

	if v := apiObject.SslClientKeyArn; v != nil {
		tfMap["ssl_client_key_arn"] = aws.StringValue(v)
	}

	if v := apiObject.SslClientKeyPassword; v != nil {
		tfMap["ssl_client_key_password"] = aws.StringValue(v)
	}

	if v := apiObject.Topic; v != nil {
		tfMap["topic"] = aws.StringValue(v)
	}

	return tfMap
}

func expandKinesisSettings(tfMap map[string]interface{}) *dms.KinesisSettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &dms.KinesisSettings{}

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
		apiObject.MessageFormat = aws.String(v)
	}

	if v, ok := tfMap["partition_include_schema_table"].(bool); ok {
		apiObject.PartitionIncludeSchemaTable = aws.Bool(v)
	}

	if v, ok := tfMap["service_access_role_arn"].(string); ok && v != "" {
		apiObject.ServiceAccessRoleArn = aws.String(v)
	}

	if v, ok := tfMap["stream_arn"].(string); ok && v != "" {
		apiObject.StreamArn = aws.String(v)
	}

	return apiObject
}

func flattenKinesisSettings(apiObject *dms.KinesisSettings) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.IncludeControlDetails; v != nil {
		tfMap["include_control_details"] = aws.BoolValue(v)
	}

	if v := apiObject.IncludeNullAndEmpty; v != nil {
		tfMap["include_null_and_empty"] = aws.BoolValue(v)
	}

	if v := apiObject.IncludePartitionValue; v != nil {
		tfMap["include_partition_value"] = aws.BoolValue(v)
	}

	if v := apiObject.IncludeTableAlterOperations; v != nil {
		tfMap["include_table_alter_operations"] = aws.BoolValue(v)
	}

	if v := apiObject.IncludeTransactionDetails; v != nil {
		tfMap["include_transaction_details"] = aws.BoolValue(v)
	}

	if v := apiObject.MessageFormat; v != nil {
		tfMap["message_format"] = aws.StringValue(v)
	}

	if v := apiObject.PartitionIncludeSchemaTable; v != nil {
		tfMap["partition_include_schema_table"] = aws.BoolValue(v)
	}

	if v := apiObject.ServiceAccessRoleArn; v != nil {
		tfMap["service_access_role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.StreamArn; v != nil {
		tfMap["stream_arn"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenMongoDBSettings(settings *dms.MongoDbSettings) []map[string]interface{} {
	if settings == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"auth_type":           aws.StringValue(settings.AuthType),
		"auth_mechanism":      aws.StringValue(settings.AuthMechanism),
		"nesting_level":       aws.StringValue(settings.NestingLevel),
		"extract_doc_id":      aws.StringValue(settings.ExtractDocId),
		"docs_to_investigate": aws.StringValue(settings.DocsToInvestigate),
		"auth_source":         aws.StringValue(settings.AuthSource),
	}

	return []map[string]interface{}{m}
}

func flattenRedshiftSettings(settings *dms.RedshiftSettings) []map[string]interface{} {
	if settings == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"bucket_folder":                     aws.StringValue(settings.BucketFolder),
		"bucket_name":                       aws.StringValue(settings.BucketName),
		"encryption_mode":                   aws.StringValue(settings.EncryptionMode),
		"server_side_encryption_kms_key_id": aws.StringValue(settings.ServerSideEncryptionKmsKeyId),
		"service_access_role_arn":           aws.StringValue(settings.ServiceAccessRoleArn),
	}

	return []map[string]interface{}{m}
}

func expandS3Settings(tfMap map[string]interface{}) *dms.S3Settings {
	if tfMap == nil {
		return nil
	}

	apiObject := &dms.S3Settings{}

	if v, ok := tfMap["add_column_name"].(bool); ok {
		apiObject.AddColumnName = aws.Bool(v)
	}
	if v, ok := tfMap["bucket_folder"].(string); ok {
		apiObject.BucketFolder = aws.String(v)
	}
	if v, ok := tfMap["bucket_name"].(string); ok {
		apiObject.BucketName = aws.String(v)
	}
	if v, ok := tfMap["canned_acl_for_objects"].(string); ok {
		apiObject.CannedAclForObjects = aws.String(v)
	}
	if v, ok := tfMap["cdc_inserts_and_updates"].(bool); ok {
		apiObject.CdcInsertsAndUpdates = aws.Bool(v)
	}
	if v, ok := tfMap["cdc_inserts_only"].(bool); ok {
		apiObject.CdcInsertsOnly = aws.Bool(v)
	}
	if v, ok := tfMap["cdc_max_batch_interval"].(int); ok {
		apiObject.CdcMaxBatchInterval = aws.Int64(int64(v))
	}
	if v, ok := tfMap["cdc_min_file_size"].(int); ok {
		apiObject.CdcMinFileSize = aws.Int64(int64(v))
	}
	if v, ok := tfMap["cdc_path"].(string); ok {
		apiObject.CdcPath = aws.String(v)
	}
	if v, ok := tfMap["compression_type"].(string); ok {
		apiObject.CompressionType = aws.String(v)
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
		apiObject.DataFormat = aws.String(v)
	}
	if v, ok := tfMap["data_page_size"].(int); ok {
		apiObject.DataPageSize = aws.Int64(int64(v))
	}
	if v, ok := tfMap["date_partition_delimiter"].(string); ok {
		apiObject.DatePartitionDelimiter = aws.String(v)
	}
	if v, ok := tfMap["date_partition_enabled"].(bool); ok {
		apiObject.DatePartitionEnabled = aws.Bool(v)
	}
	if v, ok := tfMap["date_partition_sequence"].(string); ok {
		apiObject.DatePartitionSequence = aws.String(v)
	}
	if v, ok := tfMap["dict_page_size_limit"].(int); ok {
		apiObject.DictPageSizeLimit = aws.Int64(int64(v))
	}
	if v, ok := tfMap["enable_statistics"].(bool); ok {
		apiObject.EnableStatistics = aws.Bool(v)
	}
	if v, ok := tfMap["encoding_type"].(string); ok {
		apiObject.EncodingType = aws.String(v)
	}
	if v, ok := tfMap["encryption_mode"].(string); ok {
		apiObject.EncryptionMode = aws.String(v)
	}
	if v, ok := tfMap["external_table_definition"].(string); ok {
		apiObject.ExternalTableDefinition = aws.String(v)
	}
	if v, ok := tfMap["ignore_header_rows"].(int); ok {
		apiObject.IgnoreHeaderRows = aws.Int64(int64(v))
	}
	if v, ok := tfMap["include_op_for_full_load"].(bool); ok {
		apiObject.IncludeOpForFullLoad = aws.Bool(v)
	}
	if v, ok := tfMap["max_file_size"].(int); ok {
		apiObject.MaxFileSize = aws.Int64(int64(v))
	}
	if v, ok := tfMap["parquet_timestamp_in_millisecond"].(bool); ok {
		apiObject.ParquetTimestampInMillisecond = aws.Bool(v)
	}
	if v, ok := tfMap["parquet_version"].(string); ok {
		apiObject.ParquetVersion = aws.String(v)
	}
	if v, ok := tfMap["preserve_transactions"].(bool); ok {
		apiObject.PreserveTransactions = aws.Bool(v)
	}
	if v, ok := tfMap["rfc_4180"].(bool); ok {
		apiObject.Rfc4180 = aws.Bool(v)
	}
	if v, ok := tfMap["row_group_length"].(int); ok {
		apiObject.RowGroupLength = aws.Int64(int64(v))
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

	return apiObject
}

func flattenS3Settings(apiObject *dms.S3Settings) []map[string]interface{} {
	if apiObject == nil {
		return []map[string]interface{}{}
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AddColumnName; v != nil {
		tfMap["add_column_name"] = aws.BoolValue(v)
	}
	if v := apiObject.BucketFolder; v != nil {
		tfMap["bucket_folder"] = aws.StringValue(v)
	}
	if v := apiObject.BucketName; v != nil {
		tfMap["bucket_name"] = aws.StringValue(v)
	}
	if v := apiObject.CannedAclForObjects; v != nil {
		tfMap["canned_acl_for_objects"] = aws.StringValue(v)
	}
	if v := apiObject.CdcInsertsAndUpdates; v != nil {
		tfMap["cdc_inserts_and_updates"] = aws.BoolValue(v)
	}
	if v := apiObject.CdcInsertsOnly; v != nil {
		tfMap["cdc_inserts_only"] = aws.BoolValue(v)
	}
	if v := apiObject.CdcMaxBatchInterval; v != nil {
		tfMap["cdc_max_batch_interval"] = aws.Int64Value(v)
	}
	if v := apiObject.CdcMinFileSize; v != nil {
		tfMap["cdc_min_file_size"] = aws.Int64Value(v)
	}
	if v := apiObject.CdcPath; v != nil {
		tfMap["cdc_path"] = aws.StringValue(v)
	}
	if v := apiObject.CompressionType; v != nil {
		tfMap["compression_type"] = aws.StringValue(v)
	}
	if v := apiObject.CsvDelimiter; v != nil {
		tfMap["csv_delimiter"] = aws.StringValue(v)
	}
	if v := apiObject.CsvNoSupValue; v != nil {
		tfMap["csv_no_sup_value"] = aws.StringValue(v)
	}
	if v := apiObject.CsvNullValue; v != nil {
		tfMap["csv_null_value"] = aws.StringValue(v)
	}
	if v := apiObject.CsvRowDelimiter; v != nil {
		tfMap["csv_row_delimiter"] = aws.StringValue(v)
	}
	if v := apiObject.DataFormat; v != nil {
		tfMap["data_format"] = aws.StringValue(v)
	}
	if v := apiObject.DataPageSize; v != nil {
		tfMap["data_page_size"] = aws.Int64Value(v)
	}
	if v := apiObject.DatePartitionDelimiter; v != nil {
		tfMap["date_partition_delimiter"] = aws.StringValue(v)
	}
	if v := apiObject.DatePartitionEnabled; v != nil {
		tfMap["date_partition_enabled"] = aws.BoolValue(v)
	}
	if v := apiObject.DatePartitionSequence; v != nil {
		tfMap["date_partition_sequence"] = aws.StringValue(v)
	}
	if v := apiObject.DictPageSizeLimit; v != nil {
		tfMap["dict_page_size_limit"] = aws.Int64Value(v)
	}
	if v := apiObject.EnableStatistics; v != nil {
		tfMap["enable_statistics"] = aws.BoolValue(v)
	}
	if v := apiObject.EncodingType; v != nil {
		tfMap["encoding_type"] = aws.StringValue(v)
	}
	if v := apiObject.EncryptionMode; v != nil {
		tfMap["encryption_mode"] = aws.StringValue(v)
	}
	if v := apiObject.ExternalTableDefinition; v != nil {
		tfMap["external_table_definition"] = aws.StringValue(v)
	}
	if v := apiObject.IgnoreHeaderRows; v != nil {
		tfMap["ignore_header_rows"] = aws.Int64Value(v)
	}
	if v := apiObject.IncludeOpForFullLoad; v != nil {
		tfMap["include_op_for_full_load"] = aws.BoolValue(v)
	}
	if v := apiObject.MaxFileSize; v != nil {
		tfMap["max_file_size"] = aws.Int64Value(v)
	}
	if v := apiObject.ParquetTimestampInMillisecond; v != nil {
		tfMap["parquet_timestamp_in_millisecond"] = aws.BoolValue(v)
	}
	if v := apiObject.ParquetVersion; v != nil {
		tfMap["parquet_version"] = aws.StringValue(v)
	}
	if v := apiObject.Rfc4180; v != nil {
		tfMap["rfc_4180"] = aws.BoolValue(v)
	}
	if v := apiObject.RowGroupLength; v != nil {
		tfMap["row_group_length"] = aws.Int64Value(v)
	}
	if v := apiObject.ServerSideEncryptionKmsKeyId; v != nil {
		tfMap["server_side_encryption_kms_key_id"] = aws.StringValue(v)
	}
	if v := apiObject.ServiceAccessRoleArn; v != nil {
		tfMap["service_access_role_arn"] = aws.StringValue(v)
	}
	if v := apiObject.TimestampColumnName; v != nil {
		tfMap["timestamp_column_name"] = aws.StringValue(v)
	}
	if v := apiObject.UseCsvNoSupValue; v != nil {
		tfMap["use_csv_no_sup_value"] = aws.BoolValue(v)
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

			return diff.Len() == 0 || diff.Equal(n)
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
		matchAllCap := regexp.MustCompile("([a-z])([A-Z])")
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
	input.Username = aws.String(d.Get("username").(string))
	input.Password = aws.String(d.Get("password").(string))
	input.ServerName = aws.String(d.Get("server_name").(string))
	input.Port = aws.Int64(int64(d.Get("port").(int)))

	if v, ok := d.GetOk("database_name"); ok {
		input.DatabaseName = aws.String(v.(string))
	}
}

func expandTopLevelConnectionInfoModify(d *schema.ResourceData, input *dms.ModifyEndpointInput) {
	input.Username = aws.String(d.Get("username").(string))
	input.Password = aws.String(d.Get("password").(string))
	input.ServerName = aws.String(d.Get("server_name").(string))
	input.Port = aws.Int64(int64(d.Get("port").(int)))

	if v, ok := d.GetOk("database_name"); ok {
		input.DatabaseName = aws.String(v.(string))
	}
}

func flattenTopLevelConnectionInfo(d *schema.ResourceData, endpoint *dms.Endpoint) {
	d.Set("username", endpoint.Username)
	d.Set("server_name", endpoint.ServerName)
	d.Set("port", endpoint.Port)
	d.Set("database_name", endpoint.DatabaseName)
}
