---
subcategory: "Neptune"
layout: "aws"
page_title: "AWS: aws_neptune_cluster_snapshot"
description: |-
  Manages a Neptune database cluster snapshot.
---

# Resource: aws_neptune_cluster_snapshot

Manages a Neptune database cluster snapshot.

## Example Usage

```terraform
resource "aws_neptune_cluster_snapshot" "example" {
  db_cluster_identifier          = aws_neptune_cluster.example.id
  db_cluster_snapshot_identifier = "resourcetestsnapshot1234"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `db_cluster_identifier` - (Required) The DB Cluster Identifier from which to take the snapshot.
* `db_cluster_snapshot_identifier` - (Required) The Identifier for the snapshot.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `allocated_storage` - Specifies the allocated storage size in gigabytes (GB).
* `availability_zones` - List of EC2 Availability Zones that instances in the DB cluster snapshot can be restored in.
* `db_cluster_snapshot_arn` - The Amazon Resource Name (ARN) for the DB Cluster Snapshot.
* `engine` - Specifies the name of the database engine.
* `engine_version` - Version of the database engine for this DB cluster snapshot.
* `kms_key_id` - If storage_encrypted is true, the AWS KMS key identifier for the encrypted DB cluster snapshot.
* `license_model` - License model information for the restored DB cluster.
* `port` - Port that the DB cluster was listening on at the time of the snapshot.
* `source_db_cluster_snapshot_identifier` - The DB Cluster Snapshot Arn that the DB Cluster Snapshot was copied from. It only has value in case of cross customer or cross region copy.
* `storage_encrypted` - Specifies whether the DB cluster snapshot is encrypted.
* `status` - The status of this DB Cluster Snapshot.
* `vpc_id` - The VPC ID associated with the DB cluster snapshot.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_neptune_cluster_snapshot` using the cluster snapshot identifier. For example:

```terraform
import {
  to = aws_neptune_cluster_snapshot.example
  id = "my-cluster-snapshot"
}
```

Using `terraform import`, import `aws_neptune_cluster_snapshot` using the cluster snapshot identifier. For example:

```console
% terraform import aws_neptune_cluster_snapshot.example my-cluster-snapshot
```
