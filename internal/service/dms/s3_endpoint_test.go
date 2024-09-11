// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDMSS3Endpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_s3_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccS3EndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "add_column_name", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "add_trailing_padding_character", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "bucket_folder", "folder"),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucketName, names.AttrBucketName),
					resource.TestCheckResourceAttr(resourceName, "canned_acl_for_objects", "private"),
					resource.TestCheckResourceAttr(resourceName, "cdc_inserts_and_updates", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cdc_inserts_only", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cdc_max_batch_interval", "100"),
					resource.TestCheckResourceAttr(resourceName, "cdc_min_file_size", "16"),
					resource.TestCheckResourceAttr(resourceName, "cdc_path", "cdc/path"),
					resource.TestCheckResourceAttr(resourceName, "compression_type", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "csv_delimiter", ";"),
					resource.TestCheckResourceAttr(resourceName, "csv_no_sup_value", "x"),
					resource.TestCheckResourceAttr(resourceName, "csv_null_value", "?"),
					resource.TestCheckResourceAttr(resourceName, "csv_row_delimiter", "\\r\\n"),
					resource.TestCheckResourceAttr(resourceName, "data_format", "parquet"),
					resource.TestCheckResourceAttr(resourceName, "data_page_size", "1100000"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_delimiter", "UNDERSCORE"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "date_partition_sequence", "yyyymmddhh"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_timezone", "Asia/Seoul"),
					resource.TestCheckResourceAttr(resourceName, "dict_page_size_limit", "1000000"),
					resource.TestCheckResourceAttr(resourceName, "enable_statistics", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "encoding_type", "plain"),
					resource.TestCheckResourceAttr(resourceName, "encryption_mode", "SSE_S3"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExpectedBucketOwner, "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "glue_catalog_generation", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "ignore_header_rows", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "include_op_for_full_load", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "max_file_size", "1000000"),
					resource.TestCheckResourceAttr(resourceName, "parquet_timestamp_in_millisecond", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "parquet_version", "parquet-2-0"),
					resource.TestCheckResourceAttr(resourceName, "preserve_transactions", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "rfc_4180", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "row_group_length", "11000"),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption_kms_key_id", "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "service_access_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "timestamp_column_name", "tx_commit_time"),
					resource.TestCheckResourceAttr(resourceName, "use_csv_no_sup_value", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "use_task_start_time_for_full_load_timestamp", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDMSS3Endpoint_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_s3_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccS3EndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "add_column_name", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "add_trailing_padding_character", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "bucket_folder", "folder"),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucketName, names.AttrBucketName),
					resource.TestCheckResourceAttr(resourceName, "canned_acl_for_objects", "private"),
					resource.TestCheckResourceAttr(resourceName, "cdc_inserts_and_updates", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cdc_max_batch_interval", "100"),
					resource.TestCheckResourceAttr(resourceName, "cdc_min_file_size", "16"),
					resource.TestCheckResourceAttr(resourceName, "cdc_path", "cdc/path"),
					resource.TestCheckResourceAttr(resourceName, "compression_type", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "csv_delimiter", ";"),
					resource.TestCheckResourceAttr(resourceName, "csv_no_sup_value", "x"),
					resource.TestCheckResourceAttr(resourceName, "csv_null_value", "?"),
					resource.TestCheckResourceAttr(resourceName, "csv_row_delimiter", "\\r\\n"),
					resource.TestCheckResourceAttr(resourceName, "data_format", "parquet"),
					resource.TestCheckResourceAttr(resourceName, "data_page_size", "1100000"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_delimiter", "UNDERSCORE"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "date_partition_sequence", "yyyymmddhh"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_timezone", "Asia/Seoul"),
					resource.TestCheckResourceAttr(resourceName, "dict_page_size_limit", "1000000"),
					resource.TestCheckResourceAttr(resourceName, "enable_statistics", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "encoding_type", "plain"),
					resource.TestCheckResourceAttr(resourceName, "encryption_mode", "SSE_S3"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExpectedBucketOwner, "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "glue_catalog_generation", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "ignore_header_rows", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "include_op_for_full_load", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "max_file_size", "1000000"),
					resource.TestCheckResourceAttr(resourceName, "parquet_timestamp_in_millisecond", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "parquet_version", "parquet-2-0"),
					resource.TestCheckResourceAttr(resourceName, "preserve_transactions", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "rfc_4180", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "row_group_length", "11000"),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption_kms_key_id", "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "service_access_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "timestamp_column_name", "tx_commit_time"),
					resource.TestCheckResourceAttr(resourceName, "use_csv_no_sup_value", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "use_task_start_time_for_full_load_timestamp", acctest.CtTrue),
				),
			},
			{
				Config: testAccS3EndpointConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "add_column_name", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "add_trailing_padding_character", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "bucket_folder", "folder2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucketName, "updated_name"),
					resource.TestCheckResourceAttr(resourceName, "canned_acl_for_objects", "private"),
					resource.TestCheckResourceAttr(resourceName, "cdc_inserts_and_updates", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cdc_inserts_only", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cdc_max_batch_interval", "105"),
					resource.TestCheckResourceAttr(resourceName, "cdc_min_file_size", "17"),
					resource.TestCheckResourceAttr(resourceName, "cdc_path", "cdc/path"),
					resource.TestCheckResourceAttr(resourceName, "compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "csv_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "csv_no_sup_value", "U"),
					resource.TestCheckResourceAttr(resourceName, "csv_null_value", "-"),
					resource.TestCheckResourceAttr(resourceName, "csv_row_delimiter", "\\n"),
					resource.TestCheckResourceAttr(resourceName, "data_format", "parquet"),
					resource.TestCheckResourceAttr(resourceName, "data_page_size", "1100000"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_delimiter", "SLASH"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "date_partition_sequence", "yyyymmddhh"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_timezone", "Europe/Paris"),
					resource.TestCheckResourceAttr(resourceName, "dict_page_size_limit", "1000000"),
					resource.TestCheckResourceAttr(resourceName, "enable_statistics", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "encoding_type", "plain"),
					resource.TestCheckResourceAttr(resourceName, "encryption_mode", "SSE_S3"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExpectedBucketOwner, "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "glue_catalog_generation", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ignore_header_rows", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "include_op_for_full_load", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "max_file_size", "900000"),
					resource.TestCheckResourceAttr(resourceName, "parquet_timestamp_in_millisecond", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "parquet_version", "parquet-2-0"),
					resource.TestCheckResourceAttr(resourceName, "preserve_transactions", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "rfc_4180", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "row_group_length", "13000"),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption_kms_key_id", "aws_kms_key.test2", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "service_access_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "timestamp_column_name", "tx_commit_time2"),
					resource.TestCheckResourceAttr(resourceName, "use_csv_no_sup_value", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "use_task_start_time_for_full_load_timestamp", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccDMSS3Endpoint_simple(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_s3_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccS3EndpointConfig_simple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "add_column_name", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "add_trailing_padding_character", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "bucket_folder", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucketName, "beckut_name"),
					resource.TestCheckResourceAttr(resourceName, "canned_acl_for_objects", ""),
					resource.TestCheckResourceAttr(resourceName, "cdc_inserts_and_updates", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cdc_inserts_only", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cdc_max_batch_interval", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cdc_min_file_size", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cdc_path", ""),
					resource.TestCheckResourceAttr(resourceName, "compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "csv_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "csv_no_sup_value", ""),
					resource.TestCheckResourceAttr(resourceName, "csv_null_value", ""),
					resource.TestCheckResourceAttr(resourceName, "csv_row_delimiter", "\\n"),
					resource.TestCheckResourceAttr(resourceName, "data_format", ""),
					resource.TestCheckResourceAttr(resourceName, "data_page_size", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "date_partition_delimiter", ""),
					resource.TestCheckResourceAttr(resourceName, "date_partition_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "date_partition_sequence", ""),
					resource.TestCheckResourceAttr(resourceName, "date_partition_timezone", ""),
					resource.TestCheckResourceAttr(resourceName, "dict_page_size_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "enable_statistics", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "encoding_type", ""),
					resource.TestCheckResourceAttr(resourceName, "encryption_mode", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrExpectedBucketOwner, ""),
					resource.TestCheckResourceAttr(resourceName, "glue_catalog_generation", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ignore_header_rows", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "include_op_for_full_load", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "max_file_size", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "parquet_timestamp_in_millisecond", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "parquet_version", ""),
					resource.TestCheckResourceAttr(resourceName, "preserve_transactions", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "rfc_4180", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "row_group_length", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_kms_key_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "service_access_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "timestamp_column_name", ""),
					resource.TestCheckResourceAttr(resourceName, "use_csv_no_sup_value", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "use_task_start_time_for_full_load_timestamp", acctest.CtFalse),
				),
			},
			{
				Config:   testAccS3EndpointConfig_simple(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccDMSS3Endpoint_sourceSimple(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_s3_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccS3EndpointConfig_sourceSimple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "add_column_name", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "bucket_folder", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucketName, "beckut_name"),
					resource.TestCheckResourceAttr(resourceName, "canned_acl_for_objects", ""),
					resource.TestCheckResourceAttr(resourceName, "cdc_inserts_and_updates", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cdc_inserts_only", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cdc_max_batch_interval", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cdc_min_file_size", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cdc_path", ""),
					resource.TestCheckResourceAttr(resourceName, "compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "csv_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "csv_null_value", ""),
					resource.TestCheckResourceAttr(resourceName, "csv_row_delimiter", "\\n"),
					resource.TestCheckResourceAttr(resourceName, "data_page_size", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "date_partition_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "dict_page_size_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "enable_statistics", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "encoding_type", ""),
					resource.TestCheckResourceAttr(resourceName, "endpoint_id", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEndpointType, names.AttrSource),
					resource.TestCheckResourceAttr(resourceName, names.AttrExpectedBucketOwner, ""),
					resource.TestCheckResourceAttr(resourceName, "external_table_definition", "{\"TableCount\":1,\"Tables\":[{\"TableColumns\":[{\"ColumnIsPk\":\"true\",\"ColumnName\":\"ID\",\"ColumnNullable\":\"false\",\"ColumnType\":\"INT8\"},{\"ColumnLength\":\"20\",\"ColumnName\":\"LastName\",\"ColumnType\":\"STRING\"}],\"TableColumnsTotal\":\"2\",\"TableName\":\"employee\",\"TableOwner\":\"hr\",\"TablePath\":\"hr/employee/\"}]}"),
					resource.TestCheckResourceAttr(resourceName, "ignore_header_rows", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "include_op_for_full_load", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "max_file_size", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "rfc_4180", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "row_group_length", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "service_access_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_column_name", ""),
					resource.TestCheckResourceAttr(resourceName, "use_task_start_time_for_full_load_timestamp", acctest.CtFalse),
				),
			},
			{
				Config:   testAccS3EndpointConfig_sourceSimple(rName),
				PlanOnly: true,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"compression_type", "date_partition_enabled", "parquet_timestamp_in_millisecond", "preserve_transactions", "use_csv_no_sup_value", "glue_catalog_generation"},
			},
		},
	})
}

