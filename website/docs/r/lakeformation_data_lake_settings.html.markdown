---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_data_lake_settings"
description: |-
  Manages data lake administrators and default database and table permissions
---

# Resource: aws_lakeformation_data_lake_settings

Manages Lake Formation principals designated as data lake administrators and lists of principal permission entries for default create database and default create table permissions.

## Example Usage


### Data Lake Admins

```hcl
resource "aws_iam_user" "test" {
  name = "username"
}

resource "aws_iam_role" "test" {
  name = "rolename"
}

resource "aws_lakeformation_data_lake_settings" "example" {
  data_lake_admins = [aws_iam_user.test.arn, aws_iam_role.test.arn]
}
```

### Create Default Permissions

```hcl
resource "aws_lakeformation_data_lake_settings" "example" {
  data_lake_admins = [aws_iam_user.test.arn, aws_iam_role.test.arn]

  create_database_default_permissions {
    permissions = ["SELECT", "ALTER", "DROP"]
    principal   = aws_iam_user.test.arn
  }

  create_table_default_permissions {
    permissions = ["ALL"]
    principal   = aws_iam_role.test.arn
  }
}
```

## Argument Reference

The following arguments are optional:

* `catalog_id` – (Optional) Identifier for the Data Catalog. By default, the account ID.
* `create_database_default_permissions` - (Optional) Up to three configuration blocks of principal permissions for default create database permissions. Detailed below.
* `create_table_default_permissions` - (Optional) Up to three configuration blocks of principal permissions for default create table permissions. Detailed below.
* `data_lake_admins` – (Optional) List of ARNs of AWS Lake Formation principals (IAM users or roles).
* `trusted_resource_owners` – (Optional) List of the resource-owning account IDs that the caller's account can use to share their user access details (user ARNs).

### create_database_default_permissions

The following arguments are optional:

* `permissions` - (Optional) List of permissions that are granted to the principal. Valid values include `ALL`, `SELECT`, `ALTER`, `DROP`, `DELETE`, `INSERT`, `DESCRIBE`, `CREATE_DATABASE`, `CREATE_TABLE`, and `DATA_LOCATION_ACCESS`.
* `principal` - (Optional) Identifier for the Lake Formation principal. Supported principals are IAM users or IAM roles.

### create_table_default_permissions

The following arguments are optional:

* `permissions` - (Optional) List of permissions that are granted to the principal. Valid values include `ALL`, `SELECT`, `ALTER`, `DROP`, `DELETE`, `INSERT`, `DESCRIBE`, `CREATE_DATABASE`, `CREATE_TABLE`, and `DATA_LOCATION_ACCESS`.
* `principal` - (Optional) Identifier for the Lake Formation principal. Supported principals are IAM users or IAM roles.

## Attributes Reference

In addition to all arguments above, no attributes are exported.
