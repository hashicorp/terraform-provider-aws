---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_instance"
description: |-
  Provides details about a specific Connect Instance.
---

# Resource: aws_connect_instance

Provides an Amazon Connect instance resource. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html)

!> **WARN:** Amazon Connect enforces a limit of [100 combined instance creation and deletions every 30 days](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-service-limits.html#feature-limits).  For example, if you create 80 instances and delete 20 of them, you must wait 30 days to create or delete another instance.  Use care when creating or deleting instances.

## Example Usage

```terraform
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = "friendly-name-connect"
  outbound_calls_enabled   = true
}
```

## Example Usage with Existing Active Directory

```terraform
resource "aws_connect_instance" "test" {
  directory_id             = aws_directory_service_directory.test.id
  identity_management_type = "EXISTING_DIRECTORY"
  inbound_calls_enabled    = true
  instance_alias           = "friendly-name-connect"
  outbound_calls_enabled   = true
}
```

## Example Usage with SAML

```terraform
resource "aws_connect_instance" "test" {
  identity_management_type = "SAML"
  inbound_calls_enabled    = true
  instance_alias           = "friendly-name-connect"
  outbound_calls_enabled   = true
}
```

## Argument Reference

The following arguments are supported:

* `auto_resolve_best_voices_enabled` - (Optional) Specifies whether auto resolve best voices is enabled. Defaults to `true`.
* `contact_flow_logs_enabled` - (Optional) Specifies whether contact flow logs are enabled. Defaults to `false`.
* `contact_lens_enabled` - (Optional) Specifies whether contact lens is enabled. Defaults to `true`.
* `directory_id` - (Optional) The identifier for the directory if identity_management_type is `EXISTING_DIRECTORY`.
* `early_media_enabled` - (Optional) Specifies whether early media for outbound calls is enabled . Defaults to `true` if outbound calls is enabled.
* `identity_management_type` - (Required) Specifies the identity management type attached to the instance. Allowed Values are: `SAML`, `CONNECT_MANAGED`, `EXISTING_DIRECTORY`.
* `inbound_calls_enabled` - (Required) Specifies whether inbound calls are enabled.
* `instance_alias` - (Optional) Specifies the name of the instance. Required if `directory_id` not specified.
* `outbound_calls_enabled` - (Required) Specifies whether outbound calls are enabled.
<!-- * `use_custom_tts_voices` - (Optional) Specifies Whether use custom tts voices is enabled. Defaults to `false` -->

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The identifier of the instance.
* `arn` - Amazon Resource Name (ARN) of the instance.
* `created_time` - Specifies when the instance was created.
* `service_role` - The service role of the instance.
* `status` - The state of the instance.

### Timeouts

`aws_connect_instance` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Defaults to 5 mins) Used when creating the instance.
* `delete` - (Defaults to 5 mins) Used when deleting the instance.

## Import

Connect instances can be imported using the `id`, e.g.,

```
$ terraform import aws_connect_instance.example f1288a1f-6193-445a-b47e-af739b2
```
