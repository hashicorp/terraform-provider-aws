---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function"
description: |-
  Manages an AWS Lambda Function.
---

# Resource: aws_lambda_function

Manages an AWS Lambda Function. Use this resource to create serverless functions that run code in response to events without provisioning or managing servers.

For information about Lambda and how to use it, see [What is AWS Lambda?](https://docs.aws.amazon.com/lambda/latest/dg/welcome.html). For a detailed example of setting up Lambda and API Gateway, see [Serverless Applications with AWS Lambda and API Gateway](https://learn.hashicorp.com/terraform/aws/lambda-api-gateway).

~> **Note:** Due to [AWS Lambda improved VPC networking changes that began deploying in September 2019](https://aws.amazon.com/blogs/compute/announcing-improved-vpc-networking-for-aws-lambda-functions/), EC2 subnets and security groups associated with Lambda Functions can take up to 45 minutes to successfully delete. Terraform AWS Provider version 2.31.0 and later automatically handles this increased timeout, however prior versions require setting the customizable deletion timeouts of those Terraform resources to 45 minutes (`delete = "45m"`). AWS and HashiCorp are working together to reduce the amount of time required for resource deletion and updates can be tracked in this [GitHub issue](https://github.com/hashicorp/terraform-provider-aws/issues/10329).

~> **Note:** If you get a `KMSAccessDeniedException: Lambda was unable to decrypt the environment variables because KMS access was denied` error when invoking an `aws_lambda_function` with environment variables, the IAM role associated with the function may have been deleted and recreated after the function was created. You can fix the problem two ways: 1) updating the function's role to another role and then updating it back again to the recreated role, or 2) by using Terraform to `taint` the function and `apply` your configuration again to recreate the function. (When you create a function, Lambda grants permissions on the KMS key to the function's IAM role. If the IAM role is recreated, the grant is no longer valid. Changing the function's role or recreating the function causes Lambda to update the grant.)

-> **Tip:** To give an external source (like an EventBridge Rule, SNS, or S3) permission to access the Lambda function, use the [`aws_lambda_permission`](lambda_permission.html) resource. See [Lambda Permission Model](https://docs.aws.amazon.com/lambda/latest/dg/intro-permission-model.html) for more details. On the other hand, the `role` argument of this resource is the function's execution role for identity and access to AWS services and resources.

## Example Usage

### Basic Function with Node.js

```terraform
# IAM role for Lambda execution
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "example" {
  name               = "lambda_execution_role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

# Package the Lambda function code
data "archive_file" "example" {
  type        = "zip"
  source_file = "${path.module}/lambda/index.js"
  output_path = "${path.module}/lambda/function.zip"
}

# Lambda function
resource "aws_lambda_function" "example" {
  filename         = data.archive_file.example.output_path
  function_name    = "example_lambda_function"
  role             = aws_iam_role.example.arn
  handler          = "index.handler"
  source_code_hash = data.archive_file.example.output_base64sha256

  runtime = "nodejs20.x"

  environment {
    variables = {
      ENVIRONMENT = "production"
      LOG_LEVEL   = "info"
    }
  }

  tags = {
    Environment = "production"
    Application = "example"
  }
}
```

### Container Image Function

```terraform
resource "aws_lambda_function" "example" {
  function_name = "example_container_function"
  role          = aws_iam_role.example.arn
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.example.repository_url}:latest"

  image_config {
    entry_point = ["/lambda-entrypoint.sh"]
    command     = ["app.handler"]
  }

  memory_size = 512
  timeout     = 30

  architectures = ["arm64"] # Graviton support for better price/performance
}
```

### Function with Lambda Layers

~> **Note:** The `aws_lambda_layer_version` attribute values for `arn` and `layer_arn` were swapped in version 2.0.0 of the Terraform AWS Provider. For version 2.x, use `arn` references.

```terraform
# Common dependencies layer
resource "aws_lambda_layer_version" "example" {
  filename            = "layer.zip"
  layer_name          = "example_dependencies_layer"
  description         = "Common dependencies for Lambda functions"
  compatible_runtimes = ["nodejs20.x", "python3.12"]

  compatible_architectures = ["x86_64", "arm64"]
}

# Function using the layer
resource "aws_lambda_function" "example" {
  filename      = "function.zip"
  function_name = "example_layered_function"
  role          = aws_iam_role.example.arn
  handler       = "index.handler"
  runtime       = "nodejs20.x"

  layers = [aws_lambda_layer_version.example.arn]

  tracing_config {
    mode = "Active" # Enable X-Ray tracing
  }
}
```

### VPC Function with Enhanced Networking

```terraform
resource "aws_lambda_function" "example" {
  filename      = "function.zip"
  function_name = "example_vpc_function"
  role          = aws_iam_role.example.arn
  handler       = "app.handler"
  runtime       = "python3.12"
  memory_size   = 1024
  timeout       = 30

  vpc_config {
    subnet_ids                  = [aws_subnet.example_private1.id, aws_subnet.example_private2.id]
    security_group_ids          = [aws_security_group.example_lambda.id]
    ipv6_allowed_for_dual_stack = true # Enable IPv6 support
  }

  # Increase /tmp storage to 5GB
  ephemeral_storage {
    size = 5120
  }

  # Enable SnapStart for faster cold starts
  snap_start {
    apply_on = "PublishedVersions"
  }
}
```

### Function with EFS Integration

```terraform
# EFS file system for Lambda
resource "aws_efs_file_system" "example" {
  encrypted = true

  tags = {
    Name = "lambda-efs"
  }
}

# Example subnet IDs (replace with your actual subnet IDs)
variable "subnet_ids" {
  description = "List of subnet IDs for EFS mount targets"
  type        = list(string)
  default     = ["subnet-12345678", "subnet-87654321"]
}

# Mount target in each subnet
resource "aws_efs_mount_target" "example" {
  count = length(var.subnet_ids)

  file_system_id  = aws_efs_file_system.example.id
  subnet_id       = var.subnet_ids[count.index]
  security_groups = [aws_security_group.efs.id]
}

# Access point for Lambda
resource "aws_efs_access_point" "example" {
  file_system_id = aws_efs_file_system.example.id

  root_directory {
    path = "/lambda"
    creation_info {
      owner_gid   = 1000
      owner_uid   = 1000
      permissions = "755"
    }
  }

  posix_user {
    gid = 1000
    uid = 1000
  }
}

# Lambda function with EFS
resource "aws_lambda_function" "example" {
  filename      = "function.zip"
  function_name = "example_efs_function"
  role          = aws_iam_role.example.arn
  handler       = "index.handler"
  runtime       = "nodejs20.x"

  vpc_config {
    subnet_ids         = var.subnet_ids
    security_group_ids = [aws_security_group.lambda.id]
  }

  file_system_config {
    arn              = aws_efs_access_point.example.arn
    local_mount_path = "/mnt/data"
  }

  # Ensure EFS is ready before Lambda creation
  depends_on = [aws_efs_mount_target.example]
}
```

### Function with Advanced Logging

```terraform
resource "aws_cloudwatch_log_group" "example" {
  name              = "/aws/lambda/example_function"
  retention_in_days = 14

  tags = {
    Environment = "production"
    Application = "example"
  }
}

resource "aws_lambda_function" "example" {
  filename      = "function.zip"
  function_name = "example_function"
  role          = aws_iam_role.example.arn
  handler       = "index.handler"
  runtime       = "nodejs20.x"

  # Advanced logging configuration
  logging_config {
    log_format            = "JSON"
    application_log_level = "INFO"
    system_log_level      = "WARN"
  }

  # Ensure log group exists before function
  depends_on = [aws_cloudwatch_log_group.example]
}
```

### Function with Error Handling

```terraform
# Main Lambda function
resource "aws_lambda_function" "example" {
  filename      = "function.zip"
  function_name = "example_function"
  role          = aws_iam_role.example.arn
  handler       = "index.handler"
  runtime       = "nodejs20.x"

  dead_letter_config {
    target_arn = aws_sqs_queue.dlq.arn
  }
}

# Event invoke configuration for retries
resource "aws_lambda_function_event_invoke_config" "example" {
  function_name = aws_lambda_function.example.function_name

  maximum_event_age_in_seconds = 60
  maximum_retry_attempts       = 2

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

### CloudWatch Logging and Permissions

```terraform
# Function name variable
variable "function_name" {
  description = "Name of the Lambda function"
  type        = string
  default     = "example_function"
}

# CloudWatch Log Group with retention
resource "aws_cloudwatch_log_group" "example" {
  name              = "/aws/lambda/${var.function_name}"
  retention_in_days = 14

  tags = {
    Environment = "production"
    Function    = var.function_name
  }
}

# Lambda execution role
resource "aws_iam_role" "example" {
  name = "lambda_execution_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

# CloudWatch Logs policy
resource "aws_iam_policy" "lambda_logging" {
  name        = "lambda_logging"
  path        = "/"
  description = "IAM policy for logging from Lambda"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = ["arn:aws:logs:*:*:*"]
      }
    ]
  })
}

