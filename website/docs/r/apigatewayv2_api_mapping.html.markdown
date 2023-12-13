---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_api_mapping"
description: |-
  Manages an Amazon API Gateway Version 2 API mapping.
---

# Resource: aws_apigatewayv2_api_mapping

Manages an Amazon API Gateway Version 2 API mapping.
More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-custom-domains.html).

## Example Usage

### Basic

```terraform
resource "aws_apigatewayv2_api_mapping" "example" {
  api_id      = aws_apigatewayv2_api.example.id
  domain_name = aws_apigatewayv2_domain_name.example.id
  stage       = aws_apigatewayv2_stage.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `api_id` - (Required) API identifier.
* `domain_name` - (Required) Domain name. Use the [`aws_apigatewayv2_domain_name`](/docs/providers/aws/r/apigatewayv2_domain_name.html) resource to configure a domain name.
* `stage` - (Required) API stage. Use the [`aws_apigatewayv2_stage`](/docs/providers/aws/r/apigatewayv2_stage.html) resource to configure an API stage.
* `api_mapping_key` - (Optional) The API mapping key. Refer to [REST API](https://docs.aws.amazon.com/apigateway/latest/developerguide/rest-api-mappings.html), [HTTP API](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-mappings.html) or [WebSocket API](https://docs.aws.amazon.com/apigateway/latest/developerguide/websocket-api-mappings.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - API mapping identifier.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_apigatewayv2_api_mapping` using the API mapping identifier and domain name. For example:

```terraform
import {
  to = aws_apigatewayv2_api_mapping.example
  id = "1122334/ws-api.example.com"
}
```

Using `terraform import`, import `aws_apigatewayv2_api_mapping` using the API mapping identifier and domain name. For example:

```console
% terraform import aws_apigatewayv2_api_mapping.example 1122334/ws-api.example.com
```
