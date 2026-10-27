---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_tenant_database"
description: |-
  Manages a tenant database (pluggable database, PDB) in an RDS for Oracle CDB instance in the multi-tenant configuration.
---

# Resource: aws_rds_tenant_database

Manages a tenant database (pluggable database, PDB) in an RDS for Oracle CDB instance in the multi-tenant configuration. The DB instance must use the `oracle-ee-cdb` or `oracle-se2-cdb` engine with `multi_tenant = true`. See the [AWS documentation](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Oracle.Concepts.CDBs.html) for details.

The initial tenant database is created with the DB instance, using its `db_name`, `username`, and password. Use this resource to manage additional tenant databases or to import the initial one. RDS doesn't allow deleting the last tenant database in a DB instance.

## Example Usage

### Basic Usage

```terraform
resource "aws_db_instance" "example" {
  allocated_storage   = 20
  apply_immediately   = true
  db_name             = "PDB1"
  engine              = "oracle-se2-cdb"
  engine_version      = "19.0.0.0.ru-2026-07.rur-2026-07.r1"
  identifier          = "example"
  instance_class      = "db.t3.small"
  license_model       = "bring-your-own-license"
  multi_tenant        = true
  password_wo         = var.master_password
  password_wo_version = 1
  username            = "admin"
}

resource "aws_rds_tenant_database" "example" {
  db_instance_identifier          = aws_db_instance.example.identifier
  final_snapshot_identifier       = "example-pdb2-final"
  master_user_password_wo         = var.tenant_password
  master_user_password_wo_version = 1
  master_username                 = "pdbadmin"
  tenant_db_name                  = "PDB2"
}
```

### Password Managed by AWS Secrets Manager

```terraform
resource "aws_rds_tenant_database" "example" {
  db_instance_identifier      = aws_db_instance.example.identifier
  manage_master_user_password = true
  master_username             = "pdbadmin"
  skip_final_snapshot         = true
  tenant_db_name              = "PDB2"
}
```

## Argument Reference

The following arguments are required:

* `db_instance_identifier` - (Required, Forces new resource) Identifier of the DB instance that contains the tenant database.
* `master_username` - (Required, Forces new resource) Name of the master user for the tenant database.
* `tenant_db_name` - (Required) Name of the tenant database. Changing it renames the tenant database.

The following arguments are optional:

* `character_set_name` - (Optional, Forces new resource) Character set of the tenant database. Defaults to `AL32UTF8`.
* `final_snapshot_identifier` - (Optional) Name of the DB snapshot created when the tenant database is deleted. Required unless `skip_final_snapshot` is `true`.
* `manage_master_user_password` - (Optional) Whether to manage the master user password with AWS Secrets Manager. Conflicts with `master_user_password` and `master_user_password_wo`.
* `master_user_password` - (Optional) Password for the master user. Stored in the Terraform state. Conflicts with `master_user_password_wo`.
* `master_user_password_wo` - (Optional, Write-Only) Password for the master user. Not stored in the Terraform state. Requires `master_user_password_wo_version`.
* `master_user_password_wo_version` - (Optional) Version of `master_user_password_wo`. Change it to update the password.
* `master_user_secret_kms_key_id` - (Optional) ARN, ID, or alias of the AWS KMS key used to encrypt the Secrets Manager secret when `manage_master_user_password` is `true`. Can't be changed while `manage_master_user_password` is `true`; turn it off, apply, then turn it back on with the new key. Defaults to the `aws/secretsmanager` key.
* `nchar_character_set_name` - (Optional, Forces new resource) National character set of the tenant database. Defaults to `AL16UTF16`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `skip_final_snapshot` - (Optional) Whether to skip the final DB snapshot when the tenant database is deleted. Defaults to `false`.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

~> **Note:** One of `manage_master_user_password`, `master_user_password`, or `master_user_password_wo` is required by RDS when creating a tenant database.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the tenant database.
* `master_user_secret` - Secrets Manager secret for the master user password when `manage_master_user_password` is `true`. See [`master_user_secret` Block](#master_user_secret-block) below.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `tenant_database_resource_id` - Region-unique, immutable identifier of the tenant database.

### `master_user_secret` Block

* `kms_key_id` - ARN, ID, or alias of the AWS KMS key used to encrypt the secret.
* `secret_arn` - ARN of the secret.
* `secret_status` - Status of the secret.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `60m`)
* `delete` - (Default `60m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_rds_tenant_database.example
  identity = {
    tenant_database_resource_id = "tdb-ABCDEFGHIJKLMNOPQRSTUVWXYZ"
  }
}

resource "aws_rds_tenant_database" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `tenant_database_resource_id` (String) Region-unique, immutable identifier of the tenant database.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RDS tenant databases using the `tenant_database_resource_id`. For example:

```terraform
import {
  to = aws_rds_tenant_database.example
  id = "tdb-ABCDEFGHIJKLMNOPQRSTUVWXYZ"
}
```

Using `terraform import`, import RDS tenant databases using the `tenant_database_resource_id`. For example:

```console
% terraform import aws_rds_tenant_database.example tdb-ABCDEFGHIJKLMNOPQRSTUVWXYZ
```
