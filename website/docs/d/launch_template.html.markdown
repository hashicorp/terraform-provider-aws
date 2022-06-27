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

The following arguments are supported:

* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.
* `id` - (Optional) The ID of the specific launch template to retrieve.
* `name` - (Optional) The name of the launch template.
* `tags` - (Optional) A map of tags, each pair of which must exactly match a pair on the desired Launch Template.

### filter Configuration Block

The following arguments are supported by the `filter` configuration block:

* `name` - (Required) The name of the filter field. Valid values can be found in the [EC2 DescribeLaunchTemplates API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeLaunchTemplates.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the launch template.

This resource also exports a full set of attributes corresponding to the arguments of the [`aws_launch_template`](/docs/providers/aws/r/launch_template.html) resource.
