---
subcategory: "DMS (Database Migration)"
layout: "aws"
page_title: "AWS: aws_dms_endpoint"
description: |-
  Provides a DMS (Data Migration Service) endpoint resource.
---

# Resource: aws_dms_endpoint

Provides a DMS (Data Migration Service) endpoint resource. DMS endpoints can be created, updated, deleted, and imported.

~> **Note:** All arguments including the password will be stored in the raw state as plain-text. [Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

~> **Note:** The `s3_settings` argument is deprecated, may not be maintained, and will be removed in a future version. Use the [`aws_dms_s3_endpoint`](/docs/providers/aws/r/dms_s3_endpoint.html) resource instead.

## Example Usage

```terraform
# Create a new endpoint
resource "aws_dms_endpoint" "test" {
  certificate_arn             = "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
  database_name               = "test"
  endpoint_id                 = "test-dms-endpoint-tf"
  endpoint_type               = "source"
  engine_name                 = "aurora"
  extra_connection_attributes = ""
  kms_key_arn                 = "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
  password                    = "test"
  port                        = 3306
  server_name                 = "test"
  ssl_mode                    = "none"

  tags = {
    Name = "test"
  }

  username = "test"
}
```

## Argument Reference

The following arguments are required:

* `endpoint_id` - (Required) Database endpoint identifier. Identifiers must contain from 1 to 255 alphanumeric characters or hyphens, begin with a letter, contain only ASCII letters, digits, and hyphens, not end with a hyphen, and not contain two consecutive hyphens.
* `endpoint_type` - (Required) Type of endpoint. Valid values are `source`, `target`.
* `engine_name` - (Required) Type of engine for the endpoint. Valid values are `aurora`, `aurora-postgresql`, `aurora-serverless`, `aurora-postgresql-serverless`,`azuredb`, `azure-sql-managed-instance`, `babelfish`, `db2`, `db2-zos`, `docdb`, `dynamodb`, `elasticsearch`, `kafka`, `kinesis`, `mariadb`, `mongodb`, `mysql`, `opensearch`, `oracle`, `postgres`, `redshift`,`redshift-serverless`, `s3`, `sqlserver`, `neptune` ,`sybase`. Please note that some of engine names are available only for `target` endpoint type (e.g. `redshift`).
* `kms_key_arn` - (Required when `engine_name` is `mongodb`, cannot be set when `engine_name` is `s3`, optional otherwise) ARN for the KMS key that will be used to encrypt the connection parameters. If you do not specify a value for `kms_key_arn`, then AWS DMS will use your default encryption key. AWS KMS creates the default encryption key for your AWS account. Your AWS account has a different default encryption key for each AWS region. To encrypt an S3 target with a KMS Key, use the parameter `s3_settings.server_side_encryption_kms_key_id`. When `engine_name` is `redshift`, `kms_key_arn` is the KMS Key for the Redshift target and the parameter `redshift_settings.server_side_encryption_kms_key_id` encrypts the S3 intermediate storage.

The following arguments are optional:

* `certificate_arn` - (Optional, Default: empty string) ARN for the certificate.
* `database_name` - (Optional) Name of the endpoint database.
* `elasticsearch_settings` - (Optional) Configuration block for OpenSearch settings. See below.
* `extra_connection_attributes` - (Optional) Additional attributes associated with the connection. For available attributes for a `source` Endpoint, see [Sources for data migration](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.html). For available attributes for a `target` Endpoint, see [Targets for data migration](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.html).
* `kafka_settings` - (Optional) Configuration block for Kafka settings. See below.
* `kinesis_settings` - (Optional) Configuration block for Kinesis settings. See below.
* `mongodb_settings` - (Optional) Configuration block for MongoDB settings. See below.
* `password` - (Optional) Password to be used to login to the endpoint database.
* `postgres_settings` - (Optional) Configuration block for Postgres settings. See below.
* `pause_replication_tasks` - (Optional) Whether to pause associated running replication tasks, regardless if they are managed by Terraform, prior to modifying the endpoint. Only tasks paused by the resource will be restarted after the modification completes. Default is `false`.
* `port` - (Optional) Port used by the endpoint database.
* `redshift_settings` - (Optional) Configuration block for Redshift settings. See below.
* `s3_settings` - (Optional) (**Deprecated**, use the [`aws_dms_s3_endpoint`](/docs/providers/aws/r/dms_s3_endpoint.html) resource instead) Configuration block for S3 settings. See below.
* `secrets_manager_access_role_arn` - (Optional) ARN of the IAM role that specifies AWS DMS as the trusted entity and has the required permissions to access the value in the Secrets Manager secret referred to by `secrets_manager_arn`. The role must allow the `iam:PassRole` action.

   ~> **Note:** You can specify one of two sets of values for these permissions. You can specify the values for this setting and `secrets_manager_arn`. Or you can specify clear-text values for `username`, `password` , `server_name`, and `port`. You can't specify both.

* `secrets_manager_arn` - (Optional) Full ARN, partial ARN, or friendly name of the Secrets Manager secret that contains the endpoint connection details. Supported only when `engine_name` is `aurora`, `aurora-postgresql`, `mariadb`, `mongodb`, `mysql`, `oracle`, `postgres`, `redshift`, or `sqlserver`.
* `server_name` - (Optional) Host name of the server.
* `service_access_role` - (Optional) ARN used by the service access IAM role for dynamodb endpoints.
* `ssl_mode` - (Optional, Default: `none`) SSL mode to use for the connection. Valid values are `none`, `require`, `verify-ca`, `verify-full`
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `username` - (Optional) User name to be used to login to the endpoint database.

### elasticsearch_settings

-> Additional information can be found in the [Using Amazon OpenSearch Service as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.Elasticsearch.html).

* `endpoint_uri` - (Required) Endpoint for the OpenSearch cluster.
* `error_retry_duration` - (Optional) Maximum number of seconds for which DMS retries failed API requests to the OpenSearch cluster. Default is `300`.
* `full_load_error_percentage` - (Optional) Maximum percentage of records that can fail to be written before a full load operation stops. Default is `10`.
* `service_access_role_arn` - (Required) ARN of the IAM Role with permissions to write to the OpenSearch cluster.
* `use_new_mapping_type` - (Optional) Enable to migrate documentation using the documentation type `_doc`. OpenSearch and an Elasticsearch clusters only support the _doc documentation type in versions 7.x and later. The default value is `false`.

### kafka_settings

-> Additional information can be found in the [Using Apache Kafka as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.Kafka.html).

* `broker` - (Required) Kafka broker location. Specify in the form broker-hostname-or-ip:port.
* `include_control_details` - (Optional) Shows detailed control information for table definition, column definition, and table and column changes in the Kafka message output. Default is `false`.
* `include_null_and_empty` - (Optional) Include NULL and empty columns for records migrated to the endpoint. Default is `false`.
* `include_partition_value` - (Optional) Shows the partition value within the Kafka message output unless the partition type is `schema-table-type`. Default is `false`.
* `include_table_alter_operations` - (Optional) Includes any data definition language (DDL) operations that change the table in the control data, such as `rename-table`, `drop-table`, `add-column`, `drop-column`, and `rename-column`. Default is `false`.
* `include_transaction_details` - (Optional) Provides detailed transaction information from the source database. This information includes a commit timestamp, a log position, and values for `transaction_id`, previous `transaction_id`, and `transaction_record_id` (the record offset within a transaction). Default is `false`.
* `message_format` - (Optional) Output format for the records created on the endpoint. Message format is `JSON` (default) or `JSON_UNFORMATTED` (a single line with no tab).
* `message_max_bytes` - (Optional) Maximum size in bytes for records created on the endpoint Default is `1,000,000`.
* `no_hex_prefix` - (Optional) Set this optional parameter to true to avoid adding a '0x' prefix to raw data in hexadecimal format. For example, by default, AWS DMS adds a '0x' prefix to the LOB column type in hexadecimal format moving from an Oracle source to a Kafka target. Use the `no_hex_prefix` endpoint setting to enable migration of RAW data type columns without adding the `'0x'` prefix.
* `partition_include_schema_table` - (Optional) Prefixes schema and table names to partition values, when the partition type is `primary-key-type`. Doing this increases data distribution among Kafka partitions. For example, suppose that a SysBench schema has thousands of tables and each table has only limited range for a primary key. In this case, the same primary key is sent from thousands of tables to the same partition, which causes throttling. Default is `false`.
* `sasl_mechanism` - (Optional) For SASL/SSL authentication, AWS DMS supports the `scram-sha-512` mechanism by default. AWS DMS versions 3.5.0 and later also support the PLAIN mechanism. To use the PLAIN mechanism, set this parameter to `plain`.
* `sasl_password` - (Optional) Secure password you created when you first set up your MSK cluster to validate a client identity and make an encrypted connection between server and client using SASL-SSL authentication.
* `sasl_username` - (Optional) Secure user name you created when you first set up your MSK cluster to validate a client identity and make an encrypted connection between server and client using SASL-SSL authentication.
* `security_protocol` - (Optional) Set secure connection to a Kafka target endpoint using Transport Layer Security (TLS). Options include `ssl-encryption`, `ssl-authentication`, and `sasl-ssl`. `sasl-ssl` requires `sasl_username` and `sasl_password`.
* `ssl_ca_certificate_arn` - (Optional) ARN for the private certificate authority (CA) cert that AWS DMS uses to securely connect to your Kafka target endpoint.
* `ssl_client_certificate_arn` - (Optional) ARN of the client certificate used to securely connect to a Kafka target endpoint.
* `ssl_client_key_arn` - (Optional) ARN for the client private key used to securely connect to a Kafka target endpoint.
* `ssl_client_key_password` - (Optional) Password for the client private key used to securely connect to a Kafka target endpoint.
* `topic` - (Optional) Kafka topic for migration. Default is `kafka-default-topic`.

### kinesis_settings

-> Additional information can be found in the [Using Amazon Kinesis Data Streams as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.Kinesis.html).

* `include_control_details` - (Optional) Shows detailed control information for table definition, column definition, and table and column changes in the Kinesis message output. Default is `false`.
* `include_null_and_empty` - (Optional) Include NULL and empty columns in the target. Default is `false`.
* `include_partition_value` - (Optional) Shows the partition value within the Kinesis message output, unless the partition type is schema-table-type. Default is `false`.
* `include_table_alter_operations` - (Optional) Includes any data definition language (DDL) operations that change the table in the control data. Default is `false`.
* `include_transaction_details` - (Optional) Provides detailed transaction information from the source database. Default is `false`.
* `message_format` - (Optional) Output format for the records created. Default is `json`. Valid values are `json` and `json-unformatted` (a single line with no tab).
* `partition_include_schema_table` - (Optional) Prefixes schema and table names to partition values, when the partition type is primary-key-type. Default is `false`.
* `service_access_role_arn` - (Optional) ARN of the IAM Role with permissions to write to the Kinesis data stream.
* `stream_arn` - (Optional) ARN of the Kinesis data stream.

### mongodb_settings

-> Additional information can be found in the [Using MongoDB as a Source for AWS DMS documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.MongoDB.html).

* `auth_mechanism` - (Optional) Authentication mechanism to access the MongoDB source endpoint. Default is `default`.
* `auth_source` - (Optional) Authentication database name. Not used when `auth_type` is `no`. Default is `admin`.
* `auth_type` - (Optional) Authentication type to access the MongoDB source endpoint. Default is `password`.
* `docs_to_investigate` - (Optional) Number of documents to preview to determine the document organization. Use this setting when `nesting_level` is set to `one`. Default is `1000`.
* `extract_doc_id` - (Optional) Document ID. Use this setting when `nesting_level` is set to `none`. Default is `false`.
* `nesting_level` - (Optional) Specifies either document or table mode. Default is `none`. Valid values are `one` (table mode) and `none` (document mode).

### postgres_settings

-> Additional information can be found in the [Using PostgreSQL as a Source for AWS DMS documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.PostgreSQL.html).

* `after_connect_script` - (Optional) For use with change data capture (CDC) only, this attribute has AWS DMS bypass foreign keys and user triggers to reduce the time it takes to bulk load data.
* `babelfish_database_name` - (Optional) The Babelfish for Aurora PostgreSQL database name for the endpoint.
* `capture_ddls` - (Optional) To capture DDL events, AWS DMS creates various artifacts in the PostgreSQL database when the task starts.
* `database_mode` - (Optional) Specifies the default behavior of the replication's handling of PostgreSQL- compatible endpoints that require some additional configuration, such as Babelfish endpoints.
* `ddl_artifacts_schema` - (Optional) Sets the schema in which the operational DDL database artifacts are created. Default is `public`.
* `execute_timeout` - (Optional) Sets the client statement timeout for the PostgreSQL instance, in seconds. Default value is `60`.
* `fail_tasks_on_lob_truncation` - (Optional) When set to `true`, this value causes a task to fail if the actual size of a LOB column is greater than the specified `LobMaxSize`. Default is `false`.
* `heartbeat_enable` - (Optional) The write-ahead log (WAL) heartbeat feature mimics a dummy transaction. By doing this, it prevents idle logical replication slots from holding onto old WAL logs, which can result in storage full situations on the source.
* `heartbeat_frequency` - (Optional) Sets the WAL heartbeat frequency (in minutes). Default value is `5`.
* `heartbeat_schema` - (Optional) Sets the schema in which the heartbeat artifacts are created. Default value is `public`.
* `map_boolean_as_boolean` - (Optional) You can use PostgreSQL endpoint settings to map a boolean as a boolean from your PostgreSQL source to a Amazon Redshift target. Default value is `false`.
* `map_jsonb_as_clob` - Optional When true, DMS migrates JSONB values as CLOB.
* `map_long_varchar_as` - Optional When true, DMS migrates LONG values as VARCHAR.
* `max_file_size` - (Optional) Specifies the maximum size (in KB) of any .csv file used to transfer data to PostgreSQL. Default is `32,768 KB`.
* `plugin_name` - (Optional) Specifies the plugin to use to create a replication slot. Valid values: `pglogical`, `test_decoding`.
* `slot_name` - (Optional) Sets the name of a previously created logical replication slot for a CDC load of the PostgreSQL source instance.

### redis_settings

-> Additional information can be found in the [Using Redis as a target for AWS Database Migration Service](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.Redis.html).

* `auth_password` - (Optional) The password provided with the auth-role and auth-token options of the AuthType setting for a Redis target endpoint.
* `auth_type` - (Required) The type of authentication to perform when connecting to a Redis target. Options include `none`, `auth-token`, and `auth-role`. The `auth-token` option requires an `auth_password` value to be provided. The `auth-role` option requires `auth_user_name` and `auth_password` values to be provided.
* `auth_user_name` - (Optional) The username provided with the `auth-role` option of the AuthType setting for a Redis target endpoint.
* `server_name` - (Required) Fully qualified domain name of the endpoint.
* `port` - (Required) Transmission Control Protocol (TCP) port for the endpoint.
* `ssl_ca_certificate_arn` - (Optional) The Amazon Resource Name (ARN) for the certificate authority (CA) that DMS uses to connect to your Redis target endpoint.
* `ssl_security_protocol`- (Optional) The plaintext option doesn't provide Transport Layer Security (TLS) encryption for traffic between endpoint and database. Options include `plaintext`, `ssl-encryption`. The default is `ssl-encryption`.

### redshift_settings

-> Additional information can be found in the [Using Amazon Redshift as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.Redshift.html).

* `bucket_folder` - (Optional) Custom S3 Bucket Object prefix for intermediate storage.
* `bucket_name` - (Optional) Custom S3 Bucket name for intermediate storage.
* `encryption_mode` - (Optional) The server-side encryption mode that you want to encrypt your intermediate .csv object files copied to S3. Defaults to `SSE_S3`. Valid values are `SSE_S3` and `SSE_KMS`.
* `server_side_encryption_kms_key_id` - (Required when `encryption_mode` is  `SSE_KMS`, must not be set otherwise) ARN or Id of KMS Key to use when `encryption_mode` is `SSE_KMS`.
* `service_access_role_arn` - (Optional) Amazon Resource Name (ARN) of the IAM Role with permissions to read from or write to the S3 Bucket for intermediate storage.

### s3_settings

~> **Deprecated:** This argument is deprecated, may not be maintained, and will be removed in a future version. Use the [`aws_dms_s3_endpoint`](/docs/providers/aws/r/dms_s3_endpoint.html) resource instead.

-> Additional information can be found in the [Using Amazon S3 as a Source for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.S3.html) and [Using Amazon S3 as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.S3.html).

* `add_column_name` - (Optional) Whether to add column name information to the .csv output file. Default is `false`.
* `bucket_folder` - (Optional) S3 object prefix.
* `bucket_name` - (Optional) S3 bucket name.
* `canned_acl_for_objects` - (Optional) Predefined (canned) access control list for objects created in an S3 bucket. Valid values include `none`, `private`, `public-read`, `public-read-write`, `authenticated-read`, `aws-exec-read`, `bucket-owner-read`, and `bucket-owner-full-control`. Default is `none`.
* `cdc_inserts_and_updates` - (Optional) Whether to write insert and update operations to .csv or .parquet output files. Default is `false`.
* `cdc_inserts_only` - (Optional) Whether to write insert operations to .csv or .parquet output files. Default is `false`.
* `cdc_max_batch_interval` - (Optional) Maximum length of the interval, defined in seconds, after which to output a file to Amazon S3. Default is `60`.
* `cdc_min_file_size` - (Optional) Minimum file size condition as defined in kilobytes to output a file to Amazon S3. Default is `32000`. **NOTE:** Previously, this setting was measured in megabytes but now represents kilobytes. Update configurations accordingly.
* `cdc_path` - (Optional) Folder path of CDC files. For an S3 source, this setting is required if a task captures change data; otherwise, it's optional. If `cdc_path` is set, AWS DMS reads CDC files from this path and replicates the data changes to the target endpoint. Supported in AWS DMS versions 3.4.2 and later.
* `compression_type` - (Optional) Set to compress target files. Default is `NONE`. Valid values are `GZIP` and `NONE`.
* `csv_delimiter` - (Optional) Delimiter used to separate columns in the source files. Default is `,`.
* `csv_no_sup_value` - (Optional) String to use for all columns not included in the supplemental log.
* `csv_null_value` - (Optional) String to as null when writing to the target.
* `csv_row_delimiter` - (Optional) Delimiter used to separate rows in the source files. Default is `\n`.
* `data_format` - (Optional) Output format for the files that AWS DMS uses to create S3 objects. Valid values are `csv` and `parquet`. Default is `csv`.
* `data_page_size` - (Optional) Size of one data page in bytes. Default is `1048576` (1 MiB).
* `date_partition_delimiter` - (Optional) Date separating delimiter to use during folder partitioning. Valid values are `SLASH`, `UNDERSCORE`, `DASH`, and `NONE`. Default is `SLASH`.
* `date_partition_enabled` - (Optional) Partition S3 bucket folders based on transaction commit dates. Default is `false`.
* `date_partition_sequence` - (Optional) Date format to use during folder partitioning. Use this parameter when `date_partition_enabled` is set to true. Valid values are `YYYYMMDD`, `YYYYMMDDHH`, `YYYYMM`, `MMYYYYDD`, and `DDMMYYYY`. Default is `YYYYMMDD`.
* `dict_page_size_limit` - (Optional) Maximum size in bytes of an encoded dictionary page of a column. Default is `1048576` (1 MiB).
* `enable_statistics` - (Optional) Whether to enable statistics for Parquet pages and row groups. Default is `true`.
* `encoding_type` - (Optional) Type of encoding to use. Value values are `rle_dictionary`, `plain`, and `plain_dictionary`. Default is `rle_dictionary`.
* `encryption_mode` - (Optional) Server-side encryption mode that you want to encrypt your .csv or .parquet object files copied to S3. Valid values are `SSE_S3` and `SSE_KMS`. Default is `SSE_S3`.
* `external_table_definition` - (Optional) JSON document that describes how AWS DMS should interpret the data.
* `glue_catalog_generation` - (Optional) Whether to integrate AWS Glue Data Catalog with an Amazon S3 target. See [Using AWS Glue Data Catalog with an Amazon S3 target for AWS DMS](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.S3.html#CHAP_Target.S3.GlueCatalog) for more information. Default is `false`.
* `ignore_header_rows` - (Optional) When this value is set to `1`, DMS ignores the first row header in a .csv file. Default is `0`.
* `include_op_for_full_load` - (Optional) Whether to enable a full load to write INSERT operations to the .csv output files only to indicate how the rows were added to the source database. Default is `false`.
* `max_file_size` - (Optional) Maximum size (in KB) of any .csv file to be created while migrating to an S3 target during full load. Valid values are from `1` to `1048576`. Default is `1048576` (1 GB).
* `parquet_timestamp_in_millisecond` - (Optional) - Specifies the precision of any TIMESTAMP column values written to an S3 object file in .parquet format. Default is `false`.
* `parquet_version` - (Optional) Version of the .parquet file format. Default is `parquet-1-0`. Valid values are `parquet-1-0` and `parquet-2-0`.
* `preserve_transactions` - (Optional) Whether DMS saves the transaction order for a CDC load on the S3 target specified by `cdc_path`. Default is `false`.
* `rfc_4180` - (Optional) For an S3 source, whether each leading double quotation mark has to be followed by an ending double quotation mark. Default is `true`.
* `row_group_length` - (Optional) Number of rows in a row group. Default is `10000`.
* `server_side_encryption_kms_key_id` - (Required when `encryption_mode` is  `SSE_KMS`, must not be set otherwise) ARN or Id of KMS Key to use when `encryption_mode` is `SSE_KMS`.
* `service_access_role_arn` - (Optional) ARN of the IAM Role with permissions to read from or write to the S3 Bucket.
* `timestamp_column_name` - (Optional) Column to add with timestamp information to the endpoint data for an Amazon S3 target.
* `use_csv_no_sup_value` - (Optional) Whether to use `csv_no_sup_value` for columns not included in the supplemental log.
* `use_task_start_time_for_full_load_timestamp` - (Optional) When set to true, uses the task start time as the timestamp column value instead of the time data is written to target. For full load, when set to true, each row of the timestamp column contains the task start time. For CDC loads, each row of the timestamp column contains the transaction commit time. When set to false, the full load timestamp in the timestamp column increments with the time data arrives at the target. Default is `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `endpoint_arn` - ARN for the endpoint.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import endpoints using the `endpoint_id`. For example:

```terraform
import {
  to = aws_dms_endpoint.test
  id = "test-dms-endpoint-tf"
}
```

Using `terraform import`, import endpoints using the `endpoint_id`. For example:

```console
% terraform import aws_dms_endpoint.test test-dms-endpoint-tf
```
