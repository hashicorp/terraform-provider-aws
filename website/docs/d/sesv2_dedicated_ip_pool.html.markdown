---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_dedicated_ip_pool"
description: |-
  Terraform data source for managing an AWS SESv2 (Simple Email V2) Dedicated IP Pool.
---

# Data Source: aws_sesv2_dedicated_ip_pool

Terraform data source for managing an AWS SESv2 (Simple Email V2) Dedicated IP Pool.

## Example Usage

### Basic Usage

```terraform
data "aws_sesv2_dedicated_ip_pool" "example" {
  pool_name = "my-pool"
}
```

## Argument Reference

The following arguments are required:

* `pool_name` - (Required) Name of the dedicated IP pool.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Dedicated IP Pool.
* `dedicated_ips` - A list of objects describing the pool's dedicated IP's. See [`dedicated_ips`](#dedicated_ips).
* `scaling_mode` - (Optional) IP pool scaling mode. Valid values: `STANDARD`, `MANAGED`.
* `tags` - A map of tags attached to the pool.

### dedicated_ips

* `ip` - IPv4 address.
* `warmup_percentage` - Indicates how complete the dedicated IP warm-up process is. When this value equals `1`, the address has completed the warm-up process and is ready for use.
* `warmup_status` - The warm-up status of a dedicated IP address. Valid values: `IN_PROGRESS`, `DONE`.
