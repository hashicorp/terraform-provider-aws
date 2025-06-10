---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function"
description: |-
  Provides details about an AWS Lambda Function.
---

# Data Source: aws_lambda_function

Provides details about an AWS Lambda Function. Use this data source to obtain information about an existing Lambda function for use in other resources or as a reference for function configurations.

~> **Note:** This data source returns information about the latest version or alias specified by the `qualifier`. If no `qualifier` is provided, it returns information about the most recent published version, or `$LATEST` if no published version exists.

## Example Usage

### Basic Usage

```terraform
data "aws_lambda_function" "example" {
  function_name = "my-lambda-function"
}

output "function_arn" {
  value = data.aws_lambda_function.example.arn
}
```

### Using Function Alias

```terraform
data "aws_lambda_function" "example" {
  function_name = "api-handler"
  qualifier     = "production" # Alias name
}

# Use in API Gateway integration
resource "aws_api_gateway_integration" "example" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  resource_id = aws_api_gateway_resource.example.id
  http_method = aws_api_gateway_method.example.http_method

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = data.aws_lambda_function.example.invoke_arn
}
```

### Function Configuration Reference

```terraform
# Get existing function details
data "aws_lambda_function" "reference" {
  function_name = "existing-function"
}

# Create new function with similar configuration
resource "aws_lambda_function" "example" {
  filename      = "new-function.zip"
  function_name = "new-function"
  role          = data.aws_lambda_function.reference.role
  handler       = data.aws_lambda_function.reference.handler
  runtime       = data.aws_lambda_function.reference.runtime
  memory_size   = data.aws_lambda_function.reference.memory_size
  timeout       = data.aws_lambda_function.reference.timeout
  architectures = data.aws_lambda_function.reference.architectures

  vpc_config {
    subnet_ids         = data.aws_lambda_function.reference.vpc_config[0].subnet_ids
    security_group_ids = data.aws_lambda_function.reference.vpc_config[0].security_group_ids
  }

  environment {
    variables = data.aws_lambda_function.reference.environment[0].variables
  }
}
```

### Function Version Management

```terraform
# Get details about specific version
data "aws_lambda_function" "version" {
  function_name = "my-function"
  qualifier     = "3" # Specific version number
}

# Get details about latest version
data "aws_lambda_function" "latest" {
  function_name = "my-function"
  qualifier     = "$LATEST"
}

# Compare versions
output "version_comparison" {
  value = {
    specific_version = data.aws_lambda_function.version.version
    latest_version   = data.aws_lambda_function.latest.version
    code_difference  = data.aws_lambda_function.version.code_sha256 != data.aws_lambda_function.latest.code_sha256
  }
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name of the Lambda function.

The following arguments are optional:

* `qualifier` - (Optional) Alias name or version number of the Lambda function. E.g., `$LATEST`, `my-alias`, or `1`. When not included: the data source resolves to the most recent published version; if no published version exists: it resolves to the most recent unpublished version.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `architectures` - Instruction set architecture for the Lambda function.
* `arn` - Unqualified (no `:QUALIFIER` or `:VERSION` suffix) ARN identifying your Lambda Function. See also `qualified_arn`.
* `code_sha256` - Base64-encoded representation of raw SHA-256 sum of the zip file.
* `code_signing_config_arn` - ARN for a Code Signing Configuration.
* `dead_letter_config` - Configuration for the function's dead letter queue. [See below](#dead_letter_config-attribute-reference).
* `description` - Description of what your Lambda Function does.
* `environment` - Lambda environment's configuration settings. [See below](#environment-attribute-reference).
* `ephemeral_storage` - Amount of ephemeral storage (`/tmp`) allocated for the Lambda Function. [See below](#ephemeral_storage-attribute-reference).
* `file_system_config` - Connection settings for an Amazon EFS file system. [See below](#file_system_config-attribute-reference).
* `handler` - Function entrypoint in your code.
* `image_uri` - URI of the container image.
* `invoke_arn` - ARN to be used for invoking Lambda Function from API Gateway. **Note:** Starting with `v4.51.0` of the provider, this will not include the qualifier.
* `kms_key_arn` - ARN for the KMS encryption key.
* `last_modified` - Date this resource was last modified.
* `layers` - List of Lambda Layer ARNs attached to your Lambda Function.
* `logging_config` - Advanced logging settings. [See below](#logging_config-attribute-reference).
* `memory_size` - Amount of memory in MB your Lambda Function can use at runtime.
* `qualified_arn` - Qualified (`:QUALIFIER` or `:VERSION` suffix) ARN identifying your Lambda Function. See also `arn`.
* `qualified_invoke_arn` - Qualified (`:QUALIFIER` or `:VERSION` suffix) ARN to be used for invoking Lambda Function from API Gateway. See also `invoke_arn`.
* `reserved_concurrent_executions` - Amount of reserved concurrent executions for this Lambda function or `-1` if unreserved.
* `role` - IAM role attached to the Lambda Function.
* `runtime` - Runtime environment for the Lambda function.
* `signing_job_arn` - ARN of a signing job.
* `signing_profile_version_arn` - ARN for a signing profile version.
* `source_code_hash` - (**Deprecated** use `code_sha256` instead) Base64-encoded representation of raw SHA-256 sum of the zip file.
* `source_code_size` - Size in bytes of the function .zip file.
* `tags` - Map of tags assigned to the Lambda Function.
* `timeout` - Function execution time at which Lambda should terminate the function.
* `tracing_config` - Tracing settings of the function. [See below](#tracing_config-attribute-reference).
* `version` - Version of the Lambda function returned. If `qualifier` is not set, this will resolve to the most recent published version. If no published version of the function exists, `version` will resolve to `$LATEST`.
* `vpc_config` - VPC configuration associated with your Lambda function. [See below](#vpc_config-attribute-reference).

### dead_letter_config

* `target_arn` - ARN of an SNS topic or SQS queue to notify when an invocation fails.

### environment

* `variables` - Map of environment variables that are accessible from the function code during execution.

### ephemeral_storage

* `size` - Size of the Lambda function ephemeral storage (`/tmp`) in MB.

### file_system_config

* `arn` - ARN of the Amazon EFS Access Point that provides access to the file system.
* `local_mount_path` - Path where the function can access the file system, starting with `/mnt/`.

### logging_config

* `application_log_level` - Detail level of the logs your application sends to CloudWatch when using supported logging libraries.
* `log_format` - Format for your function's logs. Valid values: `Text`, `JSON`.
* `log_group` - CloudWatch log group your function sends logs to.
* `system_log_level` - Detail level of the Lambda platform event logs sent to CloudWatch.

### tracing_config

* `mode` - Tracing mode. Valid values: `Active`, `PassThrough`.

### vpc_config

* `security_group_ids` - List of security group IDs associated with the Lambda function.
* `subnet_ids` - List of subnet IDs associated with the Lambda function.
* `vpc_id` - ID of the VPC.
