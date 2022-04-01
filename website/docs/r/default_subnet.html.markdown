---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_default_subnet"
description: |-
  Manage a default subnet resource.
---

# Resource: aws_default_subnet

Provides a resource to manage a [default subnet](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/default-vpc.html#default-vpc-basics) in the current region.

**This is an advanced resource** and has special caveats to be aware of when using it. Please read this document in its entirety before using this resource.

The `aws_default_subnet` resource behaves differently from normal resources in that if a default subnet exists in the specified Availability Zone, Terraform does not _create_ this resource, but instead "adopts" it into management.
If no default subnet exists, Terraform creates a new default subnet.
By default, `terraform destroy` does not delete the default subnet but does remove the resource from Terraform state.
Set the `force_destroy` argument to `true` to delete the default subnet.

## Example Usage

```terraform
resource "aws_default_subnet" "default_az1" {
  availability_zone = "us-west-2a"

  tags = {
    Name = "Default subnet for us-west-2a"
  }
}
```

## Argument Reference

The arguments of an `aws_default_subnet` differ slightly from those of [`aws_subnet`](subnet.html):

* `availability_zone` is required
* The `availability_zone_id`, `cidr_block` and `vpc_id` arguments become computed attributes
* The default value for `map_public_ip_on_launch` is `true`

The following additional arguments are supported:

* `force_destroy` - (Optional) Whether destroying the resource deletes the default subnet. Default: `false`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `availability_zone_id` - The AZ ID of the subnet
* `cidr_block` - The IPv4 CIDR block assigned to the subnet
* `vpc_id` - The ID of the VPC the subnet is in

## Import

Subnets can be imported using the `subnet id`, e.g.,

```
$ terraform import aws_default_subnet.public_subnet subnet-9d4a7b6c
```
