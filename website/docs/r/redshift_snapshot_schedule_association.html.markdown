---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_snapshot_schedule_association"
description: |-
  Provides an Association Redshift Cluster and Snapshot Schedule resource.
---

# Resource: aws_redshift_snapshot_schedule_association

## Example Usage

```terraform
resource "aws_redshift_cluster" "default" {
  cluster_identifier = "tf-redshift-cluster"
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "Mustbe8characters"
  node_type          = "dc1.large"
  cluster_type       = "single-node"
}

resource "aws_redshift_snapshot_schedule" "default" {
  identifier = "tf-redshift-snapshot-schedule"
  definitions = [
    "rate(12 hours)",
  ]
}

resource "aws_redshift_snapshot_schedule_association" "default" {
  cluster_identifier  = aws_redshift_cluster.default.id
  schedule_identifier = aws_redshift_snapshot_schedule.default.id
}
```

## Argument Reference

This resource supports the following arguments:

* `cluster_identifier` - (Required, Forces new resource) The cluster identifier.
* `schedule_identifier` - (Required, Forces new resource) The snapshot schedule identifier.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Snapshot Schedule Association using the `<cluster-identifier>/<schedule-identifier>`. For example:

```terraform
import {
  to = aws_redshift_snapshot_schedule_association.default
  id = "tf-redshift-cluster/tf-redshift-snapshot-schedule"
}
```

Using `terraform import`, import Redshift Snapshot Schedule Association using the `<cluster-identifier>/<schedule-identifier>`. For example:

```console
% terraform import aws_redshift_snapshot_schedule_association.default tf-redshift-cluster/tf-redshift-snapshot-schedule
```
