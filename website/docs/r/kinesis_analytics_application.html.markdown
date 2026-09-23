---
subcategory: "Kinesis Analytics"
layout: "aws"
page_title: "AWS: aws_kinesis_analytics_application"
description: |-
  Provides a AWS Kinesis Analytics Application
---

# Resource: aws_kinesis_analytics_application

Provides a Kinesis Analytics Application resource. Kinesis Analytics is a managed service that
allows processing and analyzing streaming data using standard SQL.

For more details, see the [Amazon Kinesis Analytics Documentation](https://docs.aws.amazon.com/kinesisanalytics/latest/dev/what-is.html).

!> **WARNING:** _This resource is deprecated and will be removed in a future version._ [Effective January 27, 2026](https://aws.amazon.com/blogs/big-data/migrate-from-amazon-kinesis-data-analytics-for-sql-to-amazon-managed-service-for-apache-flink-and-amazon-managed-service-for-apache-flink-studio/), AWS will [no longer support](https://docs.aws.amazon.com/kinesisanalytics/latest/dev/discontinuation.html) Amazon Kinesis Data Analytics for SQL. Use the `aws_kinesisanalyticsv2_application` resource instead to manage Amazon Kinesis Data Analytics for Apache Flink applications. AWS provides guidance for migrating from [Amazon Kinesis Data Analytics for SQL Applications to Amazon Managed Service for Apache Flink Studio](https://aws.amazon.com/blogs/big-data/migrate-from-amazon-kinesis-data-analytics-for-sql-applications-to-amazon-managed-service-for-apache-flink-studio/) including [examples](https://docs.aws.amazon.com/kinesisanalytics/latest/dev/migrating-to-kda-studio-overview.html).

-> **Note:** To manage Amazon Kinesis Data Analytics for Apache Flink applications, use the [`aws_kinesisanalyticsv2_application`](/docs/providers/aws/r/kinesisanalyticsv2_application.html) resource.

## Example Usage

### Kinesis Stream Input

```terraform
resource "aws_kinesis_stream" "test_stream" {
  name        = "terraform-kinesis-test"
  shard_count = 1
}

resource "aws_kinesis_analytics_application" "test_application" {
  name = "kinesis-analytics-application-test"

  inputs {
    name_prefix = "test_prefix"

    kinesis_stream {
      resource_arn = aws_kinesis_stream.test_stream.arn
      role_arn     = aws_iam_role.test.arn
    }

    parallelism {
      count = 1
    }

    schema {
      record_columns {
        mapping  = "$.test"
        name     = "test"
        sql_type = "VARCHAR(8)"
      }

      record_encoding = "UTF-8"

      record_format {
        mapping_parameters {
          json {
            record_row_path = "$"
          }
        }
      }
    }
  }
}
```

### Starting An Application

```terraform
resource "aws_cloudwatch_log_group" "example" {
  name = "analytics"
}

resource "aws_cloudwatch_log_stream" "example" {
  name           = "example-kinesis-application"
  log_group_name = aws_cloudwatch_log_group.example.name
}

resource "aws_kinesis_stream" "example" {
  name        = "example-kinesis-stream"
  shard_count = 1
}

resource "aws_kinesis_firehose_delivery_stream" "example" {
  name        = "example-kinesis-delivery-stream"
  destination = "extended_s3"

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.example.arn
    role_arn   = aws_iam_role.example.arn
  }
}

resource "aws_kinesis_analytics_application" "test" {
  name = "example-application"

  cloudwatch_logging_options {
    log_stream_arn = aws_cloudwatch_log_stream.example.arn
    role_arn       = aws_iam_role.example.arn
  }

  inputs {
    name_prefix = "example_prefix"

    schema {
      record_columns {
        name     = "COLUMN_1"
        sql_type = "INTEGER"
      }

      record_format {
        mapping_parameters {
          csv {
            record_column_delimiter = ","
            record_row_delimiter    = "|"
          }
        }
      }
    }

    kinesis_stream {
      resource_arn = aws_kinesis_stream.example.arn
      role_arn     = aws_iam_role.example.arn
    }

    starting_position_configuration {
      starting_position = "NOW"
    }
  }

  outputs {
    name = "OUTPUT_1"

    schema {
      record_format_type = "CSV"
    }

    kinesis_firehose {
      resource_arn = aws_kinesis_firehose_delivery_stream.example.arn
      role_arn     = aws_iam_role.example.arn
    }
  }

  start_application = true
}
```

## Argument Reference

This resource supports the following arguments:

* `cloudwatch_logging_options` - (Optional) CloudWatch log stream options to monitor application errors. See [`cloudwatch_logging_options` Block](#cloudwatch_logging_options-block) below for details.
* `code` - (Optional) SQL Code to transform input data, and generate output.
* `description` - (Optional) Description of the application.
* `inputs` - (Optional) Input configuration of the application. See [`inputs` Block](#inputs-block) below for details.
* `name` - (Required) Name of the Kinesis Analytics Application.
* `outputs` - (Optional) Output destination configuration of the application. See [`outputs` Block](#outputs-block) below for details.
* `reference_data_sources` - (Optional) S3 Reference Data Source for the application. See [`reference_data_sources` Block](#reference_data_sources-block) below for details.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `start_application` - (Optional) Whether to start or stop the Kinesis Analytics Application. To start an application, an input with a defined `starting_position` must be configured. To modify an application's starting position, first stop the application by setting `start_application = false`, then update `starting_position` and set `start_application = true`.
* `tags` - (Optional) Key-value map of tags for the Kinesis Analytics Application. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `cloudwatch_logging_options` Block

* `log_stream_arn` - (Required) ARN of the CloudWatch Log Stream.
* `role_arn` - (Required) ARN of the IAM Role used to send application messages.

### `inputs` Block

* `kinesis_firehose` - (Optional) Kinesis Firehose configuration for the streaming source. Conflicts with `kinesis_stream`. See [`inputs.kinesis_firehose` Block](#inputskinesis_firehose-block) below for details.
* `kinesis_stream` - (Optional) Kinesis Stream configuration for the streaming source. Conflicts with `kinesis_firehose`. See [`inputs.kinesis_stream` Block](#inputskinesis_stream-block) below for details.
* `name_prefix` - (Required) Name Prefix to use when creating an in-application stream.
* `parallelism` - (Optional) Number of Parallel in-application streams to create. See [`inputs.parallelism` Block](#inputsparallelism-block) below for details.
* `processing_configuration` - (Optional) Processing Configuration to transform records as they are received from the stream. See [`inputs.processing_configuration` Block](#inputsprocessing_configuration-block) below for details.
* `schema` - (Required) Schema format of the data in the streaming source. See [`inputs.schema` Block](#inputsschema-block) below for details.
* `starting_position_configuration` - (Optional) Point at which the application starts processing records from the streaming source. See [`inputs.starting_position_configuration` Block](#inputsstarting_position_configuration-block) below for details.

### `inputs.kinesis_firehose` Block

* `resource_arn` - (Required) ARN of the Kinesis Firehose delivery stream.
* `role_arn` - (Required) ARN of the IAM Role used to access the stream.

### `inputs.kinesis_stream` Block

* `resource_arn` - (Required) ARN of the Kinesis Stream.
* `role_arn` - (Required) ARN of the IAM Role used to access the stream.

### `inputs.parallelism` Block

* `count` - (Optional) Count of streams.

### `inputs.processing_configuration` Block

* `lambda` - (Required) Lambda function configuration. See [`inputs.processing_configuration.lambda` Block](#inputsprocessing_configurationlambda-block) below for details.

### `inputs.processing_configuration.lambda` Block

* `resource_arn` - (Required) ARN of the Lambda function.
* `role_arn` - (Required) ARN of the IAM Role used to access the Lambda function.

### `inputs.schema` Block

* `record_columns` - (Required) Record Column mapping for the streaming source data element. See [`inputs.schema.record_columns` Block](#inputsschemarecord_columns-block) below for details.
* `record_encoding` - (Optional) Encoding of the record in the streaming source.
* `record_format` - (Required) Record Format and mapping information to schematize a record. See [`inputs.schema.record_format` Block](#inputsschemarecord_format-block) below for details.

### `inputs.schema.record_columns` Block

* `mapping` - (Optional) Mapping reference to the data element.
* `name` - (Required) Name of the column.
* `sql_type` - (Required) SQL Type of the column.

### `inputs.schema.record_format` Block

* `mapping_parameters` - (Optional) Mapping Information for the record format. See [`inputs.schema.record_format.mapping_parameters` Block](#inputsschemarecord_formatmapping_parameters-block) below for details.

### `inputs.schema.record_format.mapping_parameters` Block

* `csv` - (Optional) Mapping information when the record format uses delimiters. See [`inputs.schema.record_format.mapping_parameters.csv` Block](#inputsschemarecord_formatmapping_parameterscsv-block) below for details.
* `json` - (Optional) Mapping information when JSON is the record format on the streaming source. See [`inputs.schema.record_format.mapping_parameters.json` Block](#inputsschemarecord_formatmapping_parametersjson-block) below for details.

### `inputs.schema.record_format.mapping_parameters.csv` Block

* `record_column_delimiter` - (Required) Column Delimiter.
* `record_row_delimiter` - (Required) Row Delimiter.

### `inputs.schema.record_format.mapping_parameters.json` Block

* `record_row_path` - (Required) Path to the top-level parent that contains the records.

### `inputs.starting_position_configuration` Block

* `starting_position` - (Optional) Starting position on the stream. Valid values: `LAST_STOPPED_POINT`, `NOW`, `TRIM_HORIZON`.

### `outputs` Block

* `kinesis_firehose` - (Optional) Kinesis Firehose configuration for the destination stream. Conflicts with `kinesis_stream`. See [`outputs.kinesis_firehose` Block](#outputskinesis_firehose-block) below for details.
* `kinesis_stream` - (Optional) Kinesis Stream configuration for the destination stream. Conflicts with `kinesis_firehose`. See [`outputs.kinesis_stream` Block](#outputskinesis_stream-block) below for details.
* `lambda` - (Optional) Lambda function destination. See [`outputs.lambda` Block](#outputslambda-block) below for details.
* `name` - (Required) Name of the in-application stream.
* `schema` - (Required) Schema format of the data written to the destination. See [`outputs.schema` Block](#outputsschema-block) below for details.

### `outputs.kinesis_firehose` Block

* `resource_arn` - (Required) ARN of the Kinesis Firehose delivery stream.
* `role_arn` - (Required) ARN of the IAM Role used to access the stream.

### `outputs.kinesis_stream` Block

* `resource_arn` - (Required) ARN of the Kinesis Stream.
* `role_arn` - (Required) ARN of the IAM Role used to access the stream.

### `outputs.lambda` Block

* `resource_arn` - (Required) ARN of the Lambda function.
* `role_arn` - (Required) ARN of the IAM Role used to access the Lambda function.

### `outputs.schema` Block

* `record_format_type` - (Required) Format Type of the records on the output stream. Can be `CSV` or `JSON`.

### `reference_data_sources` Block

* `s3` - (Required) S3 configuration for the reference data source. See [`reference_data_sources.s3` Block](#reference_data_sourcess3-block) below for details.
* `schema` - (Required) Schema format of the data in the streaming source. See [`reference_data_sources.schema` Block](#reference_data_sourcesschema-block) below for details.
* `table_name` - (Required) In-application Table Name.

### `reference_data_sources.s3` Block

* `bucket_arn` - (Required) S3 Bucket ARN.
* `file_key` - (Required) File Key name containing reference data.
* `role_arn` - (Required) IAM Role ARN to read the data.

### `reference_data_sources.schema` Block

* `record_columns` - (Required) Record Column mapping for the streaming source data element. See [`reference_data_sources.schema.record_columns` Block](#reference_data_sourcesschemarecord_columns-block) below for details.
* `record_encoding` - (Optional) Encoding of the record in the streaming source.
* `record_format` - (Required) Record Format and mapping information to schematize a record. See [`reference_data_sources.schema.record_format` Block](#reference_data_sourcesschemarecord_format-block) below for details.

### `reference_data_sources.schema.record_columns` Block

* `mapping` - (Optional) Mapping reference to the data element.
* `name` - (Required) Name of the column.
* `sql_type` - (Required) SQL Type of the column.

### `reference_data_sources.schema.record_format` Block

* `mapping_parameters` - (Optional) Mapping Information for the record format. See [`reference_data_sources.schema.record_format.mapping_parameters` Block](#reference_data_sourcesschemarecord_formatmapping_parameters-block) below for details.

### `reference_data_sources.schema.record_format.mapping_parameters` Block

* `csv` - (Optional) Mapping information when the record format uses delimiters. See [`reference_data_sources.schema.record_format.mapping_parameters.csv` Block](#reference_data_sourcesschemarecord_formatmapping_parameterscsv-block) below for details.
* `json` - (Optional) Mapping information when JSON is the record format on the streaming source. See [`reference_data_sources.schema.record_format.mapping_parameters.json` Block](#reference_data_sourcesschemarecord_formatmapping_parametersjson-block) below for details.

### `reference_data_sources.schema.record_format.mapping_parameters.csv` Block

* `record_column_delimiter` - (Required) Column Delimiter.
* `record_row_delimiter` - (Required) Row Delimiter.

### `reference_data_sources.schema.record_format.mapping_parameters.json` Block

* `record_row_path` - (Required) Path to the top-level parent that contains the records.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Kinesis Analytics Application.
* `create_timestamp` - Timestamp when the application version was created.
* `id` - ARN of the Kinesis Analytics Application.
* `inputs.schema.record_format.record_format_type` - Type of Record Format of the input streaming source.
* `inputs.stream_names` - Names of the in-application streams created for the input.
* `last_update_timestamp` - Timestamp when the application was last updated.
* `reference_data_sources.schema.record_format.record_format_type` - Type of Record Format of the reference data source.
* `status` - Status of the application.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `version` - Version of the application.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Kinesis Analytics Application using ARN. For example:

```terraform
import {
  to = aws_kinesis_analytics_application.example
  id = "arn:aws:kinesisanalytics:us-west-2:1234567890:application/example"
}
```

Using `terraform import`, import Kinesis Analytics Application using ARN. For example:

```console
% terraform import aws_kinesis_analytics_application.example arn:aws:kinesisanalytics:us-west-2:1234567890:application/example
```
