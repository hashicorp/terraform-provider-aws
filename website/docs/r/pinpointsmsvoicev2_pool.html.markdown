---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_pool"
description: |-
  Manages an AWS End User Messaging SMS Pool.
---

# Resource: aws_pinpointsmsvoicev2_pool

Manages an AWS End User Messaging SMS Pool.

## Example Usage

### Basic Usage

```terraform
resource "aws_pinpointsmsvoicev2_phone_number" "example" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_pool" "example" {
  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.example.arn]

  tags = {
    Name = "example"
  }
}
```

### Two-Way Channel

```terraform
resource "aws_sns_topic" "example" {
  name = "example-two-way-channel"
}

resource "aws_iam_role" "example" {
  name = "example-pool-two-way"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "sms-voice.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_pinpointsmsvoicev2_phone_number" "example" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_pool" "example" {
  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.example.arn]
  two_way_enabled        = true
  two_way_channel_arn    = aws_sns_topic.example.arn
  two_way_channel_role   = aws_iam_role.example.arn

  tags = {
    Name = "example"
  }
}
```

## Argument Reference

The following arguments are required:

* `message_type` - (Required) Type of message. Valid values are `TRANSACTIONAL` and `PROMOTIONAL`. Cannot be changed after creation.
* `origination_identities` - (Required) Set of origination identity ARNs (phone number ARNs or sender ID ARNs) associated with the pool. At least one identity is required at creation.

The following arguments are optional:

* `deletion_protection_enabled` - (Optional) Whether deletion protection is enabled. When `true`, the pool cannot be deleted.
* `iso_country_code` - (Optional) Two-character code, in ISO 3166-1 alpha-2 format, for the country or region of the pool. Cannot be changed after creation.
* `opt_out_list_name` - (Optional) Name of the opt-out list associated with the pool.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-options.html#cli-configure-options-region). Defaults to the region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `self_managed_opt_outs_enabled` - (Optional) Whether the pool relies on self-managed opt-out handling. When `false`, AWS auto-replies to HELP/STOP requests and manages the opt-out list.
* `shared_routes_enabled` - (Optional) Whether shared routes are enabled for the pool. When `true`, messages may use shared phone numbers or sender IDs in countries that allow it.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `two_way_channel_arn` - (Optional) Destination for incoming messages. Specify an ARN to receive incoming messages, or `connect.[region].amazonaws.com` (with `[region]` replaced by the AWS Region of the Amazon Connect instance) to set Amazon Connect as the inbound destination.
* `two_way_channel_role` - (Optional) ARN of the IAM role that End User Messaging SMS assumes to publish inbound messages to the two-way channel.
* `two_way_enabled` - (Optional) Whether inbound message reception is enabled for the pool. When `true`, `two_way_channel_arn` must be set.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the pool.
* `id` - ID of the pool.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

~> **Note:** `iso_country_code` is never returned by AWS, so importing a pool with `iso_country_code` set plans a replacement until removed from config.

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_pool.example
  identity = {
    id = "pool-abcdef0123456789abcdef0123456789"
  }
}

resource "aws_pinpointsmsvoicev2_pool" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` - (String) ID of the Pool.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an AWS End User Messaging SMS Pool using the `id`. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_pool.example
  id = "pool-abcdef0123456789abcdef0123456789"
}
```

Using `terraform import`, import an AWS End User Messaging SMS Pool using the `id`. For example:

```console
% terraform import aws_pinpointsmsvoicev2_pool.example pool-abcdef0123456789abcdef0123456789
```
