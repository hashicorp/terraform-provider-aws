---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_image_deregistration_protection"
description: |-
  Terraform resource for managing an AWS EC2 (Elastic Compute Cloud) Image Deregistration Protection.
---

# Resource: aws_ec2_image_deregistration_protection

Terraform resource for managing an AWS EC2 (Elastic Compute Cloud) Image Deregistration Protection. This prevents AMIs from being deleted.

## Example Usage

```terraform
resource "aws_ec2_image_deregistration_protection" "example" {
  ami_id        = "ami-odb953344c964484b"
  with_cooldown = true
}
```

## Argument Reference

The following arguments are required:

* `ami_id` - (Required) AMI ID to which image deregistration protection has to be enabled or disabled.

The following arguments are optional:

* `with_cooldown` - (Optional) if true, enabled image deregistration protection. The default value is `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AMI ID to which image deregistration protection has to be enabled or disabled.
* `deregistration_protection` - Image deregistration protection enabled or disabled, e.g., `enabled-without-cooldown` or `enabled-with-cooldown`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `update` - (Default `10m`)
- `delete` - (Default `15m`)

## Import

You cannot import this resource.
