---
subcategory: "CloudWatch Observability Admin"
layout: "aws"
page_title: "AWS: aws_observabilityadmin_telemetry_rule"
description: |-
  Manages an AWS CloudWatch Observability Admin Telemetry Rule.
---

# Resource: aws_observabilityadmin_telemetry_rule

Manages an AWS CloudWatch Observability Admin Telemetry Rule.

A telemetry rule defines how telemetry data (logs, metrics, or traces) should be collected for AWS resources within an AWS account. The rule can target one or more Regions and optionally configure a destination (such as CloudWatch Logs or S3) along with source-specific parameters for VPC flow logs, WAF logs, CloudTrail events, ELB access logs, and more.

~> **NOTE:** Before using this resource, telemetry evaluation must be enabled for your AWS account. Use the [`aws_observabilityadmin_telemetry_evaluation`](observabilityadmin_telemetry_evaluation.html.markdown) resource to enable it.

## Example Usage

### Basic Usage

```terraform
resource "aws_observabilityadmin_telemetry_evaluation" "example" {}

resource "aws_observabilityadmin_telemetry_rule" "example" {
  rule_name = "example-telemetry-rule"

  rule {
    telemetry_type = "Logs"
    resource_type  = "AWS::EC2::VPC"
  }

  depends_on = [aws_observabilityadmin_telemetry_evaluation.example]
}
```

### VPC Flow Logs to CloudWatch Logs

```terraform
resource "aws_observabilityadmin_telemetry_evaluation" "example" {}

resource "aws_observabilityadmin_telemetry_rule" "example" {
  rule_name = "vpc-flow-logs-rule"

  rule {
    telemetry_type         = "Logs"
    resource_type          = "AWS::EC2::VPC"
    telemetry_source_types = ["VPC_FLOW_LOGS"]
    all_regions            = true
    allow_field_updates    = true

    destination_configuration {
      destination_type    = "cloud-watch-logs"
      destination_pattern = "/aws/vpcflowlogs/<resourceId>"
      retention_in_days   = 30

      vpc_flow_log_parameters {
        traffic_type             = "ALL"
        max_aggregation_interval = 60
      }
    }
  }

  depends_on = [aws_observabilityadmin_telemetry_evaluation.example]
}
```

### Replicated Across Specific Regions

```terraform
resource "aws_observabilityadmin_telemetry_evaluation" "example" {}

resource "aws_observabilityadmin_telemetry_rule" "example" {
  rule_name = "multi-region-rule"

  rule {
    telemetry_type = "Logs"
    resource_type  = "AWS::EKS::Cluster"
    regions        = ["us-east-1", "us-west-2", "eu-west-1"]
  }

  depends_on = [aws_observabilityadmin_telemetry_evaluation.example]
}
```

### WAF Logging with Filters and Redacted Fields

```terraform
resource "aws_observabilityadmin_telemetry_evaluation" "example" {}

resource "aws_observabilityadmin_telemetry_rule" "example" {
  rule_name = "waf-logs-rule"

  rule {
    telemetry_type = "Logs"
    resource_type  = "AWS::WAFv2::WebACL"

    destination_configuration {
      destination_type    = "cloud-watch-logs"
      destination_pattern = "aws-waf-logs-<resourceId>"
      retention_in_days   = 30

      waf_logging_parameters {
        log_type = "WAF_LOGS"

        logging_filter {
          default_behavior = "KEEP"

          filters {
            behavior    = "DROP"
            requirement = "MEETS_ANY"

            conditions {
              action_condition {
                action = "ALLOW"
              }
            }
          }
        }

        redacted_fields {
          query_string = ""

          single_header {
            name = "authorization"
          }
        }
      }
    }
  }

  depends_on = [aws_observabilityadmin_telemetry_evaluation.example]
}
```

### With Tags

```terraform
resource "aws_observabilityadmin_telemetry_evaluation" "example" {}

resource "aws_observabilityadmin_telemetry_rule" "example" {
  rule_name = "tagged-rule"

  rule {
    telemetry_type = "Logs"
    resource_type  = "AWS::EC2::VPC"
  }

  tags = {
    Environment = "production"
    Purpose     = "monitoring"
  }

  depends_on = [aws_observabilityadmin_telemetry_evaluation.example]
}
```

