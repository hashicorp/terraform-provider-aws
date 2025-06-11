---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_routing_rule"
description: |-
  Terraform resource for managing an AWS API Gateway V2 Routing Rule.
---

# Resource: aws_apigatewayv2_routing_rule

Terraform resource for managing an AWS API Gateway V2 Routing Rule.

## Example Usage

### Basic Usage

```terraform
resource "aws_apigatewayv2_routing_rule" "example" {
  domain_name = "test.example.com"

  conditions {
    match_headers {
      any_of {
        header     = "X-Example-Header"
        value_glob = "example-value-*"
      }
    }
    match_base_paths {
      any_of = ["example-path", "another-path"]
    }
  }
  actions {
    invoke_api {
      api_id          = "example-api-id"
      stage           = "example-stage"
      strip_base_path = true
    }
  }
  priority = 1
}
```

## Argument Reference

The following arguments are required:

* `actions` - (Required) Configuration of resulting action based on matching routing rules condition. See below.
* `conditions` - (Required) Conditions configuration. See below.
* `domain_name` - (Required) Domain name. Must be between 1 and 512 characters in length.

### `actions`

* `invoke_api` - (Required) Configuration to invoke a stage of a target API. Only REST APIs are supported. See below.

### `conditions`

* `match_base_paths` - (Optional) The base path to be matched. See below.
* `match_headers` - (Optional) The headers to be matched. See below.

### `match_base_paths`

* `any_of` - (Required) List of strings of the case sensitive base path to be matched.

### `match_headers`

* `any_of` - (Required) Configuration of the headers to be matched. There is a match if any of the header name and header value globs are matched. See below.

### `any_of`

* `header` - (Required) The case insensitive header name to be matched. The header name must be less than 40 characters and the only allowed characters are a-z, A-Z, 0-9, and the following special characters: *?-!#$%&'.^_`|~.
* `value_glob` - (Required) The case sensitive header glob value to be matched against entire header value. The header glob value must be less than 128 characters and the only allowed characters are a-z, A-Z, 0-9, and the following special characters: \*?-!#$%&'.^_`|~. Wildcard matching is supported for header glob values but must be for \*prefix-match, suffix-match*, or \*infix*-match.

### `invoke_api`

* `api_id` - (Required) Action to invoke a stage of a target API. Only REST APIs are supported.
* `stage` - (Required) Action to invoke a stage of a target API. Only REST APIs are supported.
* `strip_base_path` - (Required) Action to invoke a stage of a target API. Only REST APIs are supported.

The following arguments are optional:

* `domain_name_id` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `priority` - (Optional) The order of rule evaluation. Priority is evaluated from the lowest value to the highest value. Rules can't have the same priority. Value must be between 1 and 1,000,000.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Routing Rule. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `id` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import API Gateway V2 Routing Rule using the `example_id_arg`. For example:

```terraform
import {
  to = aws_apigatewayv2_routing_rule.example
  id = "routing_rule-id-12345678"
}
```

Using `terraform import`, import API Gateway V2 Routing Rule using the `example_id_arg`. For example:

```console
% terraform import aws_apigatewayv2_routing_rule.example routing_rule-id-12345678
```
