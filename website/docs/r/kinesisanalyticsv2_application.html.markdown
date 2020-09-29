---
subcategory: "Kinesis Data Analytics v2 (SQL and Java Applications)"
layout: "aws"
page_title: "AWS: aws_kinesisanalyticsv2_application"
description: |-
  Manages a Kinesis Analytics v2 Application.
---

# Resource: aws_apigatewayv2_api

Manages a Kinesis Analytics v2 Application.
This resource can be used to manage both Kinesis Data Analytics for SQL applications and Kinesis Data Analytics for Apache Flink applications.

## Example Usage

### Basic SQL Application

```hcl
resource "aws_kinesisanalyticsv2_application" "example" {
  name                = "example-sql-application"
  runtime_environment = "SQL-1.0"
}
```

### Basic Apache Flink Application

```hcl
resource "aws_kinesisanalyticsv2_application" "example" {
  name                = "example-flink-application"
  runtime_environment = "FLINK-1_8"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the application.
* `runtime_environment` - (Required) The runtime environment for the application. Valid values: `SQL-1_0`, `FLINK-1_6`, `FLINK-1_8`.
* `service_execution_role` - (Required) The ARN of the [IAM role](/docs/providers/aws/r/iam_role.html) used by the application to access Kinesis data streams, Kinesis Data Firehose delivery streams, Amazon S3 objects, and other external resources.
* `application_configuration` - (Optional) The application's configuration
* `cloudwatch_logging_options` - (Optional) An Amazon CloudWatch [log stream](/docs/providers/aws/r/cloudwatch_log_stream.html) to monitor application configuration errors.
* `description` - (Optional) A summary description of the application.
* `tags` - (Optional) A map of tags to assign to the application.

The `application_configuration` object supports the following:

* `application_code_configuration` - (Required) The code location and type parameters for the application.
* `application_snapshot_configuration` - (Optional) Describes whether snapshots are enabled for a Flink-based application.
* `environment_property` - (Optional) Describes execution properties for a Flink-based application.
* `flink_application_configuration` - (Optional) The configuration of a Flink-based application.
* `sql_application_configuration` - (Optional) The configuration of a SQL-based application.

The `application_code_configuration` object supports the following:

* `code_content_type` - (Required) Specifies whether the code content is in text or zip format. Valid values: `PLAINTEXT`, `ZIPFILE`.
* `code_content` - (Optional) The location and type of the application code.

The `code_content` object supports the following:

* `s3_content_location` - (Optional) Information about the Amazon S3 bucket containing the application code.
* `text_content` - (Optional) The text-format code for the application.

The `s3_content_location` object supports the following:

* `bucket_arn` - (Required) The ARN for the S3 bucket containing the application code.
* `file_key` - (Required) The file key for the object containing the application code.
* `object_version` - (Optional) The version of the object containing the application code.

The `application_snapshot_configuration` object supports the following:

* `snapshots_enabled` - (Required) Describes whether snapshots are enabled for a Flink-based Kinesis Data Analytics application.

The `flink_application_configuration` object supports the following:

* `checkpoint_configuration` - (Optional) Describes an application's checkpointing configuration.
* `monitoring_configuration` - (Optional) Describes configuration parameters for Amazon CloudWatch logging for an application.
* `parallelism_configuration` - (Optional) Describes parameters for how an application executes multiple tasks simultaneously.

The `checkpoint_configuration` object supports the following:

* `configuration_type` - (Required) Describes whether the application uses Kinesis Data Analytics' default checkpointing behavior. Valid values: `CUSTOM`, `DEFAULT`.
<br/>Set this attribute to `CUSTOM` in order for any specified `checkpointing_enabled`, `checkpoint_interval`, or `min_pause_between_checkpoints` attribute values to be effective.
<br/>If this attribute is set to `DEFAULT`, the application will always use the following values:
  * `checkpointing_enabled = true`
  * `checkpoint_interval = 60000`
  * `min_pause_between_checkpoints = 5000`
* `checkpointing_enabled` - (Optional) Describes whether checkpointing is enabled for a Flink-based Kinesis Data Analytics application.
* `checkpoint_interval` - (Optional) Describes the interval in milliseconds between checkpoint operations.
* `min_pause_between_checkpoints` - (Optional) Describes the minimum time in milliseconds after a checkpoint operation completes that a new checkpoint operation can start.

The `monitoring_configuration` object supports the following:

* `configuration_type` - (Required) Describes whether to use the default CloudWatch logging configuration for an application. Valid values: `CUSTOM`, `DEFAULT`.
<br/>Set this attribute to `CUSTOM` in order for any specified `log_level` or `metrics_level` attribute values to be effective.
* `log_level` - (Optional) Describes the verbosity of the CloudWatch Logs for an application. Valid values: `DEBUG`, `ERROR`, `INFO`, `WARN`.
* `metrics_level` - (Optional) Describes the granularity of the CloudWatch Logs for an application. Valid values: `APPLICATION`, `OPERATOR`, `PARALLELISM`, `TASK`.

The `parallelism_configuration` object supports the following:

* `configuration_type` - (Required) Describes whether the application uses the default parallelism for the Kinesis Data Analytics service. Valid values: `CUSTOM`, `DEFAULT`.
<br/>Set this attribute to `CUSTOM` in order for any specified `auto_scaling_enabled`, `parallelism`, or `parallelism_per_kpu` attribute values to be effective.
* `auto_scaling_enabled` - (Optional) Describes whether the Kinesis Data Analytics service can increase the parallelism of the application in response to increased throughput.
* `parallelism` - (Optional) Describes the initial number of parallel tasks that a Flink-based Kinesis Data Analytics application can perform.
* `parallelism_per_kpu` - (Optional) Describes the number of parallel tasks that a Flink-based Kinesis Data Analytics application can perform per Kinesis Processing Unit (KPU) used by the application.

The `cloudwatch_logging_options` object supports the following:

* `log_stream_arn` - (Required) The ARN of the CloudWatch log stream to receive application messages.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The application identifier.
* `arn` - The ARN of the application.
* `create_timestamp` - The current timestamp when the application was created.
* `last_update_timestamp` - The current timestamp when the application was last updated.
* `status` - The status of the application.
* `version_id` - The current application version. Kinesis Data Analytics updates the `version_id` each time the application is updated.

## Import

`aws_kinesisanalyticsv2_application` can be imported by using the application ARN, e.g.

```
$ terraform import aws_kinesisanalyticsv2_application.example arn:aws:kinesisanalytics:us-west-2:123456789012:application/example-sql-application
```
