---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_integrations"
description: |-
  Provides details about integrations in an AWS API Gateway V2 API.
---

# Data Source: aws_apigatewayv2_integrations

Use this data source to get a list of integration ids of integrations in an Amazon API Gateway V2 API.

## Example Usage

```terraform
data "aws_apigatewayv2_integrations" "example" {
  api_id = aws_apigatewayv2_api.example.id
}
```

## Argument Reference

This data source supports the following arguments:

* `api_id` - (Required) ID of the associated API Gateway.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - IDs of the Integrations.
