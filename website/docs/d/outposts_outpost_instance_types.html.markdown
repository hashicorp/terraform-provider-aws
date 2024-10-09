---
subcategory: "Outposts"
layout: "aws"
page_title: "AWS: aws_outposts_outpost_instance_types"
description: |-
  Information about Outpost Instance Types.
---

# Data Source: aws_outposts_outpost_instance_types

Information about Outposts Instance Types.

## Example Usage

```terraform
data "aws_outposts_outpost_instance_types" "example" {
  arn = data.aws_outposts_outpost.example.arn
}
```

## Argument Reference

The following arguments are required:

* `arn` - (Required) Outpost ARN.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `instance_types` - Set of instance types.
