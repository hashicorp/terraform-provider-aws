---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_usage_plan_key"
description: |-
  Provides an API Gateway Usage Plan Key.
---

# Resource: aws_api_gateway_usage_plan_key

Provides an API Gateway Usage Plan Key.

## Example Usage

```terraform
resource "aws_api_gateway_rest_api" "test" {
  name = "MyDemoAPI"
}

# ...

resource "aws_api_gateway_usage_plan" "myusageplan" {
  name = "my_usage_plan"

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.foo.stage_name
  }
}

resource "aws_api_gateway_api_key" "mykey" {
  name = "my_key"
}

resource "aws_api_gateway_usage_plan_key" "main" {
  key_id        = aws_api_gateway_api_key.mykey.id
  key_type      = "API_KEY"
  usage_plan_id = aws_api_gateway_usage_plan.myusageplan.id
}
```

## Argument Reference

The following arguments are supported:

* `key_id` - (Required) Identifier of the API key resource.
* `key_type` - (Required) Type of the API key resource. Currently, the valid key type is API_KEY.
* `usage_plan_id` - (Required) Id of the usage plan resource representing to associate the key to.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of a usage plan key.
* `key_id` - Identifier of the API gateway key resource.
* `key_type` - Type of a usage plan key. Currently, the valid key type is API_KEY.
* `usage_plan_id` - ID of the API resource
* `name` - Name of a usage plan key.
* `value` - Value of a usage plan key.

## Import

AWS API Gateway Usage Plan Key can be imported using the `USAGE-PLAN-ID/USAGE-PLAN-KEY-ID`, e.g.,

```sh
$ terraform import aws_api_gateway_usage_plan_key.key 12345abcde/zzz
```
