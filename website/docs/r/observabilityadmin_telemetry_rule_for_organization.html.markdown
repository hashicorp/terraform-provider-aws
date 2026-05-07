---
subcategory: "CloudWatch Observability Admin"
layout: "aws"
page_title: "AWS: aws_observabilityadmin_telemetry_rule_for_organization"
description: |-
  Manages an AWS CloudWatch Observability Admin Telemetry Rule For Organization.
---

# Resource: aws_observabilityadmin_telemetry_rule_for_organization

Manages an AWS CloudWatch Observability Admin Telemetry Rule For Organization.

A telemetry rule for organization defines how telemetry should be configured for specific AWS resources across all member accounts in the organization. Rules specify the resource type, telemetry type, and optionally the destination configuration.

For more information, see the [AWS CloudWatch Observability Admin documentation](https://docs.aws.amazon.com/cloudwatch/latest/observabilityadmin/what-is-observabilityadmin.html).

~> **NOTE:** This resource can only be used from the management account or a delegated admin account of an AWS Organization.

## Example Usage

### Basic Usage

```terraform
resource "aws_observabilityadmin_telemetry_rule_for_organization" "example" {
  rule_name = "my-org-rule"

  rule {
    telemetry_type = "Logs"
    resource_type  = "AWS::EC2::VPC"
  }
}
```

## Argument Reference

The following arguments are required:

* `rule_name` - (Required) Unique name for the organization telemetry rule. Must be between 1 and 100 characters, and can contain alphanumeric characters, hyphens, underscores, periods, hash signs, and forward slashes.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### rule

The following arguments are required:

* `telemetry_type` - (Required) Type of telemetry to collect. Valid values: `Logs`, `Metrics`, `Traces`.

The following arguments are optional:

* `resource_type` - (Optional) Type of AWS resource to configure telemetry for (e.g., `AWS::EC2::VPC`, `AWS::EKS::Cluster`, `AWS::WAFv2::WebACL`).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `rule_arn` - ARN of the organization telemetry rule.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `update` - (Default `5m`)
- `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Observability Admin Telemetry Rule For Organization using the `rule_name`. For example:

```terraform
import {
  to = aws_observabilityadmin_telemetry_rule_for_organization.example
  id = "my-org-rule"
}
```

Using `terraform import`, import CloudWatch Observability Admin Telemetry Rule For Organization using the `rule_name`. For example:

```console
% terraform import aws_observabilityadmin_telemetry_rule_for_organization.example my-org-rule
```
