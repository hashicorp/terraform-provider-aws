---
subcategory: "Inspector Classic"
layout: "aws"
page_title: "AWS: aws_inspector_resource_group"
description: |-
  Provides an Amazon Inspector Classic Resource Group.
---

# Resource: aws_inspector_resource_group

Provides an Amazon Inspector Classic Resource Group.

## Example Usage

```terraform
resource "aws_inspector_resource_group" "example" {
  tags = {
    Name = "foo"
    Env  = "bar"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `tags` - (Required) Key-value map of tags that are used to select the EC2 instances to be included in an [Amazon Inspector assessment target](/docs/providers/aws/r/inspector_assessment_target.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The resource group ARN.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Inspector Classic Resource Groups using their Amazon Resource Name (ARN). For example:

```terraform
import {
  to = aws_inspector_resource_group.example
  id = "arn:aws:inspector:us-east-1:123456789012:resourcegroup/0-xxxxxxx"
}
```

Using `terraform import`, import Inspector Classic Resource Groups using their Amazon Resource Name (ARN). For example:

```console
% terraform import aws_inspector_resource_group.example arn:aws:inspector:us-east-1:123456789012:resourcegroup/0-xxxxxxx
```
