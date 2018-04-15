---
layout: "aws"
page_title: "AWS: aws_lambda_invoke"
sidebar_current: "docs-aws-datasource-lambda-invoke"
description: |-
  Invoke AWS Lambda Function as data source
---

# Data Source: aws_lambda_invoke

Use this data source to invoke custom lambda functions as data source.

~> **NOTE**: The `input` argument is JSON encoded and passed as payload to the
lambda function. All values in `input` map are converted to strings.
The lambda function is invoked with
[RequestResponse](https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html#API_Invoke_RequestSyntax)
invocation type. Response of lambda must be map of primitive types (string, bool or float).

## Example Usage

```hcl
data "aws_lambda_invoke" "example" {
  function_name = "${aws_lambda_function.lambda_function_test.function_name}"

  input {
    param1 = "value1"
    param2 = "value2"
  }
}

output "result" {
  value = "${data.aws_lambda_invoke.result["key1"]}"
}
```

## Argument Reference

 * `function_name` - (Required) The name of the lambda function.
 * `input` - (Required) Map of string values to send as payload to the lambda function.
 * `qualifier` - (Optional) The qualifier (a.k.a version) of the lambda function. Defaults
 to `$LATEST`.

## Attributes Reference

 * `result` - A map of string values returned from the lambda invocation.