# Attach logging policy to Lambda role
resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.example.name
  policy_arn = aws_iam_policy.lambda_logging.arn
}

# Lambda function with logging
resource "aws_lambda_function" "example" {
  filename      = "function.zip"
  function_name = var.function_name
  role          = aws_iam_role.example.arn
  handler       = "index.handler"
  runtime       = "nodejs20.x"

  # Advanced logging configuration
  logging_config {
    log_format            = "JSON"
    application_log_level = "INFO"
    system_log_level      = "WARN"
  }

  # Ensure IAM role and log group are ready
  depends_on = [
    aws_iam_role_policy_attachment.lambda_logs,
    aws_cloudwatch_log_group.example
  ]
}
```

## Specifying the Deployment Package

AWS Lambda expects source code to be provided as a deployment package whose structure varies depending on which `runtime` is in use. See [Runtimes](https://docs.aws.amazon.com/lambda/latest/dg/API_CreateFunction.html#SSS-CreateFunction-request-Runtime) for the valid values of `runtime`. The expected structure of the deployment package can be found in [the AWS Lambda documentation for each runtime](https://docs.aws.amazon.com/lambda/latest/dg/deployment-package-v2.html).

Once you have created your deployment package you can specify it either directly as a local file (using the `filename` argument) or indirectly via Amazon S3 (using the `s3_bucket`, `s3_key` and `s3_object_version` arguments). When providing the deployment package via S3 it may be useful to use [the `aws_s3_object` resource](s3_object.html) to upload it.

For larger deployment packages it is recommended by Amazon to upload via S3, since the S3 API has better support for uploading large files efficiently.

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Unique name for your Lambda Function.
* `role` - (Required) ARN of the function's execution role. The role provides the function's identity and access to AWS services and resources.

The following arguments are optional:

* `architectures` - (Optional) Instruction set architecture for your Lambda function. Valid values are `["x86_64"]` and `["arm64"]`. Default is `["x86_64"]`. Removing this attribute, function's architecture stays the same.
* `code_signing_config_arn` - (Optional) ARN of a code-signing configuration to enable code signing for this function.
* `dead_letter_config` - (Optional) Configuration block for dead letter queue. [See below](#dead_letter_config-configuration-block).
* `description` - (Optional) Description of what your Lambda Function does.
* `environment` - (Optional) Configuration block for environment variables. [See below](#environment-configuration-block).
* `ephemeral_storage` - (Optional) Amount of ephemeral storage (`/tmp`) to allocate for the Lambda Function. [See below](#ephemeral_storage-configuration-block).
* `file_system_config` - (Optional) Configuration block for EFS file system. [See below](#file_system_config-configuration-block).
* `filename` - (Optional) Path to the function's deployment package within the local filesystem. Conflicts with `image_uri` and `s3_bucket`. One of `filename`, `image_uri`, or `s3_bucket` must be specified.
* `handler` - (Optional) Function entry point in your code. Required if `package_type` is `Zip`.
* `image_config` - (Optional) Container image configuration values. [See below](#image_config-configuration-block).
* `image_uri` - (Optional) ECR image URI containing the function's deployment package. Conflicts with `filename` and `s3_bucket`. One of `filename`, `image_uri`, or `s3_bucket` must be specified.
* `kms_key_arn` - (Optional) ARN of the AWS Key Management Service key used to encrypt environment variables. If not provided when environment variables are in use, AWS Lambda uses a default service key. If provided when environment variables are not in use, the AWS Lambda API does not save this configuration.
* `layers` - (Optional) List of Lambda Layer Version ARNs (maximum of 5) to attach to your Lambda Function.
* `logging_config` - (Optional) Configuration block for advanced logging settings. [See below](#logging_config-configuration-block).
* `memory_size` - (Optional) Amount of memory in MB your Lambda Function can use at runtime. Valid value between 128 MB to 10,240 MB (10 GB), in 1 MB increments. Defaults to 128.
* `package_type` - (Optional) Lambda deployment package type. Valid values are `Zip` and `Image`. Defaults to `Zip`.
* `publish` - (Optional) Whether to publish creation/change as new Lambda Function Version. Defaults to `false`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `replace_security_groups_on_destroy` - (Optional) Whether to replace the security groups on the function's VPC configuration prior to destruction. Default is `false`.
* `replacement_security_group_ids` - (Optional) List of security group IDs to assign to the function's VPC configuration prior to destruction. Required if `replace_security_groups_on_destroy` is `true`.
* `reserved_concurrent_executions` - (Optional) Amount of reserved concurrent executions for this lambda function. A value of `0` disables lambda from being triggered and `-1` removes any concurrency limitations. Defaults to Unreserved Concurrency Limits `-1`.
* `runtime` - (Optional) Identifier of the function's runtime. Required if `package_type` is `Zip`. See [Runtimes](https://docs.aws.amazon.com/lambda/latest/dg/API_CreateFunction.html#SSS-CreateFunction-request-Runtime) for valid values.
* `s3_bucket` - (Optional) S3 bucket location containing the function's deployment package. Conflicts with `filename` and `image_uri`. One of `filename`, `image_uri`, or `s3_bucket` must be specified.
* `s3_key` - (Optional) S3 key of an object containing the function's deployment package. Required if `s3_bucket` is set.
* `s3_object_version` - (Optional) Object version containing the function's deployment package. Conflicts with `filename` and `image_uri`.
* `skip_destroy` - (Optional) Whether to retain the old version of a previously deployed Lambda Layer. Default is `false`.
* `snap_start` - (Optional) Configuration block for snap start settings. [See below](#snap_start-configuration-block).
* `source_code_hash` - (Optional) Base64-encoded SHA256 hash of the package file. Used to trigger updates when source code changes.
* `tags` - (Optional) Key-value map of tags for the Lambda function. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `timeout` - (Optional) Amount of time your Lambda Function has to run in seconds. Defaults to 3. Valid between 1 and 900.
* `tracing_config` - (Optional) Configuration block for X-Ray tracing. [See below](#tracing_config-configuration-block).
* `vpc_config` - (Optional) Configuration block for VPC. [See below](#vpc_config-configuration-block).

### dead_letter_config Configuration Block

* `target_arn` - (Required) ARN of an SNS topic or SQS queue to notify when an invocation fails.

### environment Configuration Block

* `variables` - (Optional) Map of environment variables available to your Lambda function during execution.

### ephemeral_storage Configuration Block

* `size` - (Required) Amount of ephemeral storage (`/tmp`) in MB. Valid between 512 MB and 10,240 MB (10 GB).

### file_system_config Configuration Block

* `arn` - (Required) ARN of the Amazon EFS Access Point.
* `local_mount_path` - (Required) Path where the function can access the file system. Must start with `/mnt/`.

### image_config Configuration Block

* `command` - (Optional) Parameters to pass to the container image.
* `entry_point` - (Optional) Entry point to your application.
* `working_directory` - (Optional) Working directory for the container image.

### logging_config Configuration Block

* `application_log_level` - (Optional) Detail level of application logs. Valid values: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`.
* `log_format` - (Required) Log format. Valid values: `Text`, `JSON`.
* `log_group` - (Optional) CloudWatch log group where logs are sent.
* `system_log_level` - (Optional) Detail level of Lambda platform logs. Valid values: `DEBUG`, `INFO`, `WARN`.

