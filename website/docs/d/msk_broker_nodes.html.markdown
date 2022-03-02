---
subcategory: "Managed Streaming for Kafka (MSK)"
layout: "aws"
page_title: "AWS: aws_msk_broker_nodes"
description: |-
  Get information on an Amazon MSK Broker Nodes
---

# Data Source: aws_msk_broker_nodes

Get information on an Amazon MSK Broker Nodes.

## Example Usage

```terraform
data "aws_msk_broker_nodes" "example" {
  cluster_arn = aws_msk_cluster.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `cluster_arn` - (Required) The ARN of the cluster the nodes belong to.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* [`node_info_list`](#Nodes) - List of MSK Broker Nodes, sorted by broker ID in ascending order.

### Nodes

* `attached_eni_id` - The attached elastic network interface of the broker
* `broker_id` - The ID of the broker
* `client_subnet` - The client subnet to which this broker node belongs
* `client_vpc_ip_address` - The client virtual private cloud (VPC) IP address
* `endpoints` - Set of endpoints for accessing the broker. This does not include ports
* `node_arn` - The Amazon Resource Name (ARN) of the node
