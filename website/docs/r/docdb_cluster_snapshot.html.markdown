---
subcategory: "DocumentDB"
layout: "aws"
page_title: "AWS: aws_docdb_cluster_snapshot"
description: |-
  Manages a DocumentDB database cluster snapshot.
---

# Resource: aws_docdb_cluster_snapshot

Manages a DocumentDB database cluster snapshot for DocumentDB clusters.

## Example Usage

```terraform
resource "aws_docdb_cluster_snapshot" "example" {
  db_cluster_identifier          = aws_docdb_cluster.example.id
  db_cluster_snapshot_identifier = "resourcetestsnapshot1234"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `db_cluster_identifier` - (Required) The DocumentDB Cluster Identifier from which to take the snapshot.
* `db_cluster_snapshot_identifier` - (Required) The Identifier for the snapshot.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `availability_zones` - List of EC2 Availability Zones that instances in the DocumentDB cluster snapshot can be restored in.
* `db_cluster_snapshot_arn` - The Amazon Resource Name (ARN) for the DocumentDB Cluster Snapshot.
* `engine` - Specifies the name of the database engine.
* `engine_version` - Version of the database engine for this DocumentDB cluster snapshot.
* `kms_key_id` - If storage_encrypted is true, the AWS KMS key identifier for the encrypted DocumentDB cluster snapshot.
* `port` - Port that the DocumentDB cluster was listening on at the time of the snapshot.
* `source_db_cluster_snapshot_identifier` - The DocumentDB Cluster Snapshot Arn that the DocumentDB Cluster Snapshot was copied from. It only has value in case of cross customer or cross region copy.
* `storage_encrypted` - Specifies whether the DocumentDB cluster snapshot is encrypted.
* `status` - The status of this DocumentDB Cluster Snapshot.
* `vpc_id` - The VPC ID associated with the DocumentDB cluster snapshot.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_docdb_cluster_snapshot` using the cluster snapshot identifier. For example:

```terraform
import {
  to = aws_docdb_cluster_snapshot.example
  id = "my-cluster-snapshot"
}
```

Using `terraform import`, import `aws_docdb_cluster_snapshot` using the cluster snapshot identifier. For example:

```console
% terraform import aws_docdb_cluster_snapshot.example my-cluster-snapshot
```
