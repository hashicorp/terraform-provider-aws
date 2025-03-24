---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_cluster_snapshot_copy"
description: |-
  Manages an RDS database cluster snapshot copy.
---

# Resource: aws_rds_cluster_snapshot_copy

Manages an RDS database cluster snapshot copy. For managing RDS database instance snapshot copies, see the [`aws_db_snapshot_copy` resource](/docs/providers/aws/r/db_snapshot_copy.html).

## Example Usage

```terraform
resource "aws_rds_cluster" "example" {
  cluster_identifier  = "aurora-cluster-demo"
  database_name       = "test"
  engine              = "aurora-mysql"
  master_username     = "tfacctest"
  master_password     = "avoid-plaintext-passwords"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "example" {
  db_cluster_identifier          = aws_rds_cluster.example.cluster_identifier
  db_cluster_snapshot_identifier = "example"
}

resource "aws_rds_cluster_snapshot_copy" "example" {
  source_db_cluster_snapshot_identifier = aws_db_cluster_snapshot.example.db_cluster_snapshot_arn
  target_db_cluster_snapshot_identifier = "example-copy"
}
```

## Argument Reference

The following arguments are required:

* `source_db_cluster_snapshot_identifier` - (Required) Identifier of the source snapshot.
* `target_db_cluster_snapshot_identifier` - (Required) Identifier for the snapshot.

The following arguments are optional:

* `copy_tags` - (Optional) Whether to copy existing tags. Defaults to `false`.
* `destination_region` - (Optional) The Destination region to place snapshot copy.
* `kms_key_id` - (Optional) KMS key ID.
* `presigned_url` - (Optional) URL that contains a Signature Version 4 signed request.
* `shared_accounts` - (Optional) List of AWS Account IDs to share the snapshot with. Use `all` to make the snapshot public.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `allocated_storage` - Specifies the allocated storage size in gigabytes (GB).
* `availability_zones` - Specifies the the Availability Zones the DB cluster was located in at the time of the DB snapshot.
* `db_cluster_snapshot_arn` - The Amazon Resource Name (ARN) for the DB cluster snapshot.
* `engine` - Specifies the name of the database engine.
* `engine_version` - Specifies the version of the database engine.
* `id` - Cluster snapshot identifier.
* `kms_key_id` - ARN for the KMS encryption key.
* `license_model` - License model information for the restored DB instance.
* `shared_accounts` - (Optional) List of AWS Account IDs to share the snapshot with. Use `all` to make the snapshot public.
* `source_db_cluster_snapshot_identifier` - DB snapshot ARN that the DB cluster snapshot was copied from. It only has value in case of cross customer or cross region copy.
* `storage_encrypted` - Specifies whether the DB cluster snapshot is encrypted.
* `storage_type` - Specifies the storage type associated with DB cluster snapshot.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `vpc_id` - Provides the VPC ID associated with the DB cluster snapshot.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_rds_cluster_snapshot_copy` using the snapshot identifier. For example:

```terraform
import {
  to = aws_rds_cluster_snapshot_copy.example
  id = "my-snapshot"
}
```

Using `terraform import`, import `aws_rds_cluster_snapshot_copy` using the `id`. For example:

```console
% terraform import aws_rds_cluster_snapshot_copy.example my-snapshot
```
