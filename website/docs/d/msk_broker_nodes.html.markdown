---
subcategory: "Managed Streaming for Kafka"
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

This data source supports the following arguments:

* `cluster_arn` - (Required) ARN of the cluster the nodes belong to.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* [`node_info_list`](#nodes) - List of MSK Broker Nodes, sorted by broker ID in ascending order.

### Nodes

* `attached_eni_id` - Attached elastic network interface of the broker
* `broker_id` - ID of the broker
* `client_subnet` - Client subnet to which this broker node belongs
* `client_vpc_ip_address` - The client virtual private cloud (VPC) IP address
* `endpoints` - Set of endpoints for accessing the broker. This does not include ports
* `node_arn` - ARN of the node
