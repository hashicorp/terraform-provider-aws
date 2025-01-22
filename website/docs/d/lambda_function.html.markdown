---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function"
description: |-
  Provides a Lambda Function data source.
---

# aws_lambda_function

Provides information about a Lambda Function.

## Example Usage

```terraform
variable "function_name" {
  type = string
}

data "aws_lambda_function" "existing" {
  function_name = var.function_name
}
```

## Argument Reference

This data source supports the following arguments:

* `function_name` - (Required) Name of the lambda function.
* `qualifier` - (Optional) Alias name or version number of the lambda functionE.g., `$LATEST`, `my-alias`, or `1`. When not included: the data source resolves to the most recent published version; if no published version exists: it resolves to the most recent unpublished version.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `architectures` - Instruction set architecture for the Lambda function.
* `arn` - Unqualified (no `:QUALIFIER` or `:VERSION` suffix) ARN identifying your Lambda Function. See also `qualified_arn`.
* `code_sha256` - Base64-encoded representation of raw SHA-256 sum of the zip file.
* `code_signing_config_arn` - ARN for a Code Signing Configuration.
* `dead_letter_config` - Configure the function's *dead letter queue*.
* `description` - Description of what your Lambda Function does.
* `environment` - Lambda environment's configuration settings.
* `ephemeral_storage` - Amount of Ephemeral storage(`/tmp`) allocated for the Lambda Function.
* `file_system_config` - Connection settings for an Amazon EFS file system.
* `handler` - Function entrypoint in your code.
* `image_uri` - URI of the container image.
* `invoke_arn` - ARN to be used for invoking Lambda Function from API Gateway. **NOTE:** Starting with `v4.51.0` of the provider, this will *not* include the qualifier.
* `kms_key_arn` - ARN for the KMS encryption key.
* `last_modified` - Date this resource was last modified.
* `layers` - List of Lambda Layer ARNs attached to your Lambda Function.
* `logging_config` - Advanced logging settings.
* `memory_size` - Amount of memory in MB your Lambda Function can use at runtime.
* `qualified_arn` - Qualified (`:QUALIFIER` or `:VERSION` suffix) ARN identifying your Lambda Function. See also `arn`.
* `qualified_invoke_arn` - Qualified (`:QUALIFIER` or `:VERSION` suffix) ARN to be used for invoking Lambda Function from API Gateway. See also `invoke_arn`.
* `reserved_concurrent_executions` - The amount of reserved concurrent executions for this lambda function or `-1` if unreserved.
* `role` - IAM role attached to the Lambda Function.
* `runtime` - Runtime environment for the Lambda function.
* `signing_job_arn` - ARN of a signing job.
* `signing_profile_version_arn` - The ARN for a signing profile version.
* `source_code_hash` - (**Deprecated** use `code_sha256` instead) Base64-encoded representation of raw SHA-256 sum of the zip file.
* `source_code_size` - Size in bytes of the function .zip file.
* `timeout` - Function execution time at which Lambda should terminate the function.
* `tracing_config` - Tracing settings of the function.
* `version` - The version of the Lambda function returned. If `qualifier` is not set, this will resolve to the most recent published version. If no published version of the function exists, `version` will resolve to `$LATEST`.
* `vpc_config` - VPC configuration associated with your Lambda function.
