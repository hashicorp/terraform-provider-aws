---
subcategory: "Redshift Serverless"
layout: "aws"
page_title: "AWS: aws_redshiftserverless_credentials"
description: |-
  Provides redshift serverless credentials
---

# Data Source: aws_redshiftserverless_credentials

Provides redshift serverless temporary credentials for a workgroup.

## Example Usage

```terraform
data "aws_redshiftserverless_credentials" "example" {
  workgroup_name = aws_redshiftserverless_workgroup.example.workgroup_name
}
```

## Argument Reference

The following arguments are supported:

* `workgroup_name` - (Required) The name of the workgroup associated with the database.
* `db_name` - (Optional) The name of the database to get temporary authorization to log on to.
* `duration_seconds` - (Optional) The number of seconds until the returned temporary password expires. The minimum is 900 seconds, and the maximum is 3600 seconds.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `db_password` - Temporary password that authorizes the user name returned by `db_user` to log on to the database `db_name`.
* `db_user` - A database user name that is authorized to log on to the database `db_name` using the password `db_password` . If the specified `db_user` exists in the database, the new user name has the same database privileges as the user named in `db_user` . By default, the user is added to PUBLIC. the user doesn't exist in the database.
* `expiration` - Date and time the password in `db_password` expires.