## Argument Reference

The following arguments are required:

* `rule` - (Required) Configuration block for the telemetry rule. See [`rule`](#rule-block) below.
* `rule_name` - (Required) Name of the telemetry rule. Must be between 1 and 100 characters and contain only alphanumeric characters, hyphens, underscores, periods, hash symbols, and forward slashes. Changing this argument forces a new resource to be created.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `rule` Block

* `all_regions` - (Optional) Whether to replicate the rule to every Region in the partition where CloudWatch Observability Admin is available. Mutually exclusive with `regions`.
* `allow_field_updates` - (Optional) Whether CloudWatch Observability Admin should detect and remediate configuration drift in managed telemetry resources. Currently supported for `AWS::EC2::VPC` resources (VPC flow logs).
* `destination_configuration` - (Optional) Configuration block specifying where and how the telemetry data is delivered. See [`destination_configuration`](#destination_configuration-block) below.
* `regions` - (Optional) Set of Regions to replicate the rule to. Mutually exclusive with `all_regions`. Order is not preserved.
* `resource_type` - (Optional) AWS resource type to apply the rule to (for example `AWS::EC2::VPC`, `AWS::EKS::Cluster`, `AWS::WAFv2::WebACL`).
* `scope` - (Optional) Organizational scope to which the rule applies, specified using accounts or organizational units.
* `selection_criteria` - (Optional) Criteria for selecting which resources the rule applies to, such as resource tags.
* `telemetry_source_types` - (Optional) List of telemetry source types to configure for the resource (for example `VPC_FLOW_LOGS`, `EKS_AUDIT_LOGS`). Must correlate with the chosen `resource_type`. If not provided, the API may default this value based on `resource_type` (for example `VPC_FLOW_LOGS` for `AWS::EC2::VPC`).
* `telemetry_type` - (Required) Type of telemetry data to collect. Valid values: `Logs`, `Metrics`, `Traces`.

### `destination_configuration` Block

* `cloudtrail_parameters` - (Optional) CloudTrail-specific parameters when CloudTrail is the source. See [`cloudtrail_parameters`](#cloudtrail_parameters-block) below.
* `destination_pattern` - (Optional) Pattern used to generate the destination path or name. May contain alphanumeric characters, the macros `<accountId>` and `<resourceId>`, and the symbols `_`, `/`, `-`.
* `destination_type` - (Optional) Destination type for the telemetry data (for example `cloud-watch-logs`).
* `elb_load_balancer_logging_parameters` - (Optional) ELB load balancer logging parameters when the resource is an ELB. See [`elb_load_balancer_logging_parameters`](#elb_load_balancer_logging_parameters-block) below.
* `log_delivery_parameters` - (Optional) Amazon Bedrock AgentCore log delivery parameters. See [`log_delivery_parameters`](#log_delivery_parameters-block) below.
* `msk_monitoring_parameters` - (Optional) Amazon MSK cluster monitoring parameters. See [`msk_monitoring_parameters`](#msk_monitoring_parameters-block) below.
* `retention_in_days` - (Optional) Number of days to retain the telemetry data in the destination.
* `vpc_flow_log_parameters` - (Optional) VPC Flow Logs-specific parameters when the resource is `AWS::EC2::VPC`. See [`vpc_flow_log_parameters`](#vpc_flow_log_parameters-block) below.
* `waf_logging_parameters` - (Optional) WAF logging parameters when the resource is `AWS::WAFv2::WebACL`. See [`waf_logging_parameters`](#waf_logging_parameters-block) below.

### `cloudtrail_parameters` Block

* `advanced_event_selectors` - (Optional) List of advanced event selectors used to filter CloudTrail events. See [`advanced_event_selectors`](#advanced_event_selectors-block) below.

### `advanced_event_selectors` Block

* `field_selectors` - (Optional) List of field selectors that compose the selector statement. See [`field_selectors`](#field_selectors-block) below.
* `name` - (Optional) Descriptive name for the advanced event selector.

### `field_selectors` Block

* `ends_with` - (Optional) Match if the field value ends with one of the specified values.
* `equals` - (Optional) Match if the field value equals one of the specified values.
* `field` - (Required) Name of the field to use for selection.
* `not_ends_with` - (Optional) Match if the field value does not end with one of the specified values.
* `not_equals` - (Optional) Match if the field value does not equal any of the specified values.
* `not_starts_with` - (Optional) Match if the field value does not start with any of the specified values.
* `starts_with` - (Optional) Match if the field value starts with one of the specified values.

### `elb_load_balancer_logging_parameters` Block

* `field_delimiter` - (Optional) Delimiter character used to separate fields in ELB access log entries when using plain text format.
* `output_format` - (Optional) Format for ELB access log entries. Valid values: `plain-text`, `json`.

### `log_delivery_parameters` Block

* `log_types` - (Optional) List of log types that the source is sending.

### `msk_monitoring_parameters` Block

* `enhanced_monitoring` - (Optional) Level of enhanced monitoring for the MSK cluster. Valid values: `DEFAULT`, `PER_BROKER`, `PER_TOPIC_PER_BROKER`, `PER_TOPIC_PER_PARTITION`.

### `vpc_flow_log_parameters` Block

* `log_format` - (Optional) Format string for VPC Flow Log entries.
* `max_aggregation_interval` - (Optional) Maximum interval (in seconds) between the capture of flow log records. Valid values: `60`, `600`.
* `traffic_type` - (Optional) Type of traffic to log. Valid values: `ACCEPT`, `REJECT`, `ALL`.

### `waf_logging_parameters` Block

* `log_type` - (Optional) Type of WAF logs to collect (currently `WAF_LOGS`).
* `logging_filter` - (Optional) Filter configuration that determines which WAF log records to include or exclude. See [`logging_filter`](#logging_filter-block) below.
* `redacted_fields` - (Optional) List of fields to redact from WAF logs. See [`redacted_fields`](#redacted_fields-block) below.

### `logging_filter` Block

* `default_behavior` - (Optional) Default action for log records that do not match any filter. Valid values: `KEEP`, `DROP`.
* `filters` - (Optional) List of filter configurations. See [`filters`](#filters-block) below.

### `filters` Block

* `behavior` - (Optional) Action to take for matching log records. Valid values: `KEEP`, `DROP`.
* `conditions` - (Optional) Conditions that determine if a log record matches this filter. See [`conditions`](#conditions-block) below.
* `requirement` - (Optional) Whether the log record must meet all conditions or any condition. Valid values: `MEETS_ALL`, `MEETS_ANY`.

### `conditions` Block

Exactly one of `action_condition` or `label_name_condition` must be set per `conditions` block.

* `action_condition` - (Optional) Condition that matches based on the WAF action. See [`action_condition`](#action_condition-block) below.
* `label_name_condition` - (Optional) Condition that matches based on WAF rule labels. See [`label_name_condition`](#label_name_condition-block) below.

### `action_condition` Block

* `action` - (Required) WAF action to match against. Valid values: `ALLOW`, `BLOCK`, `COUNT`, `CAPTCHA`, `CHALLENGE`, `EXCLUDED_AS_COUNT`.

### `label_name_condition` Block

* `label_name` - (Optional) Label name to match (alphanumeric, underscores, hyphens, and colons; up to 1024 characters).

### `redacted_fields` Block

* `method` - (Optional) Redact the HTTP method from WAF logs. Set to an empty string to enable redaction.
* `query_string` - (Optional) Redact the entire query string from WAF logs. Set to an empty string to enable redaction.
* `single_header` - (Optional) Redact a specific header by name from WAF logs. See [`single_header`](#single_header-block) below.
* `uri_path` - (Optional) Redact the URI path from WAF logs. Set to an empty string to enable redaction.

### `single_header` Block

* `name` - (Required) Header name to redact (up to 64 characters).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `rule_arn` - ARN of the telemetry rule.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_observabilityadmin_telemetry_rule.example
  identity = {
    rule_name = "example-telemetry-rule"
  }
}

resource "aws_observabilityadmin_telemetry_rule" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `rule_name` (String) Name of the telemetry rule.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Observability Admin Telemetry Rules using `rule_name`. For example:

```terraform
import {
  to = aws_observabilityadmin_telemetry_rule.example
  id = "example-telemetry-rule"
}
```

Using `terraform import`, import CloudWatch Observability Admin Telemetry Rules using `rule_name`. For example:

```console
% terraform import aws_observabilityadmin_telemetry_rule.example example-telemetry-rule
```
