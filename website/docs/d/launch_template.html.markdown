---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_launch_template"
description: |-
  Provides a Launch Template data source.
---

# Data Source: aws_launch_template

Provides information about a Launch Template.

## Example Usage

```terraform
data "aws_launch_template" "default" {
  name = "my-launch-template"
}
```

### Filter

```terraform
data "aws_launch_template" "test" {
  filter {
    name   = "launch-template-name"
    values = ["some-template"]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.
* `id` - (Optional) ID of the specific launch template to retrieve.
* `name` - (Optional) Name of the launch template.
* `tags` - (Optional) Map of tags, each pair of which must exactly match a pair on the desired Launch Template.

### filter Configuration Block

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter field. Valid values can be found in the [EC2 DescribeLaunchTemplates API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeLaunchTemplates.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the launch template.

This resource also exports a full set of attributes corresponding to the arguments of the [`aws_launch_template`](/docs/providers/aws/r/launch_template.html) resource.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
