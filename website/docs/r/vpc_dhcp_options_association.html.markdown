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

* `dhcp_options_id` - (Required) ID of the DHCP Options Set to associate to the VPC.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `vpc_id` - (Required) ID of the VPC to which we would like to associate a DHCP Options Set.

~> **Note:** Only one DHCP Options Set can be associated to a given VPC. Removing the association automatically sets AWS's `default` DHCP Options Set to the VPC.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the DHCP Options Set Association.

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
