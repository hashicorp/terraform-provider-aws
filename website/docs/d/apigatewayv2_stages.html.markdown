---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_stages"
description: |-
  Provides details about stages in an Amazon API Gateway V2 API.
---

# Data Source: aws_apigatewayv2_stages

Use this data source to get a list of stages in an Amazon API Gateway V2 API.

## Example Usage

```terraform
data "aws_apigatewayv2_stages" "example" {
  api_id = aws_apigatewayv2_api.example.id
}
```

## Argument Reference

This data source supports the following arguments:

* `api_id` - (Required) ID of the associated API Gateway.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) A mapping of tags, each pair of which must exactly match a pair on the desired stages.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `names` - Names of stages in the API Gateway.
