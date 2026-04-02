---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_online_evaluation_config"
description: |-
  Manages an AWS Bedrock AgentCore Online Evaluation Configuration.
---

# Resource: aws_bedrockagentcore_online_evaluation_config

Manages an AWS Bedrock AgentCore Online Evaluation Configuration. Online evaluation configurations continuously monitor agent performance by sampling live traffic from CloudWatch logs and applying evaluators to assess agent quality in production.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_role" "example" {
  name = "agentcore-eval-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = "sts:AssumeRole"
      Principal = {
        Service = "bedrock-agentcore.amazonaws.com"
      }
    }]
  })
}

resource "aws_cloudwatch_log_group" "example" {
  name = "/aws/agentcore/my-agent-traces"
}

resource "aws_bedrockagentcore_online_evaluation_config" "example" {
  online_evaluation_config_name = "my_evaluation_config"
  description                   = "Continuous evaluation of agent performance"
  enable_on_create              = true
  evaluation_execution_role_arn = aws_iam_role.example.arn

  data_source_config {
    cloud_watch_logs {
      log_group_names = [aws_cloudwatch_log_group.example.name]
      service_names   = ["my_agent_service"]
    }
  }

  evaluator {
    evaluator_id = "Builtin.Helpfulness"
  }

  evaluator {
    evaluator_id = "Builtin.GoalSuccessRate"
  }

  rule {
    sampling_config {
      sampling_percentage = 10.0
    }
  }
}
```

### With Filters and Session Config

```terraform
resource "aws_bedrockagentcore_online_evaluation_config" "filtered" {
  online_evaluation_config_name = "filtered_evaluation"
  enable_on_create              = true
  evaluation_execution_role_arn = aws_iam_role.example.arn

  data_source_config {
    cloud_watch_logs {
      log_group_names = [aws_cloudwatch_log_group.example.name]
      service_names   = ["my_agent_service"]
    }
  }

  evaluator {
    evaluator_id = "Builtin.Helpfulness"
  }

  rule {
    sampling_config {
      sampling_percentage = 50.0
    }

    filter {
      key      = "environment"
      operator = "Equals"

      value {
        string_value = "production"
      }
    }

    session_config {
      session_timeout_minutes = 30
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `online_evaluation_config_name` - (Required, Forces new resource) Name of the online evaluation configuration. Must start with a letter and contain only alphanumeric characters and underscores, up to 48 characters.
* `enable_on_create` - (Required) Whether to enable the online evaluation configuration immediately upon creation.
* `evaluation_execution_role_arn` - (Required) ARN of the IAM role that grants permissions to read from CloudWatch logs, write evaluation results, and invoke Amazon Bedrock models for evaluation.
* `data_source_config` - (Required) Data source configuration specifying where to read agent traces. See [`data_source_config`](#data_source_config) below.
* `evaluator` - (Required) List of evaluators to apply during online evaluation. Minimum 1, maximum 10. See [`evaluator`](#evaluator) below.
* `rule` - (Required) Evaluation rule defining sampling configuration, filters, and session detection settings. See [`rule`](#rule) below.

The following arguments are optional:

* `description` - (Optional) Description of the online evaluation configuration.
* `execution_status` - (Optional) Execution status to enable or disable the online evaluation. Valid values: `ENABLED`, `DISABLED`. Computed on create based on `enable_on_create`.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `data_source_config`

The `data_source_config` block supports the following:

* `cloud_watch_logs` - (Optional) CloudWatch logs configuration for reading agent traces. See [`cloud_watch_logs`](#cloud_watch_logs) below.

### `cloud_watch_logs`

The `cloud_watch_logs` block supports the following:

* `log_group_names` - (Required) List of CloudWatch log group names to monitor for agent traces. Maximum 5.
* `service_names` - (Required) List of service names to filter traces within the specified log groups.

### `evaluator`

The `evaluator` block supports the following:

* `evaluator_id` - (Required) Unique identifier of the evaluator. Can reference builtin evaluators (e.g., `Builtin.Helpfulness`, `Builtin.GoalSuccessRate`) or custom evaluator IDs.

### `rule`

The `rule` block supports the following:

* `sampling_config` - (Required) Sampling configuration determining what percentage of agent traces to evaluate. See [`sampling_config`](#sampling_config) below.
* `filter` - (Optional) List of filters determining which agent traces to evaluate. Maximum 5. See [`filter`](#filter) below.
* `session_config` - (Optional) Session configuration defining timeout settings for detecting when agent sessions are complete. See [`session_config`](#session_config) below.

### `sampling_config`

The `sampling_config` block supports the following:

* `sampling_percentage` - (Required) Percentage of agent traces to sample for evaluation, from 0.01 to 100.

### `filter`

The `filter` block supports the following:

* `key` - (Required) Key or field name to filter on within the agent trace data.
* `operator` - (Required) Comparison operator. Valid values: `Equals`, `NotEquals`, `GreaterThan`, `LessThan`, `GreaterThanOrEqual`, `LessThanOrEqual`, `Contains`, `NotContains`.
* `value` - (Required) Value to compare against. See [`value`](#value) below.

### `value`

The `value` block supports the following (exactly one must be specified):

* `string_value` - (Optional) String value for text-based filtering.
* `boolean_value` - (Optional) Boolean value for true/false filtering.
* `double_value` - (Optional) Numeric value for numerical filtering.

### `session_config`

The `session_config` block supports the following:

* `session_timeout_minutes` - (Required) Minutes of inactivity after which a session is considered complete. Between 1 and 60.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `online_evaluation_config_arn` - ARN of the online evaluation configuration.
* `online_evaluation_config_id` - Unique identifier of the online evaluation configuration.
* `failure_reason` - Reason for failure if the configuration creation or execution failed.
* `output_config` - Configuration specifying where evaluation results are written. See [`output_config`](#output_config) below.
* `status` - Status of the online evaluation configuration. Values: `ACTIVE`, `CREATING`, `CREATE_FAILED`, `UPDATING`, `UPDATE_FAILED`, `DELETING`.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `output_config`

* `cloud_watch_config` - CloudWatch configuration for evaluation results.
    * `log_group_name` - Name of the CloudWatch log group where evaluation results are written.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Online Evaluation Config using the config ID. For example:

```terraform
import {
  to = aws_bedrockagentcore_online_evaluation_config.example
  id = "my_evaluation_config-aBcDeFgHiJ"
}
```

Using `terraform import`, import Bedrock AgentCore Online Evaluation Config using the config ID. For example:

```console
% terraform import aws_bedrockagentcore_online_evaluation_config.example my_evaluation_config-aBcDeFgHiJ
```
