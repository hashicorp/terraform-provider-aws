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

~> **Note:** We discourage using `extra_connection_attributes`. AWS is planning to **deprecate** this argument. Use the `<engine>_settings` arguments (_e.g._, `s3_settings`) instead. If you do use `extra_connection_attributes`, we recommend that you duplicate any settings in the `<engine>_settings` arguments to avoid Terraform reporting incorrect differences.

## Example Usage

### Basic Example

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

### S3 Example

AWS is planning to deprecate `extra_connection_attributes` so we recommend using `s3_settings` instead as shown below.

```terraform
resource "aws_dms_endpoint" "example" {
  endpoint_type = "target"
  engine_name   = "s3"
  endpoint_id   = "milajovovich"

  s3_settings {
    bucket_folder                    = "some/prefix"
    bucket_name                      = aws_s3_bucket.test.id
    cdc_path                         = "cdc"
    data_format                      = "parquet"
    date_partition_enabled           = true   
    include_op_for_full_load         = true
    parquet_timestamp_in_millisecond = true
    parquet_version                  = "parquet-2-0"
    service_access_role_arn          = aws_iam_role.example.arn
    timestamp_column_name            = "timestamp"
  }
}
```

## Argument Reference

The following arguments are required:

* `endpoint_id` - (Required) Database endpoint identifier. Identifiers must contain from 1 to 255 alphanumeric characters or hyphens, begin with a letter, contain only ASCII letters, digits, and hyphens, not end with a hyphen, and not contain two consecutive hyphens.
* `endpoint_type` - (Required) Type of endpoint. Valid values are `source`, `target`.
* `engine_name` - (Required) Type of engine for the endpoint. Valid values are `aurora`, `aurora-postgresql`, `azuredb`, `db2`, `docdb`, `dynamodb`, `elasticsearch`, `kafka`, `kinesis`, `mariadb`, `mongodb`, `mysql`, `opensearch`, `oracle`, `postgres`, `redshift`, `s3`, `sqlserver`, `sybase`.
* `kms_key_arn` - (Required when `engine_name` is `mongodb`, optional otherwise) ARN for the KMS key that will be used to encrypt the connection parameters. If you do not specify a value for `kms_key_arn`, then AWS DMS will use your default encryption key. AWS KMS creates the default encryption key for your AWS account. Your AWS account has a different default encryption key for each AWS region.

The following arguments are optional:

* `certificate_arn` - (Optional, Default: empty string) ARN for the certificate.
* `database_name` - (Optional) Name of the endpoint database.
* `dms_transfer_settings` - (Optional) Configuration block for DMS transfer type of source endpoint. See below.
* `docdb_settings` - (Optional) Configuration block for DocDB settings. See below.
* `elasticsearch_settings` - (Optional) Configuration block for OpenSearch settings. See below.
* `external_table_definition` - (Optional) External table definition.
* `extra_connection_attributes` - (Optional) AWS is planning to **deprecate** this argument. Use the `<engine>_settings` arguments (_e.g._, `s3_settings`) instead. Additional attributes associated with the connection. For available attributes see your target type at [Targets for data migration](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.html).
* `gcp_mysql_settings` - (Optional) Configuration block for GCP MySQL settings. See below.
* `ibm_db2_settings` - (Optional) Configuration block for IBM Db2 settings. See below.
* `kafka_settings` - (Optional) Configuration block for Kafka settings. See below.
* `kinesis_settings` - (Optional) Configuration block for Kinesis settings. See below.
* `microsoft_sql_server_settings` - (Optional) Configuration block for Microsoft SQL Server settings. See below.
* `mongodb_settings` - (Optional) Configuration block for MongoDB settings. See below.
* `mysql_settings` - (Optional) Configuration block for MySQL settings. See below.
* `neptune_settings` - (Optional) Configuration block for Neptune settings. See below.
* `oracle_settings` - (Optional) Configuration block for Oracle settings. See below.
* `password` - (Optional) Password to be used to login to the endpoint database.
* `port` - (Optional) Port used by the endpoint database.
* `postgresql_settings` - (Optional) Configuration block for PostgreSQL settings. See below.
* `redis_settings` - (Optional) Configuration block for PostgreSQL settings. See below.
* `redshift_settings` - (Optional) Configuration block for PostgreSQL settings. See below.
* `s3_settings` - (Optional) Configuration block for S3 settings. See below.
* `secrets_manager_access_role_arn` - (Optional) ARN of the IAM role that specifies AWS DMS as the trusted entity and has the required permissions to access the value in SecretsManagerSecret.
* `secrets_manager_arn` - (Optional) Full ARN, partial ARN, or friendly name of the SecretsManagerSecret that contains the endpoint connection details. Supported only for `engine_name` as `oracle` and `postgres`.
* `server_name` - (Optional) Host name of the server.
* `service_access_role` - (Optional) ARN used by the service access IAM role for dynamodb endpoints.
* `ssl_mode` - (Optional, Default: none) SSL mode to use for the connection. Valid values are `none`, `require`, `verify-ca`, `verify-full`
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `username` - (Optional) User name to be used to login to the endpoint database.

