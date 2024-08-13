// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_dms_endpoint", name="Endpoint")
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

func dataSourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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

	if err := resourceEndpointSetState(d, out); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	tags, err := listTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for DMS Endpoint (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set(names.AttrTags, tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
