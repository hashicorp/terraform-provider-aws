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

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Identifier for the organization administrator account.
* `organization_id` - Identifier for the organization.

## Import

Audit Manager Organization Admin Account Registration can be imported using the `id`, e.g.,

```
$ terraform import aws_auditmanager_organization_admin_account_registration.example 012345678901 
```
