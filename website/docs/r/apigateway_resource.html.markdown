---
subcategory: "API Gateway (REST APIs)"
layout: "aws"
page_title: "AWS: aws_apigateway_resource"
description: |-
  Provides an API Gateway Resource.
---

# Resource: aws_apigateway_resource

Provides an API Gateway Resource.

## Example Usage

```terraform
resource "aws_apigateway_rest_api" "MyDemoAPI" {
  name        = "MyDemoAPI"
  description = "This is my API for demonstration purposes"
}

resource "aws_apigateway_resource" "MyDemoResource" {
  rest_api_id = aws_apigateway_rest_api.MyDemoAPI.id
  parent_id   = aws_apigateway_rest_api.MyDemoAPI.root_resource_id
  path_part   = "mydemoresource"
}
```

## Argument Reference

The following arguments are supported:

* `rest_api_id` - (Required) The ID of the associated REST API
* `parent_id` - (Required) The ID of the parent API resource
* `path_part` - (Required) The last path segment of this API resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource's identifier.
* `path` - The complete path for this API resource, including all parent paths.

## Import

`aws_apigateway_resource` can be imported using `REST-API-ID/RESOURCE-ID`, e.g.,

```
$ terraform import aws_apigateway_resource.example 12345abcde/67890fghij
```
