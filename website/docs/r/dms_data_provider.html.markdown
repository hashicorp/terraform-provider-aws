---
subcategory: "DMS (Database Migration)"
layout: "aws"
page_title: "AWS: aws_dms_data_provider"
description: |-
  Provides a DMS (Data Migration Service) data provider resource.
---

# Resource: aws_dms_data_provider

Provides a DMS (Data Migration Service) data provider resource. DMS data providers store database connection information.

## Example Usage

### PostgreSQL Data Provider

```terraform
resource "aws_dms_data_provider" "postgres" {
  data_provider_name = "my-postgres-provider"
  engine             = "postgres"

  settings {
    postgres_settings {
      server_name   = "mydb.example.com"
      port          = 5432
      database_name = "mydb"
      ssl_mode      = "require"
    }
  }

  tags = {
    Name = "postgres-provider"
  }
}
```

### MySQL Data Provider

```terraform
resource "aws_dms_data_provider" "mysql" {
  data_provider_name = "my-mysql-provider"
  engine             = "mysql"

  settings {
    mysql_settings {
      server_name = "mydb.example.com"
      port        = 3306
      ssl_mode    = "require"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `engine` - (Required) Database engine type. Valid values: `aurora`, `aurora-postgresql`, `mysql`, `oracle`, `postgres`, `sqlserver`, `redshift`, `mariadb`, `mongodb`, `db2`, `db2-zos`, `docdb`, `sybase`.
* `settings` - (Required) Configuration block for data provider settings. See [`settings`](#settings) below.
* `data_provider_name` - (Optional) User-friendly name for the data provider.
* `description` - (Optional) Description of the data provider.
* `virtual` - (Optional) Indicates whether the data provider is virtual.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### settings

The `settings` block supports one of the following:

* `docdb_settings` - (Optional) Configuration for DocumentDB. See [common settings](#common-settings) below.
* `ibm_db2_luw_settings` - (Optional) Configuration for IBM DB2 LUW. See [common settings](#common-settings) below.
* `ibm_db2_zos_settings` - (Optional) Configuration for IBM DB2 for z/OS. See [common settings](#common-settings) below.
* `mariadb_settings` - (Optional) Configuration for MariaDB. See [common settings](#common-settings) below.
* `microsoft_sql_server_settings` - (Optional) Configuration for Microsoft SQL Server. See [common settings](#common-settings) below.
* `mongodb_settings` - (Optional) Configuration for MongoDB. See [common settings](#common-settings) below.
* `mysql_settings` - (Optional) Configuration for MySQL. See [common settings](#common-settings) below.
* `oracle_settings` - (Optional) Configuration for Oracle. See [common settings](#common-settings) below.
* `postgres_settings` - (Optional) Configuration for PostgreSQL. See [common settings](#common-settings) below.
* `redshift_settings` - (Optional) Configuration for Redshift. See [common settings](#common-settings) below.
* `sybase_ase_settings` - (Optional) Configuration for SAP ASE. See [common settings](#common-settings) below.

### Common Settings

All settings blocks support the following common attributes:

* `server_name` - (Optional) Server name.
* `port` - (Optional) Port number.
* `database_name` - (Optional) Database name.
* `ssl_mode` - (Optional) SSL mode. Valid values: `none`, `require`, `verify-ca`, `verify-full`.
* `certificate_arn` - (Optional) ARN of the certificate for SSL connection.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `data_provider_arn` - ARN of the data provider.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import data providers using the `data_provider_arn`. For example:

```terraform
import {
  to = aws_dms_data_provider.example
  id = "arn:aws:dms:us-east-1:123456789012:data-provider:ABCDEFGHIJKLMNOPQRSTUVWXYZ"
}
```

Using `terraform import`, import data providers using the `data_provider_arn`. For example:

```console
% terraform import aws_dms_data_provider.example arn:aws:dms:us-east-1:123456789012:data-provider:ABCDEFGHIJKLMNOPQRSTUVWXYZ
```
