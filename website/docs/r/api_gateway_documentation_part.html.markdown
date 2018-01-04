---
layout: "aws"
page_title: "AWS: aws_api_gateway_documentation_part"
sidebar_current: "docs-aws-resource-api-gateway-documentation-part"
description: |-
  Provides a settings of an API Gateway Documentation Part.
---

# aws_api_gateway_documentation_part

Provides a settings of an API Gateway Documentation Part.

## Example Usage

```hcl
resource "aws_api_gateway_documentation_part" "example" {
  location {
    type = "METHOD"
    method = "GET"
    path = "/example"
  }
  properties = "{\"description\":\"Example description\"}"
  rest_api_id = "${aws_api_gateway_rest_api.example.id}"
}

resource "aws_api_gateway_rest_api" "example" {
  name = "example_api"
}
```

## Argument Reference

The following argument is supported:

* `location` - (Required) The location of the targeted API entity of the to-be-created documentation part. See below.
* `properties` - (Required) A content map of API-specific key-value pairs describing the targeted API entity. The map must be encoded as a JSON string, e.g., "{ \"description\": \"The API does ...\" }". Only Swagger-compliant key-value pairs can be exported and, hence, published.
* `rest_api_id` - (Required) The ID of the associated Rest API

### Nested fields

#### `location`

* `method` - (Optional) The HTTP verb of a method. It is a valid field for the API entity types of `METHOD`, `PATH_PARAMETER`, `QUERY_PARAMETER`, `REQUEST_HEADER`, `REQUEST_BODY`, `RESPONSE`, `RESPONSE_HEADER`, and `RESPONSE_BODY`. The default value is `*` for any method.
* `name` - (Optional) The name of the targeted API entity. It is a valid and required field for the API entity types of `AUTHORIZER`, `MODEL`, `PATH_PARAMETER`, `QUERY_PARAMETER`, `REQUEST_HEADER`, `REQUEST_BODY` and `RESPONSE_HEADER`.
* `path` - (Optional) The URL path of the target. It is a valid field for the API entity types of `RESOURCE`, `METHOD`, `PATH_PARAMETER`, `QUERY_PARAMETER`, `REQUEST_HEADER`, `REQUEST_BODY`, `RESPONSE`, `RESPONSE_HEADER`, and `RESPONSE_BODY`. The default value is `/` for the root resource.
* `status_code` - (Optional) The HTTP status code of a response. It is a valid field for the API entity types of `RESPONSE`, `RESPONSE_HEADER`, and `RESPONSE_BODY`. The default value is `*` for any status code.
* `type` - (Required) The type of API entity to which the documentation content applies. e.g. `API`, `METHOD` or `REQUEST_BODY`

## Attribute Reference

The following attribute is exported in addition to the arguments listed above:

* `id` - The unique ID of the Documentation Part

## Import

API Gateway documentation_parts can be imported using the word `api-gateway-documentation_part`, e.g.

```
$ terraform import aws_api_gateway_documentation_part.example 3oyy3t
```