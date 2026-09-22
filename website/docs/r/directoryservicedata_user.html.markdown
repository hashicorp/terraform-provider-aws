---
subcategory: "Directory Service Data"
layout: "aws"
page_title: "AWS: aws_directoryservicedata_user"
description: |-
  Manages a user in an AWS Directory Service directory.
---

# Resource: aws_directoryservicedata_user

Manages a user in an AWS Directory Service directory.

## Example Usage

### Basic Usage

The following example assumes an existing Directory Service directory with Directory Service Data access enabled.

```terraform
resource "aws_directoryservicedata_user" "example" {
  directory_id     = "d-1234567890"
  sam_account_name = "exampleuser"
  email_address    = "exampleuser@example.com"
  given_name       = "Example"
  surname          = "User"
}
```

## Argument Reference

The following arguments are required:

* `directory_id` - (Required) Identifier ID of the Directory that's associated with the user.
* `sam_account_name` - (Required) SAM account name of the user. Must contain only word characters, hyphens, and periods.

The following arguments are optional:

* `email_address` - (Optional) Email address of the user.
* `given_name` - (Optional) First name of the user.
* `region` - (Optional) Region where the user is managed. Defaults to the provider region.
* `surname` - (Optional) Last name of the user.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `distinguished_name` - Distinguished name of the user.
* `enabled` - Whether the user is active.
* `realm` - Realm of the user.
* `sid` - Unique Security identifier (SID) of the user.
* `user_principal_name` - User principal name of the user.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_directoryservicedata_user.example
  identity = {
    directory_id     = "d-1234567890"
    sam_account_name = "exampleuser"
  }
}

resource "aws_directoryservicedata_user" "example" {
  directory_id     = "d-1234567890"
  sam_account_name = "exampleuser"
}
```

### Identity Schema

#### Required

* `directory_id` (String) ID of the Directory Service directory where the user is managed.
* `sam_account_name` (String) SAM account name of the user.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Directory Service Data User using the `directory_id` and `sam_account_name`. For example:

```terraform
import {
  to = aws_directoryservicedata_user.example
  id = "d-1234567890,exampleuser"
}
```

Using `terraform import`, import Directory Service Data User using the directory ID and SAM account name separated by a comma. For example:

```console
% terraform import aws_directoryservicedata_user.example d-1234567890,exampleuser
```
