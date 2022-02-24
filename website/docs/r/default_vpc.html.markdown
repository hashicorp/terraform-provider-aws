---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_default_vpc"
description: |-
  Manage a default VPC resource.
---

# Resource: aws_default_vpc

Provides a resource to manage the [default AWS VPC](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/default-vpc.html)
in the current AWS Region.

If you created your AWS account after 2013-12-04 you have a default VPC in each AWS Region.

**This is an advanced resource** and has special caveats to be aware of when using it. Please read this document in its entirety before using this resource.

The `aws_default_vpc` resource behaves differently from normal resources in that if a default VPC exists, Terraform does not _create_ this resource, but instead "adopts" it into management.
If no default VPC exists, Terraform creates a new default VPC, which leads to the implicit creation of [other resources](https://docs.aws.amazon.com/vpc/latest/userguide/default-vpc.html#default-vpc-components).
By default, `terraform destroy` does not delete the default VPC but does remove the resource from Terraform state.
Set the `force_destroy` argument to `true` to delete the default VPC.

## Example Usage

Basic usage with tags:

```terraform
resource "aws_default_vpc" "default" {
  tags = {
    Name = "Default VPC"
  }
}
```

## Argument Reference

The arguments of an `aws_default_vpc` differ slightly from those of [`aws_vpc`](vpc.html):

* The `cidr_block` and `instance_tenancy` arguments become computed attributes
* The default value for `enable_dns_hostnames` is `true`

The following additional arguments are supported:

* `force_destroy` - (Optional) Whether destroying the resource deletes the default VPC. Default: `false`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `cidr_block` - The primary IPv4 CIDR block for the VPC
* `instance_tenancy` - The allowed tenancy of instances launched into the VPC

## Import

Default VPCs can be imported using the `vpc id`, e.g.,

```
$ terraform import aws_default_vpc.default vpc-a01106c2
```
