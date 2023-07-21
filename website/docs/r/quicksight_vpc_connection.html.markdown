---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_vpc_connection"
description: |-
  Terraform resource for managing an AWS QuickSight VPC Connection.
---

# Resource: aws_quicksight_vpc_connection

Terraform resource for managing an AWS QuickSight VPC Connection.

## Example Usage

### Basic Usage

```terraform

resource "aws_iam_role" "vpc_connection_role" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = "sts:AssumeRole"
        Principal = {
          Service = "quicksight.amazonaws.com"
        }
      }
    ]
  })
  inline_policy {
    name = "QuickSightVPCConnectionRolePolicy"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Effect = "Allow"
          Action = [
            "ec2:CreateNetworkInterface",
            "ec2:ModifyNetworkInterfaceAttribute",
            "ec2:DeleteNetworkInterface",
            "ec2:DescribeSubnets",
            "ec2:DescribeSecurityGroups"
          ]
          Resource = ["*"]
        }
      ]
    })
  }
}

resource "aws_quicksight_vpc_connection" "example" {
  vpc_connection_id  = "example-connection-id"
  name               = "Example Connection"
  role_arn           = aws_iam_role.vpc_connection_role.arn
  security_group_ids = ["sg-00000000000000000"]
  subnet_ids = [
    "subnet-00000000000000000",
    "subnet-00000000000000001",
  ]
}
```

## Argument Reference

The following arguments are required:

* `vpc_connection_id` - (Required) The ID of the VPC connection.
* `name` - (Required) The display name for the VPC connection.
* `role_arn` - (Required) The IAM role to associate with the VPC connection.
* `security_group_ids` - (Required) A list of security group IDs for the VPC connection.
* `subnet_ids` - (Required) A list of subnet IDs for the VPC connection.

The following arguments are optional:

* `aws_account_id` - (Optional) AWS account ID.
* `dns_resolvers` - (Optional) A list of IP addresses of DNS resolver endpoints for the VPC connection.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the VPC connection.
* `availability_status` - The availability status of the VPC connection. Valid values are `AVAILABLE`, `UNAVAILABLE` or `PARTIALLY_AVAILABLE`.
* `id` - A comma-delimited string joining AWS account ID and VPC connection ID.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import QuickSight VPC connection using the AWS account ID and VPC connection ID separated by commas (`,`). For example:

```terraform
import {
  to = aws_quicksight_vpc_connection.example
  id = "123456789012,example"
}
```

Using `terraform import`, import QuickSight VPC connection using the AWS account ID and VPC connection ID separated by commas (`,`). For example:

```console
% terraform import aws_quicksight_vpc_connection.example 123456789012,example
```
