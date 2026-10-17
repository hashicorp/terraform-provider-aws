---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_pool"
description: |-
  Provides details about an AWS End User Messaging SMS Pool.
---

# Data Source: aws_pinpointsmsvoicev2_pool

Provides details about an AWS End User Messaging SMS Pool.

## Example Usage

### Basic Usage

```terraform
data "aws_pinpointsmsvoicev2_pool" "example" {
  id = "pool-abcdef0123456789abcdef0123456789"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) ID of the pool.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the pool.
* `created_timestamp` - Time when the pool was created, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `deletion_protection_enabled` - Whether deletion protection is enabled. When `true`, the pool cannot be deleted.
* `message_type` - Type of message the pool is configured for.
* `opt_out_list_name` - Name of the opt-out list associated with the pool.
* `origination_identities` - Set of origination identity ARNs (phone number ARNs or sender ID ARNs) associated with the pool.
* `self_managed_opt_outs_enabled` - Whether the pool relies on self-managed opt-out handling. When `false`, AWS auto-replies to HELP/STOP requests and manages the opt-out list.
* `shared_routes_enabled` - Whether shared routes are enabled for the pool. When `true`, messages may use shared phone numbers or sender IDs in countries that allow it.
* `status` - Current status of the pool.
* `tags` - Map of tags assigned to the pool.
* `two_way_channel_arn` - Destination for incoming messages.
* `two_way_channel_role` - ARN of the IAM role that End User Messaging SMS assumes to publish inbound messages to the two-way channel.
* `two_way_enabled` - Whether inbound message reception is enabled for the pool.
