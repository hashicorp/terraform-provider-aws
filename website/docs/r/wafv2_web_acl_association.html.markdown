---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_web_acl_association"
description: |-
  Creates a WAFv2 Web ACL Association.
---

# Resource: aws_wafv2_web_acl_association

Creates a WAFv2 Web ACL Association.

~> **NOTE on associating a WAFv2 Web ACL with a Cloudfront distribution:** Do not use this resource to associate a WAFv2 Web ACL with a Cloudfront Distribution. The [AWS API call backing this resource][1] notes that you should use the [`web_acl_id`][2] property on the [`cloudfront_distribution`][2] instead.

[1]: https://docs.aws.amazon.com/waf/latest/APIReference/API_AssociateWebACL.html
[2]: /docs/providers/aws/r/cloudfront_distribution.html#web_acl_id

## Example Usage

```terraform
resource "aws_api_gateway_rest_api" "example" {
  body = jsonencode({
    openapi = "3.0.1"
    info = {
      title   = "example"
      version = "1.0"
    }
    paths = {
      "/path1" = {
        get = {
          "x-amazon-apigateway-integration" = {
            httpMethod           = "GET"
            payloadFormatVersion = "1.0"
            type                 = "HTTP_PROXY"
            uri                  = "https://ip-ranges.amazonaws.com/ip-ranges.json"
          }
        }
      }
    }
  })

  name = "example"
}

resource "aws_api_gateway_deployment" "example" {
  rest_api_id = aws_api_gateway_rest_api.example.id

  triggers = {
    redeployment = sha1(jsonencode(aws_api_gateway_rest_api.example.body))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "example" {
  deployment_id = aws_api_gateway_deployment.example.id
  rest_api_id   = aws_api_gateway_rest_api.example.id
  stage_name    = "example"
}

resource "aws_wafv2_web_acl" "example" {
  name  = "web-acl-association-example"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}

resource "aws_wafv2_web_acl_association" "example" {
  resource_arn = aws_api_gateway_stage.example.arn
  web_acl_arn  = aws_wafv2_web_acl.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `resource_arn` - (Required) The Amazon Resource Name (ARN) of the resource to associate with the web ACL. This must be an ARN of an Application Load Balancer, an Amazon API Gateway stage (REST only, HTTP is unsupported), an Amazon Cognito User Pool, an Amazon AppSync GraphQL API, an Amazon App Runner service, or an Amazon Verified Access instance.
* `web_acl_arn` - (Required) The Amazon Resource Name (ARN) of the Web ACL that you want to associate with the resource.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAFv2 Web ACL Association using `WEB_ACL_ARN,RESOURCE_ARN`. For example:

```terraform
import {
  to = aws_wafv2_web_acl_association.example
  id = "arn:aws:wafv2:...7ce849ea,arn:aws:apigateway:...ages/name"
}
```

Using `terraform import`, import WAFv2 Web ACL Association using `WEB_ACL_ARN,RESOURCE_ARN`. For example:

```console
% terraform import aws_wafv2_web_acl_association.example arn:aws:wafv2:...7ce849ea,arn:aws:apigateway:...ages/name
```
