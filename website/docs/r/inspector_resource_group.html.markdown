---
subcategory: "Inspector"
layout: "aws"
page_title: "AWS: aws_inspector_resource_group"
description: |-
  Provides a Inspector resource group.
---

# Resource: aws_inspector_resource_group

Provides a Inspector resource group

## Example Usage

```hcl
resource "aws_inspector_resource_group" "bar" {
  tags = {
    Name = "foo"
    Env  = "bar"
  }
}
```

## Argument Reference

The following arguments are supported:

* `tags` - (Required) The tags on your EC2 Instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The resource group ARN.
