---
subcategory: ""
layout: "aws"
page_title: "AWS: user_agent"
description: |-
  Formats a User-Agent product for use with the `user_agent` argument in the `provider` or `provider_meta` block.
---
# Function: user_agent

Formats a User-Agent product for use with the `user_agent` argument in the `provider` block.

-> Functions cannot be used in the [`terraform` block](https://developer.hashicorp.com/terraform/language/block/terraform#terraform-block), meaning this utility cannot be used with the [`provider_meta`](https://developer.hashicorp.com/terraform/language/block/terraform#provider_meta) `user_agent` argument.

## Example Usage

```terraform
# result: "example-module/0.0.1 (example comment)"
output "example" {
  value = provider::aws::user_agent("example-module", "0.0.1", "example comment")
}
```

## Signature

```text
user_agent(product_name string, product_version string, comment string) string
```

## Arguments

1. `product_name` (String) Product name.
1. `product_version` (String) Product version.
1. `comment` (String) Comment describing any additional product details.
