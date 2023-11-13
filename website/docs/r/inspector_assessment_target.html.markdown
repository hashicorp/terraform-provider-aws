---
subcategory: "Inspector Classic"
layout: "aws"
page_title: "AWS: aws_inspector_assessment_target"
description: |-
  Provides an Inspector Classic Assessment Target.
---

# Resource: aws_inspector_assessment_target

Provides an Inspector Classic Assessment Target

## Example Usage

```terraform
resource "aws_inspector_resource_group" "bar" {
  tags = {
    Name = "foo"
    Env  = "bar"
  }
}

resource "aws_inspector_assessment_target" "foo" {
  name               = "assessment target"
  resource_group_arn = aws_inspector_resource_group.bar.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the assessment target.
* `resource_group_arn` (Optional) Inspector Resource Group Amazon Resource Name (ARN) stating tags for instance matching. If not specified, all EC2 instances in the current AWS account and region are included in the assessment target.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The target assessment ARN.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Inspector Classic Assessment Targets using their Amazon Resource Name (ARN). For example:

```terraform
import {
  to = aws_inspector_assessment_target.example
  id = "arn:aws:inspector:us-east-1:123456789012:target/0-xxxxxxx"
}
```

Using `terraform import`, import Inspector Classic Assessment Targets using their Amazon Resource Name (ARN). For example:

```console
% terraform import aws_inspector_assessment_target.example arn:aws:inspector:us-east-1:123456789012:target/0-xxxxxxx
```
