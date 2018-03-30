---
layout: "aws"
page_title: "AWS: aws_default_internet_gateway"
sidebar_current: "docs-aws-resource-default-internet-gateway"
description: |-
  Provides a resource to manage a default VPC Internet Gateway.
---

# aws_default_internet_gateway

Provides a resource to manage a [default VPC Internet Gateway](https://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/default-vpc.html#default-vpc-components)
in the current region.

The `aws_default_internet_gateway` behaves differently from normal resources, in that
Terraform does not _create_ this resource, but instead "adopts" it into management.

## Example Usage

Basic usage with tags:

```hcl
resource "aws_default_internet_gateway" "default" {
  tags {
    Name = "Default IGW"
  }
}
```

## Argument Reference

The arguments of an `aws_default_internet_gateway` differ from [`aws_internet_gateway`](internet_gateway.html) resources.
Namely, the `vpc_id` argument is computed.
The following arguments are still supported:

* `tags` - (Optional) A mapping of tags to assign to the resource.

### Removing `aws_default_internet_gateway` from your configuration

The `aws_default_internet_gateway` resource allows you to manage a region's default VPC Internet Gateway,
but Terraform cannot destroy it. Removing this resource from your configuration
will remove it from your statefile and management, but will not destroy the Internet Gateway.
You can resume managing the Internet Gateway via the AWS Console.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Internet Gateway.
* `vpc_id` - The VPC ID.
