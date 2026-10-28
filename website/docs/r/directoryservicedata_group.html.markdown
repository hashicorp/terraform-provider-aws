---
subcategory: "Directory Service Data"
layout: "aws"
page_title: "AWS: aws_directoryservicedata_group"
description: |-
  Manages a group in an AWS Directory Service directory.
---

# Resource: aws_directoryservicedata_group

Manages a group in an AWS Directory Service directory.

## Example Usage

### Basic Usage

The following example assumes an existing Directory Service directory with Directory Service Data access enabled.

```terraform
resource "aws_directoryservicedata_group" "example" {
  directory_id     = "d-1234567890"
  sam_account_name = "examplegroup"
  group_scope      = "Global"
  group_type       = "Security"
}
```

## Argument Reference

The following arguments are required:

* `directory_id` - (Required) ID of the Directory that's associated with the group.
* `sam_account_name` - (Required) SAM account name of the group. Must be between 1 and 64 characters and must not contain any of the following characters: `: ; | = + " * ? < > / \ , [ ] @`.

The following arguments are optional:

* `group_scope` - (Optional) Scope of the AD group. Valid values are `DomainLocal`, `Global`, `Universal`, and `BuiltinLocal`. If omitted, the value is read from AWS. See [AD group scopes](https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/manage/understand-security-groups#group-scope) for scope restrictions.
* `group_type` - (Optional) Type of the AD group. Valid values are `Distribution` and `Security`. If omitted, the value is read from AWS. See [AD group types](https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/manage/understand-security-groups#how-active-directory-security-groups-work).
* `region` - (Optional) Region where the group is managed. Defaults to the provider region.

Changes to `group_scope` and `group_type` in the same apply are sent to AWS as separate requests, with the scope updated first.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `distinguished_name` - Distinguished name of the group.
* `sid` - Unique Security identifier (SID) of the group.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_directoryservicedata_group.example
  identity = {
    directory_id     = "d-1234567890"
    sam_account_name = "example-group"
  }
}

resource "aws_directoryservicedata_group" "example" {
  directory_id     = "d-1234567890"
  sam_account_name = "example-group"
}
```

### Identity Schema

#### Required

* `directory_id` (String) ID of the Directory Service directory where the group is managed.
* `sam_account_name` (String) SAM account name of the group.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Directory Service Data Group using the `directory_id` and `sam_account_name`. For example:

```terraform
import {
  to = aws_directoryservicedata_group.example
  id = "d-1234567890,example-group"
}
```

Using `terraform import`, import Directory Service Data Group using the directory ID and SAM account name separated by a comma. For example:

```console
% terraform import aws_directoryservicedata_group.example d-1234567890,example-group
```
