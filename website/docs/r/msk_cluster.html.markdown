---
layout: "aws"
page_title: "AWS: aws_msk_cluster"
sidebar_current: "docs-aws-resource-msk-cluster"
description: |-
  Provides a Kafka Cluster
---

# aws_msk_cluster

Provides an MSK Kafka cluster resource.

## Example Usage

```hcl
resource "aws_msk_cluster" "test_cluster" {
	name = "test-cluster"
	broker_count = 3
	broker_instance_type = "kafka.m5.large"
	broker_volume_size = 10
	client_subnets = ["${aws_subnet.test_subnet_a.id}", "${aws_subnet.test_subnet_b.id}", "${aws_subnet.test_subnet_c.id}"]
}
```

## Argument Reference

* `name` – (Required) Name of the cluster.
* `client_subnets` - (Required) List of subnets to deploy cluster in.
* `kafka_version` - (Required) Version of Kafka to use for cluster.
* `broker_count` - (Required) Number of broker nodes you want to create in each Availability Zone.
* `broker_instance_type` - (Required) Instance type for brokers from the m5 family. e.g. kafka.m5.large
* `broker_volume_size` - (Required) The size of the drive in GiBs.
* `broker_security_groups` - (Required) Security groups to attach to broker nodes.
* `encryption_key` - (Optional) The AWS KMS key used for data encryption.
* `enhanced_monitoring` - (Optional) Level of monitoring for the cluster. Possible values are DEFAULT, PER_BROKER, and PER_TOPIC_PER_BROKER.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the cluster.
* `status` - The status of the cluster. The possible states are CREATING, ACTIVE, and FAILED.
* `zookeeper_connect` - Connection string for Zookeeper.
* `bootstrap_brokers` - A list of brokers that a client application can use to bootstrap.

## Timeouts

`aws_msk_cluster` provides the following [Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

* `create` - (Default `60m`) How long to wait for cluster creation.
* `delete` - (Default `120m`) How long to wait for cluster deletion.

## Import

Clusters can be imported using the  cluster `name`, e.g.

```
$ terraform import aws_msk_cluster.my_cluster my_cluster
```
