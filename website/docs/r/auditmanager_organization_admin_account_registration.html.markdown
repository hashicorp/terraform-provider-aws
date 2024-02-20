---
subcategory: "Audit Manager"
layout: "aws"
page_title: "AWS: aws_auditmanager_organization_admin_account_registration"
description: |-
  Terraform resource for managing AWS Audit Manager Organization Admin Account Registration.
---

# Resource: aws_auditmanager_organization_admin_account_registration

Terraform resource for managing AWS Audit Manager Organization Admin Account Registration.

## Example Usage

### Basic Usage

```terraform
resource "aws_auditmanager_organization_admin_account_registration" "example" {
  admin_account_id = "012345678901"
}
```

## Argument Reference

The following arguments are required:

* `admin_account_id` - (Required) Identifier for the organization administrator account.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Identifier for the organization administrator account.
* `organization_id` - Identifier for the organization.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Audit Manager Organization Admin Account Registration using the `id`. For example:

```terraform
import {
  to = aws_auditmanager_organization_admin_account_registration.example
  id = "012345678901 "
}
```

Using `terraform import`, import Audit Manager Organization Admin Account Registration using the `id`. For example:

```console
% terraform import aws_auditmanager_organization_admin_account_registration.example 012345678901 
```
