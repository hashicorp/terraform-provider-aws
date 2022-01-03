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
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
							// InvalidParameterCombinationException: Elasticsearch endpoint cant be modified.
							ForceNew: true,
						},
						"error_retry_duration": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      300,
							ValidateFunc: validation.IntAtLeast(0),
							// API returns this error with ModifyEndpoint:
							// InvalidParameterCombinationException: Elasticsearch endpoint cant be modified.
							ForceNew: true,
						},
						"full_load_error_percentage": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntBetween(0, 100),
							// API returns this error with ModifyEndpoint:
							// InvalidParameterCombinationException: Elasticsearch endpoint cant be modified.
							ForceNew: true,
						},
						"service_access_role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
							// API returns this error with ModifyEndpoint:
							// InvalidParameterCombinationException: Elasticsearch endpoint cant be modified.
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
						"auth_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      dms.AuthTypeValuePassword,
							ValidateFunc: validation.StringInSlice(dms.AuthTypeValue_Values(), false),
						},
						"auth_mechanism": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      mongoDBAuthMechanismValueDefault,
							ValidateFunc: validation.StringInSlice(mongoDBAuthMechanismValue_Values(), false),
						},
						"nesting_level": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      dms.NestingLevelValueNone,
							ValidateFunc: validation.StringInSlice(dms.NestingLevelValue_Values(), false),
						},
						"extract_doc_id": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "false",
						},
						"docs_to_investigate": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "1000",
						},
						"auth_source": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  mongoDBAuthSourceAdmin,
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
			"s3_settings": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_access_role_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "",
							ValidateFunc: verify.ValidARN,
						},
						"external_table_definition": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"csv_row_delimiter": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "\\n",
						},
						"csv_delimiter": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  ",",
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
						"compression_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      s3SettingsCompressionTypeNone,
							ValidateFunc: validation.StringInSlice(s3SettingsCompressionType_Values(), false),
						},
						"date_partition_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"data_format": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      dms.DataFormatValueCsv,
							ValidateFunc: validation.StringInSlice(dms.DataFormatValue_Values(), false),
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
						"encryption_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      s3SettingsEncryptionModeSseS3,
							ValidateFunc: validation.StringInSlice(s3SettingsEncryptionMode_Values(), false),
						},
						"server_side_encryption_kms_key_id": {
							Type:     schema.TypeString,
							Optional: true,
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

	request := &dms.CreateEndpointInput{
		EndpointIdentifier: aws.String(d.Get("endpoint_id").(string)),
		EndpointType:       aws.String(d.Get("endpoint_type").(string)),
		EngineName:         aws.String(d.Get("engine_name").(string)),
		Tags:               Tags(tags.IgnoreAWS()),
	}

	switch d.Get("engine_name").(string) {
	case engineNameDynamoDB:
		request.DynamoDbSettings = &dms.DynamoDbSettings{
			ServiceAccessRoleArn: aws.String(d.Get("service_access_role").(string)),
		}
	case engineNameElasticsearch:
		request.ElasticsearchSettings = &dms.ElasticsearchSettings{
			ServiceAccessRoleArn:    aws.String(d.Get("elasticsearch_settings.0.service_access_role_arn").(string)),
			EndpointUri:             aws.String(d.Get("elasticsearch_settings.0.endpoint_uri").(string)),
			ErrorRetryDuration:      aws.Int64(int64(d.Get("elasticsearch_settings.0.error_retry_duration").(int))),
			FullLoadErrorPercentage: aws.Int64(int64(d.Get("elasticsearch_settings.0.full_load_error_percentage").(int))),
		}
	case engineNameKafka:
		request.KafkaSettings = expandDmsKafkaSettings(d.Get("kafka_settings").([]interface{})[0].(map[string]interface{}))
	case engineNameKinesis:
		request.KinesisSettings = expandDmsKinesisSettings(d.Get("kinesis_settings").([]interface{})[0].(map[string]interface{}))
	case engineNameMongodb:
		request.MongoDbSettings = &dms.MongoDbSettings{
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

		// Set connection info in top-level namespace as well
		request.Username = aws.String(d.Get("username").(string))
		request.Password = aws.String(d.Get("password").(string))
		request.ServerName = aws.String(d.Get("server_name").(string))
		request.Port = aws.Int64(int64(d.Get("port").(int)))
		request.DatabaseName = aws.String(d.Get("database_name").(string))
	case engineNameOracle:
		if _, ok := d.GetOk("secrets_manager_arn"); ok {
			request.OracleSettings = &dms.OracleSettings{
				SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
				SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
				DatabaseName:                aws.String(d.Get("database_name").(string)),
			}
		} else {
			request.OracleSettings = &dms.OracleSettings{
				Username:     aws.String(d.Get("username").(string)),
				Password:     aws.String(d.Get("password").(string)),
				ServerName:   aws.String(d.Get("server_name").(string)),
				Port:         aws.Int64(int64(d.Get("port").(int))),
				DatabaseName: aws.String(d.Get("database_name").(string)),
			}

			// Set connection info in top-level namespace as well
			request.Username = aws.String(d.Get("username").(string))
			request.Password = aws.String(d.Get("password").(string))
			request.ServerName = aws.String(d.Get("server_name").(string))
			request.Port = aws.Int64(int64(d.Get("port").(int)))
			request.DatabaseName = aws.String(d.Get("database_name").(string))
		}
	case engineNamePostgres:
		if _, ok := d.GetOk("secrets_manager_arn"); ok {
			request.PostgreSQLSettings = &dms.PostgreSQLSettings{
				SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
				SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
				DatabaseName:                aws.String(d.Get("database_name").(string)),
			}
		} else {
			request.PostgreSQLSettings = &dms.PostgreSQLSettings{
				Username:     aws.String(d.Get("username").(string)),
				Password:     aws.String(d.Get("password").(string)),
				ServerName:   aws.String(d.Get("server_name").(string)),
				Port:         aws.Int64(int64(d.Get("port").(int))),
				DatabaseName: aws.String(d.Get("database_name").(string)),
			}

			// Set connection info in top-level namespace as well
			request.Username = aws.String(d.Get("username").(string))
			request.Password = aws.String(d.Get("password").(string))
			request.ServerName = aws.String(d.Get("server_name").(string))
			request.Port = aws.Int64(int64(d.Get("port").(int)))
			request.DatabaseName = aws.String(d.Get("database_name").(string))
		}
	case engineNameS3:
		request.S3Settings = &dms.S3Settings{
			BucketFolder:                  aws.String(d.Get("s3_settings.0.bucket_folder").(string)),
			BucketName:                    aws.String(d.Get("s3_settings.0.bucket_name").(string)),
			CompressionType:               aws.String(d.Get("s3_settings.0.compression_type").(string)),
			CsvDelimiter:                  aws.String(d.Get("s3_settings.0.csv_delimiter").(string)),
			CsvRowDelimiter:               aws.String(d.Get("s3_settings.0.csv_row_delimiter").(string)),
			DataFormat:                    aws.String(d.Get("s3_settings.0.data_format").(string)),
			DatePartitionEnabled:          aws.Bool(d.Get("s3_settings.0.date_partition_enabled").(bool)),
			EncryptionMode:                aws.String(d.Get("s3_settings.0.encryption_mode").(string)),
			ExternalTableDefinition:       aws.String(d.Get("s3_settings.0.external_table_definition").(string)),
			ParquetTimestampInMillisecond: aws.Bool(d.Get("s3_settings.0.parquet_timestamp_in_millisecond").(bool)),
			ParquetVersion:                aws.String(d.Get("s3_settings.0.parquet_version").(string)),
			ServerSideEncryptionKmsKeyId:  aws.String(d.Get("s3_settings.0.server_side_encryption_kms_key_id").(string)),
			ServiceAccessRoleArn:          aws.String(d.Get("s3_settings.0.service_access_role_arn").(string)),
		}
	default:
		request.Password = aws.String(d.Get("password").(string))
		request.Port = aws.Int64(int64(d.Get("port").(int)))
		request.ServerName = aws.String(d.Get("server_name").(string))
		request.Username = aws.String(d.Get("username").(string))

		if v, ok := d.GetOk("database_name"); ok {
			request.DatabaseName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("kms_key_arn"); ok {
			request.KmsKeyId = aws.String(v.(string))
		}
	}

	if v, ok := d.GetOk("certificate_arn"); ok {
		request.CertificateArn = aws.String(v.(string))
	}

	// Send ExtraConnectionAttributes in the API request for all resource types
	// per https://github.com/hashicorp/terraform-provider-aws/issues/8009
	if v, ok := d.GetOk("extra_connection_attributes"); ok {
		request.ExtraConnectionAttributes = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ssl_mode"); ok {
		request.SslMode = aws.String(v.(string))
	}

	log.Println("[DEBUG] DMS create endpoint:", request)

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.CreateEndpoint(request)
		if tfawserr.ErrMessageContains(err, "AccessDeniedFault", "") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		// Successful delete
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.CreateEndpoint(request)
	}
	if err != nil {
		return fmt.Errorf("Error creating DMS endpoint: %s", err)
	}

	d.SetId(d.Get("endpoint_id").(string))
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
		return fmt.Errorf("error reading DMS Endpoint (%s): %w", d.Id(), err)
	}

	err = resourceEndpointSetState(d, endpoint)

	if err != nil {
		return err
	}

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

func resourceEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DMSConn

	request := &dms.ModifyEndpointInput{
		EndpointArn: aws.String(d.Get("endpoint_arn").(string)),
	}
	hasChanges := false

	if d.HasChange("endpoint_type") {
		request.EndpointType = aws.String(d.Get("endpoint_type").(string))
		hasChanges = true
	}

	if d.HasChange("certificate_arn") {
		request.CertificateArn = aws.String(d.Get("certificate_arn").(string))
		hasChanges = true
	}

	if d.HasChange("service_access_role") {
		request.DynamoDbSettings = &dms.DynamoDbSettings{
			ServiceAccessRoleArn: aws.String(d.Get("service_access_role").(string)),
		}
		hasChanges = true
	}

	if d.HasChange("endpoint_type") {
		request.EndpointType = aws.String(d.Get("endpoint_type").(string))
		hasChanges = true
	}

	if d.HasChange("engine_name") {
		request.EngineName = aws.String(d.Get("engine_name").(string))
		hasChanges = true
	}

	if d.HasChange("extra_connection_attributes") {
		request.ExtraConnectionAttributes = aws.String(d.Get("extra_connection_attributes").(string))
		hasChanges = true
	}

	if d.HasChange("ssl_mode") {
		request.SslMode = aws.String(d.Get("ssl_mode").(string))
		hasChanges = true
	}

	if d.HasChange("tags_all") {
		arn := d.Get("endpoint_arn").(string)
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating DMS Endpoint (%s) tags: %s", arn, err)
		}
	}

	switch engineName := d.Get("engine_name").(string); engineName {
	case engineNameDynamoDB:
		if d.HasChange("service_access_role") {
			request.DynamoDbSettings = &dms.DynamoDbSettings{
				ServiceAccessRoleArn: aws.String(d.Get("service_access_role").(string)),
			}
			hasChanges = true
		}
	case engineNameElasticsearch:
		if d.HasChanges(
			"elasticsearch_settings.0.endpoint_uri",
			"elasticsearch_settings.0.error_retry_duration",
			"elasticsearch_settings.0.full_load_error_percentage",
			"elasticsearch_settings.0.service_access_role_arn") {
			request.ElasticsearchSettings = &dms.ElasticsearchSettings{
				ServiceAccessRoleArn:    aws.String(d.Get("elasticsearch_settings.0.service_access_role_arn").(string)),
				EndpointUri:             aws.String(d.Get("elasticsearch_settings.0.endpoint_uri").(string)),
				ErrorRetryDuration:      aws.Int64(int64(d.Get("elasticsearch_settings.0.error_retry_duration").(int))),
				FullLoadErrorPercentage: aws.Int64(int64(d.Get("elasticsearch_settings.0.full_load_error_percentage").(int))),
			}
			request.EngineName = aws.String(engineName)
			hasChanges = true
		}
	case engineNameKafka:
		if d.HasChange("kafka_settings") {
			request.KafkaSettings = expandDmsKafkaSettings(d.Get("kafka_settings").([]interface{})[0].(map[string]interface{}))
			request.EngineName = aws.String(engineName)
			hasChanges = true
		}
	case engineNameKinesis:
		if d.HasChanges("kinesis_settings") {
			request.KinesisSettings = expandDmsKinesisSettings(d.Get("kinesis_settings").([]interface{})[0].(map[string]interface{}))
			request.EngineName = aws.String(engineName)
			hasChanges = true
		}
	case engineNameMongodb:
		if d.HasChanges(
			"username", "password", "server_name", "port", "database_name", "mongodb_settings.0.auth_type",
			"mongodb_settings.0.auth_mechanism", "mongodb_settings.0.nesting_level", "mongodb_settings.0.extract_doc_id",
			"mongodb_settings.0.docs_to_investigate", "mongodb_settings.0.auth_source") {
			request.MongoDbSettings = &dms.MongoDbSettings{
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
			request.EngineName = aws.String(engineName)

			// Update connection info in top-level namespace as well
			request.Username = aws.String(d.Get("username").(string))
			request.Password = aws.String(d.Get("password").(string))
			request.ServerName = aws.String(d.Get("server_name").(string))
			request.Port = aws.Int64(int64(d.Get("port").(int)))
			request.DatabaseName = aws.String(d.Get("database_name").(string))

			hasChanges = true
		}
	case engineNameOracle:
		if d.HasChanges(
			"username", "password", "server_name", "port", "database_name", "secrets_manager_access_role_arn",
			"secrets_manager_arn") {
			if _, ok := d.GetOk("secrets_manager_arn"); ok {
				request.OracleSettings = &dms.OracleSettings{
					DatabaseName:                aws.String(d.Get("database_name").(string)),
					SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
					SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
				}
			} else {
				request.OracleSettings = &dms.OracleSettings{
					Username:     aws.String(d.Get("username").(string)),
					Password:     aws.String(d.Get("password").(string)),
					ServerName:   aws.String(d.Get("server_name").(string)),
					Port:         aws.Int64(int64(d.Get("port").(int))),
					DatabaseName: aws.String(d.Get("database_name").(string)),
				}
				request.EngineName = aws.String(d.Get("engine_name").(string)) // Must be included (should be 'oracle')

				// Update connection info in top-level namespace as well
				request.Username = aws.String(d.Get("username").(string))
				request.Password = aws.String(d.Get("password").(string))
				request.ServerName = aws.String(d.Get("server_name").(string))
				request.Port = aws.Int64(int64(d.Get("port").(int)))
				request.DatabaseName = aws.String(d.Get("database_name").(string))
			}
			hasChanges = true
		}
	case engineNamePostgres:
		if d.HasChanges(
			"username", "password", "server_name", "port", "database_name", "secrets_manager_access_role_arn",
			"secrets_manager_arn") {
			if _, ok := d.GetOk("secrets_manager_arn"); ok {
				request.PostgreSQLSettings = &dms.PostgreSQLSettings{
					DatabaseName:                aws.String(d.Get("database_name").(string)),
					SecretsManagerAccessRoleArn: aws.String(d.Get("secrets_manager_access_role_arn").(string)),
					SecretsManagerSecretId:      aws.String(d.Get("secrets_manager_arn").(string)),
				}
			} else {
				request.PostgreSQLSettings = &dms.PostgreSQLSettings{
					Username:     aws.String(d.Get("username").(string)),
					Password:     aws.String(d.Get("password").(string)),
					ServerName:   aws.String(d.Get("server_name").(string)),
					Port:         aws.Int64(int64(d.Get("port").(int))),
					DatabaseName: aws.String(d.Get("database_name").(string)),
				}
				request.EngineName = aws.String(d.Get("engine_name").(string)) // Must be included (should be 'postgres')

				// Update connection info in top-level namespace as well
				request.Username = aws.String(d.Get("username").(string))
				request.Password = aws.String(d.Get("password").(string))
				request.ServerName = aws.String(d.Get("server_name").(string))
				request.Port = aws.Int64(int64(d.Get("port").(int)))
				request.DatabaseName = aws.String(d.Get("database_name").(string))
			}
			hasChanges = true
		}
	case engineNameS3:
		if d.HasChanges(
			"s3_settings.0.service_access_role_arn", "s3_settings.0.external_table_definition",
			"s3_settings.0.csv_row_delimiter", "s3_settings.0.csv_delimiter", "s3_settings.0.bucket_folder",
			"s3_settings.0.bucket_name", "s3_settings.0.compression_type", "s3_settings.0.data_format",
			"s3_settings.0.parquet_version", "s3_settings.0.parquet_timestamp_in_millisecond",
			"s3_settings.0.encryption_mode", "s3_settings.0.server_side_encryption_kms_key_id") {
			request.S3Settings = &dms.S3Settings{
				BucketFolder:                  aws.String(d.Get("s3_settings.0.bucket_folder").(string)),
				BucketName:                    aws.String(d.Get("s3_settings.0.bucket_name").(string)),
				CompressionType:               aws.String(d.Get("s3_settings.0.compression_type").(string)),
				CsvDelimiter:                  aws.String(d.Get("s3_settings.0.csv_delimiter").(string)),
				CsvRowDelimiter:               aws.String(d.Get("s3_settings.0.csv_row_delimiter").(string)),
				DataFormat:                    aws.String(d.Get("s3_settings.0.data_format").(string)),
				EncryptionMode:                aws.String(d.Get("s3_settings.0.encryption_mode").(string)),
				ExternalTableDefinition:       aws.String(d.Get("s3_settings.0.external_table_definition").(string)),
				ParquetTimestampInMillisecond: aws.Bool(d.Get("s3_settings.0.parquet_timestamp_in_millisecond").(bool)),
				ParquetVersion:                aws.String(d.Get("s3_settings.0.parquet_version").(string)),
				ServerSideEncryptionKmsKeyId:  aws.String(d.Get("s3_settings.0.server_side_encryption_kms_key_id").(string)),
				ServiceAccessRoleArn:          aws.String(d.Get("s3_settings.0.service_access_role_arn").(string)),
			}
			request.EngineName = aws.String(engineName)
			hasChanges = true
		}
	default:
		if d.HasChange("database_name") {
			request.DatabaseName = aws.String(d.Get("database_name").(string))
			hasChanges = true
		}

		if d.HasChange("password") {
			request.Password = aws.String(d.Get("password").(string))
			hasChanges = true
		}

		if d.HasChange("port") {
			request.Port = aws.Int64(int64(d.Get("port").(int)))
			hasChanges = true
		}

		if d.HasChange("server_name") {
			request.ServerName = aws.String(d.Get("server_name").(string))
			hasChanges = true
		}

		if d.HasChange("username") {
			request.Username = aws.String(d.Get("username").(string))
			hasChanges = true
		}
	}

	if hasChanges {
		log.Println("[DEBUG] DMS update endpoint:", request)

		_, err := conn.ModifyEndpoint(request)
		if err != nil {
			return err
		}

		return resourceEndpointRead(d, meta)
	}

	return nil
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
		return fmt.Errorf("error deleting DMS Endpoint (%s): %w", d.Id(), err)
	}

	_, err = waitEndpointDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for DMS Endpoint (%s) delete: %w", d.Id(), err)
	}

	return err
}

func resourceEndpointCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	switch engineName := diff.Get("engine_name").(string); engineName {
	case engineNameElasticsearch:
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
	case engineNameDynamoDB:
		if endpoint.DynamoDbSettings != nil {
			d.Set("service_access_role", endpoint.DynamoDbSettings.ServiceAccessRoleArn)
		} else {
			d.Set("service_access_role", "")
		}
	case engineNameElasticsearch:
		if err := d.Set("elasticsearch_settings", flattenDmsElasticsearchSettings(endpoint.ElasticsearchSettings)); err != nil {
			return fmt.Errorf("Error setting elasticsearch for DMS: %s", err)
		}
	case engineNameKafka:
		if endpoint.KafkaSettings != nil {
			// SASL password isn't returned in API. Propagate state value.
			tfMap := flattenDmsKafkaSettings(endpoint.KafkaSettings)
			tfMap["sasl_password"] = d.Get("kafka_settings.0.sasl_password").(string)

			if err := d.Set("kafka_settings", []interface{}{tfMap}); err != nil {
				return fmt.Errorf("error setting kafka_settings: %w", err)
			}
		} else {
			d.Set("kafka_settings", nil)
		}
	case engineNameKinesis:
		if err := d.Set("kinesis_settings", []interface{}{flattenDmsKinesisSettings(endpoint.KinesisSettings)}); err != nil {
			return fmt.Errorf("error setting kinesis_settings: %w", err)
		}
	case engineNameMongodb:
		if endpoint.MongoDbSettings != nil {
			d.Set("username", endpoint.MongoDbSettings.Username)
			d.Set("server_name", endpoint.MongoDbSettings.ServerName)
			d.Set("port", endpoint.MongoDbSettings.Port)
			d.Set("database_name", endpoint.MongoDbSettings.DatabaseName)
		} else {
			d.Set("username", endpoint.Username)
			d.Set("server_name", endpoint.ServerName)
			d.Set("port", endpoint.Port)
			d.Set("database_name", endpoint.DatabaseName)
		}
		if err := d.Set("mongodb_settings", flattenDmsMongoDbSettings(endpoint.MongoDbSettings)); err != nil {
			return fmt.Errorf("Error setting mongodb_settings for DMS: %s", err)
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
			d.Set("username", endpoint.Username)
			d.Set("server_name", endpoint.ServerName)
			d.Set("port", endpoint.Port)
			d.Set("database_name", endpoint.DatabaseName)
		}
	case engineNamePostgres:
		if endpoint.PostgreSQLSettings != nil {
			d.Set("username", endpoint.PostgreSQLSettings.Username)
			d.Set("server_name", endpoint.PostgreSQLSettings.ServerName)
			d.Set("port", endpoint.PostgreSQLSettings.Port)
			d.Set("database_name", endpoint.PostgreSQLSettings.DatabaseName)
			d.Set("secrets_manager_access_role_arn", endpoint.PostgreSQLSettings.SecretsManagerAccessRoleArn)
			d.Set("secrets_manager_arn", endpoint.PostgreSQLSettings.SecretsManagerSecretId)
		} else {
			d.Set("username", endpoint.Username)
			d.Set("server_name", endpoint.ServerName)
			d.Set("port", endpoint.Port)
			d.Set("database_name", endpoint.DatabaseName)
		}
	case engineNameS3:
		if err := d.Set("s3_settings", flattenDmsS3Settings(endpoint.S3Settings)); err != nil {
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

func flattenDmsElasticsearchSettings(settings *dms.ElasticsearchSettings) []map[string]interface{} {
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

func expandDmsKafkaSettings(tfMap map[string]interface{}) *dms.KafkaSettings {
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

func flattenDmsKafkaSettings(apiObject *dms.KafkaSettings) map[string]interface{} {
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

func expandDmsKinesisSettings(tfMap map[string]interface{}) *dms.KinesisSettings {
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

func flattenDmsKinesisSettings(apiObject *dms.KinesisSettings) map[string]interface{} {
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

func flattenDmsMongoDbSettings(settings *dms.MongoDbSettings) []map[string]interface{} {
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

func flattenDmsS3Settings(settings *dms.S3Settings) []map[string]interface{} {
	if settings == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"bucket_folder":                     aws.StringValue(settings.BucketFolder),
		"bucket_name":                       aws.StringValue(settings.BucketName),
		"compression_type":                  aws.StringValue(settings.CompressionType),
		"csv_delimiter":                     aws.StringValue(settings.CsvDelimiter),
		"csv_row_delimiter":                 aws.StringValue(settings.CsvRowDelimiter),
		"data_format":                       aws.StringValue(settings.DataFormat),
		"date_partition_enabled":            aws.BoolValue(settings.DatePartitionEnabled),
		"encryption_mode":                   aws.StringValue(settings.EncryptionMode),
		"external_table_definition":         aws.StringValue(settings.ExternalTableDefinition),
		"parquet_timestamp_in_millisecond":  aws.BoolValue(settings.ParquetTimestampInMillisecond),
		"parquet_version":                   aws.StringValue(settings.ParquetVersion),
		"server_side_encryption_kms_key_id": aws.StringValue(settings.ServerSideEncryptionKmsKeyId),
		"service_access_role_arn":           aws.StringValue(settings.ServiceAccessRoleArn),
	}

	return []map[string]interface{}{m}
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
