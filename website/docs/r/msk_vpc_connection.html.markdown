---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_vpc_connection"
description: |-
  Terraform resource for managing an AWS Managed Streaming for Kafka Vpc Connection.
---
# Resource: aws_msk_vpc_connection

Terraform resource for managing an AWS Managed Streaming for Kafka Vpc Connection.

## Example Usage

```terraform
resource "aws_msk_vpc_connection" "test" {
  authentication     = "SASL_IAM"
  target_cluster_arn = "aws_msk_cluster.arn"
  vpc_id             = aws_vpc.test.id
  client_subnets     = aws_subnet.test[*].id
  security_groups    = [aws_security_group.test.id]
}
```

## Argument Reference

The following arguments are required:

* `authentication` - (Required) The authentication type for the client VPC connection. Specify one of these auth type strings: SASL_IAM, SASL_SCRAM, or TLS.

* `client_subnets` - (Required) The list of subnets in the client VPC to connect to.

* `security_groups` - (Required) The security groups to attach to the ENIs for the broker nodes.

* `target_cluster_arn` - (Required) The Amazon Resource Name (ARN) of the cluster.

* `vpc_id` - (Required) The VPC id of the remote client.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the VPC connection.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MSK configurations using the configuration ARN. For example:

```terraform
import {
  to = aws_msk_vpc_connection.example
  id = "arn:aws:kafka:eu-west-2:123456789012:vpc-connection/123456789012/example/38173259-79cd-4ee8-87f3-682ea6023f48-2"
}
```

Using `terraform import`, import MSK configurations using the configuration ARN. For example:

```console
% terraform import aws_msk_vpc_connection.example arn:aws:kafka:eu-west-2:123456789012:vpc-connection/123456789012/example/38173259-79cd-4ee8-87f3-682ea6023f48-2
```
