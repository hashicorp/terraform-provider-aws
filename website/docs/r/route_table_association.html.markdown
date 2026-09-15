---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_route_table_association"
description: |-
  Provides a resource to create an association between a route table and a subnet or a route table and an internet gateway or virtual private gateway.
---

# Resource: aws_route_table_association

Provides a resource to create an association between a route table and a subnet or a route table and an
internet gateway or virtual private gateway.

## Example Usage

```terraform
resource "aws_route_table_association" "example" {
  subnet_id      = aws_subnet.example.id
  route_table_id = aws_route_table.example.id
}
```

```terraform
resource "aws_route_table_association" "example" {
  gateway_id     = aws_internet_gateway.example.id
  route_table_id = aws_route_table.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `subnet_id` - (Optional) The subnet ID to create an association. Conflicts with `gateway_id`.
* `gateway_id` - (Optional) The gateway ID to create an association. Conflicts with `subnet_id`.
* `route_table_id` - (Required) The ID of the routing table to associate with.

~> **NOTE:** Please note that one of either `subnet_id` or `gateway_id` is required.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the association

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `update` - (Default `2m`)
- `delete` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_route_table_association.example
  identity = {
    id = "rtbassoc-1234567890abcdef1"
  }
}

resource "aws_route_table_association" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` - (String) ID of the association.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 Route Table Associations using the associated resource ID and Route Table ID separated by a forward slash (`/`). For example:

~> **NOTE:** Attempting to associate a route table with a subnet or gateway, where either is already associated, will result in an error (e.g., `Resource.AlreadyAssociated: the specified association for route table rtb-4176657279 conflicts with an existing association`) unless you first import the original association.

With EC2 Subnets:

```terraform
import {
  to = aws_route_table_association.example
  id = "subnet-6777656e646f6c796e/rtb-656c65616e6f72"
}
```

With EC2 Internet Gateways:

```terraform
import {
  to = aws_route_table_association.example
  id = "igw-01b3a60780f8d034a/rtb-656c65616e6f72"
}
```

**Using `terraform import` to import** EC2 Route Table Associations using the associated resource ID and Route Table ID separated by a forward slash (`/`). For example:

With EC2 Subnets:

```console
% terraform import aws_route_table_association.example subnet-6777656e646f6c796e/rtb-656c65616e6f72
```

With EC2 Internet Gateways:

```console
% terraform import aws_route_table_association.example igw-01b3a60780f8d034a/rtb-656c65616e6f72
```
