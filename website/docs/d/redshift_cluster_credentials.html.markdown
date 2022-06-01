---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_cluster_credentials"
description: |-
  Provides redshift cluster credentials
---

# Data Source: aws_redshift_cluster_credentials

Provides redshift subnet group.

## Example Usage

```terraform
data "aws_redshift_cluster_credentials" "example" {
  name = aws_redshift_cluster_credentials.example.name
}
```

## Argument Reference

The following arguments are supported:

* `auto_create` - (Optional)  Create a database user with the name specified for the user named in `db_user` if one does not exist.
* `cluster_identifier` - (Required) The unique identifier of the cluster that contains the database for which your are requesting credentials.
* `db_name` - (Optional) The name of a database that DbUser is authorized to log on to. If `db_name` is not specified, `db_user` can log on to any existing database.
* `db_user` - (Required) The name of a database user. If a user name matching `db_user` exists in the database, the temporary user credentials have the same permissions as the  existing user. If `db_user` doesn't exist in the database and `auto_create` is `True`, a new user is created using the value for `db_user` with `PUBLIC` permissions.  If a database user matching the value for `db_user` doesn't exist and `not` is `False`, then the command succeeds but the connection attempt will fail because the user doesn't exist in the database.
* `db_groups` - (Optional) A list of the names of existing database groups that the user named in `db_user` will join for the current session, in addition to any group memberships for an existing user. If not specified, a new user is added only to `PUBLIC`.
* `duration_seconds` - (Optional)  The number of seconds until the returned temporary password expires. Valid values are between `900` and `3600`. Default value is `900`.


## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `db_password` - A temporary password that authorizes the user name returned by `db_user` to log on to the database `db_name`.
* `expiration` - The date and time the password in `db_password` expires.
