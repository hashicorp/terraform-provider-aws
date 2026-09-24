---
subcategory: "Kinesis Analytics V2"
layout: "aws"
page_title: "AWS: aws_kinesisanalyticsv2_application"
description: |-
  Manages a Kinesis Analytics v2 Application.
---

# Resource: aws_kinesisanalyticsv2_application

Manages a Kinesis Analytics v2 Application.
This resource can be used to manage both Kinesis Data Analytics for SQL applications and Kinesis Data Analytics for Apache Flink applications.

-> **Note:** Kinesis Data Analytics for SQL applications created using this resource cannot currently be viewed in the AWS Console. To manage Kinesis Data Analytics for SQL applications that can also be viewed in the AWS Console, use the [`aws_kinesis_analytics_application`](/docs/providers/aws/r/kinesis_analytics_application.html) resource.

## Example Usage

### Apache Flink Application

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "example-flink-application"
}

resource "aws_s3_object" "example" {
  bucket = aws_s3_bucket.example.id
  key    = "example-flink-application"
  source = "flink-app.jar"
}

resource "aws_kinesisanalyticsv2_application" "example" {
  name                   = "example-flink-application"
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.example.arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.example.arn
          file_key   = aws_s3_object.example.key
        }
      }

      code_content_type = "ZIPFILE"
    }

    environment_properties {
      property_group {
        property_group_id = "PROPERTY-GROUP-1"

        property_map = {
          Key1 = "Value1"
        }
      }

      property_group {
        property_group_id = "PROPERTY-GROUP-2"

        property_map = {
          KeyA = "ValueA"
          KeyB = "ValueB"
        }
      }
    }

    flink_application_configuration {
      checkpoint_configuration {
        configuration_type = "DEFAULT"
      }

      monitoring_configuration {
        configuration_type = "CUSTOM"
        log_level          = "DEBUG"
        metrics_level      = "TASK"
      }

      parallelism_configuration {
        auto_scaling_enabled = true
        configuration_type   = "CUSTOM"
        parallelism          = 10
        parallelism_per_kpu  = 4
      }
    }
  }

  tags = {
    Environment = "test"
  }
}
```

### SQL Application

```terraform
resource "aws_cloudwatch_log_group" "example" {
  name = "example-sql-application"
}

resource "aws_cloudwatch_log_stream" "example" {
  name           = "example-sql-application"
  log_group_name = aws_cloudwatch_log_group.example.name
}

