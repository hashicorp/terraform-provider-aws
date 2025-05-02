---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_dedicated_ip_assignment"
description: |-
  Terraform resource for managing an AWS SESv2 (Simple Email V2) Dedicated IP Assignment.
---

# Resource: aws_sesv2_dedicated_ip_assignment

Terraform resource for managing an AWS SESv2 (Simple Email V2) Dedicated IP Assignment.

This resource is used with "Standard" dedicated IP addresses. This includes addresses [requested and relinquished manually](https://docs.aws.amazon.com/ses/latest/dg/dedicated-ip-case.html) via an AWS support case, or [Bring Your Own IP](https://docs.aws.amazon.com/ses/latest/dg/dedicated-ip-byo.html) addresses. Once no longer assigned, this resource returns the IP to the [`ses-default-dedicated-pool`](https://docs.aws.amazon.com/ses/latest/dg/managing-ip-pools.html), managed by AWS.

## Example Usage

### Basic Usage

```terraform
resource "aws_sesv2_dedicated_ip_assignment" "example" {
  ip                    = "0.0.0.0"
  destination_pool_name = "my-pool"
}
```

## Argument Reference

The following arguments are required:

* `ip` - (Required) Dedicated IP address.
* `destination_pool_name` - (Required) Dedicated IP address.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A comma-separated string made up of `ip` and `destination_pool_name`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SESv2 (Simple Email V2) Dedicated IP Assignment using the `id`, which is a comma-separated string made up of `ip` and `destination_pool_name`. For example:

```terraform
import {
  to = aws_sesv2_dedicated_ip_assignment.example
  id = "0.0.0.0,my-pool"
}
```

Using `terraform import`, import SESv2 (Simple Email V2) Dedicated IP Assignment using the `id`, which is a comma-separated string made up of `ip` and `destination_pool_name`. For example:

```console
% terraform import aws_sesv2_dedicated_ip_assignment.example "0.0.0.0,my-pool"
```
