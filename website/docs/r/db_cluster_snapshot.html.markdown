---
layout: "aws"
page_title: "AWS: aws_db_cluster_snapshot"
sidebar_current: "docs-aws-resource-db-cluster-snapshot"
description: |-
  Provides a DB Cluster Snapshot.
---

# aws_db_cluster_snapshot

Creates a Snapshot of a DB Cluster.

## Example Usage

```hcl
resource "aws_vpc" "aurora" {
    cidr_block = "192.168.0.0/16"
    tags {
        Name = "resource_aws_db_cluster_snapshot_test"
    }
}

resource "aws_subnet" "aurora1" {
    vpc_id = "${aws_vpc.aurora.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
    tags {
        Name = "resource_aws_db_cluster_snapshot_test"
    }
}

resource "aws_subnet" "aurora2" {
    vpc_id = "${aws_vpc.aurora.id}"
    cidr_block = "192.168.16.0/20"
    availability_zone = "us-west-2b"
    tags {
        Name = "resource_aws_db_cluster_snapshot_test"
    }
}

resource "aws_db_subnet_group" "aurora" {
  subnet_ids = [
    "${aws_subnet.aurora1.id}",
    "${aws_subnet.aurora2.id}"
  ]
}

resource "aws_rds_cluster" "aurora" {
  master_username         = "foo"
  master_password         = "barbarbarbar"
  db_subnet_group_name = "${aws_db_subnet_group.aurora.name}"
  backup_retention_period = 1
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "aurora" {
  count=1
  cluster_identifier = "${aws_rds_cluster.aurora.id}"
  instance_class = "db.t2.small"
  db_subnet_group_name = "${aws_db_subnet_group.aurora.name}"
}

resource "aws_db_cluster_snapshot" "test" {
	db_cluster_identifier = "${aws_rds_cluster.aurora.id}"
	db_cluster_snapshot_identifier = "resourcetestsnapshot1234"
}
```

## Argument Reference

The following arguments are supported:

* `db_cluster_identifier` - (Required) The DB Cluster Identifier from which to take the snapshot.
* `db_cluster_snapshot_identifier` - (Required) The Identifier for the snapshot.


## Attributes Reference

The following attributes are exported:

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
