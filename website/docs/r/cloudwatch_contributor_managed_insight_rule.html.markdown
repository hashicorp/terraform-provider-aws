---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_contributor_managed_insight_rule"
description: |-
  Terraform resource for managing an AWS CloudWatch Contributor Managed Insight Rule.
---

# Resource: aws_cloudwatch_contributor_managed_insight_rule

Terraform resource for managing an AWS CloudWatch Contributor Managed Insight Rule.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudwatch_contributor_managed_insight_rule" "example" {
  resource_arn  = aws_vpc_endpoint_service.test.arn
  template_name = "VpcEndpointService-BytesByEndpointId-v1"
  rule_state    = "DISABLED"
}
```

## Argument Reference

The following arguments are required:

* `resource_arn` - (Required) ARN of an Amazon Web Services resource that has managed Contributor Insights rules.
* `template_name` - (Required) Template name for the managed Contributor Insights rule, as returned by ListManagedInsightRules.

The following arguments are optional:

* `rule_state` - (Optional) State of the rule. Valid values are `ENABLED` and `DISABLED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Contributor Managed Insight Rule.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Contributor Managed Insight Rule using the `resource_arn`. For example:

```terraform
import {
  to = aws_cloudwatch_contributor_managed_insight_rule.example
  id = "contributor_managed_insight_rule-id-12345678"
}
```

Using `terraform import`, import CloudWatch Contributor Managed Insight Rule using the `resource_arn`. For example:

```console
% terraform import aws_cloudwatch_contributor_managed_insight_rule.example contributor_managed_insight_rule-id-12345678
```
