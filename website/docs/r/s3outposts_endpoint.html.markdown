---
subcategory: "S3 on Outposts"
layout: "aws"
page_title: "AWS: aws_s3outposts_endpoint"
description: |-
  Manages an S3 Outposts Endpoint.
---

# Resource: aws_s3outposts_endpoint

Provides a resource to manage an S3 Outposts Endpoint.

## Example Usage

```terraform
resource "aws_s3outposts_endpoint" "example" {
  outpost_id        = data.aws_outposts_outpost.example.id
  security_group_id = aws_security_group.example.id
  subnet_id         = aws_subnet.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `outpost_id` - (Required) Identifier of the Outpost to contain this endpoint.
* `security_group_id` - (Required) Identifier of the EC2 Security Group.
* `subnet_id` - (Required) Identifier of the EC2 Subnet.
* `access_type` - (Optional) Type of access for the network connectivity. Valid values are `Private` or `CustomerOwnedIp`.
* `customer_owned_ipv4_pool` - (Optional) The ID of a Customer Owned IP Pool. For more on customer owned IP addresses see the [User Guide](https://docs.aws.amazon.com/outposts/latest/userguide/local-rack.html#local-gateway-subnet).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the endpoint.
* `cidr_block` - VPC CIDR block of the endpoint.
* `creation_time` - UTC creation time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `id` - Amazon Resource Name (ARN) of the endpoint.
* `network_interfaces` - Set of nested attributes for associated Elastic Network Interfaces (ENIs).
    * `network_interface_id` - Identifier of the Elastic Network Interface (ENI).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Outposts Endpoints using Amazon Resource Name (ARN), EC2 Security Group identifier, and EC2 Subnet identifier, separated by commas (`,`). For example:

```terraform
import {
  to = aws_s3outposts_endpoint.example
  id = "arn:aws:s3-outposts:us-east-1:123456789012:outpost/op-12345678/endpoint/0123456789abcdef,sg-12345678,subnet-12345678"
}
```

Using `terraform import`, import S3 Outposts Endpoints using Amazon Resource Name (ARN), EC2 Security Group identifier, and EC2 Subnet identifier, separated by commas (`,`). For example:

```console
% terraform import aws_s3outposts_endpoint.example arn:aws:s3-outposts:us-east-1:123456789012:outpost/op-12345678/endpoint/0123456789abcdef,sg-12345678,subnet-12345678
```
