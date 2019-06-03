---
subcategory: "API Gateway v2"
layout: "aws"
page_title: "AWS: aws_api_gateway_v2_model"
description: |-
  Manages an Amazon API Gateway Version 2 model.
---

# Resource: aws_api_gateway_v2_model

Manages an Amazon API Gateway Version 2 [model](https://docs.aws.amazon.com/apigateway/latest/developerguide/models-mappings.html#models-mappings-models).

## Example Usage

### Basic

```hcl
resource "aws_api_gateway_v2_model" "example" {
  api_id       = "${aws_api_gateway_v2_api.example.id}"
  content_type = "application/json"
  name         = "example"

  schema = <<EOF
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "ExampleModel",
  "type": "object",
  "properties": {
    "id": { "type": "string" }
  }
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The API identifier.
* `content_type` - (Required)  The content-type for the model, for example, `application/json`.
* `name` - (Required) The name of the model. Must be alphanumeric.
* `schema` - (Required) The schema for the model. This should be a [JSON schema draft 4](https://tools.ietf.org/html/draft-zyp-json-schema-04) model.
* `description` - (Optional) The description of the model.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The model identifier.

## Import

`aws_api_gateway_v2_model` can be imported by using the API identifier and model identifier, e.g.

```
$ terraform import aws_api_gateway_v2_model.example aabbccddee/1122334
```
