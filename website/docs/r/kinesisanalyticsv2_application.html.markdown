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
* `runtime_environment` - (Required) The runtime environment for the application. Valid values: `SQL-1.0`, `FLINK-1_6`, `FLINK-1_8`.
* `service_execution_role` - (Required) The ARN of the [IAM role](/docs/providers/aws/r/iam_role.html) used by the application to access Kinesis data streams, Kinesis Data Firehose delivery streams, Amazon S3 objects, and other external resources.
* `application_configuration` - (Optional) The application's configuration
* `cloudwatch_logging_options` - (Optional) An Amazon CloudWatch [log stream](/docs/providers/aws/r/cloudwatch_log_stream.html) to monitor application configuration errors.
* `description` - (Optional) A summary description of the application.
* `tags` - (Optional) A map of tags to assign to the application.

The `application_configuration` object supports the following:

* `application_code_configuration` - (Optional) The code location and type parameters for a Flink-based application.
* `application_snapshot_configuration` - (Optional) Describes whether snapshots are enabled for a Flink-based application.
* `environment_property` - (Optional) Describes execution properties for a Flink-based application.
* `flink_application_configuration` - (Optional) The configuration of a Flink-based application.
* `sql_application_configuration` - (Optional) The configuration of a SQL-based application.

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
