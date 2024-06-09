---
subcategory: ""
layout: "aws"
page_title: "AWS: arn_parse"
description: |-
  Parses an ARN into its constituent parts.
---

# Function: arn_parse

~> Provider-defined functions are supported in Terraform 1.8 and later.

Parses an ARN into its constituent parts.

See the [AWS documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference-arns.html) for additional information on Amazon Resource Names.

## Example Usage

```terraform
# result: 
# {
#   "partition": "aws",
#   "service": "iam",
#   "region": "",
#   "account_id": "444455556666",
#   "resource": "role/example",
# }
output "example" {
  value = provider::aws::arn_parse("arn:aws:iam::444455556666:role/example")
}
```

## Signature

```text
arn_parse(arn string) object
```

## Arguments

1. `arn` (String) ARN (Amazon Resource Name) to parse.
