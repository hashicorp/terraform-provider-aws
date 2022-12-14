---
subcategory: "Audit Manager"
layout: "aws"
page_title: "AWS: aws_auditmanager_account_registration"
description: |-
  Terraform resource for managing AWS Audit Manager Account Registration.
---

# Resource: aws_auditmanager_account_registration

Terraform resource for managing AWS Audit Manager Account Registration.

## Example Usage

### Basic Usage

```terraform
resource "aws_auditmanager_account_registration" "example" {}
```

### Deregister On Destroy

```terraform
resource "aws_auditmanager_account_registration" "example" {
  deregister_on_destroy = true
}
```

## Argument Reference

The following arguments are optional:

* `delegated_admin_account` - (Optional) Identifier for the delegated administrator account.
* `deregister_on_destroy` - (Optional) Flag to deregister AuditManager in the account upon destruction. Defaults to `false` (ie. AuditManager will remain active in the account, even if this resource is removed).
* `kms_key` - (Optional) KMS key identifier.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique identifier for the account registration. Since registration is applied per AWS region, this will be the active region name (ex. `us-east-1`).
* `status` - Status of the account registration request.

## Import

Audit Manager Account Registration resources can be imported using the `id`, e.g.,

```
$ terraform import aws_auditmanager_account_registration.example us-east-1
```
