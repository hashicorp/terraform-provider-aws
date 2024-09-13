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
  grant_arn = "arn:aws:license-manager::123456789012:grant:g-1cf9fba4ba2f42dcab11c686c4b4d329"
}
```

## Argument Reference

This resource supports the following arguments:

* `grant_arn` - (Required) The ARN of the grant to accept.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

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

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_licensemanager_grant_accepter` using the grant arn. For example:

```terraform
import {
  to = aws_licensemanager_grant_accepter.test
  id = "arn:aws:license-manager::123456789012:grant:g-1cf9fba4ba2f42dcab11c686c4b4d329"
}
```

Using `terraform import`, import `aws_licensemanager_grant_accepter` using the grant arn. For example:

```console
% terraform import aws_licensemanager_grant_accepter.test arn:aws:license-manager::123456789012:grant:g-1cf9fba4ba2f42dcab11c686c4b4d329
```
