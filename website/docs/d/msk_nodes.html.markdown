---
subcategory: "Managed Streaming for Kafka (MSK)"
layout: "aws"
page_title: "AWS: aws_msk_nodes"
description: |-
  Get information on an Amazon MSK Broker Nodes
---

# Data Source: aws_msk_nodes

Get information on an Amazon MSK Broker Nodes.

## Example Usage

```terraform
data "aws_msk_nodes" "example" {
  cluster_arn = aws_msk_cluster.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `cluster_arn` - (Required) The ARN of the cluster the nodes belong to.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* [`nodes`](#Nodes) - List of MSK Broker Nodes, sorted by broker ID in ascending order.

### Nodes
* `broker_id` - Numeric ID of the broker node
* `attached_eni_id` - The ENI associated with the broker node
* `client_subnet` - The subnet name the broker resides in
* `endpoints` - List of hostnames associated with the broker node. This does not include ports.