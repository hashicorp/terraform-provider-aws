---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_permission"
description: |-
  Lists Lambda permissions.
---

# List Resource: aws_lambda_permission

Lists Lambda permissions for a function.

## Example Usage

```terraform
list "aws_lambda_permission" "example" {
  provider = aws
  config {
    function_name = aws_lambda_function.example.function_name
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `function_name` - (Required) Name or ARN of the Lambda function.
* `qualifier` - (Optional) Function version or alias name.
* `region` - (Optional) Region to query. Defaults to provider region.
