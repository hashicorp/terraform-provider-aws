---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_alias"
description: |-
  Lists Lambda Alias resources.
---

# List Resource: aws_lambda_alias

Lists Lambda Alias resources.

## Example Usage

```terraform
list "aws_lambda_alias" "example" {
  provider = aws
  config {
    function_name = "example-function"
  }
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name or ARN of the Lambda function.

The following arguments are optional:

* `region` - (Optional) Region to query. Defaults to provider region.
