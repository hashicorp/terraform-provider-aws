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

* `data_provider_name` - (Optional) User-friendly name for the data provider.
* `description` - (Optional) Description of the data provider.
* `engine` - (Required) Database engine type. Valid values: `aurora`, `aurora-postgresql`, `mysql`, `oracle`, `postgres`, `sqlserver`, `redshift`, `mariadb`, `mongodb`, `db2`, `db2-zos`, `docdb`, `sybase`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `settings` - (Required) Configuration block for data provider settings. See [`settings`](#settings-block) below.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `virtual` - (Optional) Whether the data provider is virtual.

### `settings` Block

The `settings` block supports one of the following:

* `docdb_settings` - (Optional) Configuration for DocumentDB. See [`docdb_settings` Block](#docdb_settings-block) below.
* `ibm_db2_luw_settings` - (Optional) Configuration for IBM DB2 LUW. See [`ibm_db2_luw_settings` Block](#ibm_db2_luw_settings-block) below.
* `ibm_db2_zos_settings` - (Optional) Configuration for IBM DB2 for z/OS. See [`ibm_db2_zos_settings` Block](#ibm_db2_zos_settings-block) below.
* `mariadb_settings` - (Optional) Configuration for MariaDB. See [`mariadb_settings` Block](#mariadb_settings-block) below.
* `microsoft_sql_server_settings` - (Optional) Configuration for Microsoft SQL Server. See [`microsoft_sql_server_settings` Block](#microsoft_sql_server_settings-block) below.
* `mongodb_settings` - (Optional) Configuration for MongoDB. See [`mongodb_settings` Block](#mongodb_settings-block) below.
* `mysql_settings` - (Optional) Configuration for MySQL. See [`mysql_settings` Block](#mysql_settings-block) below.
* `oracle_settings` - (Optional) Configuration for Oracle. See [`oracle_settings` Block](#oracle_settings-block) below.
* `postgres_settings` - (Optional) Configuration for PostgreSQL. See [`postgres_settings` Block](#postgres_settings-block) below.
* `redshift_settings` - (Optional) Configuration for Redshift. See [`redshift_settings` Block](#redshift_settings-block) below.
* `sybase_ase_settings` - (Optional) Configuration for SAP ASE. See [`sybase_ase_settings` Block](#sybase_ase_settings-block) below.

### `docdb_settings` Block

The `docdb_settings` block supports the following:

* `certificate_arn` - (Optional) ARN of the certificate for SSL connection.
* `database_name` - (Optional) Database name.
* `port` - (Optional) Port number.
* `server_name` - (Optional) Server name.
* `ssl_mode` - (Optional) SSL mode. Valid values: `none`, `require`, `verify-ca`, `verify-full`.

### `ibm_db2_luw_settings` Block

The `ibm_db2_luw_settings` block supports the following:

* `certificate_arn` - (Optional) ARN of the certificate for SSL connection.
* `database_name` - (Optional) Database name.
* `port` - (Optional) Port number.
* `server_name` - (Optional) Server name.
* `ssl_mode` - (Optional) SSL mode. Valid values: `none`, `require`, `verify-ca`, `verify-full`.

### `ibm_db2_zos_settings` Block

The `ibm_db2_zos_settings` block supports the following:

* `certificate_arn` - (Optional) ARN of the certificate for SSL connection.
* `database_name` - (Optional) Database name.
* `port` - (Optional) Port number.
* `server_name` - (Optional) Server name.
* `ssl_mode` - (Optional) SSL mode. Valid values: `none`, `require`, `verify-ca`, `verify-full`.

### `mariadb_settings` Block

The `mariadb_settings` block supports the following:

* `certificate_arn` - (Optional) ARN of the certificate for SSL connection.
* `database_name` - (Optional) Database name.
* `port` - (Optional) Port number.
* `server_name` - (Optional) Server name.
* `ssl_mode` - (Optional) SSL mode. Valid values: `none`, `require`, `verify-ca`, `verify-full`.

### `microsoft_sql_server_settings` Block

The `microsoft_sql_server_settings` block supports the following:

* `certificate_arn` - (Optional) ARN of the certificate for SSL connection.
* `database_name` - (Optional) Database name.
* `port` - (Optional) Port number.
* `server_name` - (Optional) Server name.
* `ssl_mode` - (Optional) SSL mode. Valid values: `none`, `require`, `verify-ca`, `verify-full`.

### `mongodb_settings` Block

The `mongodb_settings` block supports the following:

* `certificate_arn` - (Optional) ARN of the certificate for SSL connection.
* `database_name` - (Optional) Database name.
* `port` - (Optional) Port number.
* `server_name` - (Optional) Server name.
* `ssl_mode` - (Optional) SSL mode. Valid values: `none`, `require`, `verify-ca`, `verify-full`.

### `mysql_settings` Block

The `mysql_settings` block supports the following:

* `certificate_arn` - (Optional) ARN of the certificate for SSL connection.
* `database_name` - (Optional) Database name.
* `port` - (Optional) Port number.
* `server_name` - (Optional) Server name.
* `ssl_mode` - (Optional) SSL mode. Valid values: `none`, `require`, `verify-ca`, `verify-full`.

### `oracle_settings` Block

The `oracle_settings` block supports the following:

* `certificate_arn` - (Optional) ARN of the certificate for SSL connection.
* `database_name` - (Optional) Database name.
* `port` - (Optional) Port number.
* `server_name` - (Optional) Server name.
* `ssl_mode` - (Optional) SSL mode. Valid values: `none`, `require`, `verify-ca`, `verify-full`.

### `postgres_settings` Block

The `postgres_settings` block supports the following:

* `certificate_arn` - (Optional) ARN of the certificate for SSL connection.
* `database_name` - (Optional) Database name.
* `port` - (Optional) Port number.
* `server_name` - (Optional) Server name.
* `ssl_mode` - (Optional) SSL mode. Valid values: `none`, `require`, `verify-ca`, `verify-full`.

### `redshift_settings` Block

The `redshift_settings` block supports the following:

* `certificate_arn` - (Optional) ARN of the certificate for SSL connection.
* `database_name` - (Optional) Database name.
* `port` - (Optional) Port number.
* `server_name` - (Optional) Server name.
* `ssl_mode` - (Optional) SSL mode. Valid values: `none`, `require`, `verify-ca`, `verify-full`.

### `sybase_ase_settings` Block

The `sybase_ase_settings` block supports the following:

* `certificate_arn` - (Optional) ARN of the certificate for SSL connection.
* `database_name` - (Optional) Database name.
* `port` - (Optional) Port number.
* `server_name` - (Optional) Server name.
* `ssl_mode` - (Optional) SSL mode. Valid values: `none`, `require`, `verify-ca`, `verify-full`.

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
