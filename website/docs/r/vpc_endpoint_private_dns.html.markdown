---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_private_dns"
description: |-
  Terraform resource for enabling private DNS on an AWS VPC (Virtual Private Cloud) Endpoint.
---
# Resource: aws_vpc_endpoint_private_dns

Terraform resource for enabling private DNS on an AWS VPC (Virtual Private Cloud) Endpoint.

~> When using this resource, the `private_dns_enabled` argument should be omitted on the parent `aws_vpc_endpoint` resource.
Setting the value both places can lead to unintended behavior and persistent differences.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpc_endpoint_private_dns" "example" {
  vpc_endpoint_id     = aws_vpc_endpoint.example.id
  private_dns_enabled = true
}
```

## Argument Reference

The following arguments are required:

* `private_dns_enabled` - (Required) Indicates whether a private hosted zone is associated with the VPC. Only applicable for `Interface` endpoints.
* `vpc_endpoint_id` - (Required) VPC endpoint identifier.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a VPC (Virtual Private Cloud) Endpoint Private DNS using the `vpc_endpoint_id`. For example:

```terraform
import {
  to = aws_vpc_endpoint_private_dns.example
  id = "vpce-abcd-1234"
}
```

Using `terraform import`, import a VPC (Virtual Private Cloud) Endpoint Private DNS using the `vpc_endpoint_id`. For example:

```console
% terraform import aws_vpc_endpoint_private_dns.example vpce-abcd-1234
```