resource "aws_kinesisanalyticsv2_application" "example" {
  name                   = "example-sql-application"
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.example.arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      input {
        name_prefix = "PREFIX_1"

        input_parallelism {
          count = 3
        }

        input_schema {
          record_column {
            name     = "COLUMN_1"
            sql_type = "VARCHAR(8)"
            mapping  = "MAPPING-1"
          }

          record_column {
            name     = "COLUMN_2"
            sql_type = "DOUBLE"
          }

          record_encoding = "UTF-8"

          record_format {
            record_format_type = "CSV"

            mapping_parameters {
              csv_mapping_parameters {
                record_column_delimiter = ","
                record_row_delimiter    = "\n"
              }
            }
          }
        }

        kinesis_streams_input {
          resource_arn = aws_kinesis_stream.example.arn
        }
      }

      output {
        name = "OUTPUT_1"

        destination_schema {
          record_format_type = "JSON"
        }

        lambda_output {
          resource_arn = aws_lambda_function.example.arn
        }
      }

      output {
        name = "OUTPUT_2"

        destination_schema {
          record_format_type = "CSV"
        }

        kinesis_firehose_output {
          resource_arn = aws_kinesis_firehose_delivery_stream.example.arn
        }
      }

      reference_data_source {
        table_name = "TABLE-1"

        reference_schema {
          record_column {
            name     = "COLUMN_1"
            sql_type = "INTEGER"
          }

          record_format {
            record_format_type = "JSON"

            mapping_parameters {
              json_mapping_parameters {
                record_row_path = "$"
              }
            }
          }
        }

        s3_reference_data_source {
          bucket_arn = aws_s3_bucket.example.arn
          file_key   = "KEY-1"
        }
      }
    }
  }

  cloudwatch_logging_options {
    log_stream_arn = aws_cloudwatch_log_stream.example.arn
  }
}
```

### VPC Configuration

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "example-flink-application"
}

resource "aws_s3_object" "example" {
  bucket = aws_s3_bucket.example.id
  key    = "example-flink-application"
  source = "flink-app.jar"
}

resource "aws_kinesisanalyticsv2_application" "example" {
  name                   = "example-flink-application"
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.example.arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.example.arn
          file_key   = aws_s3_object.example.key
        }
      }

      code_content_type = "ZIPFILE"
    }

    vpc_configuration {
      security_group_ids = [aws_security_group.example[0].id, aws_security_group.example[1].id]
      subnet_ids         = [aws_subnet.example.id]
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the application.
* `runtime_environment` - (Required) Runtime environment for the application. Valid values: `SQL-1_0`, `FLINK-1_6`, `FLINK-1_8`, `FLINK-1_11`, `FLINK-1_13`, `FLINK-1_15`, `FLINK-1_18`, `FLINK-1_19`, `FLINK-1_20`, `FLINK-2_2`.
* `service_execution_role` - (Required) ARN of the [IAM role](/docs/providers/aws/r/iam_role.html) used by the application to access Kinesis data streams, Kinesis Data Firehose delivery streams, Amazon S3 objects, and other external resources.

The following arguments are optional:

* `application_configuration` - (Optional) Application configuration. See [`application_configuration` Block](#application_configuration-block) below.
* `application_mode` - (Optional) Application's mode. Valid values are `STREAMING`, `INTERACTIVE`.
* `cloudwatch_logging_options` - (Optional) CloudWatch log stream to monitor application configuration errors. See [`cloudwatch_logging_options` Block](#cloudwatch_logging_options-block) below.
* `description` - (Optional) Summary description of the application.
* `force_stop` - (Optional) Whether to force stop an unresponsive Flink-based application.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `start_application` - (Optional) Whether to start or stop the application.
* `tags` - (Optional) Map of tags to assign to the application. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `application_configuration` Block

The `application_configuration` configuration block supports the following arguments:

* `application_code_configuration` - (Required) Code location and type parameters for the application. See [`application_code_configuration` Block](#application_code_configuration-block) below.
* `application_encryption_configuration` - (Optional) Encryption configuration for the application. Use this to encrypt data at rest in the application. See [`application_encryption_configuration` Block](#application_encryption_configuration-block) below.
* `application_snapshot_configuration` - (Optional) Snapshot configuration for a Flink-based application. See [`application_snapshot_configuration` Block](#application_snapshot_configuration-block) below.
* `environment_properties` - (Optional) Execution properties for a Flink-based application. See [`environment_properties` Block](#environment_properties-block) below.
* `flink_application_configuration` - (Optional) Configuration of a Flink-based application. See [`flink_application_configuration` Block](#flink_application_configuration-block) below.
* `run_configuration` - (Optional) Starting properties for a Flink-based application. See [`run_configuration` Block](#run_configuration-block) below.
* `sql_application_configuration` - (Optional) Configuration of a SQL-based application. See [`sql_application_configuration` Block](#sql_application_configuration-block) below.
* `vpc_configuration` - (Optional) VPC configuration of a Flink-based application. See [`vpc_configuration` Block](#vpc_configuration-block) below.

### `application_code_configuration` Block

The `application_code_configuration` configuration block supports the following arguments:

* `code_content` - (Optional) Location and type of the application code. See [`code_content` Block](#code_content-block) below.
* `code_content_type` - (Required) Whether the code content is in text or zip format. Valid values: `PLAINTEXT`, `ZIPFILE`.

### `code_content` Block

The `code_content` configuration block supports the following arguments:

* `s3_content_location` - (Optional) Information about the Amazon S3 bucket containing the application code. See [`s3_content_location` Block](#s3_content_location-block) below.
* `text_content` - (Optional) Text-format code for the application.

### `s3_content_location` Block

The `s3_content_location` configuration block supports the following arguments:

* `bucket_arn` - (Required) ARN for the S3 bucket containing the application code.
* `file_key` - (Required) File key for the object containing the application code.
* `object_version` - (Optional) Version of the object containing the application code.

### `application_encryption_configuration` Block

The `application_encryption_configuration` configuration block supports the following arguments:

* `key_id` - (Optional) ARN of the KMS key to use for encryption. Required when `key_type` is set to `CUSTOMER_MANAGED_KEY`. The KMS key must be in the same Region as the application.
* `key_type` - (Required) Type of encryption key to use. Valid values: `CUSTOMER_MANAGED_KEY`, `AWS_OWNED_KEY`.

### `application_snapshot_configuration` Block

The `application_snapshot_configuration` configuration block supports the following arguments:

* `snapshots_enabled` - (Required) Whether snapshots are enabled for a Flink-based application.

### `environment_properties` Block

The `environment_properties` configuration block supports the following arguments:

* `property_group` - (Required) Execution property groups. See [`property_group` Block](#property_group-block) below.

### `property_group` Block

The `property_group` configuration block supports the following arguments:

* `property_group_id` - (Required) Key of the application execution property key-value map.
* `property_map` - (Required) Application execution property key-value map.

### `flink_application_configuration` Block

The `flink_application_configuration` configuration block supports the following arguments:

* `checkpoint_configuration` - (Optional) Application's checkpointing configuration. See [`checkpoint_configuration` Block](#checkpoint_configuration-block) below.
* `monitoring_configuration` - (Optional) Configuration parameters for CloudWatch logging for an application. See [`monitoring_configuration` Block](#monitoring_configuration-block) below.
* `parallelism_configuration` - (Optional) Parameters for how an application executes multiple tasks simultaneously. See [`parallelism_configuration` Block](#parallelism_configuration-block) below.

### `checkpoint_configuration` Block

The `checkpoint_configuration` configuration block supports the following arguments:

* `checkpoint_interval` - (Optional) Interval in milliseconds between checkpoint operations.
* `checkpointing_enabled` - (Optional) Whether checkpointing is enabled for a Flink-based application.
* `configuration_type` - (Required) Whether the application uses Kinesis Data Analytics' default checkpointing behavior. Valid values: `CUSTOM`, `DEFAULT`. Set this attribute to `CUSTOM` in order for any specified `checkpointing_enabled`, `checkpoint_interval`, or `min_pause_between_checkpoints` attribute values to be effective. If this attribute is set to `DEFAULT`, the application will always use the following values: `checkpointing_enabled = true`, `checkpoint_interval = 60000`, and `min_pause_between_checkpoints = 5000`.
* `min_pause_between_checkpoints` - (Optional) Minimum time in milliseconds after a checkpoint operation completes that a new checkpoint operation can start.

### `monitoring_configuration` Block

The `monitoring_configuration` configuration block supports the following arguments:

* `configuration_type` - (Required) Whether to use the default CloudWatch logging configuration for an application. Valid values: `CUSTOM`, `DEFAULT`. Set this attribute to `CUSTOM` in order for any specified `log_level` or `metrics_level` attribute values to be effective.
* `log_level` - (Optional) Verbosity of the CloudWatch Logs for an application. Valid values: `DEBUG`, `ERROR`, `INFO`, `WARN`.
* `metrics_level` - (Optional) Granularity of the CloudWatch Logs for an application. Valid values: `APPLICATION`, `OPERATOR`, `PARALLELISM`, `TASK`.

### `parallelism_configuration` Block

The `parallelism_configuration` configuration block supports the following arguments:

* `auto_scaling_enabled` - (Optional) Whether the Kinesis Data Analytics service can increase the parallelism of the application in response to increased throughput.
* `configuration_type` - (Required) Whether the application uses the default parallelism for the Kinesis Data Analytics service. Valid values: `CUSTOM`, `DEFAULT`. Set this attribute to `CUSTOM` in order for any specified `auto_scaling_enabled`, `parallelism`, or `parallelism_per_kpu` attribute values to be effective.
* `parallelism` - (Optional) Initial number of parallel tasks that a Flink-based application can perform.
* `parallelism_per_kpu` - (Optional) Number of parallel tasks that a Flink-based application can perform per Kinesis Processing Unit (KPU) used by the application.

### `run_configuration` Block

The `run_configuration` configuration block supports the following arguments:

* `application_restore_configuration` - (Optional) Restore behavior of a restarting application. See [`application_restore_configuration` Block](#application_restore_configuration-block) below.
* `flink_run_configuration` - (Optional) Starting parameters for a Flink-based application. See [`flink_run_configuration` Block](#flink_run_configuration-block) below.

### `application_restore_configuration` Block

The `application_restore_configuration` configuration block supports the following arguments:

* `application_restore_type` - (Optional) How the application should be restored. Valid values: `RESTORE_FROM_CUSTOM_SNAPSHOT`, `RESTORE_FROM_LATEST_SNAPSHOT`, `SKIP_RESTORE_FROM_SNAPSHOT`.
* `snapshot_name` - (Optional) Identifier of an existing snapshot of application state to use to restart an application. The application uses this value if `RESTORE_FROM_CUSTOM_SNAPSHOT` is specified for `application_restore_type`.

### `flink_run_configuration` Block

The `flink_run_configuration` configuration block supports the following arguments:

* `allow_non_restored_state` - (Optional) Whether the runtime is allowed to skip a state that cannot be mapped to the new program when restoring from a snapshot. Default is `false`.

### `sql_application_configuration` Block

The `sql_application_configuration` configuration block supports the following arguments:

* `input` - (Optional) Input stream used by the application. See [`input` Block](#input-block) below.
* `output` - (Optional) Destination streams used by the application. See [`output` Block](#output-block) below.
* `reference_data_source` - (Optional) Reference data source used by the application. See [`reference_data_source` Block](#reference_data_source-block) below.

### `input` Block

The `input` configuration block supports the following arguments:

* `input_parallelism` - (Optional) Number of in-application streams to create. See [`input_parallelism` Block](#input_parallelism-block) below.
* `input_processing_configuration` - (Optional) Input processing configuration for the input. An input processor transforms records as they are received from the stream, before the application's SQL code executes. See [`input_processing_configuration` Block](#input_processing_configuration-block) below.
* `input_schema` - (Required) Format of the data in the streaming source, and how each data element maps to corresponding columns in the in-application stream that is being created. See [`input_schema` Block](#input_schema-block) below.
* `input_starting_position_configuration` - (Optional) Point at which the application starts processing records from the streaming source. See [`input_starting_position_configuration` Block](#input_starting_position_configuration-block) below.
* `kinesis_firehose_input` - (Optional) If the streaming source is a [Kinesis Data Firehose delivery stream](/docs/providers/aws/r/kinesis_firehose_delivery_stream.html), identifies the delivery stream's ARN. See [`kinesis_firehose_input` Block](#kinesis_firehose_input-block) below.
* `kinesis_streams_input` - (Optional) If the streaming source is a [Kinesis data stream](/docs/providers/aws/r/kinesis_stream.html), identifies the stream's ARN. See [`kinesis_streams_input` Block](#kinesis_streams_input-block) below.
* `name_prefix` - (Required) Name prefix to use when creating an in-application stream.

### `input_parallelism` Block

The `input_parallelism` configuration block supports the following arguments:

* `count` - (Optional) Number of in-application streams to create.

### `input_processing_configuration` Block

The `input_processing_configuration` configuration block supports the following arguments:

* `input_lambda_processor` - (Required) [Lambda function](/docs/providers/aws/r/lambda_function.html) used to preprocess the records in the stream before being processed by your application code. See [`input_lambda_processor` Block](#input_lambda_processor-block) below.

### `input_lambda_processor` Block

The `input_lambda_processor` configuration block supports the following arguments:

* `resource_arn` - (Required) ARN of the Lambda function that operates on records in the stream.

### `input_schema` Block

The `input_schema` configuration block supports the following arguments:

* `record_column` - (Required) Mapping of each data element in the streaming source to the corresponding column in the in-application stream. See [`record_column` Block](#record_column-block) below.
* `record_encoding` - (Optional) Encoding of the records in the streaming source. For example, `UTF-8`.
* `record_format` - (Required) Format of the records on the streaming source. See [`record_format` Block](#record_format-block) below.

### `record_column` Block

The `record_column` configuration block supports the following arguments:

* `mapping` - (Optional) Reference to the data element in the streaming input or the reference data source.
* `name` - (Required) Name of the column that is created in the in-application input stream or reference table.
* `sql_type` - (Required) Type of column created in the in-application input stream or reference table.

### `record_format` Block

The `record_format` configuration block supports the following arguments:

* `mapping_parameters` - (Required) Additional mapping information specific to the record format (such as JSON, CSV, or record fields delimited by some delimiter) on the streaming source. See [`mapping_parameters` Block](#mapping_parameters-block) below.
* `record_format_type` - (Required) Type of record format. Valid values: `CSV`, `JSON`.

### `mapping_parameters` Block

The `mapping_parameters` configuration block supports the following arguments:

* `csv_mapping_parameters` - (Optional) Additional mapping information when the record format uses delimiters (for example, CSV). See [`csv_mapping_parameters` Block](#csv_mapping_parameters-block) below.
* `json_mapping_parameters` - (Optional) Additional mapping information when JSON is the record format on the streaming source. See [`json_mapping_parameters` Block](#json_mapping_parameters-block) below.

### `csv_mapping_parameters` Block

The `csv_mapping_parameters` configuration block supports the following arguments:

* `record_column_delimiter` - (Required) Column delimiter. For example, in a CSV format, a comma (`,`) is the typical column delimiter.
* `record_row_delimiter` - (Required) Row delimiter. For example, in a CSV format, `\n` is the typical row delimiter.

### `json_mapping_parameters` Block

The `json_mapping_parameters` configuration block supports the following arguments:

* `record_row_path` - (Required) Path to the top-level parent that contains the records.

### `input_starting_position_configuration` Block

~> **NOTE:** To modify an application's starting position, first stop the application by setting `start_application = false`, then update `starting_position` and set `start_application = true`.

The `input_starting_position_configuration` configuration block supports the following arguments:

* `input_starting_position` - (Optional) Starting position on the stream. Valid values: `LAST_STOPPED_POINT`, `NOW`, `TRIM_HORIZON`.

### `kinesis_firehose_input` Block

The `kinesis_firehose_input` configuration block supports the following arguments:

* `resource_arn` - (Required) ARN of the delivery stream.

### `kinesis_streams_input` Block

The `kinesis_streams_input` configuration block supports the following arguments:

* `resource_arn` - (Required) ARN of the input Kinesis data stream to read.

### `output` Block

The `output` configuration block supports the following arguments:

* `destination_schema` - (Required) Data format when records are written to the destination. See [`destination_schema` Block](#destination_schema-block) below.
* `kinesis_firehose_output` - (Optional) Destination [Kinesis Data Firehose delivery stream](/docs/providers/aws/r/kinesis_firehose_delivery_stream.html). See [`kinesis_firehose_output` Block](#kinesis_firehose_output-block) below.
* `kinesis_streams_output` - (Optional) Destination [Kinesis data stream](/docs/providers/aws/r/kinesis_stream.html). See [`kinesis_streams_output` Block](#kinesis_streams_output-block) below.
* `lambda_output` - (Optional) Destination [Lambda function](/docs/providers/aws/r/lambda_function.html). See [`lambda_output` Block](#lambda_output-block) below.
* `name` - (Required) Name of the in-application stream.

### `destination_schema` Block

The `destination_schema` configuration block supports the following arguments:

* `record_format_type` - (Required) Format of the records on the output stream. Valid values: `CSV`, `JSON`.

### `kinesis_firehose_output` Block

The `kinesis_firehose_output` configuration block supports the following arguments:

* `resource_arn` - (Required) ARN of the destination delivery stream to write to.

### `kinesis_streams_output` Block

The `kinesis_streams_output` configuration block supports the following arguments:

* `resource_arn` - (Required) ARN of the destination Kinesis data stream to write to.

### `lambda_output` Block

The `lambda_output` configuration block supports the following arguments:

* `resource_arn` - (Required) ARN of the destination Lambda function to write to.

### `reference_data_source` Block

The `reference_data_source` configuration block supports the following arguments:

* `reference_schema` - (Required) Format of the data in the streaming source, and how each data element maps to corresponding columns created in the in-application stream. See [`reference_schema` Block](#reference_schema-block) below.
* `s3_reference_data_source` - (Required) S3 bucket and object that contains the reference data. See [`s3_reference_data_source` Block](#s3_reference_data_source-block) below.
* `table_name` - (Required) Name of the in-application table to create.

### `reference_schema` Block

The `reference_schema` configuration block supports the following arguments:

* `record_column` - (Required) Mapping of each data element in the streaming source to the corresponding column in the in-application stream. See [`record_column` Block](#record_column-block) above.
* `record_encoding` - (Optional) Encoding of the records in the streaming source. For example, `UTF-8`.
* `record_format` - (Required) Format of the records on the streaming source. See [`record_format` Block](#record_format-block) above.

### `s3_reference_data_source` Block

The `s3_reference_data_source` configuration block supports the following arguments:

* `bucket_arn` - (Required) ARN of the S3 bucket.
* `file_key` - (Required) Object key name containing the reference data.

### `vpc_configuration` Block

The `vpc_configuration` configuration block supports the following arguments:

* `security_group_ids` - (Required) [Security Group](/docs/providers/aws/r/security_group.html) IDs used by the VPC configuration.
* `subnet_ids` - (Required) [Subnet](/docs/providers/aws/r/subnet.html) IDs used by the VPC configuration.

### `cloudwatch_logging_options` Block

The `cloudwatch_logging_options` configuration block supports the following arguments:

* `log_stream_arn` - (Required) ARN of the CloudWatch log stream to receive application messages.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the application.
* `create_timestamp` - Current timestamp when the application was created.
* `id` - Application identifier.
* `last_update_timestamp` - Current timestamp when the application was last updated.
* `status` - Status of the application.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `version_id` - Current application version. Kinesis Data Analytics updates the `version_id` each time the application is updated.

### `input` Block

The `input` block exports the following attributes in addition to the arguments above:

* `in_app_stream_names` - In-application stream names.
* `input_id` - Identifier of the input configuration.

### `output` Block

The `output` block exports the following attributes in addition to the arguments above:

* `output_id` - Identifier of the output configuration.

### `reference_data_source` Block

The `reference_data_source` block exports the following attributes in addition to the arguments above:

* `reference_id` - Identifier of the reference data source.

### `vpc_configuration` Block

The `vpc_configuration` block exports the following attributes in addition to the arguments above:

* `vpc_configuration_id` - Identifier of the VPC configuration.
* `vpc_id` - Identifier of the VPC.

### `cloudwatch_logging_options` Block

The `cloudwatch_logging_options` block exports the following attributes in addition to the arguments above:

* `cloudwatch_logging_option_id` - Identifier of the CloudWatch logging option.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `update` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_kinesisanalyticsv2_application` using the application ARN. For example:

```terraform
import {
  to = aws_kinesisanalyticsv2_application.example
  id = "arn:aws:kinesisanalytics:us-west-2:123456789012:application/example-sql-application"
}
```

Using `terraform import`, import `aws_kinesisanalyticsv2_application` using the application ARN. For example:

```console
% terraform import aws_kinesisanalyticsv2_application.example arn:aws:kinesisanalytics:us-west-2:123456789012:application/example-sql-application
```
