---
layout: "aws"
page_title: "AWS: aws_dax_cluster"
sidebar_current: "docs-aws-resource-dax-cluster"
description: |-
  Provides an DAX Cluster resource.
---

# aws\_dax\_cluster

Provides a DAX Cluster resource.

## Example Usage

```hcl
resource "aws_dax_cluster" "bar" {
  cluster_id         = "cluster-example"
  iam_role_arn       = "arn:aws:iams:us-east-1:012345678999"
  node_type          = "dax.r3.large"
  replication_factor = 1
}
```

## Argument Reference

The following arguments are supported:

* `cluster_name` – (Required) Group identifier. DAX converts this name to
lowercase

* `iam_role_arn` - (Required) A valid Amazon Resource Name (ARN) that identifies
an IAM role. At runtime, DAX will assume this role and use the role's
permissions to access DynamoDB on your behalf.

* `node_type` – (Required) The compute and memory capacity of the nodes. See
[Nodes][1] for supported node types.

* `replication_factor` – (Required) The number of nodes in the DAX cluster. A
replication factor of 1 will create a single-node cluster, without any read
replicas.

* `availability_zones` - (Optional) List of Availability Zones in which the
nodes will be created

* `description` – (Optional) Description for the cluster

* `notification_topic_arn` – (Optional) An Amazon Resource Name (ARN) of an
SNS topic to send DAX notifications to. Example:
`arn:aws:sns:us-east-1:012345678999:my_sns_topic`

* `parameter_group_name` – (Optional) Name of the parameter group to associate
with this DAX cluster

* `maintenance_window` – (Optional) Specifies the weekly time range for when
maintenance on the cluster is performed. The format is `ddd:hh24:mi-ddd:hh24:mi`
(24H Clock UTC). The minimum maintenance window is a 60 minute period. Example:
`sun:05:00-sun:09:00`

* `security_group_ids` – (Optional) One or more VPC security groups associated
with the cluster

* `subnet_group_name` – (Optional) Name of the subnet group to be used for the
cluster.

* `tags` - (Optional) A mapping of tags to assign to the resource

## Attributes Reference

The following attributes are exported:

* `nodes` - List of node objects including `id`, `address`, `port` and
`availability_zone`. Referenceable e.g. as
`${aws_dax_cluster.test.nodes.0.address}`

* `configuration_endpoint` - The configuration endpoint for this DAX cluster,
consisting of a DNS name and a port number.

* `cluster_address` - The DNS name of the DAX cluster without the port appended.

* `port` - The port used by the configuration endpoint

## Import

DAX Clusters can be imported using the `cluster_id`, e.g.

```
$ terraform import aws_dax_cluster.my_cluster my_cluster
```

[1]: http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DAX.concepts.cluster.html#DAX.concepts.nodes
