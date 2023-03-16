---
subcategory: "License Manager"
layout: "aws"
page_title: "AWS: aws_licensemanager_grant_accepter"
description: |-
  Accepts a License Manager grant resource.
---

# Resource: aws_licensemanager_grant_accepter

Accepts a License Manager grant. This allows for sharing licenses with other aws accounts.

## Example Usage

```terraform
resource "aws_licensemanager_grant_accepter" "test" {
  name = "arn:aws:license-manager::123456789012:grant:g-1cf9fba4ba2f42dcab11c686c4b4d329"
}
```

## Argument Reference

The following arguments are supported:

* `grant_arn` - (Required) The ARN of the grant to accept.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The grant ARN (Same as `arn`).
* `arn` - The grant ARN.
* `name` - The Name of the grant.
* `allowed_operations` - A list of the allowed operations for the grant.
* `license_arn` - The ARN of the license for the grant.
* `principal` - The target account for the grant.
* `home_region` - The home region for the license.
* `parent_arn` - The parent ARN.
* `status` - The grant status.
* `version` - The grant version.

## Import

`aws_licensemanager_grant_accepter` can be imported using the grant arn.

```shell
$ terraform import aws_licensemanager_grant_accepter.test arn:aws:license-manager::123456789012:grant:g-1cf9fba4ba2f42dcab11c686c4b4d329
```
