---
layout: "aws"
page_title: "AWS: aws_api_gateway_v2_deployment"
sidebar_current: "docs-aws-resource-api-gateway-v2-deployment"
description: |-
  Manages an Amazon API Gateway Version 2 deployment.
---

# Resource: aws_api_gateway_v2_deployment

Manages an Amazon API Gateway Version 2 deployment.
More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api.html).

-> **Note:** Creating a deployment for an API requires at least one `aws_api_gateway_v2_route` resource associated with that API.

## Example Usage

### Basic

```hcl
resource "aws_api_gateway_v2_deployment" "example" {
  api_id      = "${aws_api_gateway_v2_route.example.api_id}"
  description = "Example deployment"
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The API identifier.
* `description` - (Optional) The description for the deployment resource.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The deployment identifier.

## Import

`aws_api_gateway_v2_deployment` can be imported by using the API identifier and deployment identifier, e.g.

```
$ terraform import aws_api_gateway_v2_deployment.example aabbccddee/1122334
```
