---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function_event_invoke_config"
description: |-
  Manages an AWS Lambda Function Event Invoke Config.
---

# Resource: aws_lambda_function_event_invoke_config

Manages an AWS Lambda Function Event Invoke Config. Use this resource to configure error handling and destinations for asynchronous Lambda function invocations.

More information about asynchronous invocations and the configurable values can be found in the [Lambda Developer Guide](https://docs.aws.amazon.com/lambda/latest/dg/invocation-async.html).

## Example Usage

### Complete Error Handling and Destinations

~> **Note:** Ensure the Lambda Function IAM Role has necessary permissions for the destination, such as `sqs:SendMessage` or `sns:Publish`, otherwise the API will return a generic `InvalidParameterValueException: The destination ARN arn:PARTITION:SERVICE:REGION:ACCOUNT:RESOURCE is invalid.` error.

```terraform
# SQS queue for failed invocations
resource "aws_sqs_queue" "dlq" {
  name = "lambda-dlq"

  tags = {
    Environment = "production"
    Purpose     = "lambda-error-handling"
  }
}

# SNS topic for successful invocations
resource "aws_sns_topic" "success" {
  name = "lambda-success-notifications"

  tags = {
    Environment = "production"
    Purpose     = "lambda-success-notifications"
  }
}

# Complete event invoke configuration
resource "aws_lambda_function_event_invoke_config" "example" {
  function_name                = aws_lambda_function.example.function_name
  maximum_event_age_in_seconds = 300 # 5 minutes
  maximum_retry_attempts       = 1   # Retry once on failure

  destination_config {
    on_failure {
      destination = aws_sqs_queue.dlq.arn
    }

    on_success {
      destination = aws_sns_topic.success.arn
    }
  }
}
```

### Error Handling Only

```terraform
resource "aws_lambda_function_event_invoke_config" "example" {
  function_name                = aws_lambda_function.example.function_name
  maximum_event_age_in_seconds = 60 # 1 minute - fail fast
  maximum_retry_attempts       = 0  # No retries
}
```

### Configuration for Lambda Alias

```terraform
resource "aws_lambda_alias" "example" {
  name             = "production"
  description      = "Production alias"
  function_name    = aws_lambda_function.example.function_name
  function_version = aws_lambda_function.example.version
}

resource "aws_lambda_function_event_invoke_config" "example" {
  function_name = aws_lambda_function.example.function_name
  qualifier     = aws_lambda_alias.example.name

  maximum_event_age_in_seconds = 1800 # 30 minutes for production
  maximum_retry_attempts       = 2    # Default retry behavior

  destination_config {
    on_failure {
      destination = aws_sqs_queue.production_dlq.arn
    }
  }
}
```

### Configuration for Published Version

```terraform
resource "aws_lambda_function_event_invoke_config" "example" {
  function_name = aws_lambda_function.example.function_name
  qualifier     = aws_lambda_function.example.version

  # Conservative settings for specific version
  maximum_event_age_in_seconds = 21600 # 6 hours maximum
  maximum_retry_attempts       = 2

  destination_config {
    on_failure {
      destination = aws_sqs_queue.version_dlq.arn
    }

    on_success {
      destination = aws_sns_topic.version_success.arn
    }
  }
}
```

### Configuration for Latest Version

```terraform
resource "aws_lambda_function_event_invoke_config" "example" {
  function_name = aws_lambda_function.example.function_name
  qualifier     = "$LATEST"

  # Development settings - fail fast
  maximum_event_age_in_seconds = 120 # 2 minutes
  maximum_retry_attempts       = 0   # No retries in development

  destination_config {
    on_failure {
      destination = aws_sqs_queue.dev_dlq.arn
    }
  }
}
```

### Multiple Destination Types

```terraform
# S3 bucket for archiving successful events
resource "aws_s3_bucket" "lambda_success_archive" {
  bucket = "lambda-success-archive-${random_id.bucket_suffix.hex}"
}

# EventBridge custom bus for failed events
resource "aws_cloudwatch_event_bus" "lambda_failures" {
  name = "lambda-failure-events"
}

resource "aws_lambda_function_event_invoke_config" "example" {
  function_name = aws_lambda_function.example.function_name

  destination_config {
    on_failure {
      destination = aws_cloudwatch_event_bus.lambda_failures.arn
    }

    on_success {
      destination = aws_s3_bucket.lambda_success_archive.arn
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name or ARN of the Lambda Function, omitting any version or alias qualifier.

The following arguments are optional:

* `destination_config` - (Optional) Configuration block with destination configuration. [See below](#destination_config-configuration-block).
* `maximum_event_age_in_seconds` - (Optional) Maximum age of a request that Lambda sends to a function for processing in seconds. Valid values between 60 and 21600.
* `maximum_retry_attempts` - (Optional) Maximum number of times to retry when the function returns an error. Valid values between 0 and 2. Defaults to 2.
* `qualifier` - (Optional) Lambda Function published version, `$LATEST`, or Lambda Alias name.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### destination_config Configuration Block

~> **Note:** At least one of `on_failure` or `on_success` must be configured when using this configuration block, otherwise remove it completely to prevent perpetual differences in Terraform runs.

* `on_failure` - (Optional) Configuration block with destination configuration for failed asynchronous invocations. [See below](#destination_config-on_failure-configuration-block).
* `on_success` - (Optional) Configuration block with destination configuration for successful asynchronous invocations. [See below](#destination_config-on_success-configuration-block).

#### destination_config on_failure Configuration Block

* `destination` - (Required) ARN of the destination resource. See the [Lambda Developer Guide](https://docs.aws.amazon.com/lambda/latest/dg/invocation-async.html#invocation-async-destinations) for acceptable resource types and associated IAM permissions.

#### destination_config on_success Configuration Block

* `destination` - (Required) ARN of the destination resource. See the [Lambda Developer Guide](https://docs.aws.amazon.com/lambda/latest/dg/invocation-async.html#invocation-async-destinations) for acceptable resource types and associated IAM permissions.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Fully qualified Lambda Function name or ARN.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda Function Event Invoke Configs using the fully qualified Function name or ARN. For example:

ARN without qualifier (all versions and aliases):

```terraform
import {
  to = aws_lambda_function_event_invoke_config.example
  id = "arn:aws:lambda:us-east-1:123456789012:function:example"
}
```

ARN with qualifier:

```terraform
import {
  to = aws_lambda_function_event_invoke_config.example
  id = "arn:aws:lambda:us-east-1:123456789012:function:example:production"
}
```

Name without qualifier (all versions and aliases):

```terraform
import {
  to = aws_lambda_function_event_invoke_config.example
  id = "example"
}
```

Name with qualifier:

```terraform
import {
  to = aws_lambda_function_event_invoke_config.example
  id = "example:production"
}
```

For backwards compatibility, the following legacy `terraform import` commands are also supported:

Using ARN without qualifier:

```console
% terraform import aws_lambda_function_event_invoke_config.example arn:aws:lambda:us-east-1:123456789012:function:example
```

Using ARN with qualifier:

```console
% terraform import aws_lambda_function_event_invoke_config.example arn:aws:lambda:us-east-1:123456789012:function:example:production
```

Name without qualifier (all versions and aliases):

```console
% terraform import aws_lambda_function_event_invoke_config.example example
```

Name with qualifier:

```console
% terraform import aws_lambda_function_event_invoke_config.example example:production
```
