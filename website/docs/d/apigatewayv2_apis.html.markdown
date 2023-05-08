---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_apis"
description: |-
  Provides details about multiple Amazon API Gateway Version 2 APIs.
---

# Data Source: aws_apigatewayv2_apis

Provides details about multiple Amazon API Gateway Version 2 APIs.

## Example Usage

```terraform
data "aws_apigatewayv2_apis" "example" {
  protocol_type = "HTTP"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) API name.
* `protocol_type` - (Optional) API protocol.
* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired APIs.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `ids` - Set of API identifiers.
