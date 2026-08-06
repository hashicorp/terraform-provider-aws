---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_tenant_database"
description: |-
  Manages an RDS Oracle tenant database (PDB) within a Container Database (CDB) instance.
---

# Resource: aws_rds_tenant_database

Manages an RDS Oracle tenant database (Pluggable Database / PDB) within a Container Database (CDB) instance. Requires an `aws_db_instance` using a CDB engine (`oracle-ee-cdb` or `oracle-se2-cdb`) with `multi_tenant = true`.

## Example Usage

```terraform
resource "aws_db_instance" "cdb" {
  identifier        = "my-oracle-cdb"
  engine            = "oracle-ee-cdb"
  engine_version    = "19.0.0.0.ru-2024-01.rur-2024-01.r1"
  instance_class    = "db.m5.large"
  allocated_storage = 200
  storage_type      = "gp3"
  username          = "admin"
  password          = "changeme"
  license_model     = "bring-your-own-license"
  multi_tenant      = true
  db_name           = "MYPDB"

  skip_final_snapshot = true
}

resource "aws_rds_tenant_database" "example" {
  db_instance_identifier = aws_db_instance.cdb.identifier
  tenant_db_name         = "MYPDB"
  username               = "pdbadmin"
  master_password        = "changeme2"
}
```

### With AWS Secrets Manager Password Management

```terraform
resource "aws_rds_tenant_database" "example" {
  db_instance_identifier      = aws_db_instance.cdb.identifier
  tenant_db_name              = "MYPDB"
  username                    = "pdbadmin"
  manage_master_user_password = true
}
```

## Argument Reference

The following arguments are required:

* `db_instance_identifier` - (Required, Forces new resource) The identifier of the CDB instance that will contain this tenant database.
* `tenant_db_name` - (Required) The name of the tenant database (PDB). Maximum 8 characters. Can be updated to rename the PDB.
* `username` - (Required, Forces new resource) The master username for the tenant database.

The following arguments are optional:

* `character_set_name` - (Optional, Forces new resource) The character set for the tenant database. Defaults to `AL32UTF8`.
* `manage_master_user_password` - (Optional) Whether to manage the master user password with AWS Secrets Manager. Conflicts with `master_password`.
* `master_password` - (Optional, Sensitive) The master user password for the tenant database. Conflicts with `manage_master_user_password`.
* `master_user_secret_kms_key_id` - (Optional) The ARN, key ID, or alias of the KMS key used to encrypt the Secrets Manager secret for the master user password.
* `nchar_character_set_name` - (Optional, Forces new resource) The `NCHAR` character set for the tenant database.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the tenant database.
* `master_user_secret` - A block describing the Secrets Manager secret for the master user password. Only populated when `manage_master_user_password` is `true`. See [master_user_secret](#master_user_secret) below.
* `status` - The status of the tenant database.
* `tenant_database_resource_id` - The AWS Region-unique, immutable identifier for the tenant database.

### master_user_secret

* `kms_key_id` - The KMS key ID used to encrypt the secret.
* `secret_arn` - The ARN of the Secrets Manager secret.
* `secret_status` - The status of the secret. Valid values: `active`, `rotating`, `impersonating-replication-role`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RDS tenant databases using the `tenant_database_resource_id`. For example:

```terraform
import {
  to = aws_rds_tenant_database.example
  id = "tdb-12345678abcdefgh"
}
```

Using `terraform import`, import RDS tenant databases using the `tenant_database_resource_id`:

```console
% terraform import aws_rds_tenant_database.example tdb-12345678abcdefgh
```
