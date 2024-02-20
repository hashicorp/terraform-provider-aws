---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_dhcp_options_association"
description: |-
  Provides a VPC DHCP Options Association resource.
---

# Resource: aws_vpc_dhcp_options_association

Provides a VPC DHCP Options Association resource.

## Example Usage

```terraform
resource "aws_vpc_dhcp_options_association" "dns_resolver" {
  vpc_id          = aws_vpc.foo.id
  dhcp_options_id = aws_vpc_dhcp_options.foo.id
}
```

## Argument Reference

This resource supports the following arguments:

* `vpc_id` - (Required) The ID of the VPC to which we would like to associate a DHCP Options Set.
* `dhcp_options_id` - (Required) The ID of the DHCP Options Set to associate to the VPC.

## Remarks

* You can only associate one DHCP Options Set to a given VPC ID.
* Removing the DHCP Options Association automatically sets AWS's `default` DHCP Options Set to the VPC.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the DHCP Options Set Association.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DHCP associations using the VPC ID associated with the options. For example:

```terraform
import {
  to = aws_vpc_dhcp_options_association.imported
  id = "vpc-0f001273ec18911b1"
}
```

Using `terraform import`, import DHCP associations using the VPC ID associated with the options. For example:

```console
% terraform import aws_vpc_dhcp_options_association.imported vpc-0f001273ec18911b1
```
