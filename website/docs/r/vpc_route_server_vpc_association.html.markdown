---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_route_server_vpc_association"
description: |-
  Terraform resource for managing a VPC (Virtual Private Cloud) Route Server Association.
---
# Resource: aws_vpc_route_server_vpc_association

  Provides a resource for managing association between VPC (Virtual Private Cloud) route server and a VPC.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpc_route_server_vpc_association" "example" {
  route_server_id = aws_vpc_route_server.example.route_server_id
  vpc_id          = aws_vpc.example.id
}
```

## Argument Reference

The following arguments are required:

* `route_server_id` - (Required) The unique identifier for the route server to be associated.
* `vpc_id` - (Required) The ID of the VPC to associate with the route server.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to  to import VPC (Virtual Private Cloud) Route Server Association using the associated resource ID and VPC Id separated by a comma (,). For example:

```terraform
import {
  to = aws_vpc_route_server_vpc_association.example
  id = "rs-12345678,vpc-0f001273ec18911b1"
}
```

Using `terraform import`, to  to import VPC (Virtual Private Cloud) Route Server Association using the associated resource ID and VPC Id separated by a comma (,). For example:

```console
% terraform import aws_vpc_route_server_vpc_association.example rs-12345678,vpc-0f001273ec18911b1
```
