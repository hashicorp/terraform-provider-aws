---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_online_evaluation_config"
description: |-
  Manages an AWS Bedrock AgentCore Online Evaluation Configuration.
---

# Resource: aws_bedrockagentcore_online_evaluation_config

Manages an AWS Bedrock AgentCore Online Evaluation Configuration. Online evaluation configurations continuously monitor agent performance by sampling live traffic from CloudWatch logs and applying evaluators to assess agent quality in production.

-> **Note:** CloudWatch Transaction Serach must be [enabled](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Enable-TransactionSearch.html) before using this resource.

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
    cloudwatch_logs {
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

  clustering_config {
    frequencies = ["DAILY", "WEEKLY"]
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
    cloudwatch_logs {
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

* `data_source_config` - (Required) Data source configuration specifying where to read agent traces. See [`data_source_config` Block](#data_source_config-block) below.
* `enable_on_create` - (Required) Whether to enable the online evaluation configuration immediately upon creation.
* `evaluation_execution_role_arn` - (Required) ARN of the IAM role that grants permissions to read from CloudWatch logs, write evaluation results, and invoke Amazon Bedrock models for evaluation.
* `online_evaluation_config_name` - (Required, Forces new resource) Name of the online evaluation configuration. Must start with a letter and contain only alphanumeric characters and underscores, up to 48 characters.
* `rule` - (Required) Evaluation rule defining sampling configuration, filters, and session detection settings. See [`rule` Block](#rule-block) below.

Exactly one of the following arguments is required:

* `evaluator` - (Optional) List of evaluators to apply during online evaluation. Minimum 1, maximum 10. Exactly one of `evaluator` or `insight` must be specified. See [`evaluator` Block](#evaluator-block) below.
* `insight` - (Optional) List of insight analyses to run against sessions during evaluation. Maximum 10. Exactly one of `evaluator` or `insight` must be specified. When `insight` is specified, `clustering_config` is required. See [`insight` Block](#insight-block) below.

The following arguments are optional:

* `clustering_config` - (Optional) Configuration for periodic batch evaluation clustering, specifying how often clustering batch evaluations are triggered. Required when `insight` is specified. See [`clustering_config` Block](#clustering_config-block) below.
* `description` - (Optional) Description of the online evaluation configuration.
* `execution_status` - (Optional) Execution status to enable or disable the online evaluation. Valid values: `ENABLED`, `DISABLED`. Computed on create based on `enable_on_create`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `data_source_config` Block

The `data_source_config` block supports the following:

* `cloudwatch_logs` - (Optional) CloudWatch logs configuration for reading agent traces. See [`cloudwatch_logs` Block](#cloudwatch_logs-block) below.

### `cloudwatch_logs` Block

The `cloudwatch_logs` block supports the following:

* `log_group_names` - (Required) List of CloudWatch log group names to monitor for agent traces. Maximum 5.
* `service_names` - (Required) List of service names to filter traces within the specified log groups.

### `evaluator` Block

The `evaluator` block supports the following:

* `evaluator_id` - (Required) Unique identifier of the evaluator. Can reference builtin evaluators (e.g., `Builtin.Helpfulness`, `Builtin.GoalSuccessRate`) or custom evaluator IDs.

### `insight` Block

The `insight` block supports the following:

* `insight_id` - (Required) Unique identifier of the insight to run. Can reference builtin insights using the `Builtin.Insight.*` naming convention or custom insight IDs.

### `clustering_config` Block

The `clustering_config` block supports the following:

* `frequencies` - (Required) List of frequencies at which clustering batch evaluations are triggered. Maximum 3. Valid values: `DAILY`, `WEEKLY`, `MONTHLY`.

### `rule` Block

The `rule` block supports the following:

* `filter` - (Optional) List of filters determining which agent traces to evaluate. Maximum 5. See [`filter` Block](#filter-block) below.
* `sampling_config` - (Required) Sampling configuration determining what percentage of agent traces to evaluate. See [`sampling_config` Block](#sampling_config-block) below.
* `session_config` - (Optional) Session configuration defining timeout settings for detecting when agent sessions are complete. See [`session_config` Block](#session_config-block) below.

### `sampling_config` Block

The `sampling_config` block supports the following:

* `sampling_percentage` - (Required) Percentage of agent traces to sample for evaluation, from 0.01 to 100.

### `filter` Block

The `filter` block supports the following:

* `key` - (Required) Key or field name to filter on within the agent trace data.
* `operator` - (Required) Comparison operator. Valid values: `Equals`, `NotEquals`, `GreaterThan`, `LessThan`, `GreaterThanOrEqual`, `LessThanOrEqual`, `Contains`, `NotContains`.
* `value` - (Required) Value to compare against. See [`value` Block](#value-block) below.

### `value` Block

The `value` block supports the following (exactly one must be specified):

* `boolean_value` - (Optional) Boolean value for true/false filtering.
* `double_value` - (Optional) Numeric value for numerical filtering.
* `string_value` - (Optional) String value for text-based filtering.

### `session_config` Block

The `session_config` block supports the following:

* `session_timeout_minutes` - (Required) Minutes of inactivity after which a session is considered complete. Between 1 and 60.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `online_evaluation_config_arn` - ARN of the online evaluation configuration.
* `online_evaluation_config_id` - Unique identifier of the online evaluation configuration.
* `output_config` - Configuration specifying where evaluation results are written. See [`output_config` Block](#output_config-block) below.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `output_config` Block

* `cloudwatch_config` - CloudWatch configuration for evaluation results. See [`cloudwatch_config` Block](#cloudwatch_config-block) below.

### `cloudwatch_config` Block

* `log_group_name` - Name of the CloudWatch log group where evaluation results are written.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockagentcore_online_evaluation_config.example
  identity = {
    online_evaluation_config_id = "my_evaluation_config-aBcDeFgHiJ"
  }
}

resource "aws_bedrockagentcore_online_evaluation_config" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `online_evaluation_config_id` (String) ID of the online evaluation config.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Online Evaluation Configs using `online_evaluation_config_id`. For example:

```terraform
import {
  to = aws_bedrockagentcore_online_evaluation_config.example
  id = "my_evaluation_config-aBcDeFgHiJ"
}
```

Using `terraform import`, import Bedrock AgentCore Online Evaluation Configs using `online_evaluation_config_id`. For example:

```console
% terraform import aws_bedrockagentcore_online_evaluation_config.example my_evaluation_config-aBcDeFgHiJ
```
