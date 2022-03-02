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
  alias           = "test-alias"
  directory_id    = "d-xxxxxxx"
}
```

## Argument Reference

The following arguments are supported:

- `alias` - (Required) The alias for the organization.
- `directory_id` - (Optional) The AWS Directory Service directory ID.
- `enable_interoperability` - (Optional) When `true` , allows organization interoperability between Amazon WorkMail and Microsoft Exchange. Can only be set to true if an AD Connector directory ID is included in the request. Defaults to `false`. Valid values: `true`, `false`.
- `domains` - (Optional) The email domains to associate with the organization.
- `kms_key_arn` - (Optional) The Amazon Resource Name (ARN) of a customer managed master key from AWS KMS.

**domains** supports the following attributes:

- `domain_name` - (Optional) The fully qualified domain name.
- `hosted_zone_id` - (Optional) The hosted zone ID for a domain hosted in Route 53. Required when configuring a domain hosted in Route 53.

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
