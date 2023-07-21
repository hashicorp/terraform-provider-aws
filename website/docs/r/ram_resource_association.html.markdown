---
subcategory: "RAM (Resource Access Manager)"
layout: "aws"
page_title: "AWS: aws_ram_resource_association"
description: |-
  Manages a Resource Access Manager (RAM) Resource Association.
---

# Resource: aws_ram_resource_association

Manages a Resource Access Manager (RAM) Resource Association.

~> *NOTE:* Certain AWS resources (e.g., EC2 Subnets) can only be shared in an AWS account that is a member of an AWS Organizations organization with organization-wide Resource Access Manager functionality enabled. See the [Resource Access Manager User Guide](https://docs.aws.amazon.com/ram/latest/userguide/what-is.html) and AWS service specific documentation for additional information.

## Example Usage

```terraform
resource "aws_ram_resource_association" "example" {
  resource_arn       = aws_subnet.example.arn
  resource_share_arn = aws_ram_resource_share.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `resource_arn` - (Required) Amazon Resource Name (ARN) of the resource to associate with the RAM Resource Share.
* `resource_share_arn` - (Required) Amazon Resource Name (ARN) of the RAM Resource Share.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) of the resource share.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RAM Resource Associations using their Resource Share ARN and Resource ARN separated by a comma. For example:

```terraform
import {
  to = aws_ram_resource_association.example
  id = "arn:aws:ram:eu-west-1:123456789012:resource-share/73da1ab9-b94a-4ba3-8eb4-45917f7f4b12,arn:aws:ec2:eu-west-1:123456789012:subnet/subnet-12345678"
}
```

Using `terraform import`, import RAM Resource Associations using their Resource Share ARN and Resource ARN separated by a comma. For example:

```console
% terraform import aws_ram_resource_association.example arn:aws:ram:eu-west-1:123456789012:resource-share/73da1ab9-b94a-4ba3-8eb4-45917f7f4b12,arn:aws:ec2:eu-west-1:123456789012:subnet/subnet-12345678
```
