---
subcategory: "DynamoDB Accelerator (DAX)"
layout: "aws"
page_title: "AWS: aws_dax_cluster"
description: |-
  Provides an DAX Cluster resource.
---

# Resource: aws_dax_cluster

Provides a DAX Cluster resource.

## Example Usage

```terraform
resource "aws_dax_cluster" "bar" {
  cluster_name       = "cluster-example"
  iam_role_arn       = data.aws_iam_role.example.arn
  node_type          = "dax.r4.large"
  replication_factor = 1
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cluster_endpoint_encryption_type` - (Optional) The type of encryption the
cluster's endpoint should support. Valid values are: `NONE` and `TLS`.
Default value is `NONE`.
* `cluster_name` - (Required) Group identifier. DAX converts this name to
lowercase
* `iam_role_arn` - (Required) A valid Amazon Resource Name (ARN) that identifies
an IAM role. At runtime, DAX will assume this role and use the role's
permissions to access DynamoDB on your behalf
* `node_type` - (Required) The compute and memory capacity of the nodes. See
[Nodes][1] for supported node types
* `replication_factor` - (Required) The number of nodes in the DAX cluster. A
replication factor of 1 will create a single-node cluster, without any read
replicas
* `availability_zones` - (Optional) List of Availability Zones in which the
nodes will be created
* `description` - (Optional) Description for the cluster
* `notification_topic_arn` - (Optional) An Amazon Resource Name (ARN) of an
SNS topic to send DAX notifications to. Example:
`arn:aws:sns:us-east-1:012345678999:my_sns_topic`
* `parameter_group_name` - (Optional) Name of the parameter group to associate
with this DAX cluster
* `maintenance_window` - (Optional) Specifies the weekly time range for when
maintenance on the cluster is performed. The format is `ddd:hh24:mi-ddd:hh24:mi`
(24H Clock UTC). The minimum maintenance window is a 60 minute period. Example:
`sun:05:00-sun:09:00`
* `security_group_ids` - (Optional) One or more VPC security groups associated
with the cluster
* `server_side_encryption` - (Optional) Encrypt at rest options
* `subnet_group_name` - (Optional) Name of the subnet group to be used for the
cluster
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

The `server_side_encryption` object supports the following:

* `enabled` - (Optional) Whether to enable encryption at rest. Defaults to `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the DAX cluster

* `nodes` - List of node objects including `id`, `address`, `port` and
`availability_zone`. Referenceable e.g., as
`${aws_dax_cluster.test.nodes.0.address}`

* `configuration_endpoint` - The configuration endpoint for this DAX cluster,
consisting of a DNS name and a port number

* `cluster_address` - The DNS name of the DAX cluster without the port appended

* `port` - The port used by the configuration endpoint

* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `45m`)
- `update` - (Default `45m`)
- `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DAX Clusters using the `cluster_name`. For example:

```terraform
import {
  to = aws_dax_cluster.my_cluster
  id = "my_cluster"
}
```

Using `terraform import`, import DAX Clusters using the `cluster_name`. For example:

```console
% terraform import aws_dax_cluster.my_cluster my_cluster
```

[1]: http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DAX.concepts.cluster.html#DAX.concepts.nodes
