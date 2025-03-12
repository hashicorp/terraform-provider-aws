---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_contributor_managed_insight_rules"
description: |-
  Terraform data source for managing an AWS CloudWatch Contributor Managed Insight Rules.
---

# Data Source: aws_cloudwatch_contributor_managed_insight_rules

Terraform data source for managing an AWS CloudWatch Contributor Managed Insight Rules.

## Example Usage

### Basic Usage

```terraform
data "aws_cloudwatch_contributor_managed_insight_rules" "example" {
  resource_arn = "arn:aws:ec2:us-west-2:123456789012:resource-name/resourceid"
}
```

## Argument Reference

The following arguments are required:

* `resource_arn` - (Required) ARN of an Amazon Web Services resource that has managed Contributor Insights rules.

The following arguments are optional:

There are no optional arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `managed_rules` - Managed rules that are available for the specified Amazon Web Services resource. See [`managed_rules reference`](#managed_rules-reference) below for details.

### `managed_rules` Reference

* `template_name` - Template name for the managed rule. Used to enable managed rules using `PutManagedInsightRules`.
* `resource_arn` - If a managed rule is enabled, this is the ARN for the related Amazon Web Services resource.
* `rule_state` - Describes the state of a managed rule. If the rule is enabled, it contains information about the Contributor Insights rule that contains information about the related Amazon Web Services resource. See [`rule_state reference`](#rule_state-reference) below for details.

### `rule_state` Reference

* `rule_name` - Name of the Contributor Insights rule that contains data for the specified Amazon Web Services resource.
* `state` - Indicates whether the rule is enabled or disabled.
