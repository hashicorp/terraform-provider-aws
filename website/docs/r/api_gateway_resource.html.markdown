---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_resource"
description: |-
  Provides an API Gateway Resource.
---

# Resource: aws_api_gateway_resource

Provides an API Gateway Resource.

## Example Usage

### Basic Usage

```terraform
resource "aws_api_gateway_rest_api" "MyDemoAPI" {
  name        = "MyDemoAPI"
  description = "This is my API for demonstration purposes"
}

resource "aws_api_gateway_resource" "MyDemoResource" {
  rest_api_id = aws_api_gateway_rest_api.MyDemoAPI.id
  parent_id   = aws_api_gateway_rest_api.MyDemoAPI.root_resource_id
  path_part   = "mydemoresource"
}
```

### Full Path Usage

The `full_path` parameter allows you to declare nested API Gateway resources in a single block, eliminating boilerplate and preventing parent-ID mistakes:

```terraform
resource "aws_api_gateway_rest_api" "example" {
  name = "example-api"
}

# Create the entire path /users/123/orders/456 in one resource block
resource "aws_api_gateway_resource" "users_orders" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  full_path   = "/users/123/orders/456"
}
```

This replaces the need for multiple resource blocks:

```terraform
# Traditional approach (replaced by full_path above)
resource "aws_api_gateway_resource" "users" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  parent_id   = aws_api_gateway_rest_api.example.root_resource_id
  path_part   = "users"
}

resource "aws_api_gateway_resource" "user_id" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  parent_id   = aws_api_gateway_resource.users.id
  path_part   = "123"
}

resource "aws_api_gateway_resource" "orders" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  parent_id   = aws_api_gateway_resource.user_id.id
  path_part   = "orders"
}

resource "aws_api_gateway_resource" "order_id" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  parent_id   = aws_api_gateway_resource.orders.id
  path_part   = "456"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `rest_api_id` - (Required) ID of the associated REST API

**Note**: Either `full_path` or both `parent_id` and `path_part` must be specified.

* `full_path` - (Optional) Complete path for this API resource. Must start with `/`. Cannot be used with `parent_id` or `path_part`. When specified, all necessary parent resources will be created automatically.
* `parent_id` - (Optional) ID of the parent API resource. Required when not using `full_path`.
* `path_part` - (Optional) Last path segment of this API resource. Required when not using `full_path`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Resource's identifier.
* `path` - Complete path for this API resource, including all parent paths.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_api_gateway_resource` using `REST-API-ID/RESOURCE-ID`. For example:

```terraform
import {
  to = aws_api_gateway_resource.example
  id = "12345abcde/67890fghij"
}
```

Using `terraform import`, import `aws_api_gateway_resource` using `REST-API-ID/RESOURCE-ID`. For example:

```console
% terraform import aws_api_gateway_resource.example 12345abcde/67890fghij
```
