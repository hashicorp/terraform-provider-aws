---
subcategory: "CloudWatch Observability Admin"
layout: "aws"
page_title: "AWS: aws_observabilityadmin_telemetry_rule"
description: |-
  Manages an AWS CloudWatch Observability Admin Telemetry Rule.
---

# Resource: aws_observabilityadmin_telemetry_rule

Manages an AWS CloudWatch Observability Admin Telemetry Rule.

~> **NOTE:** Before using this resource, telemetry evaluation must be enabled for your AWS account. You can use the [`aws_observabilityadmin_telemetry_evaluation`](observabilityadmin_telemetry_evaluation.html) or [`aws_observabilityadmin_telemetry_evaluation_for_organization`](observabilityadmin_telemetry_evaluation_for_organization.html) resource to enable it.

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

### With Tags

```terraform
resource "aws_observabilityadmin_telemetry_evaluation" "example" {}

resource "aws_observabilityadmin_telemetry_rule" "example" {
  rule_name = "vpc-logs-rule"

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

This resource supports the following arguments:

* `rule_name` - (Required) Name of the telemetry rule. Must be between 1 and 100 characters and contain only alphanumeric characters, hyphens, underscores, periods, hash symbols, and forward slashes.
* `rule` - (Required) Configuration block for the telemetry rule. See [rule](#rule) below.
* `region` - (Optional) AWS region. If not specified, the provider region is used.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### rule

* `telemetry_type` - (Required) Type of telemetry data. Valid values: `Logs`, `Metrics`, `Traces`.
* `resource_type` - (Required) AWS resource type to apply the rule to. Currently supported: `AWS::EC2::VPC` with `Logs`.

~> **Note:** This resource is currently in early development. Additional resource types and configuration options will be added in future releases.

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
    "rule_name" = "example-telemetry-rule"
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
