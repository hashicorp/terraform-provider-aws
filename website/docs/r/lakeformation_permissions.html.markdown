---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_permissions"
description: |-
    Grants permissions to the principal to access metadata in the Data Catalog and data organized in underlying data storage such as Amazon S3.
---

# Resource: aws_lakeformation_permissions

Grants permissions to the principal to access metadata in the Data Catalog and data organized in underlying data storage such as Amazon S3. Permissions are granted to a principal, in a Data Catalog, relative to a Lake Formation resource, which includes the Data Catalog, databases, tables, LF-tags, and LF-tag policies. For more information, see [Security and Access Control to Metadata and Data in Lake Formation](https://docs.aws.amazon.com/lake-formation/latest/dg/security-data-access.html).

!> **WARNING:** Lake Formation permissions are not in effect by default within AWS. Using this resource will not secure your data and will result in errors if you do not change the security settings for existing resources and the default security settings for new resources. See [Default Behavior and `IAMAllowedPrincipals`](#default-behavior-and-iamallowedprincipals) for additional details.

~> **NOTE:** In general, the `principal` should _NOT_ be a Lake Formation administrator or the entity (e.g., IAM role) that is running Terraform. Administrators have implicit permissions. These should be managed by granting or not granting administrator rights using `aws_lakeformation_data_lake_settings`, _not_ with this resource.

## Default Behavior and `IAMAllowedPrincipals`

**_Lake Formation permissions are not in effect by default within AWS._** `IAMAllowedPrincipals` (i.e., `IAM_ALLOWED_PRINCIPALS`) conflicts with individual Lake Formation permissions (i.e., non-`IAMAllowedPrincipals` permissions), will cause unexpected behavior, and may result in errors.

When using Lake Formation, choose ONE of the following options as they are mutually exclusive:

1. Use this resource (`aws_lakeformation_permissions`), change the default security settings using [`aws_lakeformation_data_lake_settings`](/docs/providers/aws/r/lakeformation_data_lake_settings.html), and remove existing `IAMAllowedPrincipals` permissions
2. Use `IAMAllowedPrincipals` without `aws_lakeformation_permissions`

This example shows removing the `IAMAllowedPrincipals` default security settings and making the caller a Lake Formation admin. Since `create_database_default_permissions` and `create_table_default_permissions` are not set in the [`aws_lakeformation_data_lake_settings`](/docs/providers/aws/r/lakeformation_data_lake_settings.html) resource, they are cleared.

```terraform
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}
```

To remove existing `IAMAllowedPrincipals` permissions, use the [AWS Lake Formation Console](https://console.aws.amazon.com/lakeformation/) or [AWS CLI](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/lakeformation/batch-revoke-permissions.html).

`IAMAllowedPrincipals` is a hook to maintain backwards compatibility with AWS Glue. `IAMAllowedPrincipals` is a pseudo-entity group that acts like a Lake Formation principal. The group includes any IAM users and roles that are allowed access to your Data Catalog resources by your IAM policies.

This is Lake Formation's default behavior:

* Lake Formation grants `Super` permission to `IAMAllowedPrincipals` on all existing AWS Glue Data Catalog resources.
* Lake Formation enables "Use only IAM access control" for new Data Catalog resources.

For more details, see [Changing the Default Security Settings for Your Data Lake](https://docs.aws.amazon.com/lake-formation/latest/dg/change-settings.html).

### Problem Using `IAMAllowedPrincipals`

AWS does not support combining `IAMAllowedPrincipals` permissions and non-`IAMAllowedPrincipals` permissions. Doing so results in unexpected permissions and behaviors. For example, this configuration grants a user `SELECT` on a column in a table.

```terraform
resource "aws_glue_catalog_database" "example" {
  name = "sadabate"
}

resource "aws_glue_catalog_table" "example" {
  name          = "abelt"
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }
  }
}

resource "aws_lakeformation_permissions" "example" {
  permissions = ["SELECT"]
  principal   = "arn:aws:iam:us-east-1:123456789012:user/SanHolo"

  table_with_columns {
    database_name = aws_glue_catalog_table.example.database_name
    name          = aws_glue_catalog_table.example.name
    column_names  = ["event"]
  }
}
```

The resulting permissions depend on whether the table had `IAMAllowedPrincipals` (IAP) permissions or not.

| Result With IAP | Result Without IAP |
| ---- | ---- |
| `SELECT` column wildcard (i.e., all columns) | `SELECT` on `"event"` (as expected) |

## Using Lake Formation Permissions

Lake Formation grants implicit permissions to data lake administrators, database creators, and table creators. These implicit permissions cannot be revoked _per se_. If this resource reads implicit permissions, it will attempt to revoke them, which causes an error when the resource is destroyed.

There are two ways to avoid these errors. First, and the way we recommend, is to avoid using this resource with principals that have implicit permissions. A second, error-prone option, is to grant explicit permissions (and `permissions_with_grant_option`) to "overwrite" a principal's implicit permissions, which you can then revoke with this resource. For more information, see [Implicit Lake Formation Permissions](https://docs.aws.amazon.com/lake-formation/latest/dg/implicit-permissions.html).

If the `principal` is also a data lake administrator, AWS grants implicit permissions that can cause errors using this resource. For example, AWS implicitly grants a `principal`/administrator `permissions` and `permissions_with_grant_option` of `ALL`, `ALTER`, `DELETE`, `DESCRIBE`, `DROP`, `INSERT`, and `SELECT` on a table. If you use this resource to explicitly grant the `principal`/administrator `permissions` but _not_ `permissions_with_grant_option` of `ALL`, `ALTER`, `DELETE`, `DESCRIBE`, `DROP`, `INSERT`, and `SELECT` on the table, this resource will read the implicit `permissions_with_grant_option` and attempt to revoke them when the resource is destroyed. Doing so will cause an `InvalidInputException: No permissions revoked` error because you cannot revoke implicit permissions _per se_. To workaround this problem, explicitly grant the `principal`/administrator `permissions` _and_ `permissions_with_grant_option`, which can then be revoked. Similarly, granting a `principal`/administrator permissions on a table with columns and providing `column_names`, will result in a `InvalidInputException: Permissions modification is invalid` error because you are narrowing the implicit permissions. Instead, set `wildcard` to `true` and remove the `column_names`.

## Example Usage

### Grant Permissions For A Lake Formation S3 Resource

```terraform
resource "aws_lakeformation_permissions" "example" {
  principal   = aws_iam_role.workflow_role.arn
  permissions = ["ALL"]

  data_location {
    arn = aws_lakeformation_resource.example.arn
  }
}
```

### Grant Permissions For A Glue Catalog Database

```terraform
resource "aws_lakeformation_permissions" "example" {
  principal   = aws_iam_role.workflow_role.arn
  permissions = ["CREATE_TABLE", "ALTER", "DROP"]

  database {
    name       = aws_glue_catalog_database.example.name
    catalog_id = "110376042874"
  }
}
```

### Grant Permissions Using Tag-Based Access Control

```terraform
resource "aws_lakeformation_permissions" "test" {
  role        = aws_iam_role.sales_role.arn
  permissions = ["CREATE_TABLE", "ALTER", "DROP"]
  lf_tag_policy {
    resource_type = "DATABASE"
    expression {
      key    = "Team"
      values = ["Sales"]
    }
    expression {
      key    = "Environment"
      values = ["Dev", "Production"]
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `permissions` – (Required) List of permissions granted to the principal. Valid values may include `ALL`, `ALTER`, `ASSOCIATE`, `CREATE_DATABASE`, `CREATE_TABLE`, `DATA_LOCATION_ACCESS`, `DELETE`, `DESCRIBE`, `DROP`, `INSERT`, and `SELECT`. For details on each permission, see [Lake Formation Permissions Reference](https://docs.aws.amazon.com/lake-formation/latest/dg/lf-permissions-reference.html).
* `principal` – (Required) Principal to be granted the permissions on the resource. Supported principals include `IAM_ALLOWED_PRINCIPALS` (see [Default Behavior and `IAMAllowedPrincipals`](#default-behavior-and-iamallowedprincipals) above), IAM roles, users, groups, SAML groups and users, QuickSight groups, OUs, and organizations as well as AWS account IDs for cross-account permissions. For more information, see [Lake Formation Permissions Reference](https://docs.aws.amazon.com/lake-formation/latest/dg/lf-permissions-reference.html).

~> **NOTE:** We highly recommend that the `principal` _NOT_ be a Lake Formation administrator (granted using `aws_lakeformation_data_lake_settings`). The entity (e.g., IAM role) running Terraform will most likely need to be a Lake Formation administrator. As such, the entity will have implicit permissions and does not need permissions granted through this resource.

One of the following is required:

* `catalog_resource` - (Optional) Whether the permissions are to be granted for the Data Catalog. Defaults to `false`.
* `data_location` - (Optional) Configuration block for a data location resource. Detailed below.
* `database` - (Optional) Configuration block for a database resource. Detailed below.
* `lf_tag` - (Optional) Configuration block for an LF-tag resource. Detailed below.
* `lf_tag_policy` - (Optional) Configuration block for an LF-tag policy resource. Detailed below.
* `table` - (Optional) Configuration block for a table resource. Detailed below.
* `table_with_columns` - (Optional) Configuration block for a table with columns resource. Detailed below.

The following arguments are optional:

* `catalog_id` – (Optional) Identifier for the Data Catalog. By default, the account ID. The Data Catalog is the persistent metadata store. It contains database definitions, table definitions, and other control information to manage your Lake Formation environment.
* `permissions_with_grant_option` - (Optional) Subset of `permissions` which the principal can pass.

### data_location

The following argument is required:

* `arn` – (Required) Amazon Resource Name (ARN) that uniquely identifies the data location resource.

The following argument is optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog where the location is registered with Lake Formation. By default, it is the account ID of the caller.

### database

The following argument is required:

* `name` – (Required) Name of the database resource. Unique to the Data Catalog.

The following argument is optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.

### lf_tag

The following arguments are required:

* `key` – (Required) The key-name for the tag.
* `values` - (Required) A list of possible values an attribute can take.

The following argument is optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.

### lf_tag_policy

The following arguments are required:

* `resource_type` – (Required) The resource type for which the tag policy applies. Valid values are `DATABASE` and `TABLE`.
* `expression` - (Required) A list of tag conditions that apply to the resource's tag policy. Configuration block for tag conditions that apply to the policy. See [`expression`](#expression) below.

The following argument is optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.

#### expression

* `key` – (Required) The key-name of an LF-Tag.
* `values` - (Required) A list of possible values of an LF-Tag.

### table

The following argument is required:

* `database_name` – (Required) Name of the database for the table. Unique to a Data Catalog.
* `name` - (Required, at least one of `name` or `wildcard`) Name of the table.
* `wildcard` - (Required, at least one of `name` or `wildcard`) Whether to use a wildcard representing every table under a database. Defaults to `false`.

The following arguments are optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.

### table_with_columns

The following arguments are required:

* `column_names` - (Required, at least one of `column_names` or `wildcard`) Set of column names for the table.
* `database_name` – (Required) Name of the database for the table with columns resource. Unique to the Data Catalog.
* `name` – (Required) Name of the table resource.
* `wildcard` - (Required, at least one of `column_names` or `wildcard`) Whether to use a column wildcard. If `excluded_column_names` is included, `wildcard` must be set to `true` to avoid Terraform reporting a difference.

The following arguments are optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.
* `excluded_column_names` - (Optional) Set of column names for the table to exclude. If `excluded_column_names` is included, `wildcard` must be set to `true` to avoid Terraform reporting a difference.

## Attributes Reference

No additional attributes are exported.
