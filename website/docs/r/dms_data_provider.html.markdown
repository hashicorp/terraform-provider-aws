---
subcategory: "DMS (Database Migration)"
layout: "aws"
page_title: "AWS: aws_dms_data_provider"
description: |-
  Manages an AWS DMS (Database Migration) Data Provider.
---

# Resource: aws_dms_data_provider

Manages an AWS DMS (Database Migration) Data Provider. A data provider stores the database engine and connection settings for a database used in a migration project. Creating a data provider does not create a database.

## Example Usage

### Basic Usage

```terraform
resource "aws_dms_data_provider" "example" {
  engine = "postgres"

  settings {
    postgresql_settings {
      server_name   = "example.com"
      port          = 5432
      database_name = "example"
      ssl_mode      = "none"
    }
  }
}
```

### MySQL Data Provider

```terraform
resource "aws_dms_data_provider" "example" {
  name        = "example-mysql"
  description = "Example MySQL data provider"
  engine      = "mysql"

  settings {
    mysql_settings {
      server_name = "mysql.example.com"
      port        = 3306
      ssl_mode    = "require"
    }
  }

  tags = {
    Environment = "example"
  }
}
```

## Argument Reference

The following arguments are required:

* `engine` - (Required) Database engine for the data provider. Valid values: `aurora`, `aurora-postgresql`, `db2`, `db2-zos`, `docdb`, `mariadb`, `mongodb`, `mysql`, `oracle`, `postgres`, `redshift`, `sqlserver`, and `sybase`. Use `aurora` for Amazon Aurora MySQL-Compatible Edition.
* `settings` - (Required) Database connection settings. Configure exactly one block matching `engine`. See [`settings` Block](#settings-block) below.

The following arguments are optional:

* `description` - (Optional) Description of the data provider. Defaults to an empty string. Removing this argument clears the description.
* `name` - (Optional) Name of the data provider. AWS generates a name when omitted.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `virtual` - (Optional) Whether to create a virtual data provider, which does not require a connection to a database. Defaults to `false`.

### `settings` Block

Configure exactly one of the following blocks:

* `doc_db_settings` - (Optional) Settings for the `docdb` engine. See [`doc_db_settings` Block](#doc_db_settings-block) below.
* `ibm_db2_luw_settings` - (Optional) Settings for the `db2` engine. See [`ibm_db2_luw_settings` Block](#ibm_db2_luw_settings-block) below.
* `ibm_db2_zos_settings` - (Optional) Settings for the `db2-zos` engine. See [`ibm_db2_zos_settings` Block](#ibm_db2_zos_settings-block) below.
* `maria_db_settings` - (Optional) Settings for the `mariadb` engine. See [`maria_db_settings` Block](#maria_db_settings-block) below.
* `microsoft_sql_server_settings` - (Optional) Settings for the `sqlserver` engine. See [`microsoft_sql_server_settings` Block](#microsoft_sql_server_settings-block) below.
* `mongo_db_settings` - (Optional) Settings for the `mongodb` engine. See [`mongo_db_settings` Block](#mongo_db_settings-block) below.
* `mysql_settings` - (Optional) Settings for the `mysql` and `aurora` engines. See [`mysql_settings` Block](#mysql_settings-block) below.
* `oracle_settings` - (Optional) Settings for the `oracle` engine. See [`oracle_settings` Block](#oracle_settings-block) below.
* `postgresql_settings` - (Optional) Settings for the `postgres` and `aurora-postgresql` engines. See [`postgresql_settings` Block](#postgresql_settings-block) below.
* `redshift_settings` - (Optional) Settings for the `redshift` engine. See [`redshift_settings` Block](#redshift_settings-block) below.
* `sybase_ase_settings` - (Optional) Settings for the `sybase` engine. See [`sybase_ase_settings` Block](#sybase_ase_settings-block) below.

### `doc_db_settings` Block

The following arguments are optional:

* `certificate_arn` - (Optional) ARN of the DMS certificate used for the SSL connection.
* `database_name` - (Optional) Database name on the DocumentDB data provider.
* `port` - (Optional) Port of the DocumentDB server. Valid values are between `1` and `65535`.
* `server_name` - (Optional) Hostname of the DocumentDB server.
* `ssl_mode` - (Optional) SSL mode for the connection. Valid values: `none`, `require`, `verify-ca`, and `verify-full`. Defaults to `none`.

### `ibm_db2_luw_settings` Block

The following arguments are optional:

* `certificate_arn` - (Optional) ARN of the DMS certificate used for the SSL connection.
* `database_name` - (Optional) Database name on the IBM DB2 LUW data provider.
* `encryption_algorithm` - (Optional) Integer identifying the encryption algorithm for the connection. When omitted, AWS uses its default behavior.
* `port` - (Optional) Port of the IBM DB2 LUW server. Valid values are between `1` and `65535`.
* `s3_access_role_arn` - (Optional) ARN of the IAM role used to access the S3 bucket containing the user-defined schema.
* `s3_path` - (Optional) S3 path containing the user-defined schema.
* `security_mechanism` - (Optional) Integer identifying the authentication mechanism for the connection. When omitted, AWS uses its default behavior.
* `server_name` - (Optional) Hostname of the IBM DB2 LUW server.
* `ssl_mode` - (Optional) SSL mode for the connection. Valid values: `none` and `verify-ca`. Defaults to `none`.

### `ibm_db2_zos_settings` Block

The following arguments are optional:

* `certificate_arn` - (Optional) ARN of the DMS certificate used for the SSL connection.
* `database_name` - (Optional) Database name on the IBM DB2 for z/OS data provider.
* `port` - (Optional) Port of the IBM DB2 for z/OS server. Valid values are between `1` and `65535`.
* `s3_access_role_arn` - (Optional) ARN of the IAM role used to access the S3 bucket containing the user-defined schema.
* `s3_path` - (Optional) S3 path containing the user-defined schema.
* `server_name` - (Optional) Hostname of the IBM DB2 for z/OS server.
* `ssl_mode` - (Optional) SSL mode for the connection. Valid values: `none` and `verify-ca`. Defaults to `none`.

### `maria_db_settings` Block

The following arguments are optional:

* `certificate_arn` - (Optional) ARN of the DMS certificate used for the SSL connection.
* `port` - (Optional) Port of the MariaDB server. Valid values are between `1` and `65535`.
* `s3_access_role_arn` - (Optional) ARN of the IAM role used to access the S3 bucket containing the user-defined schema.
* `s3_path` - (Optional) S3 path containing the user-defined schema.
* `server_name` - (Optional) Hostname of the MariaDB server.
* `ssl_mode` - (Optional) SSL mode for the connection. Valid values: `none`, `require`, `verify-ca`, and `verify-full`. Defaults to `none`.

### `microsoft_sql_server_settings` Block

The following arguments are optional:

* `certificate_arn` - (Optional) ARN of the DMS certificate used for the SSL connection.
* `database_name` - (Optional) Database name on the Microsoft SQL Server data provider.
* `port` - (Optional) Port of the Microsoft SQL Server instance. Valid values are between `1` and `65535`.
* `s3_access_role_arn` - (Optional) ARN of the IAM role used to access the S3 bucket containing the user-defined schema.
* `s3_path` - (Optional) S3 path containing the user-defined schema.
* `server_name` - (Optional) Hostname of the Microsoft SQL Server instance.
* `ssl_mode` - (Optional) SSL mode for the connection. Valid values: `none`, `require`, `verify-ca`, and `verify-full`. Defaults to `none`.

### `mongo_db_settings` Block

The following arguments are optional:

* `auth_mechanism` - (Optional) Authentication mechanism for the connection. Valid values: `default`, `mongodb_cr`, and `scram_sha_1`.
* `auth_source` - (Optional) Database used to verify credentials. Defaults to `admin`. Not used when `auth_type` is `no`.
* `auth_type` - (Optional) Authentication type for the connection. Valid values: `no` and `password`.
* `certificate_arn` - (Optional) ARN of the DMS certificate used for the SSL connection.
* `database_name` - (Optional) Database name on the MongoDB data provider.
* `port` - (Optional) Port of the MongoDB server. Valid values are between `1` and `65535`.
* `server_name` - (Optional) Hostname of the MongoDB server.
* `ssl_mode` - (Optional) SSL mode for the connection. Valid values: `none`, `require`, `verify-ca`, and `verify-full`. Defaults to `none`.

### `mysql_settings` Block

The following arguments are optional:

* `certificate_arn` - (Optional) ARN of the DMS certificate used for the SSL connection.
* `port` - (Optional) Port of the MySQL server. Valid values are between `1` and `65535`.
* `s3_access_role_arn` - (Optional) ARN of the IAM role used to access the S3 bucket containing the user-defined schema.
* `s3_path` - (Optional) S3 path containing the user-defined schema.
* `server_name` - (Optional) Hostname of the MySQL server.
* `ssl_mode` - (Optional) SSL mode for the connection. Valid values: `none`, `require`, `verify-ca`, and `verify-full`. Defaults to `none`.

### `oracle_settings` Block

The following arguments are optional:

* `asm_server` - (Optional) Address of the Oracle Automatic Storage Management (ASM) server used with Binary Reader. See [Oracle change data capture configuration](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.Oracle.html#CHAP_Source.Oracle.CDC.Configuration).
* `certificate_arn` - (Optional) ARN of the DMS certificate used for the SSL connection.
* `database_name` - (Optional) Database name on the Oracle data provider.
* `port` - (Optional) Port of the Oracle server. Valid values are between `1` and `65535`.
* `s3_access_role_arn` - (Optional) ARN of the IAM role used to access the S3 bucket containing the user-defined schema.
* `s3_path` - (Optional) S3 path containing the user-defined schema.
* `secrets_manager_oracle_asm_access_role_arn` - (Optional) ARN of the IAM role that grants access to the Secrets Manager secret containing Oracle ASM connection details.
* `secrets_manager_oracle_asm_secret_id` - (Optional) Identifier of the Secrets Manager secret containing Oracle ASM connection details. Required when the data provider uses an Oracle ASM server.
* `secrets_manager_security_db_encryption_access_role_arn` - (Optional) ARN of the IAM role that grants access to the Secrets Manager secret containing the transparent data encryption (TDE) password.
* `secrets_manager_security_db_encryption_secret_id` - (Optional) Identifier of the Secrets Manager secret containing the TDE password used by Binary Reader to access encrypted Oracle redo logs.
* `server_name` - (Optional) Hostname of the Oracle server.
* `ssl_mode` - (Optional) SSL mode for the connection. Valid values: `none`, `require`, `verify-ca`, and `verify-full`. Defaults to `none`.

### `postgresql_settings` Block

The following arguments are optional:

* `certificate_arn` - (Optional) ARN of the DMS certificate used for the SSL connection.
* `database_name` - (Optional) Database name on the PostgreSQL data provider.
* `port` - (Optional) Port of the PostgreSQL server. Valid values are between `1` and `65535`.
* `s3_access_role_arn` - (Optional) ARN of the IAM role used to access the S3 bucket containing the user-defined schema.
* `s3_path` - (Optional) S3 path containing the user-defined schema.
* `server_name` - (Optional) Hostname of the PostgreSQL server.
* `ssl_mode` - (Optional) SSL mode for the connection. Valid values: `none`, `require`, `verify-ca`, and `verify-full`. Defaults to `none`.

### `redshift_settings` Block

The following arguments are optional:

* `database_name` - (Optional) Database name on the Amazon Redshift data provider.
* `port` - (Optional) Port of the Amazon Redshift server. Valid values are between `1` and `65535`.
* `s3_access_role_arn` - (Optional) ARN of the IAM role used to access the S3 bucket containing the user-defined schema.
* `s3_path` - (Optional) S3 path containing the user-defined schema.
* `server_name` - (Optional) Hostname of the Amazon Redshift server.

### `sybase_ase_settings` Block

The following arguments are optional:

* `certificate_arn` - (Optional) ARN of the DMS certificate used for the SSL connection.
* `database_name` - (Optional) Database name on the SAP ASE data provider.
* `encrypt_password` - (Optional) Whether to encrypt the connection password during transmission. Defaults to `true`.
* `port` - (Optional) Port of the SAP ASE server. Valid values are between `1` and `65535`.
* `server_name` - (Optional) Hostname of the SAP ASE server.
* `ssl_mode` - (Optional) SSL mode for the connection. Valid values: `none`, `require`, `verify-ca`, and `verify-full`. Defaults to `none`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the data provider.
* `creation_time` - Creation time of the data provider in RFC3339 format.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_dms_data_provider.example
  identity = {
    arn = "arn:aws:dms:us-east-1:123456789012:data-provider:EXAMPLEABCDEFGHIJKLMNOPQRS"
  }
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the data provider.

#### Optional

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a DMS data provider using its full ARN. For example:

```terraform
import {
  to = aws_dms_data_provider.example
  id = "arn:aws:dms:us-east-1:123456789012:data-provider:EXAMPLEABCDEFGHIJKLMNOPQRS"
}
```

Using `terraform import`, import a DMS data provider using its full ARN. For example:

```console
% terraform import aws_dms_data_provider.example arn:aws:dms:us-east-1:123456789012:data-provider:EXAMPLEABCDEFGHIJKLMNOPQRS
```
