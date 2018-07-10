---
layout: "aws"
page_title: "AWS: aws_redshift_snapshot_copy_grant"
sidebar_current: "docs-aws-resource-redshift-snapshot-copy-grant"
description: |-
  Creates a snapshot copy grant that allows AWS Redshift to encrypt copied snapshots with a customer master key from AWS KMS in a destination region.
---

# aws_redshift_snapshot_copy_grant

Creates a snapshot copy grant that allows AWS Redshift to encrypt copied snapshots with a customer master key from AWS KMS in a destination region.

Note that the grant must exist in the destination region, and not in the region of the cluster.

## Example Usage

```hcl
resource "aws_redshift_snapshot_copy_grant" "test" {
    snapshot_copy_grant_name = "my-grant"
}

resource "aws_redshift_parameter_group" "test" {
  name = "main"
  family = "redshift-1.0"

  parameter {
    name = "require_ssl"
    value = true
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  enable_dns_hostnames = true
  tags {
    Name = "test"
  }
}

variable "test_subnet_count" {
  default = 2
}

resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.test.id}"
  count = "${var.test_subnet_count}"
  cidr_block = "${cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)}"
  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"

  tags {
    Name = "${data.aws_availability_zones.available.names[count.index]}-private-test"
  }
}

resource "aws_redshift_subnet_group" "test" {
  name = "test"
  subnet_ids = ["${aws_subnet.test.*.id}"]

  tags = {
    Name = "test-subnet-group"
  }
}

resource "aws_security_group" "test" {
  name = "test"

  description = "Managed by Terraform"
  vpc_id = "${aws_vpc.test.id}"

  # only postgres in
  ingress {
    from_port = 5432
    to_port = 5432
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # allow all outbound traffic
  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier = "test"
  database_name = "test"
  node_type = "dc1.large"
  cluster_type = "single-node"
  master_username = "test"
  master_password = "Test12345Test"
  vpc_security_group_ids = ["${aws_security_group.test.id}"]
  cluster_subnet_group_name = "${aws_redshift_subnet_group.test.id}"
  cluster_parameter_group_name = "${aws_redshift_parameter_group.test.name}"
  automated_snapshot_retention_period = 3 # days
  port = 5432
  allow_version_upgrade = true
  publicly_accessible = false
  encrypted = true
  snapshot_copy {
    destination_region = "us-east-2"
	grant_name = "${aws_redshift_snapshot_copy_grant.test.snapshot_copy_grant_name}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `snapshot_copy_grant_name` - (Required, Forces new resource) A friendly name for identifying the grant.
* `kms_key_id` - (Optional, Forces new resource) The unique identifier for the customer master key (CMK) that the grant applies to. Specify the key ID or the Amazon Resource Name (ARN) of the CMK. To specify a CMK in a different AWS account, you must use the key ARN. If not specified, the default key is used.

## Attributes Reference

No additional attributes beyond the arguments above are exported.
