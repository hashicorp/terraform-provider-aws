---
layout: "aws"
page_title: "AWS: aws_ram_principal_association"
sidebar_current: "docs-aws-resource-ram-principal-association"
description: |-
  Provides a Resource Access Manager (RAM) principal association.
---

# aws_ram_principal_association

Provides a Resource Access Manager (RAM) principal association.

## Example Usage

```hcl
resource "aws_ram_principal_association" "example" {
  resource_share_arn = "arn:aws:ram:eu-west-1:123456789012:resource-share/73da1ab9-b94a-4ba3-8eb4-45917f7f4b12"
  principal          = "111111111111"
}
```

## Argument Reference

The following arguments are supported:

* `resource_share_arn` - (Required) The Amazon Resource Names (ARN) of the resources to associate with the resource share.
* `principal` - (Required) The principal to associate with the resource share. Possible values are the ID of an AWS account, the ARN of an OU or organization from AWS Organizations.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) of the resource share.

## Import

Resource Shares can be imported using the `id`, e.g.

```
$ terraform import aws_ram_resource_share.example arn:aws:ram:eu-west-1:123456789012:resource-share/73da1ab9-b94a-4ba3-8eb4-45917f7f4b12
```