### dms_transfer_settings

* `bucket_name` - (Optional) Name of the S3 bucket to use.
* `service_access_role_arn` - (Optional) ARN used by the service access IAM role. The role must allow the `iam:PassRole` action`.

### docdb_settings

* `docs_to_investigate` - (Optional) Number of documents to preview to determine the document organization. Use this setting when `nesting_level` is `one`. Valid values are integers greater than 0. Default is `1000`.
* `extract_doc_id` - (Optional) Whether to extract the document ID. Use this setting when `nesting_level` is `none`. Default is `false`.
* `nesting_level` - (Optional) Either document (`none`) or table (`one`) mode. Default is `none`.

### elasticsearch_settings

-> Additional information can be found in the [Using Amazon OpenSearch Service as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.Elasticsearch.html).

* `endpoint_uri` - (Required) Endpoint for the OpenSearch cluster.
* `error_retry_duration` - (Optional) Maximum number of seconds for which DMS retries failed API requests to the OpenSearch cluster. Default is `300`.
* `full_load_error_percentage` - (Optional) Maximum percentage of records that can fail to be written before a full load operation stops. Default is `10`.
* `service_access_role_arn` - (Required) ARN of the IAM Role with permissions to write to the OpenSearch cluster.

### gcp_mysql_settings

-> Additional information can be found in the [Using Apache Kafka as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.Kafka.html).

* `after_connect_script` - (Optional) Script to run immediately after AWS DMS connects to the endpoint. The migration task continues running regardless if the SQL statement succeeds or fails. Provide the code of the script itself, not the name of a file containing the script.
* `clean_source_metadata_on_mismatch` - (Optional) Whether to adjust the behavior of AWS DMS when migrating from an SQL Server source database that is hosted as part of an Always On availability group cluster. If you need AWS DMS to poll all the nodes in the Always On cluster for transaction backups, set this attribute to `false`.
* `events_poll_interval` - (Optional) How often (in seconds) to check the binary log for new changes/events when the database is idle. Default is `5`.
* `max_file_size` - (Optional) Maximum size (in KB) of any .csv file used to transfer data to a MySQL-compatible database.
* `parallel_load_threads` - (Optional) How many threads to use to load the data into the MySQL-compatible target database. Setting a large number of threads can have an adverse effect on database performance, because a separate connection is required for each thread. Default is `1`.
* `server_timezone` - (Optional) Time zone for the source MySQL database. _E.g._, `US/Pacific`.
* `target_db_type` - (Optional) Where to migrate source tables on the target, either to a single database or multiple databases. Valid values are `specific-database`, `multiple-databases`.

### ibm_db2_settings

-> Additional information can be found in the [Using Apache Kafka as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.Kafka.html).

* `current_lsn` - (Optional) For ongoing replication (CDC), a log sequence number (LSN) where you want the replication to start.
* `max_k_bytes_per_read` - (Optional) Maximum number of bytes per read. The default is `64`.
* `set_data_capture_changes` - (Optional) Whether to enables ongoing replication (CDC). The default is `true`.

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
* `message_format` - (Optional) Output format for the records created. Default is `json`. Valid values are `json` and `json_unformatted` (a single line with no tab).
* `no_hex_prefix` - (Optional) Set this optional parameter to true to avoid adding a '0x' prefix to raw data in hexadecimal format. For example, by default, AWS DMS adds a '0x' prefix to the LOB column type in hexadecimal format moving from an Oracle source to a Kinesis target. Use the `no_hex_prefix` endpoint setting to enable migration of RAW data type columns without adding the `'0x'` prefix.
* `partition_include_schema_table` - (Optional) Prefixes schema and table names to partition values, when the partition type is primary-key-type. Default is `false`.
* `service_access_role_arn` - (Optional) ARN of the IAM Role with permissions to write to the Kinesis data stream.
* `stream_arn` - (Optional) ARN of the Kinesis data stream.

### microsoft_sql_server_settings

-> Additional information can be found in the [Using MongoDB as a Source for AWS DMS documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.MongoDB.html).

* `bcp_packet_size` - (Optional) Maximum size of the packets (in bytes) used to transfer data using BCP.
* `control_tables_file_group` - (Optional) File group for the AWS DMS internal tables. When the replication task starts, all the internal AWS DMS control tables (awsdms_ apply_exception, awsdms_apply, awsdms_changes) are created for the specified file group.
* `query_string_always_on_mode` - (Optional) Cleans and recreates table metadata information on the replication instance when a mismatch occurs. An example is a situation where running an alter DDL statement on a table might result in different information about the table cached in the replication instance.
* `read_backup_only` - (Optional) Whether AWS DMS only reads changes from transaction log backups and doesn't read from the active transaction log file during ongoing replication. Setting this parameter to `true` enables you to control active transaction log file growth during full load and ongoing replication tasks. However, it can add some source latency to ongoing replication.
* `safeguard_policy` - (Optional) Use this attribute to minimize the need to access the backup log and enable AWS DMS to prevent truncation using one of the following two methods. Valid values are `rely-on-sql-server-replication-agent`, `exclusive-automatic-truncation`, `shared-automatic-truncation`. Default is `rely-on-sql-server-replication-agent`.
* `use_bcp_full_load` - (Optional) ARN of the Kinesis data stream.
* `use_third_party_backup_device` - (Optional) ARN of the Kinesis data stream.

### mongodb_settings

-> Additional information can be found in the [Using MongoDB as a Source for AWS DMS documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.MongoDB.html).

* `auth_mechanism` - (Optional) Authentication mechanism to access the MongoDB source endpoint. Default is `default`.
* `auth_source` - (Optional) Authentication database name. Not used when `auth_type` is `no`. Default is `admin`.
* `auth_type` - (Optional) Authentication type to access the MongoDB source endpoint. Default is `password`.
* `docs_to_investigate` - (Optional) Number of documents to preview to determine the document organization. Use this setting when `nesting_level` is set to `one`. Default is `1000`.
* `extract_doc_id` - (Optional) Document ID. Use this setting when `nesting_level` is set to `none`. Default is `false`.
* `nesting_level` - (Optional) Specifies either document or table mode. Default is `none`. Valid values are `one` (table mode) and `none` (document mode).

### mysql_settings

-> Additional information can be found in the [Using Apache Kafka as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.Kafka.html).

* `after_connect_script` - (Optional) Script to run immediately after AWS DMS connects to the endpoint. The migration task continues running regardless if the SQL statement succeeds or fails. Provide the code of the script itself, not the name of a file containing the script.
* `clean_source_metadata_on_mismatch` - (Optional) Whether to adjust the behavior of AWS DMS when migrating from an SQL Server source database that is hosted as part of an Always On availability group cluster. If you need AWS DMS to poll all the nodes in the Always On cluster for transaction backups, set this attribute to `false`.
* `events_poll_interval` - (Optional) How often (in seconds) to check the binary log for new changes/events when the database is idle. Default is `5`.
* `max_file_size` - (Optional) Maximum size (in KB) of any .csv file used to transfer data to a MySQL-compatible database.
* `parallel_load_threads` - (Optional) How many threads to use to load the data into the MySQL-compatible target database. Setting a large number of threads can have an adverse effect on database performance, because a separate connection is required for each thread. Default is `1`.
* `server_timezone` - (Optional) Time zone for the source MySQL database. _E.g._, `US/Pacific`.
* `target_db_type` - (Optional) Where to migrate source tables on the target, either to a single database or multiple databases. Valid values are `specific-database`, `multiple-databases`.

### neptune_settings

-> Additional information can be found in the [Using Apache Kafka as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.Kafka.html).

* `error_retry_duration` - (Optional) Number of milliseconds for AWS DMS to wait to retry a bulk-load of migrated graph data to the Neptune target database before raising an error. The default is `250`.
* `iam_auth_enabled` - (Optional) Whether to enable IAM authorization for this endpoint. Then attach the appropriate IAM policy document to your service role specified by `service_access_role_arn`. Default is `false`.
* `max_file_size` - (Optional) Maximum size (in KB) of migrated graph data stored in a .csv file before AWS DMS bulk-loads the data to the Neptune target database. The default is `1048576` (KB). If the bulk load is successful, AWS DMS clears the bucket, ready to store the next batch of migrated graph data.
* `max_retry_count` - (Optional) Number of times for AWS DMS to retry a bulk load of migrated graph data to the Neptune target database before raising an error. The default is `5`.
* `s3_bucket_folder` - (Required) Folder path where you want AWS DMS to store migrated graph data in the S3 bucket specified by `s3_bucket_name`.
* `s3_bucket_name` - (Required) Name of the S3 bucket where AWS DMS can temporarily store migrated graph data in .csv files before bulk-loading it to the Neptune target database. AWS DMS maps the SQL source data to graph data before storing it in these .csv files.
* `service_access_role_arn` - (Optional) ARN of the service role that you created for the Neptune target endpoint. The role must allow the `iam:PassRole` action.

### oracle_settings

-> Additional information can be found in the [Using Apache Kafka as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.Kafka.html).

* `access_alternate_directly` - (Optional) Set to `false` in order to use the Binary Reader to capture change data for an Amazon RDS for Oracle as the source. This tells the DMS instance to not access redo logs through any specified path prefix replacement using direct file access.
* `additional_archived_log_dest_id` - (Optional) Use with `archived_log_dest_id` in a primary/standby setup. This attribute is useful in the case of a switchover. In this case, AWS DMS needs to know which destination to get archive redo logs from to read changes. This need arises because the previous primary instance is now a standby instance after switchover.
* `add_supplemental_logging` - (Optional) Set up table-level supplemental logging for the Oracle database. This enables PRIMARY KEY supplemental logging on all tables selected for a migration task. If you use this option, you still need to enable database-level supplemental logging.
* `allow_select_nested_tables` - (Optional) Whether to enable replication of Oracle tables containing columns that are nested tables or defined types.
* `archived_log_dest_id` - (Optional) ID of the destination for the archived redo logs. This value should be the same as a number in the `dest_id` column of the `v$archived_log view`. If you work with an additional redo log destination, use the `additional_archived_log_dest_id` option to specify the additional destination ID. Doing this improves performance by ensuring that the correct logs are accessed from the outset.
* `archived_logs_only` - (Optional) When this field is set to `true`, AWS DMS only accesses the archived redo logs. If the archived redo logs are stored on Oracle ASM only, the AWS DMS user account needs to be granted ASM privileges.
* `asm_password` - (Optional) ASM password. Conflicts with `secrets_manager_oracle_asm_access_role_arn` and `secrets_manager_oracle_asm_secret_id`.
* `asm_server` - (Optional) ASM server address. Conflicts with `secrets_manager_oracle_asm_access_role_arn` and `secrets_manager_oracle_asm_secret_id`.
* `asm_user` - (Optional) ASM user name. Conflicts with `secrets_manager_oracle_asm_access_role_arn` and `secrets_manager_oracle_asm_secret_id`.
* `char_length_semantics` - (Optional) Length of a character column. `char` indicates the character column length is in characters. Otherwise, the character column length is in bytes. Valid values `default`, `char`, `byte`. Default is `default`.
* `direct_path_no_log` - (Optional) When set to `true`, this attribute helps to increase the commit rate on the Oracle target database by writing directly to tables and not writing a trail to database logs.
* `direct_path_parallel_load` - (Optional) Whether, when `use_direct_path_full_load` is `true`, to parallel load. This attribute also only applies when you use the AWS DMS parallel load feature. Note that the target table cannot have any constraints or indexes.
* `enable_homogenous_tablespace` - (Optional) Whether to enable homogenous tablespace replication and create existing tables or indexes under the same tablespace on the target.
* `extra_archived_log_dest_ids` - (Optional) List of IDs of one or more destinations for one or more archived redo logs. These IDs are the values of the `dest_id` column in the `v$archived_log` view. Use this setting with the `archived_log_dest_id` attribute in a primary-to-single setup or a primary-to-multiple-standby setup. For example, in a primary-to-single standby setup you might apply these settings: `archived_Log_Dest_Id = 1` and `extra_archived_log_dest_ids = [2]`. In a primary-to-multiple-standby setup, you might apply these settings: `archived_log_dest_id = 1` and `extra_archived_log_dest_ids = [2,3,4]`.
* `fail_tasks_on_lob_truncation` - (Optional) Whether to fail if the actual size of an LOB column is greater than LobMaxSize.
* `number_datatype_scale` - (Optional) Number scale. You can select a scale up to `38`. By default, the NUMBER data type is converted to precision 38, scale 10.
* `oracle_path_prefix` - (Optional) Default Oracle root used to access the redo logs. Set this in order to use the Binary Reader to capture change data for an Amazon RDS for Oracle as the source.
* `parallel_asm_read_threads` - (Optional) Number of threads that DMS configures to perform a change data capture (CDC) load using Oracle Automatic Storage Management (ASM). Valid values are between `2`-`8`. Default is `2`. Use this attribute together with the `read_ahead_blocks` attribute.
* `read_ahead_blocks` - (Optional) Number of read-ahead blocks that DMS configures to perform a change data capture (CDC) load using Oracle Automatic Storage Management (ASM). Valid values are `1000`-`200000`.
* `read_table_space_name` - (Optional) Whether to support tablespace replication.
* `replace_path_prefix` - (Optional) Whether to use the Binary Reader to capture change data for an Amazon RDS for Oracle as the source. This setting tells DMS instance to replace the default Oracle root with the specified `use_path_prefix` setting to access the redo logs.
* `retry_interval` - (Optional) Number of seconds that the system waits before resending a query.
* `secrets_manager_oracle_asm_access_role_arn` - (Optional) Required only if your Oracle endpoint uses ASM. The full ARN of the IAM role that specifies AWS DMS as the trusted entity and grants the required permissions to access the `secrets_manager_oracle_asm_secret`. This `secrets_manager_oracle_asm_secret` has the secret value that allows access to the Oracle ASM of the endpoint. Conflicts with `asm_user_name`, `asm_password`, and `asm_server_name`.
* `secrets_manager_oracle_asm_secret_id` - (Optional) Required only if your Oracle endpoint uses ASM. The full ARN, partial ARN, or friendly name of the `secrets_manager_oracle_asm_secret` that contains the Oracle ASM connection details for the Oracle endpoint. Conflicts with `asm_user_name`, `asm_password`, and `asm_server_name`.
* `security_db_encryption` - (Optional) Transparent data encryption (TDE) password required by AWM DMS to access Oracle redo logs encrypted by TDE using Binary Reader. It is also the TDE_Password part of the comma-separated value you set to the Password request parameter when you create the endpoint. The `security_db_encryptian` setting is related to this `security_db_encryption_name` setting.
* `security_db_encryption_name` - (Optional) Name of a key used for the transparent data encryption (TDE) of the columns and tablespaces in an Oracle source database that is encrypted using TDE. The key value is the value of the `security_db_encryption` setting.
* `spatial_data_option_to_geo_json_function_name` - (Optional) Custom function to convert SDO_GEOMETRY to GEOJSON format. By default, DMS calls the SDO2GEOJSON custom function if present and accessible.
* `standby_delay_time` - (Optional) Time in minutes for the delay in standby sync. If the source is an Oracle Active Data Guard standby database, use this attribute to specify the time lag between primary and standby databases. In AWS DMS, you can create an Oracle CDC task that uses an Active Data Guard standby instance as a source for replicating ongoing changes. Doing this eliminates the need to connect to an active database that might be in production.
* `use_alternate_folder_for_online` - (Optional) Whether to use any specified prefix replacement to access all online redo logs. Set to `true` in order to use the Binary Reader to capture change data for an Amazon RDS for Oracle as the source.
* `use_b_file` - (Optional) Whether to capture change data using the Binary Reader utility. Set `use_logminer_reader` to `false` to set this attribute to `true`.
* `use_direct_path_full_load` - (Optional) Whether to have AWS DMS use a direct path full load. Specify this value to use the direct path protocol in the Oracle Call Interface (OCI). By using this OCI protocol, you can bulk-load Oracle target tables during a full load.
* `use_logminer_reader` - (Optional) Whether to capture change data using the Oracle LogMiner utility (the default). Set to `false` if you want to access the redo logs as a binary file. When you set `use_logminer_reader` to `false`, also set `use_b_file` to `true`.
* `use_path_prefix` - (Optional) Path prefix used to replace the default Oracle root to access the redo logs. Set this in order to use the Binary Reader to capture change data for an Amazon RDS for Oracle as the source.

### postgresql_settings

-> Additional information can be found in the [Using Amazon S3 as a Source for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.S3.html) and [Using Amazon S3 as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.S3.html).

* `after_connect_script` - (Optional) For use with change data capture (CDC) only, this attribute has AWS DMS bypass foreign keys and user triggers to reduce the time it takes to bulk load data.
* `capture_ddls` - (Optional) To capture DDL events, AWS DMS creates various artifacts in the PostgreSQL database when the task starts. You can later remove these artifacts. If this value is set to `false`, you don't have to create tables or triggers on the source database.
* `ddl_artifacts_schema` - (Optional) Schema in which the operational DDL database artifacts are created.
* `execute_timeout` - (Optional) Client statement timeout in seconds. The default value is `60` seconds.
* `fail_tasks_on_lob_truncation` - (Optional) Whether a task fails if the actual size of a LOB column is greater than the specified LobMaxSize. If task is set to Limited LOB mode and this option is set to `true`, the task fails instead of truncating the LOB data.
* `heartbeat_enable` - (Optional) Write-ahead log (WAL) heartbeat feature mimics a dummy transaction. By doing this, it prevents idle logical replication slots from holding onto old WAL logs, which can result in storage full situations on the source. This heartbeat keeps restart_lsn moving and prevents storage full scenarios.
* `heartbeat_frequency` - (Optional) Sets the WAL heartbeat frequency (in minutes).
* `heartbeat_schema` - (Optional) Sets the schema in which the heartbeat artifacts are created.
* `max_file_size` - (Optional) Maximum size (in KB) of any .csv file used to transfer data to PostgreSQL.
* `plugin_name` - (Optional) Plugin to use to create a replication slot. Valid values are `no-preference`, `test-decoding`, `pglogical`.
* `slot_name` - (Optional) Name of a previously created logical replication slot for a change data capture (CDC) load of the PostgreSQL source instance. When used with the `cdc_start_position` replication task parameter, this attribute also makes it possible to use native CDC start points.

### redis_settings

-> Additional information can be found in the [Using Amazon S3 as a Source for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.S3.html) and [Using Amazon S3 as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.S3.html).

* `auth_password` - (Optional) Password provided with the `auth-role` and `auth-token` options of the `auth_type` argument.
* `auth_type` - (Optional) Type of authentication to perform. Valid values are `none`, `auth-token`, and `auth-role`. The `auth-token` option requires an `auth_password` value to be provided. The `auth-role` option requires `auth_user_name` and `auth_password` values to be provided.
* `auth_user_name` - (Optional) User name provided with the `auth-role` option of the `auth_type` setting 
* `ssl_ca_certificate_arn` - (Optional) ARN for the certificate authority (CA) that DMS uses to connect to your Redis target endpoint. If you don't provide an ARN and set `ssl_security_protocol = "ssl-encryption"`, DMS uses the Amazon root CA.
* `ssl_security_protocol` - (Optional) Type of connection to a Redis target endpoint. Valid values include `plaintext` and `ssl-encryption`. The default is `ssl-encryption`. The `ssl-encryption` option makes an encrypted connection. Optionally, you can identify an ARN for an SSL certificate authority (CA) using the `ssl_ca_certificate_arn` setting. The `plaintext` option doesn't provide TLS encryption for traffic between endpoint and database.

### s3_settings

-> Additional information can be found in the [Using Amazon S3 as a Source for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.S3.html) and [Using Amazon S3 as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.S3.html).

* `add_column_name` - (Optional) Whether to add column name information to the .csv output file. Default is `false`.
* `bucket_folder` - (Optional) S3 object prefix.
* `bucket_name` - (Optional) S3 bucket name.
* `canned_acl_for_objects` - (Optional) Predefined (canned) access control list for objects created in an S3 bucket. Valid values include `NONE`, `PRIVATE`, `PUBLIC_READ`, `PUBLIC_READ_WRITE`, `AUTHENTICATED_READ`, `AWS_EXEC_READ`, `BUCKET_OWNER_READ`, and `BUCKET_OWNER_FULL_CONTROL`. Default is `NONE`.
* `cdc_inserts_and_updates` - (Optional) Whether to write insert and update operations to .csv or .parquet output files. Default is `false`.
* `cdc_inserts_only` - (Optional) Whether to write insert operations to .csv or .parquet output files. Default is `false`.
* `cdc_max_batch_interval` - (Optional) Maximum length of the interval, defined in seconds, after which to output a file to Amazon S3. Default is `60`.
* `cdc_min_file_size` - (Optional) Minimum file size, defined in megabytes, to reach for a file output. Default is `32`.
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
* `date_partition_timezone` - (Optional) Convert the current UTC time into a time zone. The conversion occurs when a date partition folder is created and a CDC filename is generated. The time zone format is Area/Location (_e.g._, `Asia/Seoul`). Use this parameter when `date_partitioned_enabled` is set to `true`.
* `dict_page_size_limit` - (Optional) Maximum size in bytes of an encoded dictionary page of a column. Default is `1048576` (1 MiB).
* `enable_statistics` - (Optional) Whether to enable statistics for Parquet pages and row groups. Default is `true`.
* `encoding_type` - (Optional) Type of encoding to use. Value values are `rle_dictionary`, `plain`, and `plain_dictionary`. Default is `rle_dictionary`.
* `encryption_mode` - (Optional) Server-side encryption mode that you want to encrypt your .csv or .parquet object files copied to S3. Valid values are `SSE_S3` and `SSE_KMS`. Default is `SSE_S3`.
* `external_table_definition` - (Optional) JSON document that describes how AWS DMS should interpret the data.
* `ignore_headers_row` - (Optional) When this value is set to `1`, DMS ignores the first row header in a .csv file. Default is `0`.
* `include_op_for_full_load` - (Optional) Whether to enable a full load to write INSERT operations to the .csv output files only to indicate how the rows were added to the source database. Default is `false`.
* `max_file_size` - (Optional) Maximum size (in KB) of any .csv file to be created while migrating to an S3 target during full load. Valid values are from `1` to `1048576`. Default is `1048576` (1 GB).
* `parquet_timestamp_in_millisecond` - (Optional) - Specifies the precision of any TIMESTAMP column values written to an S3 object file in .parquet format. Default is `false`.
* `parquet_version` - (Optional) Version of the .parquet file format. Default is `parquet-1-0`. Valid values are `parquet-1-0` and `parquet-2-0`.
* `preserve_transactions` - (Optional) Whether DMS saves the transaction order for a CDC load on the S3 target specified by `cdc_path`. Default is `false`.
* `rfc_4180` - (Optional) For an S3 source, whether each leading double quotation mark has to be followed by an ending double quotation mark. Default is `true`.
* `row_group_length` - (Optional) Number of rows in a row group. Default is `10000`.
* `server_side_encryption_kms_key_id` - (Optional) If you set encryptionMode to `SSE_KMS`, set this parameter to the ARN for the AWS KMS key.
* `service_access_role_arn` - (Optional) ARN of the IAM Role with permissions to read from or write to the S3 Bucket.
* `timestamp_column_name` - (Optional) Column to add with timestamp information to the endpoint data for an Amazon S3 target.
* `use_csv_no_sup_value` - (Optional) Whether to use `csv_no_sup_value` for columns not included in the supplemental log.
* `use_task_start_time_for_full_load_timestamp` - (Optional) Whether to use the task start time as the timestamp column value instead of the time data is written to target. For full load, when `true`, each row of the timestamp column contains the task start time. For CDC loads, each row of the timestamp column contains the transaction commit time. When `false`, the full load timestamp in the timestamp column increments with the time data arrives at the target. Default is `false`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `endpoint_arn` - ARN for the endpoint.
* `engine_display_name` - Expanded name for the engine name. For example, if the `engine_name` parameter is `aurora`, this value would be `Amazon Aurora MySQL`.
* `external_id` - Identifier that can be used for cross-account validation.
* `status` - Status of the endpoint.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Endpoints can be imported using the `endpoint_id`, e.g.,

```
$ terraform import aws_dms_endpoint.test test-dms-endpoint-tf
```