### snap_start Configuration Block

* `apply_on` - (Required) When to apply snap start optimization. Valid value: `PublishedVersions`.

### tracing_config Configuration Block

* `mode` - (Required) X-Ray tracing mode. Valid values: `Active`, `PassThrough`.

### vpc_config Configuration Block

* `ipv6_allowed_for_dual_stack` - (Optional) Whether to allow outbound IPv6 traffic on VPC functions connected to dual-stack subnets. Default: `false`.
* `security_group_ids` - (Required) List of security group IDs associated with the Lambda function.
* `subnet_ids` - (Required) List of subnet IDs associated with the Lambda function.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN identifying your Lambda Function.
* `code_sha256` - Base64-encoded representation of raw SHA-256 sum of the zip file.
* `invoke_arn` - ARN to be used for invoking Lambda Function from API Gateway - to be used in [`aws_api_gateway_integration`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/api_gateway_integration)'s `uri`.
* `last_modified` - Date this resource was last modified.
* `qualified_arn` - ARN identifying your Lambda Function Version (if versioning is enabled via `publish = true`).
* `qualified_invoke_arn` - Qualified ARN (ARN with lambda version number) to be used for invoking Lambda Function from API Gateway - to be used in [`aws_api_gateway_integration`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/api_gateway_integration)'s `uri`.
* `signing_job_arn` - ARN of the signing job.
* `signing_profile_version_arn` - ARN of the signing profile version.
* `snap_start.optimization_status` - Optimization status of the snap start configuration. Valid values are `On` and `Off`.
* `source_code_size` - Size in bytes of the function .zip file.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `version` - Latest published version of your Lambda Function.
* `vpc_config.vpc_id` - ID of the VPC.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda Functions using the `function_name`. For example:

```terraform
import {
  to = aws_lambda_function.example
  id = "example"
}
```

Using `terraform import`, import Lambda Functions using the `function_name`. For example:

```console
% terraform import aws_lambda_function.example example
```
