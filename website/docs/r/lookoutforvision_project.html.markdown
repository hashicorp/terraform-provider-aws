---
subcategory: "Lookout for Vision"
layout: "aws"
page_title: "AWS: aws_lookoutforvision_project"
description: |-
  Provides a Lookout for Vision project.
---

# Resource: aws_lookoutforvision_project

Provides a Lookout for Vision project.

## Example Usage

Basic usage:

```hcl
resource "aws_lookoutforvision_project" "demo" {
  name = "demo"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the project

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this project

## Import

Projects can be imported using the `name`, e.g.

```
$ terraform import aws_lookoutforvision_project.test_project my-project
```
