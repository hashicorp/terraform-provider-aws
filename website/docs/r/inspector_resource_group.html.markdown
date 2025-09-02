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

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Required) Key-value map of tags that are used to select the EC2 instances to be included in an [Amazon Inspector assessment target](/docs/providers/aws/r/inspector_assessment_target.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The resource group ARN.
