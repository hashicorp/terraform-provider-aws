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
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `authentication` - Authentication type for the client VPC Connection.
* `client_subnets` - List of subnets in the client VPC.
* `security_groups` - Security groups attached to the ENIs for the broker nodes.
* `tags` - Map of key-value pairs assigned to the VPC Connection.
* `target_cluster_arn` - ARN of the cluster.
* `vpc_id` - VPC ID of the remote client.
