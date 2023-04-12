---
subcategory: "CloudWatch Synthetics"
layout: "aws"
page_title: "AWS: aws_synthetics_group"
description: |-
  Provides a Synthetics Group resource
---

# Resource: aws_synthetics_group

Provides a Synthetics Group resource.

## Example Usage

### Basic Usage

```terraform
resource "aws_synthetics_group" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Group.
* `id` - ARN of the Group.

## Import

CloudWatch Synthetics Group can be imported using the `name`, e.g.,

```
$ terraform import aws_synthetics_group.example example
```
