---
subcategory: "Database Migration Service (DMS)"
layout: "aws"
page_title: "AWS: aws_dms_endpoint"
description: |-
  Provides a DMS (Data Migration Service) endpoint resource.
---

# Resource: aws_dms_endpoint

Provides a DMS (Data Migration Service) endpoint resource. DMS endpoints can be created, updated, deleted, and imported.

~> **Note:** All arguments including the password will be stored in the raw state as plain-text.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

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

The following arguments are supported:

* `certificate_arn` - (Optional, Default: empty string) The Amazon Resource Name (ARN) for the certificate.
* `database_name` - (Optional) The name of the endpoint database.
* `elasticsearch_settings` - (Optional) Configuration block with Elasticsearch settings. Detailed below.
* `endpoint_id` - (Required) The database endpoint identifier.

    - Must contain from 1 to 255 alphanumeric characters or hyphens.
    - Must begin with a letter
    - Must contain only ASCII letters, digits, and hyphens
    - Must not end with a hyphen
    - Must not contain two consecutive hyphens

* `endpoint_type` - (Required) The type of endpoint. Can be one of `source | target`.
* `engine_name` - (Required) The type of engine for the endpoint. Can be one of `aurora | aurora-postgresql| azuredb | db2 | docdb | dynamodb | elasticsearch | kafka | kinesis | mariadb | mongodb | mysql | oracle | postgres | redshift | s3 | sqlserver | sybase`.
* `extra_connection_attributes` - (Optional) Additional attributes associated with the connection. For available attributes see [Using Extra Connection Attributes with AWS Database Migration Service](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.PostgreSQL.html#CHAP_Source.PostgreSQL.ConnectionAttrib).
* `kafka_settings` - (Optional) Configuration block with Kafka settings. Detailed below.
* `kinesis_settings` - (Optional) Configuration block with Kinesis settings. Detailed below.
* `kms_key_arn` - (Required when `engine_name` is `mongodb`, optional otherwise) The Amazon Resource Name (ARN) for the KMS key that will be used to encrypt the connection parameters. If you do not specify a value for `kms_key_arn`, then AWS DMS will use your default encryption key. AWS KMS creates the default encryption key for your AWS account. Your AWS account has a different default encryption key for each AWS region.
* `mongodb_settings` - (Optional) Configuration block with MongoDB settings. Detailed below.
* `password` - (Optional) The password to be used to login to the endpoint database.
* `port` - (Optional) The port used by the endpoint database.
* `s3_settings` - (Optional) Configuration block with S3 settings. Detailed below.
* `secrets_manager_access_role_arn` - (Optional) Amazon Resource Name (ARN) of the IAM role that specifies AWS DMS as the trusted entity and has the required permissions to access the value in SecretsManagerSecret.
* `secrets_manager_arn` - (Optional) The full ARN, partial ARN, or friendly name of the SecretsManagerSecret that contains the endpoint connection details. Supported only for `engine_name` as `oracle` and `postgres`.
* `server_name` - (Optional) The host name of the server.
* `service_access_role` - (Optional) The Amazon Resource Name (ARN) used by the service access IAM role for dynamodb endpoints.
* `ssl_mode` - (Optional, Default: none) The SSL mode to use for the connection. Can be one of `none | require | verify-ca | verify-full`
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `username` - (Optional) The user name to be used to login to the endpoint database.

### elasticsearch_settings Arguments

-> Additional information can be found in the [Using Amazon Elasticsearch Service as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.Elasticsearch.html).

The `elasticsearch_settings` configuration block supports the following arguments:

* `endpoint_uri` - (Required) Endpoint for the Elasticsearch cluster.
* `error_retry_duration` - (Optional) Maximum number of seconds for which DMS retries failed API requests to the Elasticsearch cluster. Defaults to `300`.
* `full_load_error_percentage` - (Optional) Maximum percentage of records that can fail to be written before a full load operation stops. Defaults to `10`.
* `service_access_role_arn` - (Required) Amazon Resource Name (ARN) of the IAM Role with permissions to write to the Elasticsearch cluster.

### kafka_settings Arguments

-> Additional information can be found in the [Using Apache Kafka as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.Kafka.html).

The `kafka_settings` configuration block supports the following arguments:

* `broker` - (Required) Kafka broker location. Specify in the form broker-hostname-or-ip:port.
* `include_control_details` - (Optional) Shows detailed control information for table definition, column definition, and table and column changes in the Kafka message output. The default is `false`.
* `include_null_and_empty` - (Optional) Include NULL and empty columns for records migrated to the endpoint. The default is `false`.
* `include_partition_value` - (Optional) Shows the partition value within the Kafka message output unless the partition type is `schema-table-type`. The default is `false`.
* `include_table_alter_operations` - (Optional) Includes any data definition language (DDL) operations that change the table in the control data, such as `rename-table`, `drop-table`, `add-column`, `drop-column`, and `rename-column`. The default is `false`.
* `include_transaction_details` - (Optional) Provides detailed transaction information from the source database. This information includes a commit timestamp, a log position, and values for `transaction_id`, previous `transaction_id`, and `transaction_record_id` (the record offset within a transaction). The default is `false`.
* `message_format` - (Optional) The output format for the records created on the endpoint. The message format is `JSON` (default) or `JSON_UNFORMATTED` (a single line with no tab).
* `message_max_bytes` - (Optional) The maximum size in bytes for records created on the endpoint The default is `1,000,000`.
* `no_hex_prefix` - (Optional) Set this optional parameter to true to avoid adding a '0x' prefix to raw data in hexadecimal format. For example, by default, AWS DMS adds a '0x' prefix to the LOB column type in hexadecimal format moving from an Oracle source to a Kafka target. Use the `no_hex_prefix` endpoint setting to enable migration of RAW data type columns without adding the `'0x'` prefix.
* `partition_include_schema_table` - (Optional) Prefixes schema and table names to partition values, when the partition type is `primary-key-type`. Doing this increases data distribution among Kafka partitions. For example, suppose that a SysBench schema has thousands of tables and each table has only limited range for a primary key. In this case, the same primary key is sent from thousands of tables to the same partition, which causes throttling. The default is `false`.
* `sasl_password` - (Optional) The secure password you created when you first set up your MSK cluster to validate a client identity and make an encrypted connection between server and client using SASL-SSL authentication.
* `sasl_username` - (Optional) The secure user name you created when you first set up your MSK cluster to validate a client identity and make an encrypted connection between server and client using SASL-SSL authentication.
* `security_protocol` - (Optional) Set secure connection to a Kafka target endpoint using Transport Layer Security (TLS). Options include `ssl-encryption`, `ssl-authentication`, and `sasl-ssl`. `sasl-ssl` requires `sasl_username` and `sasl_password`.
* `ssl_ca_certificate_arn` - (Optional) The Amazon Resource Name (ARN) for the private certificate authority (CA) cert that AWS DMS uses to securely connect to your Kafka target endpoint.
* `ssl_client_certificate_arn` - (Optional) The Amazon Resource Name (ARN) of the client certificate used to securely connect to a Kafka target endpoint.
* `ssl_client_key_arn` - (Optional) The Amazon Resource Name (ARN) for the client private key used to securely connect to a Kafka target endpoint.
* `ssl_client_key_password` - (Optional) The password for the client private key used to securely connect to a Kafka target endpoint.
* `topic` - (Optional) Kafka topic for migration. Defaults to `kafka-default-topic`.

### kinesis_settings Arguments

-> Additional information can be found in the [Using Amazon Kinesis Data Streams as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.Kinesis.html).

The `kinesis_settings` configuration block supports the following arguments:

* `include_control_details` - (Optional) Shows detailed control information for table definition, column definition, and table and column changes in the Kinesis message output. The default is `false`.
* `include_null_and_empty` - (Optional) Include NULL and empty columns in the target. The default is `false`.
* `include_partition_value` - (Optional) Shows the partition value within the Kinesis message output, unless the partition type is schema-table-type. The default is `false`.
* `include_table_alter_operations` - (Optional) Includes any data definition language (DDL) operations that change the table in the control data. The default is `false`.
* `include_transaction_details` - (Optional) Provides detailed transaction information from the source database. The default is `false`.
* `message_format` - (Optional) Output format for the records created. Defaults to `json`. Valid values are `json` and `json_unformatted` (a single line with no tab).
* `partition_include_schema_table` - (Optional) Prefixes schema and table names to partition values, when the partition type is primary-key-type. The default is `false`.
* `service_access_role_arn` - (Optional) Amazon Resource Name (ARN) of the IAM Role with permissions to write to the Kinesis data stream.
* `stream_arn` - (Optional) Amazon Resource Name (ARN) of the Kinesis data stream.

### mongodb_settings Arguments

-> Additional information can be found in the [Using MongoDB as a Source for AWS DMS documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.MongoDB.html).

The `mongodb_settings` configuration block supports the following arguments:

* `auth_mechanism` - (Optional) Authentication mechanism to access the MongoDB source endpoint. Defaults to `default`.
* `auth_source` - (Optional) Authentication database name. Not used when `auth_type` is `no`. Defaults to `admin`.
* `auth_type` - (Optional) Authentication type to access the MongoDB source endpoint. Defaults to `password`.
* `docs_to_investigate` - (Optional) Number of documents to preview to determine the document organization. Use this setting when `nesting_level` is set to `one`. Defaults to `1000`.
* `extract_doc_id` - (Optional) Document ID. Use this setting when `nesting_level` is set to `none`. Defaults to `false`.
* `nesting_level` - (Optional) Specifies either document or table mode. Defaults to `none`. Valid values are `one` (table mode) and `none` (document mode).

### s3_settings Arguments

-> Additional information can be found in the [Using Amazon S3 as a Source for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.S3.html) and [Using Amazon S3 as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.S3.html).

The `s3_settings` configuration block supports the following arguments:

* `bucket_folder` - (Optional) S3 Bucket Object prefix.
* `bucket_name` - (Optional) S3 Bucket name.
* `compression_type` - (Optional) Set to compress target files. Defaults to `NONE`. Valid values are `GZIP` and `NONE`.
* `csv_delimiter` - (Optional) Delimiter used to separate columns in the source files. Defaults to `,`.
* `csv_row_delimiter` - (Optional) Delimiter used to separate rows in the source files. Defaults to `\n`.
* `data_format` - (Optional) The output format for the files that AWS DMS uses to create S3 objects. Defaults to `csv`. Valid values are `csv` and `parquet`.
* `date_partition_enabled` - (Optional) Partition S3 bucket folders based on transaction commit dates. Defaults to `false`.
* `encryption_mode` - (Optional) The server-side encryption mode that you want to encrypt your .csv or .parquet object files copied to S3. Defaults to `SSE_S3`. Valid values are `SSE_S3` and `SSE_KMS`.
* `external_table_definition` - (Optional) JSON document that describes how AWS DMS should interpret the data.
* `parquet_timestamp_in_millisecond` - (Optional) - Specifies the precision of any TIMESTAMP column values written to an S3 object file in .parquet format. Defaults to `false`.
* `parquet_version` - (Optional) The version of the .parquet file format. Defaults to `parquet-1-0`. Valid values are `parquet-1-0` and `parquet-2-0`.
* `server_side_encryption_kms_key_id` - (Optional) If you set encryptionMode to `SSE_KMS`, set this parameter to the Amazon Resource Name (ARN) for the AWS KMS key.
* `service_access_role_arn` - (Optional) Amazon Resource Name (ARN) of the IAM Role with permissions to read from or write to the S3 Bucket.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `endpoint_arn` - The Amazon Resource Name (ARN) for the endpoint.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Endpoints can be imported using the `endpoint_id`, e.g.,

```
$ terraform import aws_dms_endpoint.test test-dms-endpoint-tf
```
