// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_dms_endpoint")
func DataSourceEndpoint() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEndpointRead,

		Schema: map[string]*schema.Schema{
			"certificate_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"elasticsearch_settings": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint_uri": {
							Type:     schema.TypeString,
							Required: true,
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
							Required: true,
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
			"endpoint_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"extra_connection_attributes": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kafka_settings": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"broker": {
							Type:     schema.TypeString,
							Required: true,
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
						"stream_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"kms_key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mongodb_settings": {
				Type:     schema.TypeList,
				Optional: true,
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
			"password": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
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
						"port": {
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
						"bucket_name": {
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
						"bucket_name": {
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
			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameEndpoint = "Endpoint Data Source"
)

func dataSourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DMSConn(ctx)
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	endptID := d.Get("endpoint_id").(string)

	out, err := FindEndpointByID(ctx, conn, endptID)
	if err != nil {
		create.DiagError(names.DMS, create.ErrActionReading, DSNameEndpoint, d.Id(), err)
	}

	d.SetId(aws.StringValue(out.EndpointIdentifier))

	d.Set("endpoint_id", out.EndpointIdentifier)
	d.Set("endpoint_arn", out.EndpointArn)
	d.Set("endpoint_type", out.EndpointType)
	d.Set("database_name", out.DatabaseName)
	d.Set("engine_name", out.EngineName)
	d.Set("port", out.Port)
	d.Set("server_name", out.ServerName)
	d.Set("ssl_mode", out.SslMode)
	d.Set("username", out.Username)

	err = resourceEndpointSetState(d, out)
	if err != nil {
		create.DiagError(names.DMS, create.ErrActionReading, DSNameEndpoint, d.Id(), err)
	}

	tags, err := listTags(ctx, conn, aws.StringValue(out.EndpointArn))
	if err != nil {
		return create.DiagError(names.DMS, create.ErrActionReading, DSNameEndpoint, d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.DMS, create.ErrActionSetting, DSNameEndpoint, d.Id(), err)
	}

	return nil
}
