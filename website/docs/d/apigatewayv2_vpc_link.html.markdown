---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_vpc_link"
description: |-
  Terraform data source for managing an AWS API Gateway V2 VPC Link.
---

# Data Source: aws_apigatewayv2_vpc_link

Terraform data source for managing an AWS API Gateway V2 VPC Link.

## Example Usage

### Basic Usage

```terraform
data "aws_apigatewayv2_vpc_link" "example" {
  vpc_link_id = "example"
}
```

## Argument Reference

The following arguments are required:

* `vpc_link_id` - (Required) VPC Link ID

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the VPC Link.
* `id` - VPC Link ID.
* `name` - VPC Link Name.
* `security_group_ids` - List of security groups associated with the VPC Link.
* `subnet_ids` - List of subnets attached to the VPC Link.
* `tags` - VPC Link Tags.
