---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_instance"
description: |-
  Provides details about a specific Connect Instance.
---

# Data Source: aws_connect_instance

Provides details about a specific Amazon Connect Instance.

## Example Usage

By instance_alias

```hcl
data "aws_connect_instance" "foo" {
  instance_alias = "foo"
}
```

By instance_id

```hcl
data "aws_connect_instance" "foo" {
  instance_id = "97afc98d-101a-ba98-ab97-ae114fc115ec"
}
```

## Argument Reference

~> **NOTE:** One of either `instance_id` or `instance_alias` is required.

This data source supports the following arguments:

* `instance_id` - (Optional) Returns information on a specific connect instance by id

* `instance_alias` - (Optional) Returns information on a specific connect instance by alias

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `created_time` - When the instance was created.
* `arn` - ARN of the instance.
* `identity_management_type` - Specifies The identity management type attached to the instance.
* `inbound_calls_enabled` - Whether inbound calls are enabled.
* `outbound_calls_enabled` - Whether outbound calls are enabled.
* `early_media_enabled` - Whether early media for outbound calls is enabled .
* `contact_flow_logs_enabled` - Whether contact flow logs are enabled.
* `contact_lens_enabled` - Whether contact lens is enabled.
* `auto_resolve_best_voices` - Whether auto resolve best voices is enabled.
* `multi_party_conference_enabled` - Whether multi-party calls/conference is enabled.
* `use_custom_tts_voices` - Whether use custom tts voices is enabled.
* `status` - State of the instance.
* `service_role` - Service role of the instance.
