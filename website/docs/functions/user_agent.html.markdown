---
subcategory: ""
layout: "aws"
page_title: "AWS: user_agent"
description: |-
  Formats a User-Agent product for use with the user_agent argument in the provider or provider_meta block..
---
# Function: user_agent

~> Provider-defined functions are supported in Terraform 1.8 and later.

Formats a User-Agent product for use with the user_agent argument in the provider or provider_meta block..

## Example Usage

```terraform
# result: foo-bar
output "example" {
  value = provider::aws::user_agent("foo")
}
```

## Signature

```text
user_agent(arg string) string
```

## Arguments

1. `arg` (String) Example argument description.
