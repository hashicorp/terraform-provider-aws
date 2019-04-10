---
layout: "aws"
page_title: "AWS: aws_msk_cluster"
sidebar_current: "docs-aws-datasource-msk-cluster"
description: |-
    Provides information about an MSK Kafka cluster
---

# Data Source: aws_msk_cluster

Provides information about an MSK Kafka cluster.

## Example Usage

```hcl
data "aws_msk_cluster" "cluster" {
  name = "test-cluster"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the cluster.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the cluster.
* `status` - The status of the cluster. The possible states are CREATING, ACTIVE, and FAILED.
* `encryption_key` - The AWS KMS key used for data encryption.
* `zookeeper_connect` - Connection string for Zookeeper.
* `bootstrap_brokers` - A list of brokers that a client application can use to bootstrap.
