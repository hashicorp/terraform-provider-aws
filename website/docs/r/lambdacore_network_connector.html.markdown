---
subcategory: "Lambda Core"
layout: "aws"
page_title: "AWS: aws_lambdacore_network_connector"
description: |-
  Manages an AWS Lambda Network Connector for routing MicroVM egress traffic through a VPC.
---

# Resource: aws_lambdacore_network_connector

Manages an AWS Lambda Network Connector. A network connector provisions elastic network interfaces (ENIs) in the subnets you specify, routing outbound traffic from [Lambda MicroVMs](https://docs.aws.amazon.com/lambda/latest/dg/microvms-networking.html) through your VPC — for example to reach private resources, or to give MicroVM traffic a stable source IP by exiting through your NAT gateway.

## Example Usage

### Basic Usage

```terraform
resource "aws_lambdacore_network_connector" "example" {
  name          = "example"
  operator_role = aws_iam_role.example.arn

  configuration {
    vpc_egress_configuration {
      subnet_ids         = aws_subnet.example[*].id
      security_group_ids = [aws_security_group.example.id]
    }
  }
}

resource "aws_iam_role" "example" {
  name = "example-network-connector-operator"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "network-connectors.lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "example" {
  name = "example-network-connector-operator"
  role = aws_iam_role.example.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "CreateENI"
        Effect = "Allow"
        Action = "ec2:CreateNetworkInterface"
        Resource = [
          "arn:aws:ec2:*:*:network-interface/*",
          "arn:aws:ec2:*:*:subnet/*",
          "arn:aws:ec2:*:*:security-group/*",
        ]
      },
      {
        Sid      = "TagENI"
        Effect   = "Allow"
        Action   = "ec2:CreateTags"
        Resource = "arn:aws:ec2:*:*:network-interface/*"
        Condition = {
          StringEquals = {
            "ec2:ManagedResourceOperator" = "network-connectors.lambda.amazonaws.com"
          }
        }
      },
    ]
  })
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the network connector, unique within the account and Region. Changing this forces a new resource.
* `configuration` - (Required) Network configuration of the connector. See [`configuration` Block](#configuration-block) below.

The following arguments are optional:

* `operator_role` - (Optional) ARN of the IAM role that the network connector service assumes to manage elastic network interfaces in your VPC. The role must trust `network-connectors.lambda.amazonaws.com` and allow `ec2:CreateNetworkInterface`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region).

### `configuration` Block

The `configuration` block supports the following:

* `vpc_egress_configuration` - (Required) Configuration for routing egress traffic through a VPC. See [`vpc_egress_configuration` Block](#vpc_egress_configuration-block) below.

### `vpc_egress_configuration` Block

The `vpc_egress_configuration` block supports the following:

* `associated_compute_resource_types` - (Optional) Compute resource types that may use this connector. Valid values: `MicroVm`.
* `network_protocol` - (Optional) Network protocol. Valid values: `IPv4`, `DualStack`.
* `security_group_ids` - (Required) Set of security group IDs applied to the connector's ENIs.
* `subnet_ids` - (Required) Set of subnet IDs where the connector provisions its ENIs.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the network connector.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_lambdacore_network_connector.example
  identity = {
    "arn" = "arn:aws:lambda:us-east-1:123456789012:network-connector:example"
  }
}

resource "aws_lambdacore_network_connector" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) ARN of the network connector.

In Terraform v1.12.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute. For example:

```terraform
import {
  to = aws_lambdacore_network_connector.example

  identity = {
    arn = "arn:aws:lambda:us-east-1:123456789012:network-connector:example"
  }
}
```

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda Network Connectors using the `arn`. For example:

```terraform
import {
  to = aws_lambdacore_network_connector.example
  id = "arn:aws:lambda:us-east-1:123456789012:network-connector:example"
}
```

Using `terraform import`, import Lambda Network Connectors using the `arn`. For example:

```console
% terraform import aws_lambdacore_network_connector.example arn:aws:lambda:us-east-1:123456789012:network-connector:example
```
