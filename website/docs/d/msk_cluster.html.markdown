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

See the [`aws_msk_cluster` resource](/docs/providers/aws/r/msk_cluster.html) for details on the returned attributes.
