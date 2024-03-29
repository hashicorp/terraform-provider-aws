---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_cluster_credentials"
description: |-
  Provides redshift cluster credentials
---

# Data Source: aws_redshift_cluster_credentials

Provides redshift cluster temporary credentials.

## Example Usage

```terraform
data "aws_redshift_cluster_credentials" "example" {
  cluster_identifier = aws_redshift_cluster.example.cluster_identifier
  db_user            = aws_redshift_cluster.example.master_username
}
```

## Argument Reference

This data source supports the following arguments:

* `auto_create` - (Optional)  Create a database user with the name specified for the user named in `db_user` if one does not exist.
* `cluster_identifier` - (Required) Unique identifier of the cluster that contains the database for which your are requesting credentials.
* `db_name` - (Optional) Name of a database that DbUser is authorized to log on to. If `db_name` is not specified, `db_user` can log on to any existing database.
* `db_user` - (Required) Name of a database user. If a user name matching `db_user` exists in the database, the temporary user credentials have the same permissions as the  existing user. If `db_user` doesn't exist in the database and `auto_create` is `True`, a new user is created using the value for `db_user` with `PUBLIC` permissions.  If a database user matching the value for `db_user` doesn't exist and `not` is `False`, then the command succeeds but the connection attempt will fail because the user doesn't exist in the database.
* `db_groups` - (Optional) List of the names of existing database groups that the user named in `db_user` will join for the current session, in addition to any group memberships for an existing user. If not specified, a new user is added only to `PUBLIC`.
* `duration_seconds` - (Optional) The number of seconds until the returned temporary password expires. Valid values are between `900` and `3600`. Default value is `900`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `db_password` - Temporary password that authorizes the user name returned by `db_user` to log on to the database `db_name`.
* `expiration` - Date and time the password in `db_password` expires.
