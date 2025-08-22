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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Optional) API name.
* `protocol_type` - (Optional) API protocol.
* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired APIs.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - Set of API identifiers.
