---
subcategory: "Shield"
layout: "aws"
page_title: "AWS: aws_shield_protection"
description: |-
  Enables AWS Shield Advanced for a specific AWS resource.
---

# Resource: aws_shield_protection

Enables AWS Shield Advanced for a specific AWS resource.
The resource can be an Amazon CloudFront distribution, Elastic Load Balancing load balancer, AWS Global Accelerator accelerator, Elastic IP Address, or an Amazon Route 53 hosted zone.

## Example Usage

### Create protection

```terraform
data "aws_availability_zones" "available" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_eip" "example" {
  domain = "vpc"
}

resource "aws_shield_protection" "example" {
  name         = "example"
  resource_arn = "arn:aws:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:eip-allocation/${aws_eip.example.id}"

  tags = {
    Environment = "Dev"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) A friendly name for the Protection you are creating.
* `resource_arn` - (Required) The ARN (Amazon Resource Name) of the resource to be protected.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The unique identifier (ID) for the Protection object that is created.
* `arn` - The ARN of the Protection.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Shield protection resources using specifying their ID. For example:

```terraform
import {
  to = aws_shield_protection.example
  id = "ff9592dc-22f3-4e88-afa1-7b29fde9669a"
}
```

Using `terraform import`, import Shield protection resources using specifying their ID. For example:

```console
% terraform import aws_shield_protection.example ff9592dc-22f3-4e88-afa1-7b29fde9669a
```
