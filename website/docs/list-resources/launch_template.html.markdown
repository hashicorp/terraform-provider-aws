---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_launch_template"
description: |-
  Lists EC2 Launch Template resources.
---

# List Resource: aws_launch_template

Lists EC2 Launch Template resources.

## Example Usage

### Basic Usage

```terraform
list "aws_launch_template" "example" {
  provider = aws
}
```

### Filter Usage

This example returns Launch Templates by ID.

```terraform
list "aws_launch_template" "example" {
  provider = aws

  config {
    launch_template_ids = ["lt-0123456789abcdef0", "lt-0123456789abcdef1"]
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `filter` - (Optional) One or more filters to apply to the search. If multiple `filter` blocks are provided, they all must be true. See [`filter` Block](#filter-block) below.
* `launch_template_ids` - (Optional) List of Launch Template IDs to query.
* `launch_template_names` - (Optional) List of Launch Template names to query.
* `region` - (Optional) Region to query. Defaults to provider region.

### `filter` Block

The `filter` block supports the following arguments:

* `name` - (Required) Name of the filter. For a full reference of filter names, see [describe-launch-templates in the AWS CLI reference](https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-launch-templates.html).
* `values` - (Required) One or more values to match.
