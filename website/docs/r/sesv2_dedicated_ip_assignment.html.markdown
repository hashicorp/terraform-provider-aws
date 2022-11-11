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

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A comma-separated string made up of `ip` and `destination_pool_name`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

SESv2 (Simple Email V2) Dedicated IP Assignment can be imported using the `id`, which is a comma-separated string made up of `ip` and `destination_pool_name`, e.g.,

```
$ terraform import aws_sesv2_dedicated_ip_assignment.example "0.0.0.0,my-pool"
```