func TestAccDMSS3Endpoint_source(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_s3_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccS3EndpointConfig_source(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket_folder", "folder"),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucketName, names.AttrBucketName),
					resource.TestCheckResourceAttr(resourceName, "cdc_path", "cdc/path"),
					resource.TestCheckResourceAttr(resourceName, "csv_delimiter", ";"),
					resource.TestCheckResourceAttr(resourceName, "csv_row_delimiter", "\\r\\n"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_id", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEndpointType, names.AttrSource),
					resource.TestCheckResourceAttr(resourceName, "ignore_header_rows", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rfc_4180", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "service_access_role_arn", "aws_iam_role.test", names.AttrARN),

					resource.TestCheckResourceAttr(resourceName, "add_column_name", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "canned_acl_for_objects", "private"),
					resource.TestCheckResourceAttr(resourceName, "cdc_inserts_and_updates", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cdc_max_batch_interval", "100"),
					resource.TestCheckResourceAttr(resourceName, "cdc_min_file_size", "16"),
					resource.TestCheckResourceAttr(resourceName, "compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "csv_null_value", "?"),
					resource.TestCheckResourceAttr(resourceName, "data_page_size", "1100000"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "dict_page_size_limit", "1000000"),
					resource.TestCheckResourceAttr(resourceName, "enable_statistics", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "encoding_type", "plain"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExpectedBucketOwner, "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "external_table_definition", "{\"TableCount\":1,\"Tables\":[{\"TableColumns\":[{\"ColumnIsPk\":\"true\",\"ColumnName\":\"ID\",\"ColumnNullable\":\"false\",\"ColumnType\":\"INT8\"},{\"ColumnLength\":\"20\",\"ColumnName\":\"LastName\",\"ColumnType\":\"STRING\"}],\"TableColumnsTotal\":\"2\",\"TableName\":\"employee\",\"TableOwner\":\"hr\",\"TablePath\":\"hr/employee/\"}]}"),
					resource.TestCheckResourceAttr(resourceName, "include_op_for_full_load", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "max_file_size", "1000000"),
					resource.TestCheckResourceAttr(resourceName, "row_group_length", "11000"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_column_name", "tx_commit_time"),
					resource.TestCheckResourceAttr(resourceName, "use_task_start_time_for_full_load_timestamp", acctest.CtTrue),
				),
			},
			{
				Config: testAccS3EndpointConfig_sourceUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket_folder", "folder2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucketName, "beckut_name"),
					resource.TestCheckResourceAttr(resourceName, "cdc_path", "cdc/path2"),
					resource.TestCheckResourceAttr(resourceName, "csv_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "csv_row_delimiter", "\\n"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_id", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEndpointType, names.AttrSource),
					resource.TestCheckResourceAttr(resourceName, "ignore_header_rows", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rfc_4180", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "service_access_role_arn", "aws_iam_role.test", names.AttrARN),

					resource.TestCheckResourceAttr(resourceName, "add_column_name", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "canned_acl_for_objects", "authenticated-read"),
					resource.TestCheckResourceAttr(resourceName, "cdc_inserts_and_updates", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cdc_inserts_only", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cdc_max_batch_interval", "101"),
					resource.TestCheckResourceAttr(resourceName, "cdc_min_file_size", "17"),
					resource.TestCheckResourceAttr(resourceName, "compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "csv_null_value", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "data_page_size", "1000000"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "dict_page_size_limit", "830000"),
					resource.TestCheckResourceAttr(resourceName, "enable_statistics", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "encoding_type", "plain-dictionary"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExpectedBucketOwner, "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "external_table_definition", "{\"TableCount\":1,\"Tables\":[{\"TableColumns\":[{\"ColumnIsPk\":\"true\",\"ColumnName\":\"ID\",\"ColumnNullable\":\"false\",\"ColumnType\":\"INT8\"}],\"TableColumnsTotal\":\"1\",\"TableName\":\"employee\",\"TableOwner\":\"hr\",\"TablePath\":\"hr/employee/\"}]}"),
					resource.TestCheckResourceAttr(resourceName, "include_op_for_full_load", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "max_file_size", "100"),
					resource.TestCheckResourceAttr(resourceName, "row_group_length", "10000"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_column_name", "tx_commit_time2"),
					resource.TestCheckResourceAttr(resourceName, "use_task_start_time_for_full_load_timestamp", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccDMSS3Endpoint_detachTargetOnLobLookupFailureParquet(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_s3_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccS3EndpointConfig_simple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "detach_target_on_lob_lookup_failure_parquet"),
				),
			},
			{
				Config: testAccS3EndpointConfig_detachTargetOnLobLookupFailureParquet(rName, "cdc/path", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cdc_path", "cdc/path"),
					resource.TestCheckResourceAttr(resourceName, "detach_target_on_lob_lookup_failure_parquet", acctest.CtTrue),
				),
			},
			{
				Config: testAccS3EndpointConfig_detachTargetOnLobLookupFailureParquet(rName, "cdc/path2", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cdc_path", "cdc/path2"),
					resource.TestCheckResourceAttr(resourceName, "detach_target_on_lob_lookup_failure_parquet", acctest.CtTrue),
				),
			},
			{
				Config: testAccS3EndpointConfig_detachTargetOnLobLookupFailureParquet(rName, "cdc/path3", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cdc_path", "cdc/path3"),
					resource.TestCheckResourceAttr(resourceName, "detach_target_on_lob_lookup_failure_parquet", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccS3EndpointConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "dms.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "s3:CreateBucket",
        "s3:ListBucket",
        "s3:DeleteBucket",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject",
        "s3:GetObjectVersion",
        "s3:GetBucketPolicy",
        "s3:PutBucketPolicy",
        "s3:DeleteBucketPolicy"
      ]
      Resource = "*"
    }]
  })
}
`, rName)
}

func testAccS3EndpointConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccS3EndpointConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = %[1]q
    Statement = [{
      Sid    = %[1]q
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Action   = "kms:*"
      Resource = "*"
    }]
  })
}

resource "aws_dms_s3_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  ssl_mode      = "none"

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }

  add_column_name                             = true
  add_trailing_padding_character              = false
  bucket_folder                               = "folder"
  bucket_name                                 = "bucket_name"
  canned_acl_for_objects                      = "private"
  cdc_inserts_and_updates                     = true
  cdc_inserts_only                            = false
  cdc_max_batch_interval                      = 100
  cdc_min_file_size                           = 16
  cdc_path                                    = "cdc/path"
  compression_type                            = "GZIP"
  csv_delimiter                               = ";"
  csv_no_sup_value                            = "x"
  csv_null_value                              = "?"
  csv_row_delimiter                           = "\\r\\n"
  data_format                                 = "parquet"
  data_page_size                              = 1100000
  date_partition_delimiter                    = "UNDERSCORE"
  date_partition_enabled                      = true
  date_partition_sequence                     = "yyyymmddhh"
  date_partition_timezone                     = "Asia/Seoul"
  dict_page_size_limit                        = 1000000
  enable_statistics                           = false
  encoding_type                               = "plain"
  encryption_mode                             = "SSE_S3"
  expected_bucket_owner                       = data.aws_caller_identity.current.account_id
  ignore_header_rows                          = 1
  include_op_for_full_load                    = true
  max_file_size                               = 1000000
  parquet_timestamp_in_millisecond            = true
  parquet_version                             = "parquet-2-0"
  preserve_transactions                       = false
  rfc_4180                                    = false
  row_group_length                            = 11000
  server_side_encryption_kms_key_id           = aws_kms_key.test.arn
  service_access_role_arn                     = aws_iam_role.test.arn
  timestamp_column_name                       = "tx_commit_time"
  use_csv_no_sup_value                        = false
  use_task_start_time_for_full_load_timestamp = true
  glue_catalog_generation                     = true

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccS3EndpointConfig_update(rName string) string {
	return acctest.ConfigCompose(
		testAccS3EndpointConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test2" {
  description = %[1]q

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = %[1]q
    Statement = [{
      Sid    = %[1]q
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Action   = "kms:*"
      Resource = "*"
    }]
  })
}

resource "aws_dms_s3_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  ssl_mode      = "none"

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }

  add_column_name                             = false
  add_trailing_padding_character              = true
  bucket_folder                               = "folder2"
  bucket_name                                 = "updated_name"
  canned_acl_for_objects                      = "private"
  cdc_inserts_and_updates                     = false
  cdc_inserts_only                            = true
  cdc_max_batch_interval                      = 105
  cdc_min_file_size                           = 17
  cdc_path                                    = "cdc/path"
  compression_type                            = "NONE"
  csv_delimiter                               = ","
  csv_no_sup_value                            = "U"
  csv_null_value                              = "-"
  csv_row_delimiter                           = "\\n"
  data_format                                 = "parquet"
  data_page_size                              = 1100000
  date_partition_delimiter                    = "SLASH"
  date_partition_enabled                      = true
  date_partition_sequence                     = "yyyymmddhh"
  date_partition_timezone                     = "Europe/Paris"
  dict_page_size_limit                        = 1000000
  enable_statistics                           = true
  encoding_type                               = "plain"
  encryption_mode                             = "SSE_S3"
  expected_bucket_owner                       = data.aws_caller_identity.current.account_id
  ignore_header_rows                          = 1
  include_op_for_full_load                    = false
  max_file_size                               = 900000
  parquet_timestamp_in_millisecond            = true
  parquet_version                             = "parquet-2-0"
  preserve_transactions                       = false
  rfc_4180                                    = true
  row_group_length                            = 13000
  server_side_encryption_kms_key_id           = aws_kms_key.test2.arn
  service_access_role_arn                     = aws_iam_role.test.arn
  timestamp_column_name                       = "tx_commit_time2"
  use_csv_no_sup_value                        = true
  use_task_start_time_for_full_load_timestamp = false
  glue_catalog_generation                     = false

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccS3EndpointConfig_simple(rName string) string {
	return acctest.ConfigCompose(
		testAccS3EndpointConfig_base(rName),
		fmt.Sprintf(`
resource "aws_dms_s3_endpoint" "test" {
  endpoint_id             = %[1]q
  endpoint_type           = "target"
  bucket_name             = "beckut_name"
  service_access_role_arn = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccS3EndpointConfig_sourceSimple(rName string) string {
	return acctest.ConfigCompose(
		testAccS3EndpointConfig_base(rName),
		fmt.Sprintf(`
resource "aws_dms_s3_endpoint" "test" {
  bucket_name             = "beckut_name"
  endpoint_id             = %[1]q
  endpoint_type           = "source"
  service_access_role_arn = aws_iam_role.test.arn

  external_table_definition = jsonencode({
    TableCount = 1
    Tables = [{
      TableName  = "employee"
      TablePath  = "hr/employee/"
      TableOwner = "hr"
      TableColumns = [{
        ColumnName     = "ID"
        ColumnType     = "INT8"
        ColumnNullable = "false"
        ColumnIsPk     = "true"
        }, {
        ColumnName   = "LastName"
        ColumnType   = "STRING"
        ColumnLength = "20"
      }]
      TableColumnsTotal = "2"
    }]
  })

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccS3EndpointConfig_source(rName string) string {
	return acctest.ConfigCompose(
		testAccS3EndpointConfig_base(rName),
		fmt.Sprintf(`
resource "aws_dms_s3_endpoint" "test" {
  bucket_folder           = "folder"
  bucket_name             = "bucket_name"
  cdc_path                = "cdc/path"
  csv_delimiter           = ";"
  csv_row_delimiter       = "\\r\\n"
  endpoint_id             = %[1]q
  endpoint_type           = "source"
  ignore_header_rows      = 1
  rfc_4180                = false
  service_access_role_arn = aws_iam_role.test.arn

  external_table_definition = jsonencode({
    TableCount = 1
    Tables = [{
      TableName  = "employee"
      TablePath  = "hr/employee/"
      TableOwner = "hr"
      TableColumns = [{
        ColumnName     = "ID"
        ColumnType     = "INT8"
        ColumnNullable = "false"
        ColumnIsPk     = "true"
        }, {
        ColumnName   = "LastName"
        ColumnType   = "STRING"
        ColumnLength = "20"
      }]
      TableColumnsTotal = "2"
    }]
  })

  add_column_name                             = true
  canned_acl_for_objects                      = "private"
  cdc_inserts_and_updates                     = true
  cdc_inserts_only                            = false
  cdc_max_batch_interval                      = 100
  cdc_min_file_size                           = 16
  csv_null_value                              = "?"
  data_page_size                              = 1100000
  date_partition_enabled                      = true
  dict_page_size_limit                        = 1000000
  enable_statistics                           = false
  encoding_type                               = "plain"
  expected_bucket_owner                       = data.aws_caller_identity.current.account_id
  include_op_for_full_load                    = true
  max_file_size                               = 1000000
  row_group_length                            = 11000
  timestamp_column_name                       = "tx_commit_time"
  use_task_start_time_for_full_load_timestamp = true

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccS3EndpointConfig_sourceUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccS3EndpointConfig_base(rName),
		fmt.Sprintf(`
resource "aws_dms_s3_endpoint" "test" {
  bucket_folder           = "folder2"
  bucket_name             = "beckut_name"
  cdc_path                = "cdc/path2"
  csv_delimiter           = ","
  csv_row_delimiter       = "\\n"
  endpoint_id             = %[1]q
  endpoint_type           = "source"
  ignore_header_rows      = 1
  rfc_4180                = true
  service_access_role_arn = aws_iam_role.test.arn

  external_table_definition = jsonencode({
    TableCount = 1
    Tables = [{
      TableName  = "employee"
      TablePath  = "hr/employee/"
      TableOwner = "hr"
      TableColumns = [{
        ColumnName     = "ID"
        ColumnType     = "INT8"
        ColumnNullable = "false"
        ColumnIsPk     = "true"
      }]
      TableColumnsTotal = "1"
    }]
  })

  add_column_name                             = false
  canned_acl_for_objects                      = "authenticated-read"
  cdc_inserts_and_updates                     = false
  cdc_inserts_only                            = true
  cdc_max_batch_interval                      = 101
  cdc_min_file_size                           = 17
  csv_null_value                              = "0"
  data_page_size                              = 1000000
  date_partition_enabled                      = false
  dict_page_size_limit                        = 830000
  enable_statistics                           = true
  encoding_type                               = "plain-dictionary"
  expected_bucket_owner                       = data.aws_caller_identity.current.account_id
  include_op_for_full_load                    = false
  max_file_size                               = 100
  row_group_length                            = 10000
  timestamp_column_name                       = "tx_commit_time2"
  use_task_start_time_for_full_load_timestamp = false

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccS3EndpointConfig_detachTargetOnLobLookupFailureParquet(rName string, cdcp string, dt bool) string {
	return acctest.ConfigCompose(
		testAccS3EndpointConfig_base(rName),
		fmt.Sprintf(`
resource "aws_dms_s3_endpoint" "test" {
  endpoint_id                                 = %[1]q
  endpoint_type                               = "target"
  bucket_name                                 = "beckut_name"
  cdc_path                                    = %[2]q
  detach_target_on_lob_lookup_failure_parquet = %[3]t
  service_access_role_arn                     = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy.test]
}
`, rName, cdcp, dt))
}
