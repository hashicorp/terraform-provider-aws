---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_internet_gateway"
description: |-
  Provides a resource to create a VPC Internet Gateway.
---

# Resource: aws_internet_gateway

Provides a resource to create a VPC Internet Gateway.

## Example Usage

```terraform
resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "main"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `vpc_id` - (Optional) The VPC ID to create in.  See the [aws_internet_gateway_attachment](internet_gateway_attachment.html) resource for an alternate way to attach an Internet Gateway to a VPC.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

-> **Note:** It's recommended to denote that the AWS Instance or Elastic IP depends on the Internet Gateway. For example:

```terraform
resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.main.id
}

resource "aws_instance" "foo" {
  # ... other arguments ...

  depends_on = [aws_internet_gateway.gw]
}
```

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the Internet Gateway.
* `arn` - The ARN of the Internet Gateway.
* `owner_id` - The ID of the AWS account that owns the internet gateway.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `20m`)
- `update` - (Default `20m`)
- `delete` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Internet Gateways using the `id`. For example:

```terraform
import {
  to = aws_internet_gateway.gw
  id = "igw-c0a643a9"
}
```

Using `terraform import`, import Internet Gateways using the `id`. For example:

```console
% terraform import aws_internet_gateway.gw igw-c0a643a9
```
