---
subcategory: "WorkMail"
layout: "aws"
page_title: "AWS: aws_workmail_domain"
description: |-
  Manages a mail domain registered to an AWS WorkMail organization.
---

# Resource: aws_workmail_domain

Manages a mail domain registered to an AWS WorkMail organization.

## Example Usage

### Basic Usage

```terraform
resource "aws_workmail_domain" "example" {
  organization_id = aws_workmail_organization.example.id
  domain_name     = "example.com"
}
```

## Argument Reference

This resource supports the following arguments:

* `domain_name` - (Required) Mail domain name to register. Changing this forces a new resource.
* `organization_id` - (Required) Identifier of the WorkMail organization. Changing this forces a new resource.

### `records`

Each `records` block exports the following:

* `hostname` - DNS record hostname.
* `type` - DNS record type (e.g. `CNAME`, `MX`, `TXT`).
* `value` - DNS record value.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Composite ID in the form `organization_id,domain_name`.
* `dkim_verification_status` - DKIM verification status. Values: `PENDING`, `VERIFIED`, `FAILED`.
* `is_default` - Whether this domain is the default mail domain for the organization.
* `is_test_domain` - Whether this is the auto-provisioned test domain.
* `ownership_verification_status` - Domain ownership verification status. Values: `PENDING`, `VERIFIED`, `FAILED`.
* `records` - List of DNS records required for domain verification. See [`records`](#records) below.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkMail Domain using `organization_id,domain_name`. For example:

```terraform
import {
  to = aws_workmail_domain.example
  id = "m-1234567890abcdef0,example.com"
}
```

Using `terraform import`, import WorkMail Domain using `organization_id,domain_name`. For example:

```console
% terraform import aws_workmail_domain.example "m-1234567890abcdef0,example.com"
```
