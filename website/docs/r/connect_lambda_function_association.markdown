---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_lambda_function_association"
description: |-
  Provides details about a specific Connect Lambda Function Association.
---

# Resource: aws_connect_lambda_function_association

Allows the specified Amazon Connect instance to access the specified Lambda Funtion. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html)

[Invoke AWS Lambda functions](https://docs.aws.amazon.com/connect/latest/adminguide/connect-lambda-functions.html)


## Example Usage
### Basic

```hcl
resource "aws_connect_lambda_function_association" "test" {
  instance_id  = aws_connect_instance.test.id
  function_arn = aws_lambda_function.test_lambda.arn
}
```

### Including a sample Lambda function

```hcl
resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "test_lambda" {
  filename      = "lambda_function_payload.zip"
  function_name = "lambda_function_name"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.test"

  # The filebase64sha256() function is available in Terraform 0.11.12 and later
  # For Terraform 0.11.11 and earlier, use the base64sha256() function and the file() function:
  # source_code_hash = "${base64sha256(file("lambda_function_payload.zip"))}"
  source_code_hash = filebase64sha256("lambda_function_payload.zip")

  runtime = "nodejs12.x"

  environment {
    variables = {
      foo = "bar"
    }
  }
}

resource "aws_connect_lambda_function_association" "test" {
  instance_id  = aws_connect_instance.test.id
  function_arn = aws_lambda_function.test_lambda.arn
}
```

## Argument Reference

The following arguments are supported:

* `function_arn` - (Required) Name or Amazon Resource Name (ARN) of the Lambda Function, omitting any version or alias qualifier.
* `instance_id` - (Required) The identifier of the Amazon Connect instance. You can find the instanceId in the ARN of the instance.

## Attributes Reference

No additional attributes are exported.
