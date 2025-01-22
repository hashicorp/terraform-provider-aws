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

~> **NOTE:** Must specify at least one of `dynamodb_target`, `jdbc_target`, `s3_target`, `mongodb_target` or `catalog_target`.

This resource supports the following arguments:

* `database_name` (Required) Glue database where results are written.
* `name` (Required) Name of the crawler.
* `role` (Required) The IAM role friendly name (including path without leading slash), or ARN of an IAM role, used by the crawler to access other resources.
* `classifiers` (Optional) List of custom classifiers. By default, all AWS classifiers are included in a crawl, but these custom classifiers always override the default classifiers for a given classification.
* `configuration` (Optional) JSON string of configuration information. For more details see [Setting Crawler Configuration Options](https://docs.aws.amazon.com/glue/latest/dg/crawler-configuration.html).
* `description` (Optional) Description of the crawler.
* `delta_target` (Optional) List of nested Delta Lake target arguments. See [Delta Target](#delta-target) below.
* `dynamodb_target` (Optional) List of nested DynamoDB target arguments. See [Dynamodb Target](#dynamodb-target) below.
* `jdbc_target` (Optional) List of nested JDBC target arguments. See [JDBC Target](#jdbc-target) below.
* `s3_target` (Optional) List of nested Amazon S3 target arguments. See [S3 Target](#s3-target) below.
* `catalog_target` (Optional) List of nested AWS Glue Data Catalog target arguments. See [Catalog Target](#catalog-target) below.
* `mongodb_target` (Optional) List of nested MongoDB target arguments. See [MongoDB Target](#mongodb-target) below.
* `hudi_target` (Optional) List of nested Hudi target arguments. See [Iceberg Target](#hudi-target) below.
* `iceberg_target` (Optional) List of nested Iceberg target arguments. See [Iceberg Target](#iceberg-target) below.
* `schedule` (Optional) A cron expression used to specify the schedule. For more information, see [Time-Based Schedules for Jobs and Crawlers](https://docs.aws.amazon.com/glue/latest/dg/monitor-data-warehouse-schedule.html). For example, to run something every day at 12:15 UTC, you would specify: `cron(15 12 * * ? *)`.
* `schema_change_policy` (Optional) Policy for the crawler's update and deletion behavior. See [Schema Change Policy](#schema-change-policy) below.
* `lake_formation_configuration` (Optional) Specifies Lake Formation configuration settings for the crawler. See [Lake Formation Configuration](#lake-formation-configuration) below.
* `lineage_configuration` (Optional) Specifies data lineage configuration settings for the crawler. See [Lineage Configuration](#lineage-configuration) below.
* `recrawl_policy` (Optional)  A policy that specifies whether to crawl the entire dataset again, or to crawl only folders that were added since the last crawler run.. See [Recrawl Policy](#recrawl-policy) below.
* `security_configuration` (Optional) The name of Security Configuration to be used by the crawler
* `table_prefix` (Optional) The table prefix used for catalog tables that are created.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Dynamodb Target

* `path` - (Required) The name of the DynamoDB table to crawl.
* `scan_all` - (Optional) Indicates whether to scan all the records, or to sample rows from the table. Scanning all the records can take a long time when the table is not a high throughput table.  defaults to `true`.
* `scan_rate` - (Optional) The percentage of the configured read capacity units to use by the AWS Glue crawler. The valid values are null or a value between 0.1 to 1.5.

### JDBC Target

* `connection_name` - (Required) The name of the connection to use to connect to the JDBC target.
* `path` - (Required) The path of the JDBC target.
* `exclusions` - (Optional) A list of glob patterns used to exclude from the crawl.
* `enable_additional_metadata` - (Optional) Specify a value of `RAWTYPES` or `COMMENTS` to enable additional metadata intable responses. `RAWTYPES` provides the native-level datatype. `COMMENTS` provides comments associated with a column or table in the database.

### S3 Target

* `path` - (Required) The path to the Amazon S3 target.
* `connection_name` - (Optional) The name of a connection which allows crawler to access data in S3 within a VPC.
* `exclusions` - (Optional) A list of glob patterns used to exclude from the crawl.
* `sample_size` - (Optional) Sets the number of files in each leaf folder to be crawled when crawling sample files in a dataset. If not set, all the files are crawled. A valid value is an integer between 1 and 249.
* `event_queue_arn` - (Optional) The ARN of the SQS queue to receive S3 notifications from.
* `dlq_event_queue_arn` - (Optional) The ARN of the dead-letter SQS queue.

### Catalog Target

* `connection_name` - (Optional) The name of the connection for an Amazon S3-backed Data Catalog table to be a target of the crawl when using a Catalog connection type paired with a `NETWORK` Connection type.
* `database_name` - (Required) The name of the Glue database to be synchronized.
* `tables` - (Required) A list of catalog tables to be synchronized.
* `event_queue_arn` - (Optional)  A valid Amazon SQS ARN.
* `dlq_event_queue_arn` - (Optional)  A valid Amazon SQS ARN.

~> **Note:** `deletion_behavior` of catalog target doesn't support `DEPRECATE_IN_DATABASE`.

-> **Note:** `configuration` for catalog target crawlers will have `{ ... "Grouping": { "TableGroupingPolicy": "CombineCompatibleSchemas"} }` by default.

### MongoDB Target

* `connection_name` - (Required) The name of the connection to use to connect to the Amazon DocumentDB or MongoDB target.
* `path` - (Required) The path of the Amazon DocumentDB or MongoDB target (database/collection).
* `scan_all` - (Optional) Indicates whether to scan all the records, or to sample rows from the table. Scanning all the records can take a long time when the table is not a high throughput table. Default value is `true`.

### Hudi Target

* `connection_name` - (Optional) The name of the connection to use to connect to the Hudi target.
* `paths` - (Required) One or more Amazon S3 paths that contains Hudi metadata folders as s3://bucket/prefix.
* `exclusions` - (Optional) A list of glob patterns used to exclude from the crawl.
* `maximum_traversal_depth` - (Required) The maximum depth of Amazon S3 paths that the crawler can traverse to discover the Hudi metadata folder in your Amazon S3 path. Used to limit the crawler run time. Valid values are between `1` and `20`.

### Iceberg Target

* `connection_name` - (Optional) The name of the connection to use to connect to the Iceberg target.
* `paths` - (Required) One or more Amazon S3 paths that contains Iceberg metadata folders as s3://bucket/prefix.
* `exclusions` - (Optional) A list of glob patterns used to exclude from the crawl.
* `maximum_traversal_depth` - (Required) The maximum depth of Amazon S3 paths that the crawler can traverse to discover the Iceberg metadata folder in your Amazon S3 path. Used to limit the crawler run time. Valid values are between `1` and `20`.

### Delta Target

* `connection_name` - (Optional) The name of the connection to use to connect to the Delta table target.
* `create_native_delta_table` (Optional) Specifies whether the crawler will create native tables, to allow integration with query engines that support querying of the Delta transaction log directly.
* `delta_tables` - (Required) A list of the Amazon S3 paths to the Delta tables.
* `write_manifest` - (Required) Specifies whether to write the manifest files to the Delta table path.

### Schema Change Policy

* `delete_behavior` - (Optional) The deletion behavior when the crawler finds a deleted object. Valid values: `LOG`, `DELETE_FROM_DATABASE`, or `DEPRECATE_IN_DATABASE`. Defaults to `DEPRECATE_IN_DATABASE`.
* `update_behavior` - (Optional) The update behavior when the crawler finds a changed schema. Valid values: `LOG` or `UPDATE_IN_DATABASE`. Defaults to `UPDATE_IN_DATABASE`.

### Lake Formation Configuration

* `account_id` - (Optional) Required for cross account crawls. For same account crawls as the target data, this can omitted.
* `use_lake_formation_credentials` - (Optional) Specifies whether to use Lake Formation credentials for the crawler instead of the IAM role credentials.

### Lineage Configuration

* `crawler_lineage_settings` - (Optional) Specifies whether data lineage is enabled for the crawler. Valid values are: `ENABLE` and `DISABLE`. Default value is `DISABLE`.

### Recrawl Policy

* `recrawl_behavior` - (Optional) Specifies whether to crawl the entire dataset again, crawl only folders that were added since the last crawler run, or crawl what S3 notifies the crawler of via SQS. Valid Values are: `CRAWL_EVENT_MODE`, `CRAWL_EVERYTHING` and `CRAWL_NEW_FOLDERS_ONLY`. Default value is `CRAWL_EVERYTHING`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Crawler name
* `arn` - The ARN of the crawler
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
