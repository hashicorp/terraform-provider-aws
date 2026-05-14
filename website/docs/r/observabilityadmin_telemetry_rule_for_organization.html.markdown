---
subcategory: "CloudWatch Observability Admin"
layout: "aws"
page_title: "AWS: aws_observabilityadmin_telemetry_rule_for_organization"
description: |-
  Manages an AWS CloudWatch Observability Admin Telemetry Rule for Organization.
---

# Resource: aws_observabilityadmin_telemetry_rule_for_organization

Manages an AWS CloudWatch Observability Admin Telemetry Rule for Organization.

~> **NOTE:** Before using this resource, telemetry evaluation for organization must be enabled for your AWS organization. You can use the [`aws_observabilityadmin_telemetry_evaluation_for_organization`](observabilityadmin_telemetry_evaluation_for_organization.html) resource to enable it.

~> **NOTE:** This resource can only be used in the organization management account.

## Example Usage

### Basic Usage

```terraform
resource "aws_observabilityadmin_telemetry_evaluation_for_organization" "example" {}

resource "aws_observabilityadmin_telemetry_rule_for_organization" "example" {
  rule_name = "example-org-telemetry-rule"

  rule {
    telemetry_type = "Logs"
    resource_type  = "AWS::EC2::VPC"
  }

  depends_on = [aws_observabilityadmin_telemetry_evaluation_for_organization.example]
}
```

### With Tags

```terraform
resource "aws_observabilityadmin_telemetry_evaluation_for_organization" "example" {}

resource "aws_observabilityadmin_telemetry_rule_for_organization" "example" {
  rule_name = "vpc-logs-org-rule"

  rule {
    telemetry_type = "Logs"
    resource_type  = "AWS::EC2::VPC"
  }

  tags = {
    Environment = "production"
    Purpose     = "organization-monitoring"
  }

  depends_on = [aws_observabilityadmin_telemetry_evaluation_for_organization.example]
}
```

## Argument Reference

This resource supports the following arguments:

* `rule_name` - (Required) Name of the telemetry rule for organization. Must be between 1 and 100 characters and contain only alphanumeric characters, hyphens, underscores, periods, hash symbols, and forward slashes.
* `rule` - (Required) Configuration block for the telemetry rule. See [rule](#rule) below.
* `region` - (Optional) AWS region. If not specified, the provider region is used.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### rule

* `telemetry_type` - (Required) Type of telemetry data. Valid values: `Logs`, `Metrics`, `Traces`.
* `resource_type` - (Required) AWS resource type to apply the rule to. Currently supported: `AWS::EC2::VPC` with `Logs`.

~> **Note:** This resource is currently in early development. Additional resource types and configuration options will be added in future releases.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Name of the telemetry rule for organization.
* `rule_arn` - ARN of the telemetry rule for organization.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Observability Admin Telemetry Rules for Organization using the `rule_name`. For example:

```terraform
import {
  to = aws_observabilityadmin_telemetry_rule_for_organization.example
  id = "example-org-telemetry-rule"
}
```

Using `terraform import`, import CloudWatch Observability Admin Telemetry Rules for Organization using the `rule_name`. For example:

```console
% terraform import aws_observabilityadmin_telemetry_rule_for_organization.example example-org-telemetry-rule
```
