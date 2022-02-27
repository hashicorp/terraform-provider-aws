---
subcategory: "WorkMail"
layout: "aws"
page_title: "AWS: aws_workmail_organization"
description: |-
  Provides a WorkMail organization.
---

# Resource: aws_workmail_organization

Creates a WorkMail organization.

## Example Usage

```terraform
resource "aws_workmail_organization" "example" {
  alias           = "Example"
  directory_id    = "d-xxxxxxx"
}
```

## Argument Reference

The following arguments are supported:

- `alias` - (Required) The name of the accelerator.
- `directory_id` - (Optional) The value for the address type. Defaults to `IPV4`. Valid values: `IPV4`.
- `enable_interoperability` - (Optional) Indicates whether the accelerator is enabled. Defaults to `true`. Valid values: `true`, `false`.
- `attributes` - (Optional) The attributes of the accelerator. Fields documented below.
- `kms_key_arn` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

**domains** supports the following attributes:

- `domain_name` - (Optional) Indicates whether flow logs are enabled. Defaults to `false`. Valid values: `true`, `false`.
- `hosted_zone_id` - (Optional) The name of the Amazon S3 bucket for the flow logs. Required if `flow_logs_enabled` is `true`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `organization_id` - The identifier of the WorkMail organization.
- `directory_type` - The type of directory associated with the WorkMail organization.
- `state` - The state of the WorkMail organization.

## Import

WorkMail organizations can be imported using the `organization_id`, e.g.,

```
$ terraform import aws_workmail_organization.example m-xxxxxxxxxxxxxxxxxxxx
```
