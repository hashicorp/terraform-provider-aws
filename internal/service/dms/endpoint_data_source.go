// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package dms

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_dms_endpoint", name="Endpoint")
// @Tags(identifierAttribute="endpoint_arn")
func dataSourceEndpoint() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEndpointRead,

		Schema: map[string]*schema.Schema{
			names.AttrCertificateARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"elasticsearch_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint_uri": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"error_retry_duration": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"full_load_error_percentage": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"service_access_role_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"endpoint_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrEndpointType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"extra_connection_attributes": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kafka_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"broker": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"include_control_details": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_null_and_empty": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_partition_value": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_table_alter_operations": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_transaction_details": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"message_format": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"message_max_bytes": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"no_hex_prefix": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"partition_include_schema_table": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"sasl_mechanism": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"sasl_password": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"sasl_username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"security_protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ssl_ca_certificate_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ssl_client_certificate_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ssl_client_key_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ssl_client_key_password": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"topic": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"kinesis_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include_control_details": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_null_and_empty": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_partition_value": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_table_alter_operations": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_transaction_details": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"message_format": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"partition_include_schema_table": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"service_access_role_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStreamARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"use_large_integer_value": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			names.AttrKMSKeyARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mongodb_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth_mechanism": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"auth_source": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"auth_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"docs_to_investigate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"extract_doc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"nesting_level": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"use_update_lookup": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"mysql_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"after_connect_script": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"authentication_method": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"clean_source_metadata_on_mismatch": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"events_poll_interval": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"execute_timeout": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"max_file_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"parallel_load_threads": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"server_timezone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_access_role_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target_db_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrPassword: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"postgres_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"after_connect_script": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"authentication_method": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"babelfish_database_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"capture_ddls": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"database_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ddl_artifacts_schema": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"execute_timeout": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"fail_tasks_on_lob_truncation": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"heartbeat_enable": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"heartbeat_frequency": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"heartbeat_schema": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"map_boolean_as_boolean": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"map_jsonb_as_clob": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"map_long_varchar_as": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"max_file_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"plugin_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_access_role_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"slot_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"redis_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth_password": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"auth_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"auth_user_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrPort: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"server_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ssl_ca_certificate_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ssl_security_protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"redshift_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_folder": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrBucketName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"encryption_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"server_side_encryption_kms_key_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_access_role_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"s3_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"add_column_name": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"bucket_folder": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrBucketName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"canned_acl_for_objects": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cdc_inserts_and_updates": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"cdc_inserts_only": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"cdc_max_batch_interval": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cdc_min_file_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cdc_path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"compression_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"csv_delimiter": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"csv_no_sup_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"csv_null_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"csv_row_delimiter": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"data_format": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"data_page_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"date_partition_delimiter": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"date_partition_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"date_partition_sequence": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dict_page_size_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"enable_statistics": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"encoding_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"encryption_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"external_table_definition": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"glue_catalog_generation": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"ignore_headers_row": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ignore_header_rows": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"include_op_for_full_load": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"max_file_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"parquet_timestamp_in_millisecond": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"parquet_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"preserve_transactions": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"rfc_4180": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"row_group_length": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"server_side_encryption_kms_key_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_access_role_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"timestamp_column_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"use_csv_no_sup_value": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"use_task_start_time_for_full_load_timestamp": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"secrets_manager_access_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secrets_manager_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_access_role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssl_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrUsername: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	endptID := d.Get("endpoint_id").(string)
	out, err := findEndpointByID(ctx, conn, endptID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Endpoint (%s): %s", endptID, err)
	}

	d.SetId(aws.ToString(out.EndpointIdentifier))
	d.Set("endpoint_id", out.EndpointIdentifier)
	arn := aws.ToString(out.EndpointArn)
	d.Set("endpoint_arn", arn)
	d.Set(names.AttrEndpointType, out.EndpointType)
	d.Set(names.AttrDatabaseName, out.DatabaseName)
	d.Set("engine_name", out.EngineName)
	d.Set(names.AttrPort, out.Port)
	d.Set("server_name", out.ServerName)
	d.Set("ssl_mode", out.SslMode)
	d.Set(names.AttrUsername, out.Username)

	if err := resourceEndpointDataSourceSetState(d, out); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func resourceEndpointDataSourceSetState(d *schema.ResourceData, endpoint *awstypes.Endpoint) error {
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
		if err := d.Set("mysql_settings", flattenMySQLSettings(endpoint.MySQLSettings)); err != nil {
			return fmt.Errorf("setting mysql_settings: %w", err)
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
		if err := d.Set("elasticsearch_settings", flattenElasticsearchSettings(endpoint.ElasticsearchSettings)); err != nil {
			return fmt.Errorf("setting elasticsearch_settings: %w", err)
		}
	case engineNameKafka:
		if endpoint.KafkaSettings != nil {
			// SASL password isn't returned in API. Propagate state value.
			tfMap := flattenKafkaSettings(endpoint.KafkaSettings)
			tfMap["sasl_password"] = d.Get("kafka_settings.0.sasl_password").(string)

			if err := d.Set("kafka_settings", []any{tfMap}); err != nil {
				return fmt.Errorf("setting kafka_settings: %w", err)
			}
		} else {
			d.Set("kafka_settings", nil)
		}
	case engineNameKinesis:
		if err := d.Set("kinesis_settings", []any{flattenKinesisSettings(endpoint.KinesisSettings)}); err != nil {
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

		if err := d.Set("redis_settings", []any{tfMap}); err != nil {
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
			return fmt.Errorf("setting s3_settings for DMS: %w", err)
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

func flattenS3Settings(apiObject *awstypes.S3Settings) []map[string]any {
	if apiObject == nil {
		return []map[string]any{}
	}

	tfMap := map[string]any{}

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

	return []map[string]any{tfMap}
}
