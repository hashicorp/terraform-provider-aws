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

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `rest_api_id` - (Required) ID of the associated REST API
* `parent_id` - (Required) ID of the parent API resource
* `path_part` - (Required) Last path segment of this API resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Resource's identifier.
* `path` - Complete path for this API resource, including all parent paths.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_api_gateway_resource.example
  identity = {
    rest_api_id = "12345abcde"
    id          = "67890fghij"
  }
}

resource "aws_api_gateway_resource" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `rest_api_id` (String) ID of the associated REST API.
* `id` (String) Resource's identifier.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

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
