---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_crawler"
description: |-
  Manages a Glue Crawler
---

# Resource: aws_glue_crawler

Manages a Glue Crawler. More information can be found in the [AWS Glue Developer Guide](https://docs.aws.amazon.com/glue/latest/dg/add-crawler.html)

## Example Usage

### DynamoDB Target Example

```terraform
resource "aws_glue_crawler" "example" {
  database_name = aws_glue_catalog_database.example.name
  name          = "example"
  role          = aws_iam_role.example.arn

  dynamodb_target {
    path = "table-name"
  }
}
```

### JDBC Target Example

```terraform
resource "aws_glue_crawler" "example" {
  database_name = aws_glue_catalog_database.example.name
  name          = "example"
  role          = aws_iam_role.example.arn

  jdbc_target {
    connection_name = aws_glue_connection.example.name
    path            = "database-name/%"
  }
}
```

### S3 Target Example

```terraform
resource "aws_glue_crawler" "example" {
  database_name = aws_glue_catalog_database.example.name
  name          = "example"
  role          = aws_iam_role.example.arn

  s3_target {
    path = "s3://${aws_s3_bucket.example.bucket}"
  }
}
```

### Catalog Target Example

```terraform
resource "aws_glue_crawler" "example" {
  database_name = aws_glue_catalog_database.example.name
  name          = "example"
  role          = aws_iam_role.example.arn

  catalog_target {
    database_name = aws_glue_catalog_database.example.name
    tables        = [aws_glue_catalog_table.example.name]
  }

  schema_change_policy {
    delete_behavior = "LOG"
  }

  configuration = <<EOF
{
  "Version":1.0,
  "Grouping": {
    "TableGroupingPolicy": "CombineCompatibleSchemas"
  }
}
EOF
}
```

### MongoDB Target Example

```terraform
resource "aws_glue_crawler" "example" {
  database_name = aws_glue_catalog_database.example.name
  name          = "example"
  role          = aws_iam_role.example.arn

  mongodb_target {
    connection_name = aws_glue_connection.example.name
    path            = "database-name/%"
  }
}
```

### Configuration Settings Example

```terraform
resource "aws_glue_crawler" "events_crawler" {
  database_name = aws_glue_catalog_database.glue_database.name
  schedule      = "cron(0 1 * * ? *)"
  name          = "events_crawler_${var.environment_name}"
  role          = aws_iam_role.glue_role.arn
  tags          = var.tags

  configuration = jsonencode(
    {
      Grouping = {
        TableGroupingPolicy = "CombineCompatibleSchemas"
      }
      CrawlerOutput = {
        Partitions = { AddOrUpdateBehavior = "InheritFromTable" }
      }
      Version = 1
    }
  )

  s3_target {
    path = "s3://${aws_s3_bucket.data_lake_bucket.bucket}"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `catalog_target` - (Optional) List of nested AWS Glue Data Catalog target arguments. See [`catalog_target`](#catalog_target-block) below.
* `classifiers` - (Optional) List of custom classifiers. By default, all AWS classifiers are included in a crawl, but these custom classifiers always override the default classifiers for a given classification.
* `configuration` - (Optional) JSON string of configuration information. For more details see [Setting Crawler Configuration Options](https://docs.aws.amazon.com/glue/latest/dg/crawler-configuration.html).
* `database_name` - (Required) Glue database where results are written.
* `delta_target` - (Optional) List of nested Delta Lake target arguments. See [`delta_target`](#delta_target-block) below.
* `description` - (Optional) Description of the crawler.
* `dynamodb_target` - (Optional) List of nested DynamoDB target arguments. See [`dynamodb_target`](#dynamodb_target-block) below.
* `hudi_target` - (Optional) List of nested Hudi target arguments. See [`hudi_target`](#hudi_target-block) below.
* `iceberg_target` - (Optional) List of nested Iceberg target arguments. See [`iceberg_target`](#iceberg_target-block) below.
* `jdbc_target` - (Optional) List of nested JDBC target arguments. See [`jdbc_target`](#jdbc_target-block) below.
* `lake_formation_configuration` - (Optional) Lake Formation configuration settings for the crawler. See [`lake_formation_configuration`](#lake_formation_configuration-block) below.
* `lineage_configuration` - (Optional) Data lineage configuration settings for the crawler. See [`lineage_configuration`](#lineage_configuration-block) below.
* `mongodb_target` - (Optional) List of nested MongoDB target arguments. See [`mongodb_target`](#mongodb_target-block) below.
* `name` - (Required) Name of the crawler.
* `recrawl_policy` - (Optional) Policy that specifies whether to crawl the entire dataset again, or to crawl only folders that were added since the last crawler run. See [`recrawl_policy`](#recrawl_policy-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `role` - (Required) IAM role friendly name (including path without leading slash), or ARN of an IAM role, used by the crawler to access other resources.
* `s3_target` - (Optional) List of nested Amazon S3 target arguments. See [`s3_target`](#s3_target-block) below.
* `schedule` - (Optional) Cron expression used to specify the schedule. For more information, see [Time-Based Schedules for Jobs and Crawlers](https://docs.aws.amazon.com/glue/latest/dg/monitor-data-warehouse-schedule.html). For example, to run something every day at 12:15 UTC, you would specify: `cron(15 12 * * ? *)`.
* `schema_change_policy` - (Optional) Policy for the crawler's update and deletion behavior. See [`schema_change_policy`](#schema_change_policy-block) below.
* `security_configuration` - (Optional) Name of Security Configuration to be used by the crawler.
* `table_prefix` - (Optional) Table prefix used for catalog tables that are created.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

~> **NOTE:** Must specify at least one of `dynamodb_target`, `jdbc_target`, `s3_target`, `mongodb_target` or `catalog_target`.

### `dynamodb_target` Block

* `path` - (Required) Name of the DynamoDB table to crawl.
* `scan_all` - (Optional) Whether to scan all the records, or to sample rows from the table. Scanning all the records can take a long time when the table is not a high throughput table.  defaults to `true`.
* `scan_rate` - (Optional) Percentage of the configured read capacity units to use by the AWS Glue crawler. The valid values are null or a value between 0.1 to 1.5.

### `jdbc_target` Block

* `connection_name` - (Required) Name of the connection to use to connect to the JDBC target.
* `enable_additional_metadata` - (Optional) Specify a value of `RAWTYPES` or `COMMENTS` to enable additional metadata intable responses. `RAWTYPES` provides the native-level datatype. `COMMENTS` provides comments associated with a column or table in the database.
* `exclusions` - (Optional) List of glob patterns used to exclude from the crawl.
* `path` - (Required) Path of the JDBC target.

### `s3_target` Block

* `connection_name` - (Optional) Name of a connection which allows crawler to access data in S3 within a VPC.
* `dlq_event_queue_arn` - (Optional) ARN of the dead-letter SQS queue.
* `event_queue_arn` - (Optional) ARN of the SQS queue to receive S3 notifications from.
* `exclusions` - (Optional) List of glob patterns used to exclude from the crawl.
* `path` - (Required) Path to the Amazon S3 target.
* `sample_size` - (Optional) Sets the number of files in each leaf folder to be crawled when crawling sample files in a dataset. If not set, all the files are crawled. A valid value is an integer between 1 and 249.

### `catalog_target` Block

* `connection_name` - (Optional) Name of the connection for an Amazon S3-backed Data Catalog table to be a target of the crawl when using a Catalog connection type paired with a `NETWORK` Connection type.
* `database_name` - (Required) Name of the Glue database to be synchronized.
* `dlq_event_queue_arn` - (Optional) Valid Amazon SQS ARN.
* `event_queue_arn` - (Optional) Valid Amazon SQS ARN.
* `tables` - (Required) List of catalog tables to be synchronized.

~> **Note:** `deletion_behavior` of catalog target doesn't support `DEPRECATE_IN_DATABASE`.

-> **Note:** `configuration` for catalog target crawlers will have `{ ... "Grouping": { "TableGroupingPolicy": "CombineCompatibleSchemas"} }` by default.

### `mongodb_target` Block

* `connection_name` - (Required) Name of the connection to use to connect to the Amazon DocumentDB or MongoDB target.
* `path` - (Required) Path of the Amazon DocumentDB or MongoDB target (database/collection).
* `scan_all` - (Optional) Whether to scan all the records, or to sample rows from the table. Scanning all the records can take a long time when the table is not a high throughput table. Default value is `true`.

### `hudi_target` Block

* `connection_name` - (Optional) Name of the connection to use to connect to the Hudi target.
* `exclusions` - (Optional) List of glob patterns used to exclude from the crawl.
* `maximum_traversal_depth` - (Required) Maximum depth of Amazon S3 paths that the crawler can traverse to discover the Hudi metadata folder in your Amazon S3 path. Used to limit the crawler run time. Valid values are between `1` and `20`.
* `paths` - (Required) One or more Amazon S3 paths that contains Hudi metadata folders as s3://bucket/prefix.

### `iceberg_target` Block

* `connection_name` - (Optional) Name of the connection to use to connect to the Iceberg target.
* `exclusions` - (Optional) List of glob patterns used to exclude from the crawl.
* `maximum_traversal_depth` - (Required) Maximum depth of Amazon S3 paths that the crawler can traverse to discover the Iceberg metadata folder in your Amazon S3 path. Used to limit the crawler run time. Valid values are between `1` and `20`.
* `paths` - (Required) One or more Amazon S3 paths that contains Iceberg metadata folders as s3://bucket/prefix.

### `delta_target` Block

* `connection_name` - (Optional) Name of the connection to use to connect to the Delta table target.
* `create_native_delta_table` - (Optional) Whether the crawler will create native tables, to allow integration with query engines that support querying of the Delta transaction log directly.
* `delta_tables` - (Required) List of the Amazon S3 paths to the Delta tables.
* `write_manifest` - (Required) Whether to write the manifest files to the Delta table path.

### `schema_change_policy` Block

* `delete_behavior` - (Optional) Deletion behavior when the crawler finds a deleted object. Valid values: `LOG`, `DELETE_FROM_DATABASE`, or `DEPRECATE_IN_DATABASE`. Defaults to `DEPRECATE_IN_DATABASE`.
* `update_behavior` - (Optional) Update behavior when the crawler finds a changed schema. Valid values: `LOG` or `UPDATE_IN_DATABASE`. Defaults to `UPDATE_IN_DATABASE`.

### `lake_formation_configuration` Block

* `account_id` - (Optional) Required for cross account crawls. For same account crawls as the target data, this can omitted.
* `use_lake_formation_credentials` - (Optional) Whether to use Lake Formation credentials for the crawler instead of the IAM role credentials.

### `lineage_configuration` Block

* `crawler_lineage_settings` - (Optional) Whether data lineage is enabled for the crawler. Valid values are: `ENABLE` and `DISABLE`. Default value is `DISABLE`.

### `recrawl_policy` Block

* `recrawl_behavior` - (Optional) Whether to crawl the entire dataset again, crawl only folders that were added since the last crawler run, or crawl what S3 notifies the crawler of via SQS. Valid Values are: `CRAWL_EVENT_MODE`, `CRAWL_EVERYTHING` and `CRAWL_NEW_FOLDERS_ONLY`. Default value is `CRAWL_EVERYTHING`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the crawler
* `id` - Crawler name
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Glue Crawlers using `name`. For example:

```terraform
import {
  to = aws_glue_crawler.MyJob
  id = "MyJob"
}
```

Using `terraform import`, import Glue Crawlers using `name`. For example:

```console
% terraform import aws_glue_crawler.MyJob MyJob
```
