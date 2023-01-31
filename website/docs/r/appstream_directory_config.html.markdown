---
subcategory: "AppStream 2.0"
layout: "aws"
page_title: "AWS: aws_appstream_directory_config"
description: |-
  Provides an AppStream Directory Config
---

# Resource: aws_appstream_directory_config

Provides an AppStream Directory Config.

## Example Usage

```terraform
resource "aws_appstream_directory_config" "example" {
  directory_name                          = "NAME OF DIRECTORY"
  organizational_unit_distinguished_names = ["DISTINGUISHED NAME"]

  service_account_credentials {
    account_name     = "NAME OF ACCOUNT"
    account_password = "PASSWORD OF ACCOUNT"
  }
}
```

## Argument Reference

The following arguments are required:

* `directory_name` - (Required) Fully qualified name of the directory.
* `organizational_unit_distinguished_names` - (Required) Distinguished names of the organizational units for computer accounts.
* `service_account_credentials` - (Required) Configuration block for the name of the directory and organizational unit (OU) to use to join the directory config to a Microsoft Active Directory domain. See [`service_account_credentials`](#service_account_credentials) below.

### `service_account_credentials`

* `account_name` - (Required) User name of the account. This account must have the following privileges: create computer objects, join computers to the domain, and change/reset the password on descendant computer objects for the organizational units specified.
* `account_password` - (Required) Password for the account.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique identifier (ID) of the appstream directory config.
* `created_time` -  Date and time, in UTC and extended RFC 3339 format, when the directory config was created.

## Import

`aws_appstream_directory_config` can be imported using the id, e.g.,

```
$ terraform import aws_appstream_directory_config.example directoryNameExample
```
