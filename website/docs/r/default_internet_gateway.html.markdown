---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_default_internet_gateway"
description: |-
  Manage a default internet gateway resource.
---

# Resource: aws_default_internet_gateway

Provides a resource to manage the [internet gateway](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/default-vpc.html#default-vpc-basics) attached to the default VPC in the current region.

**This is an advanced resource** and has special caveats to be aware of when using it. Please read this document in its entirety before using this resource.

Please be aware that there is no such thing as "default Internet Gateway" on AWS. When creating a default VPC, it will automatically create an Internet Gateway and attach it to the default VPC. The new IGW has no properties that is specific for being part of the default VPC's infrastructure. The aws_default_internet_gateway resource assumes that the Internet Gateway attached to the default VPC is the "default Internet Gateway".

The `aws_default_internet_gateway` resource behaves differently from normal resources in that, if an Internet Gateway exists and attached to the default VPC, Terraform does not _create_ this resource, but instead "adopts" it into management.
If no Internet Gateway exists that is attached to the default VPC, Terraform creates a new internet gateway and attaches to the default VPC.
By default, `terraform destroy` does not delete the internet gateway, but does remove the resource from Terraform state.
Set the `force_destroy` argument to `true` to delete the "default" Internet Gateway".

## Example Usage

```terraform
resource "aws_default_internet_gateway" "default" {
  force_destroy = true
  depends_on    = [aws_default_vpc.default]
}
```

## Argument Reference

This resource supports the following additional arguments:

* `force_destroy` - (Optional) Whether destroying the resource deletes the Internet Gateway attached to the default VPC. Default: `false`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the Internet Gateway.
* `arn` - The ARN of the Internet Gateway.
* `owner_id` - The ID of the AWS account that owns the internet gateway.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Internet Gateways using the `id`. For example:

```terraform
import {
  to = aws_default_internet_gateway.gw
  id = "igw-c0a643a9"
}
```

Using `terraform import`, import default Internet Gateways using the `id`. For example:

```console
% terraform import aws_default_internet_gateway.gw igw-c0a643a9
```
