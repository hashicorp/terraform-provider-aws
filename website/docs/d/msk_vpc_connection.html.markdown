---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_vpc_connection"
description: |-
  Get information on an Amazon MSK VPC Connection.
---
# Data Source: aws_msk_vpc_connection

Get information on an Amazon MSK VPC Connection.

## Example Usage

```terraform
data "aws_msk_vpc_connection" "example" {
  arn = aws_msk_vpc_connection.example.arn
}
```

## Argument Reference

This data source supports the following arguments:

* `arn` - (Required) ARN of the VPC Connection.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `authentication` - The authentication type for the client VPC Connection.
* `client_subnets` - The list of subnets in the client VPC.
* `security_groups` - The security groups attached to the ENIs for the broker nodes.
* `tags` - Map of key-value pairs assigned to the VPC Connection.
* `target_cluster_arn` - The Amazon Resource Name (ARN) of the cluster.
* `vpc_id` - The VPC ID of the remote client.
