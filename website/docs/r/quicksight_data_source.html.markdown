---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_data_source"
description: |-
  Manages a Resource QuickSight Data Source.
---

# Resource: aws_quicksight_data_source

Resource for managing QuickSight Data Source

## Example Usage

```terraform
resource "aws_quicksight_data_source" "default" {
  data_source_id = "example-id"
  name           = "My Cool Data in S3"

  parameters {
    s3 {
      manifest_file_location {
        bucket = "my-bucket"
        key    = "path/to/manifest.json"
      }
    }
  }

  type = "S3"
}
```

## Argument Reference

The following arguments are required:

* `data_source_id` - (Required, Forces new resource) An identifier for the data source.
* `name` - (Required) A name for the data source, maximum of 128 characters.
* `parameters` - (Required) The [parameters](#parameters-argument-reference) used to connect to this data source (exactly one).
* `type` - (Required) The type of the data source. See the [AWS Documentation](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CreateDataSource.html#QS-CreateDataSource-request-Type) for the complete list of valid values.

The following arguments are optional:

* `aws_account_id` - (Optional, Forces new resource) The ID for the AWS account that the data source is in. Currently, you use the ID for the AWS account that contains your Amazon QuickSight account.
* `credentials` - (Optional) The credentials Amazon QuickSight uses to connect to your underlying source. Currently, only credentials based on user name and password are supported. See [Credentials](#credentials-argument-reference) below for more details.
* `permission` - (Optional) A set of resource permissions on the data source. Maximum of 64 items. See [Permission](#permission-argument-reference) below for more details.
* `ssl_properties` - (Optional) Secure Socket Layer (SSL) properties that apply when Amazon QuickSight connects to your underlying source. See [SSL Properties](#ssl_properties-argument-reference) below for more details.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_connection_properties`- (Optional) Use this parameter only when you want Amazon QuickSight to use a VPC connection when connecting to your underlying source. See [VPC Connection Properties](#vpc_connection_properties-argument-reference) below for more details.

### credentials Argument Reference

* `copy_source_arn` (Optional, Conflicts with `credential_pair`) - The Amazon Resource Name (ARN) of a data source that has the credential pair that you want to use.
When the value is not null, the `credential_pair` from the data source in the ARN is used.
* `credential_pair` (Optional, Conflicts with `copy_source_arn`) - Credential pair. See [Credential Pair](#credential_pair-argument-reference) below for more details.

### credential_pair Argument Reference

* `password` - (Required) Password, maximum length of 1024 characters.
* `username` - (Required) User name, maximum length of 64 characters.

### parameters Argument Reference

To specify data source connection parameters, exactly one of the following sub-objects must be provided.

* `amazon_elasticsearch` - (Optional) [Parameters](#amazon_elasticsearch-argument-reference) for connecting to Amazon Elasticsearch.
* `athena` - (Optional) [Parameters](#athena-argument-reference) for connecting to Athena.
* `aurora` - (Optional) [Parameters](#aurora-argument-reference) for connecting to Aurora MySQL.
* `aurora_postgresql` - (Optional) [Parameters](#aurora_postgresql-argument-reference) for connecting to Aurora Postgresql.
* `aws_iot_analytics` - (Optional) [Parameters](#aws_iot_analytics-argument-reference) for connecting to AWS IOT Analytics.
* `jira` - (Optional) [Parameters](#jira-fargument-reference) for connecting to Jira.
* `maria_db` - (Optional) [Parameters](#maria_db-argument-reference) for connecting to MariaDB.
* `mysql` - (Optional) [Parameters](#mysql-argument-reference) for connecting to MySQL.
* `oracle` - (Optional) [Parameters](#oracle-argument-reference) for connecting to Oracle.
* `postgresql` - (Optional) [Parameters](#postgresql-argument-reference) for connecting to Postgresql.
* `presto` - (Optional) [Parameters](#presto-argument-reference) for connecting to Presto.
* `rds` - (Optional) [Parameters](#rds-argument-reference) for connecting to RDS.
* `redshift` - (Optional) [Parameters](#redshift-argument-reference) for connecting to Redshift.
* `s3` - (Optional) [Parameters](#s3-argument-reference) for connecting to S3.
* `service_now` - (Optional) [Parameters](#service_now-argument-reference) for connecting to ServiceNow.
* `snowflake` - (Optional) [Parameters](#snowflake-argument-reference) for connecting to Snowflake.
* `spark` - (Optional) [Parameters](#spark-argument-reference) for connecting to Spark.
* `sql_server` - (Optional) [Parameters](#sql_server-argument-reference) for connecting to SQL Server.
* `teradata` - (Optional) [Parameters](#teradata-argument-reference) for connecting to Teradata.
* `twitter` - (Optional) [Parameters](#twitter-argument-reference) for connecting to Twitter.

### permission Argument Reference

* `actions` - (Required) Set of IAM actions to grant or revoke permissions on. Max of 16 items.
* `principal` - (Required) The Amazon Resource Name (ARN) of the principal.

### ssl_properties Argument Reference

* `disable_ssl` - (Required) A Boolean option to control whether SSL should be disabled.

### vpc_connection_properties Argument Reference

* `vpc_connection_arn` - (Required) The Amazon Resource Name (ARN) for the VPC connection.

### amazon_elasticsearch Argument Reference

* `domain` - (Required) The OpenSearch domain.

### athena Argument Reference

* `work_group` - (Optional) The work-group to which to connect.

### aurora Argument Reference

* `database` - (Required) The database to which to connect.
* `host` - (Required) The host to which to connect.
* `port` - (Required) The port to which to connect.

### aurora_postgresql Argument Reference

* `database` - (Required) The database to which to connect.
* `host` - (Required) The host to which to connect.
* `port` - (Required) The port to which to connect.

### aws_iot_analytics Argument Reference

* `data_set_name` - (Required) The name of the data set to which to connect.

### jira fArgument Reference

* `site_base_url` - (Required) The base URL of the Jira instance's site to which to connect.

### maria_db Argument Reference

* `database` - (Required) The database to which to connect.
* `host` - (Required) The host to which to connect.
* `port` - (Required) The port to which to connect.

### mysql Argument Reference

* `database` - (Required) The database to which to connect.
* `host` - (Required) The host to which to connect.
* `port` - (Required) The port to which to connect.

### oracle Argument Reference

* `database` - (Required) The database to which to connect.
* `host` - (Required) The host to which to connect.
* `port` - (Required) The port to which to connect.

### postgresql Argument Reference

* `database` - (Required) The database to which to connect.
* `host` - (Required) The host to which to connect.
* `port` - (Required) The port to which to connect.

### presto Argument Reference

* `catalog` - (Required) The catalog to which to connect.
* `host` - (Required) The host to which to connect.
* `port` - (Required) The port to which to connect.

### rds Argument Reference

* `database` - (Required) The database to which to connect.
* `instance_id` - (Optional) The instance ID to which to connect.

### redshift Argument Reference

* `cluster_id` - (Optional, Required if `host` and `port` are not provided) The ID of the cluster to which to connect.
* `database` - (Required) The database to which to connect.
* `host` - (Optional, Required if `cluster_id` is not provided) The host to which to connect.
* `port` - (Optional, Required if `cluster_id` is not provided) The port to which to connect.

### s3 Argument Reference

* `manifest_file_location` - (Required) An [object containing the S3 location](#manifest_file_location-argument-reference) of the S3 manifest file.

### manifest_file_location Argument Reference

* `bucket` - (Required) The name of the bucket that contains the manifest file.
* `key` - (Required) The key of the manifest file within the bucket.

### service_now Argument Reference

* `site_base_url` - (Required) The base URL of the Jira instance's site to which to connect.

### snowflake Argument Reference

* `database` - (Required) The database to which to connect.
* `host` - (Required) The host to which to connect.
* `warehouse` - (Required) The warehouse to which to connect.

### spark Argument Reference

* `host` - (Required) The host to which to connect.
* `port` - (Required) The warehouse to which to connect.

### sql_server Argument Reference

* `database` - (Required) The database to which to connect.
* `host` - (Required) The host to which to connect.
* `port` - (Required) The warehouse to which to connect.

### teradata Argument Reference

* `database` - (Required) The database to which to connect.
* `host` - (Required) The host to which to connect.
* `port` - (Required) The warehouse to which to connect.

#### twitter Argument Reference

* `max_rows` - (Required) The maximum number of rows to query.
* `query` - (Required) The Twitter query to retrieve the data.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the data source
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

A QuickSight data source can be imported using the AWS account ID, and data source ID name separated by a slash (`/`) e.g.,

```
$ terraform import aws_quicksight_data_source.example 123456789123/my-data-source-id
```
