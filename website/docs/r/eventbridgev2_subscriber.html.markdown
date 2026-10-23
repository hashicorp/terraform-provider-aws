---
subcategory: "EventBridge V2"
layout: "aws"
page_title: "AWS: aws_eventbridgev2_subscriber"
description: |-
  Manages an Amazon EventBridge V2 subscriber.
---

# Resource: aws_eventbridgev2_subscriber

Manages an Amazon EventBridge V2 subscriber. A subscriber delivers events that match its filters from an event bus to a target.

## Example Usage

### Basic Usage

```terraform
resource "aws_eventbridgev2_event_bus" "example" {
  name = "example"
}

resource "aws_sqs_queue" "example" {
  name = "example"
}

resource "aws_iam_role" "example" {
  name = "example"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "events.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "example" {
  name = "example"
  role = aws_iam_role.example.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action   = "sqs:SendMessage"
      Effect   = "Allow"
      Resource = aws_sqs_queue.example.arn
    }]
  })
}

resource "aws_eventbridgev2_subscriber" "example" {
  name          = "example"
  event_bus_arn = aws_eventbridgev2_event_bus.example.arn

  filter_configuration {
    filter {
      scope   = "METADATA"
      pattern = jsonencode({ source = ["example"] })
    }
  }

  invoke_configuration {
    role_arn   = aws_iam_role.example.arn
    target_arn = aws_sqs_queue.example.arn
  }

  depends_on = [aws_iam_role_policy.example]
}
```

### Batching, Retries, and Logging

```terraform
resource "aws_eventbridgev2_subscriber" "example" {
  name          = "example"
  event_bus_arn = aws_eventbridgev2_event_bus.example.arn

  batch_configuration {
    max_batch_size = 10
  }

  filter_configuration {
    filter {
      scope   = "METADATA"
      pattern = jsonencode({ source = ["example"] })
    }
  }

  invoke_configuration {
    role_arn   = aws_iam_role.example.arn
    target_arn = aws_sqs_queue.example.arn

    sqs_parameters {
      message_attributes {
        name         = "origin"
        data_type    = "String"
        string_value = "eventbridge"
      }
    }
  }

  log_configuration {
    level = "ERROR"
  }

  on_failure_configuration {
    arn = aws_sqs_queue.dlq.arn
  }

  retry_policy {
    max_retry_attempts = 3
  }
}
```

## Argument Reference

The following arguments are required:

