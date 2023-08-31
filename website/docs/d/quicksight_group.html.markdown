---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_group"
description: |-
  Use this data source to fetch information about a QuickSight Group.
---

# Data Source: aws_quicksight_group

This data source can be used to fetch information about a specific
QuickSight group. By using this data source, you can reference QuickSight group
properties without having to hard code ARNs or unique IDs as input.

## Example Usage

### Basic Usage

```terraform
data "aws_quicksight_group" "example" {
  group_name = "example"
}
```

## Argument Reference

The following arguments are required:

* `group_name` - (Required) The name of the group that you want to match.

The following arguments are optional:

* `aws_account_id` - (Optional) AWS account ID.
* `namespace` - (Optional) QuickSight namespace. Defaults to `default`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) for the group.
* `description` - The group description.
* `principal_id` - The principal ID of the group.
