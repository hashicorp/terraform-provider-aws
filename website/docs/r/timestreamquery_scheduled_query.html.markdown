---
subcategory: "Timestream Query"
layout: "aws"
page_title: "AWS: aws_timestreamquery_scheduled_query"
description: |-
  Terraform resource for managing an AWS Timestream Query Scheduled Query.
---

# Resource: aws_timestreamquery_scheduled_query

Terraform resource for managing an AWS Timestream Query Scheduled Query.

## Example Usage

### Basic Usage

Before creating a scheduled query, you must have a source database and table with ingested data. Below is a [multi-step example](#multi-step-example), providing an opportunity for data ingestion.

If your infrastructure is already set up—including the source database and table with data, results database and table, error report S3 bucket, SNS topic, and IAM role—you can create a scheduled query as follows:

```terraform
resource "aws_timestreamquery_scheduled_query" "example" {
  execution_role_arn = aws_iam_role.example.arn
  name               = aws_timestreamwrite_table.example.table_name
  query_string       = <<EOF
SELECT region, az, hostname, BIN(time, 15s) AS binned_timestamp,
	ROUND(AVG(cpu_utilization), 2) AS avg_cpu_utilization,
	ROUND(APPROX_PERCENTILE(cpu_utilization, 0.9), 2) AS p90_cpu_utilization,
	ROUND(APPROX_PERCENTILE(cpu_utilization, 0.95), 2) AS p95_cpu_utilization,
	ROUND(APPROX_PERCENTILE(cpu_utilization, 0.99), 2) AS p99_cpu_utilization
FROM exampledatabase.exampletable
WHERE measure_name = 'metrics' AND time > ago(2h)
GROUP BY region, hostname, az, BIN(time, 15s)
ORDER BY binned_timestamp ASC
LIMIT 5
EOF

  error_report_configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.example.bucket
    }
  }

  notification_configuration {
    sns_configuration {
      topic_arn = aws_sns_topic.example.arn
    }
  }

  schedule_configuration {
    schedule_expression = "rate(1 hour)"
  }

  target_configuration {
    timestream_configuration {
      database_name = aws_timestreamwrite_database.results.database_name
      table_name    = aws_timestreamwrite_table.results.table_name
      time_column   = "binned_timestamp"

      dimension_mapping {
        dimension_value_type = "VARCHAR"
        name                 = "az"
      }

      dimension_mapping {
        dimension_value_type = "VARCHAR"
        name                 = "region"
      }

      dimension_mapping {
        dimension_value_type = "VARCHAR"
        name                 = "hostname"
      }

      multi_measure_mappings {
        target_multi_measure_name = "multi-metrics"

        multi_measure_attribute_mapping {
          measure_value_type = "DOUBLE"
          source_column      = "avg_cpu_utilization"
        }

        multi_measure_attribute_mapping {
          measure_value_type = "DOUBLE"
          source_column      = "p90_cpu_utilization"
        }

        multi_measure_attribute_mapping {
          measure_value_type = "DOUBLE"
          source_column      = "p95_cpu_utilization"
        }

        multi_measure_attribute_mapping {
          measure_value_type = "DOUBLE"
          source_column      = "p99_cpu_utilization"
        }
      }
    }
  }
}
```

### Multi-step Example

To ingest data before creating a scheduled query, this example provides multiple steps:

1. [Create the prerequisite infrastructure](#step-1-create-the-prerequisite-infrastructure)
2. [Ingest data](#step-2-ingest-data)
3. [Create the scheduled query](#step-3-create-the-scheduled-query)

#### Step 1. Create the prerequisite infrastructure

```terraform
resource "aws_s3_bucket" "test" {
  bucket        = "example"
  force_destroy = true
}

resource "aws_sns_topic" "test" {
  name = "example"
}

resource "aws_sqs_queue" "test" {
  name = "example"

  sqs_managed_sse_enabled = true
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn = aws_sns_topic.test.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.test.arn
}

resource "aws_sqs_queue_policy" "test" {
  queue_url = aws_sqs_queue.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Action   = ["sqs:SendMessage"]
      Resource = aws_sqs_queue.test.arn
      Condition = {
        ArnEquals = {
          "aws:SourceArn" = aws_sns_topic.test.arn
        }
      }
    }]
  })
}

resource "aws_iam_role" "test" {
  name = "example"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "timestream.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })

  tags = {
    Name = "example"
  }
}

resource "aws_iam_role_policy" "test" {
  name = "example"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = [
        "kms:Decrypt",
        "sns:Publish",
        "timestream:describeEndpoints",
        "timestream:Select",
        "timestream:SelectValues",
        "timestream:WriteRecords",
        "s3:PutObject",
      ]
      Resource = "*"
      Effect   = "Allow"
    }]
  })
}

resource "aws_timestreamwrite_database" "test" {
  database_name = "exampledatabase"
}

resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = "exampletable"

  magnetic_store_write_properties {
    enable_magnetic_store_writes = true
  }

  retention_properties {
    magnetic_store_retention_period_in_days = 1
    memory_store_retention_period_in_hours  = 1
  }
}

resource "aws_timestreamwrite_database" "results" {
  database_name = "exampledatabase-results"
}

resource "aws_timestreamwrite_table" "results" {
  database_name = aws_timestreamwrite_database.results.database_name
  table_name    = "exampletable-results"

  magnetic_store_write_properties {
    enable_magnetic_store_writes = true
  }

  retention_properties {
    magnetic_store_retention_period_in_days = 1
    memory_store_retention_period_in_hours  = 1
  }
}
```

#### Step 2. Ingest data

This is done with Amazon Timestream Write [WriteRecords](https://docs.aws.amazon.com/timestream/latest/developerguide/API_WriteRecords.html).

#### Step 3. Create the scheduled query

```terraform
resource "aws_timestreamquery_scheduled_query" "example" {
  execution_role_arn = aws_iam_role.example.arn
  name               = aws_timestreamwrite_table.example.table_name
  query_string       = <<EOF
SELECT region, az, hostname, BIN(time, 15s) AS binned_timestamp,
	ROUND(AVG(cpu_utilization), 2) AS avg_cpu_utilization,
	ROUND(APPROX_PERCENTILE(cpu_utilization, 0.9), 2) AS p90_cpu_utilization,
	ROUND(APPROX_PERCENTILE(cpu_utilization, 0.95), 2) AS p95_cpu_utilization,
	ROUND(APPROX_PERCENTILE(cpu_utilization, 0.99), 2) AS p99_cpu_utilization
FROM exampledatabase.exampletable
WHERE measure_name = 'metrics' AND time > ago(2h)
GROUP BY region, hostname, az, BIN(time, 15s)
ORDER BY binned_timestamp ASC
LIMIT 5
EOF

  error_report_configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.example.bucket
    }
  }

  notification_configuration {
    sns_configuration {
      topic_arn = aws_sns_topic.example.arn
    }
  }

  schedule_configuration {
    schedule_expression = "rate(1 hour)"
  }

  target_configuration {
    timestream_configuration {
      database_name = aws_timestreamwrite_database.results.database_name
      table_name    = aws_timestreamwrite_table.results.table_name
      time_column   = "binned_timestamp"

      dimension_mapping {
        dimension_value_type = "VARCHAR"
        name                 = "az"
      }

      dimension_mapping {
        dimension_value_type = "VARCHAR"
        name                 = "region"
      }

      dimension_mapping {
        dimension_value_type = "VARCHAR"
        name                 = "hostname"
      }

      multi_measure_mappings {
        target_multi_measure_name = "multi-metrics"

        multi_measure_attribute_mapping {
          measure_value_type = "DOUBLE"
          source_column      = "avg_cpu_utilization"
        }

        multi_measure_attribute_mapping {
          measure_value_type = "DOUBLE"
          source_column      = "p90_cpu_utilization"
        }

        multi_measure_attribute_mapping {
          measure_value_type = "DOUBLE"
          source_column      = "p95_cpu_utilization"
        }

        multi_measure_attribute_mapping {
          measure_value_type = "DOUBLE"
          source_column      = "p99_cpu_utilization"
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `error_report_configuration` - (Required) Configuration block for error reporting configuration. [See below.](#error_report_configuration)
* `execution_role_arn` - (Required) ARN for the IAM role that Timestream will assume when running the scheduled query.
* `name` - (Required) Name of the scheduled query.
* `notification_configuration` - (Required) Configuration block for notification configuration for a scheduled query. A notification is sent by Timestream when a scheduled query is created, its state is updated, or when it is deleted. [See below.](#notification_configuration)
* `query_string` - (Required) Query string to run. Parameter names can be specified in the query string using the `@` character followed by an identifier. The named parameter `@scheduled_runtime` is reserved and can be used in the query to get the time at which the query is scheduled to run. The timestamp calculated according to the `schedule_configuration` parameter, will be the value of `@scheduled_runtime` paramater for each query run. For example, consider an instance of a scheduled query executing on 2021-12-01 00:00:00. For this instance, the `@scheduled_runtime` parameter is initialized to the timestamp 2021-12-01 00:00:00 when invoking the query.
* `schedule_configuration` - (Required) Configuration block for schedule configuration for the query. [See below.](#schedule_configuration)
* `target_configuration` - (Required) Configuration block for writing the result of a query. [See below.](#target_configuration)

The following arguments are optional:

* `kms_key_id` - (Optional) Amazon KMS key used to encrypt the scheduled query resource, at-rest. If not specified, the scheduled query resource will be encrypted with a Timestream owned Amazon KMS key. To specify a KMS key, use the key ID, key ARN, alias name, or alias ARN. When using an alias name, prefix the name with "alias/". If `error_report_configuration` uses `SSE_KMS` as the encryption type, the same `kms_key_id` is used to encrypt the error report at rest.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `error_report_configuration`

* `s3_configuration` - (Required) Configuration block for the S3 configuration for the error reports. [See below.](#s3_configuration)

#### `s3_configuration`

* `bucket_name` - (Required) Name of the S3 bucket under which error reports will be created.
* `encryption_option` - (Optional) Encryption at rest options for the error reports. If no encryption option is specified, Timestream will choose `SSE_S3` as default. Valid values are `SSE_S3`, `SSE_KMS`.
* `object_key_prefix` - (Optional) Prefix for the error report key.

### `notification_configuration`

* `sns_configuration` - (Required) Configuration block for details about the Amazon Simple Notification Service (SNS) configuration. [See below.](#sns_configuration)

#### `sns_configuration`

* `topic_arn` - (Required) SNS topic ARN that the scheduled query status notifications will be sent to.

### `schedule_configuration`

* `schedule_expression` - (Required) When to trigger the scheduled query run. This can be a cron expression or a rate expression.

### `target_configuration`

* `timestream_configuration` - (Required) Configuration block for information needed to write data into the Timestream database and table. [See below.](#timestream_configuration)

#### `timestream_configuration`

* `database_name` - (Required) Name of Timestream database to which the query result will be written.
* `dimension_mapping` - (Required) Configuration block for mapping of column(s) from the query result to the dimension in the destination table. [See below.](#dimension_mapping)
* `table_name` - (Required) Name of Timestream table that the query result will be written to. The table should be within the same database that is provided in Timestream configuration.
* `time_column` - (Required) Column from query result that should be used as the time column in destination table. Column type for this should be TIMESTAMP.
* `measure_name_column` - (Optional) Name of the measure column.
* `mixed_measure_mapping` - (Optional) Configuration block for how to map measures to multi-measure records. [See below.](#mixed_measure_mapping)
* `multi_measure_mappings` - (Optional) Configuration block for multi-measure mappings. Only one of `mixed_measure_mappings` or `multi_measure_mappings` can be provided. `multi_measure_mappings` can be used to ingest data as multi measures in the derived table. [See below.](#multi_measure_mappings)

##### `dimension_mapping`

* `dimension_value_type` - (Required) Type for the dimension. Valid value: `VARCHAR`.
* `name` - (Required) Column name from query result.

##### `mixed_measure_mapping`

* `measure_name` - (Optional) Refers to the value of measure_name in a result row. This field is required if `measure_name_column` is provided.
* `multi_measure_attribute_mapping` - (Optional) Configuration block for attribute mappings for `MULTI` value measures. Required when `measure_value_type` is `MULTI`. [See below.](#multi_measure_attribute_mapping)
* `measure_value_type` - (Required) Type of the value that is to be read from `source_column`. Valid values are `BIGINT`, `BOOLEAN`, `DOUBLE`, `VARCHAR`, `MULTI`.
* `source_column` - (Optional) Source column from which measure-value is to be read for result materialization.
* `target_measure_name` - (Optional) Target measure name to be used. If not provided, the target measure name by default is `measure_name`, if provided, or `source_column` otherwise.

##### `multi_measure_attribute_mapping`

* `measure_value_type` - (Required) Type of the attribute to be read from the source column. Valid values are `BIGINT`, `BOOLEAN`, `DOUBLE`, `VARCHAR`, `TIMESTAMP`.
* `source_column` - (Required) Source column from where the attribute value is to be read.
* `target_multi_measure_attribute_name` - (Optional) Custom name to be used for attribute name in derived table. If not provided, `source_column` is used.

##### `multi_measure_mappings`

* `multi_measure_attribute_mapping` - (Required) Attribute mappings to be used for mapping query results to ingest data for multi-measure attributes. [See above.](#multi_measure_attribute_mapping)
* `target_multi_measure_name` - (Optional) Name of the target multi-measure name in the derived table. This input is required when `measure_name_column` is not provided. If `measure_name_column` is provided, then the value from that column will be used as the multi-measure name.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Scheduled Query.
* `creation_time` - Creation time for the scheduled query.
* `next_invocation_time` - Next time the scheduled query is scheduled to run.
* `previous_invocation_time` - Last time the scheduled query was run.
* `state` - State of the scheduled query, either `ENABLED` or `DISABLED`.
* `last_run_summary` - Runtime summary for the last scheduled query run.
    * `error_report_location` - Contains the location of the error report for a single scheduled query call.
        * `s3_report_location` - S3 report location for the scheduled query run.
            * `bucket_name` - S3 bucket name.
            * `object_key` - S3 key.
    * `execution_stats` - Statistics for a single scheduled query run.
        * `bytes_metered` - Bytes metered for a single scheduled query run.
        * `cumulative_bytes_scanned` - Bytes scanned for a single scheduled query run.
        * `data_writes` - Data writes metered for records ingested in a single scheduled query run.
        * `execution_time_in_millis` - Total time, measured in milliseconds, that was needed for the scheduled query run to complete.
        * `query_result_rows` - Number of rows present in the output from running a query before ingestion to destination data source.
        * `records_ingested` - Number of records ingested for a single scheduled query run.
    * `failure_reason` - Error message for the scheduled query in case of failure. You might have to look at the error report to get more detailed error reasons.
    * `invocation_time` - InvocationTime for this run. This is the time at which the query is scheduled to run. Parameter `@scheduled_runtime` can be used in the query to get the value.
    * `query_insights_response` - Provides various insights and metrics related to the run summary of the scheduled query.
        * `output_bytes` - Size of query result set in bytes. You can use this data to validate if the result set has changed as part of the query tuning exercise.
        * `output_rows` - Total number of rows returned as part of the query result set. You can use this data to validate if the number of rows in the result set have changed as part of the query tuning exercise.
        * `query_spatial_coverage` - Insights into the spatial coverage of the query, including the table with sub-optimal (max) spatial pruning. This information can help you identify areas for improvement in your partitioning strategy to enhance spatial pruning.
            * `max` - Insights into the spatial coverage of the executed query and the table with the most inefficient spatial pruning.
                * `partition_key` - Partition key used for partitioning, which can be a default measure_name or a customer defined partition key.
                * `table_arn` - ARN of the table with the most sub-optimal spatial pruning.
                * `value` - Maximum ratio of spatial coverage.
        * `query_table_count` - Number of tables in the query.
        * `query_temporal_range` - Insights into the temporal range of the query, including the table with the largest (max) time range. Following are some of the potential options for optimizing time-based pruning: add missing time-predicates, remove functions around the time predicates, add time predicates to all the sub-queries.
            * `max` - Insights into the temporal range of the query, including the table with the largest (max) time range.
                * `table_arn` - ARN of the table table which is queried with the largest time range.
                * `value` - Maximum duration in nanoseconds between the start and end of the query.
    * `run_status` - Status of a scheduled query run. Valid values: `AUTO_TRIGGER_SUCCESS`, `AUTO_TRIGGER_FAILURE`, `MANUAL_TRIGGER_SUCCESS`, `MANUAL_TRIGGER_FAILURE`.
    * `trigger_time` - Actual time when the query was run.
* `recently_failed_runs` - Runtime summary for the last five failed scheduled query runs.
    * `error_report_location` - S3 location for error report.
        * `s3_report_location` - S3 location where error reports are written.
            * `bucket_name` - S3 bucket name.
            * `object_key` - S3 key.
    * `execution_stats` - Statistics for a single scheduled query run.
        * `bytes_metered` - Bytes metered for a single scheduled query run.
        * `cumulative_bytes_scanned` - Bytes scanned for a single scheduled query run.
        * `data_writes` - Data writes metered for records ingested in a single scheduled query run.
        * `execution_time_in_millis` - Total time, measured in milliseconds, that was needed for the scheduled query run to complete.
        * `query_result_rows` - Number of rows present in the output from running a query before ingestion to destination data source.
        * `records_ingested` - Number of records ingested for a single scheduled query run.
    * `failure_reason` - Error message for the scheduled query in case of failure. You might have to look at the error report to get more detailed error reasons.
    * `invocation_time` - InvocationTime for this run. This is the time at which the query is scheduled to run. Parameter `@scheduled_runtime` can be used in the query to get the value.
    * `query_insights_response` - Various insights and metrics related to the run summary of the scheduled query.
        * `output_bytes` - Size of query result set in bytes. You can use this data to validate if the result set has changed as part of the query tuning exercise.
        * `output_rows` - Total number of rows returned as part of the query result set. You can use this data to validate if the number of rows in the result set have changed as part of the query tuning exercise.
        * `query_spatial_coverage` - Insights into the spatial coverage of the query, including the table with sub-optimal (max) spatial pruning. This information can help you identify areas for improvement in your partitioning strategy to enhance spatial pruning.
            * `max` - Insights into the spatial coverage of the executed query and the table with the most inefficient spatial pruning.
                * `partition_key` - Partition key used for partitioning, which can be a default measure_name or a customer defined partition key.
                * `table_arn` - ARN of the table with the most sub-optimal spatial pruning.
                * `value` - Maximum ratio of spatial coverage.
        * `query_table_count` - Number of tables in the query.
        * `query_temporal_range` - Insights into the temporal range of the query, including the table with the largest (max) time range. Following are some of the potential options for optimizing time-based pruning: add missing time-predicates, remove functions around the time predicates, add time predicates to all the sub-queries.
            * `max` - Insights into the most sub-optimal performing table on the temporal axis:
                * `table_arn` - ARN of the table which is queried with the largest time range.
                * `value` - Maximum duration in nanoseconds between the start and end of the query.
    * `run_status` - Status of a scheduled query run. Valid values: `AUTO_TRIGGER_SUCCESS`, `AUTO_TRIGGER_FAILURE`, `MANUAL_TRIGGER_SUCCESS`, `MANUAL_TRIGGER_FAILURE`.
    * `trigger_time` - Actual time when the query was run.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Timestream Query Scheduled Query using the `arn`. For example:

```terraform
import {
  to = aws_timestreamquery_scheduled_query.example
  id = "arn:aws:timestream:us-west-2:012345678901:scheduled-query/tf-acc-test-7774188528604787105-e13659544fe66c8d"
}
```

Using `terraform import`, import Timestream Query Scheduled Query using the `arn`. For example:

```console
% terraform import aws_timestreamquery_scheduled_query.example arn:aws:timestream:us-west-2:012345678901:scheduled-query/tf-acc-test-7774188528604787105-e13659544fe66c8d
```