* `event_bus_arn` - (Required) ARN of the event bus the subscriber receives events from. Can be an event bus shared with this account through AWS RAM. Changing this value creates a new resource.
* `filter_configuration` - (Required) Filters that select which events are delivered. See [`filter_configuration` Block](#filter_configuration-block) below.
* `invoke_configuration` - (Required) Configuration of how the subscriber invokes its target. See [`invoke_configuration` Block](#invoke_configuration-block) below.
* `name` - (Required) Name of the subscriber. Must start with an alphanumeric character and contain only alphanumeric characters, periods (`.`), hyphens (`-`), and underscores (`_`). Changing this value creates a new resource.

The following arguments are optional:

* `batch_configuration` - (Optional) Batching configuration. If omitted, the service defaults are applied and not tracked. See [`batch_configuration` Block](#batch_configuration-block) below.
* `description` - (Optional) Description of the subscriber.
* `log_configuration` - (Optional) Logging configuration. If omitted, the service defaults are applied and not tracked. See [`log_configuration` Block](#log_configuration-block) below.
* `on_failure_configuration` - (Optional) Destination for events that could not be delivered. See [`on_failure_configuration` Block](#on_failure_configuration-block) below.
* `point_in_time_configuration` - (Optional) Point in time from which delivery starts. Only applicable when `starting_position` is `POINT_IN_TIME`. Changing this value creates a new resource. See [`point_in_time_configuration` Block](#point_in_time_configuration-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resume_position` - (Optional) Position from which delivery resumes when `state` changes from `STOPPED` to `RUNNING`. Ignored for any other change. Allowed values are: `LAST_PROCESSED` and `LATEST`. Default value: `LAST_PROCESSED`.
* `retry_policy` - (Optional) Retry policy. If omitted, the service defaults are applied and not tracked. See [`retry_policy` Block](#retry_policy-block) below.
* `starting_position` - (Optional) Position in the event bus from which delivery starts. Allowed values are: `LATEST` and `POINT_IN_TIME`. Changing this value creates a new resource.
* `state` - (Optional) Run state of the subscriber. Allowed values are: `RUNNING` and `STOPPED`.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `transformer` - (Optional) Transformation applied to events before delivery. Not supported for universal (`aws-sdk`) targets. If omitted, the service default is applied and not tracked. See [`transformer` Block](#transformer-block) below.
* `type` - (Optional) Delivery ordering mode. Allowed values are: `FIFO` and `UNORDERED`. Changing this value creates a new resource.

### `batch_configuration` Block

The `batch_configuration` block supports the following:

* `max_batch_size` - (Optional) Maximum number of events delivered to the target in a single batch. Valid value is between `1` and `500`; some targets have a lower maximum, for example `10` for Amazon SQS.
* `max_batch_window_in_seconds` - (Optional) Maximum time, in seconds, to wait for a batch to fill before delivery. Valid value is between `0` and `300`.

### `filter_configuration` Block

The `filter_configuration` block supports the following:

* `filter` - (Required) One or more filters. An event must match every filter to be delivered. See [`filter` Block](#filter-block) below.
* `language` - (Optional) Filter language. Allowed values are: `EVENT_BRIDGE_PATTERN`. Default value: `EVENT_BRIDGE_PATTERN`.

### `filter` Block

The `filter` block supports the following:

* `pattern` - (Required) JSON event pattern that selects matching events.
* `scope` - (Required) Part of the event the pattern applies to. Allowed values are: `DATA`, `METADATA`, and `SYSTEM_METADATA`.

### `invoke_configuration` Block

The `invoke_configuration` block supports the following. At most one target-specific parameters block may be specified, and it must match the target type. Unless noted otherwise, string arguments of the target-specific parameters blocks accept either a literal value or a JSONata expression, for example `{% $events.Data.groupId %}`.

* `event_bus_v2_parameters` - (Optional) Parameters for forwarding events to another EventBridge V2 event bus. See [`event_bus_v2_parameters` Block](#event_bus_v2_parameters-block) below.
* `http_parameters` - (Optional) Parameters for Amazon API Gateway and API destination targets. See [`http_parameters` Block](#http_parameters-block) below.
* `kinesis_parameters` - (Optional) Parameters for Amazon Kinesis Data Streams targets. See [`kinesis_parameters` Block](#kinesis_parameters-block) below.
* `lambda_parameters` - (Optional) Parameters for AWS Lambda targets. See [`lambda_parameters` Block](#lambda_parameters-block) below.
* `role_arn` - (Required) ARN of the IAM role the service assumes to invoke the target. Must belong to the same account as the subscriber.
* `sns_parameters` - (Optional) Parameters for Amazon SNS targets. See [`sns_parameters` Block](#sns_parameters-block) below.
* `sqs_parameters` - (Optional) Parameters for Amazon SQS targets. See [`sqs_parameters` Block](#sqs_parameters-block) below.
* `step_functions_parameters` - (Optional) Parameters for AWS Step Functions targets. See [`step_functions_parameters` Block](#step_functions_parameters-block) below.
* `target_arn` - (Required) ARN of the target. Changing this value creates a new resource.
* `universal_target_parameters` - (Optional) Parameters for universal targets, whose `target_arn` has the form `arn:aws:events:::aws-sdk:<service>:<apiAction>`. See [`universal_target_parameters` Block](#universal_target_parameters-block) below.

### `event_bus_v2_parameters` Block

The `event_bus_v2_parameters` block supports the following:

* `deduplication_configuration` - (Optional) Deduplication settings applied to forwarded events on the downstream event bus. See [`deduplication_configuration` Block](#deduplication_configuration-block) below.
* `metadata` - (Optional) Map of metadata forwarded with each event.
* `system_metadata` - (Optional) System metadata attached to each forwarded event. See [`system_metadata` Block](#system_metadata-block) below.

### `deduplication_configuration` Block

The `deduplication_configuration` block supports the following:

* `deduplication_type` - (Required) How duplicate events are detected. Allowed values are: `CONTENT_BASED`.

### `system_metadata` Block

The `system_metadata` block supports the following:

* `deduplication_id` - (Optional) Deduplication ID for FIFO deduplication on the downstream event bus.
* `event_group_id` - (Optional) Event group ID for FIFO ordering on the downstream event bus.

### `http_parameters` Block

The `http_parameters` block supports the following:

* `header_parameters` - (Optional) Map of HTTP header names to values.
* `invocation_timeout_seconds` - (Optional) Timeout, in seconds, of each invocation. Valid value is between `1` and `30`.
* `path_parameter_values` - (Optional) Values substituted, in order, for the path parameter placeholders in the endpoint path.
* `query_string_parameters` - (Optional) Map of query string parameter names to values.

### `kinesis_parameters` Block

The `kinesis_parameters` block supports the following:

* `explicit_hash_key` - (Optional) Explicit hash key.
* `partition_key` - (Optional) Partition key.

### `lambda_parameters` Block

The `lambda_parameters` block supports the following:

* `durable_execution_name` - (Optional) Durable execution name.
* `invocation_timeout_seconds` - (Optional) Timeout, in seconds, of each invocation.
* `invocation_type` - (Optional) Invocation type. `EVENT` invokes the function asynchronously and `REQUEST_RESPONSE` waits for its result. Allowed values are: `EVENT` and `REQUEST_RESPONSE`.
* `qualifier` - (Optional) Function qualifier: `$LATEST`, `$LATEST.PUBLISHED`, a version number, or an alias.
* `tenant_id` - (Optional) Tenant identifier.

### `sns_parameters` Block

The `sns_parameters` block supports the following:

* `message_attributes` - (Optional) Custom message attributes. See [`message_attributes` Block](#message_attributes-block) below.
* `message_deduplication_id` - (Optional) Message deduplication ID for FIFO topics.
* `message_group_id` - (Optional) Message group ID for FIFO topics.
* `message_structure` - (Optional) Per-protocol message formatting mode.
* `subject` - (Optional) Subject line for email subscriptions.

### `sqs_parameters` Block

The `sqs_parameters` block supports the following:

* `delay_seconds` - (Optional) Delay, in seconds, before the message becomes visible. Standard queues only.
* `message_attributes` - (Optional) Custom message attributes. See [`message_attributes` Block](#message_attributes-block) below.
* `message_deduplication_id` - (Optional) Message deduplication ID for FIFO queues.
* `message_group_id` - (Optional) Message group ID for FIFO queues.
* `message_system_attributes` - (Optional) System message attributes, for example `AWSTraceHeader`. See [`message_system_attributes` Block](#message_system_attributes-block) below.

### `message_attributes` Block

The `message_attributes` block supports the following:

* `binary_value` - (Optional) Base64-encoded binary value.
* `data_type` - (Optional) Data type of the attribute, for example `String`.
* `name` - (Required) Name of the attribute.
* `string_value` - (Optional) String value.

### `message_system_attributes` Block

The `message_system_attributes` block supports the following:

* `binary_value` - (Optional) Base64-encoded binary value.
* `data_type` - (Optional) Data type of the attribute, for example `String`.
* `name` - (Required) Name of the attribute.
* `string_value` - (Optional) String value.

### `step_functions_parameters` Block

The `step_functions_parameters` block supports the following:

* `invocation_timeout_seconds` - (Optional) Timeout, in seconds, of each invocation. Valid value is between `1` and `30`.
* `invocation_type` - (Optional) `EVENT` starts an asynchronous execution and `REQUEST_RESPONSE` starts a synchronous execution. Allowed values are: `EVENT` and `REQUEST_RESPONSE`.
* `name` - (Optional) Name of the execution. Must be unique per account, region, and state machine.
* `trace_header` - (Optional) AWS X-Ray trace header.

### `universal_target_parameters` Block

The `universal_target_parameters` block supports the following:

* `input` - (Required) JSON string or JSONata expression that produces the API request.
* `invocation_timeout_seconds` - (Optional) Timeout, in seconds, of each invocation. Valid value is between `1` and `30`. Default value: `30`.

### `log_configuration` Block

The `log_configuration` block supports the following:

* `include_payload` - (Optional) Whether the event payload is included in log records. Allowed values are: `FULL` and `ON_ERROR_ONLY`. Default value: `ON_ERROR_ONLY`.
* `level` - (Optional) Minimum log level. Allowed values are: `OFF`, `ERROR`, and `INFO`. Default value: `OFF`.

### `on_failure_configuration` Block

The `on_failure_configuration` block supports the following:

* `arn` - (Required) ARN of the Amazon SQS queue that receives events that could not be delivered.

### `point_in_time_configuration` Block

The `point_in_time_configuration` block supports the following:

* `end_point` - (Optional) Time, in RFC3339 format, at which delivery stops.
* `point_type` - (Required) Whether delivery starts from the oldest stored event or a specific time. Allowed values are: `HORIZON` and `TIMESTAMP`.
* `starting_point` - (Optional) Time, in RFC3339 format, from which delivery starts. Required when `point_type` is `TIMESTAMP`.

### `retry_policy` Block

The `retry_policy` block supports the following:

* `max_event_age_in_seconds` - (Optional) Maximum age, in seconds, of an event before it is discarded. Valid value is between `60` and `86400`. Default value: `300`.
* `max_retry_attempts` - (Optional) Maximum number of retry attempts. Valid value is between `0` and `185`. Default value: `5`.
* `retry_strategy` - (Optional) Strategy that determines which errors are retried. Allowed values are: `ALL`. Default value: `ALL`.

### `transformer` Block

The `transformer` block supports the following:

* `jsonata_configuration` - (Optional) JSONata transformation. Required when `type` is `JSONATA`. See [`jsonata_configuration` Block](#jsonata_configuration-block) below.
* `type` - (Optional) Transformation type. Allowed values are: `RAW`, `WITH_METADATA`, and `JSONATA`.

### `jsonata_configuration` Block

The `jsonata_configuration` block supports the following:

* `expression` - (Required) JSONata expression that transforms the event. Must be wrapped in `{% %}` delimiters.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the subscriber.
* `creation_time` - Time the subscriber was created, in RFC3339 format.
* `last_modified_time` - Time the subscriber was last modified, in RFC3339 format.
* `revoked` - Whether the event bus owner has revoked the subscriber.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_eventbridgev2_subscriber.example
  identity = {
    "arn" = "arn:aws:events:us-east-1:123456789012:subscriberv2/example/example"
  }
}

resource "aws_eventbridgev2_subscriber" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) ARN of the subscriber.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EventBridge V2 subscribers using the `arn`. For example:

```terraform
import {
  to = aws_eventbridgev2_subscriber.example
  id = "arn:aws:events:us-east-1:123456789012:subscriberv2/example/example"
}
```

Using `terraform import`, import EventBridge V2 subscribers using the `arn`. For example:

```console
% terraform import aws_eventbridgev2_subscriber.example arn:aws:events:us-east-1:123456789012:subscriberv2/example/example
```
