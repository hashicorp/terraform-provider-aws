---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_invocation"
description: |-
  Invokes an AWS Lambda Function as an ephemeral resource.
---

# Ephemeral: aws_lambda_invocation

Invokes an AWS Lambda Function as an ephemeral resource. Use this ephemeral resource to execute Lambda functions during Terraform operations without persisting results in state, ideal for generating sensitive data or performing lightweight operations.

The Lambda function is invoked with [RequestResponse](https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html#API_Invoke_RequestSyntax) invocation type.

~> **Note:** Ephemeral resources are a new feature and may evolve as we continue to explore their most effective uses. [Learn more](https://developer.hashicorp.com/terraform/language/v1.10.x/resources/ephemeral).

~> **Note:** The `aws_lambda_invocation` ephemeral resource invokes the function during every `plan` and `apply` when the function is known. A common use case for this functionality is when invoking a lightweight function—where repeated invocations are acceptable—that produces sensitive information you do not want to store in the state.

~> **Note:** If you get a `KMSAccessDeniedException: Lambda was unable to decrypt the environment variables because KMS access was denied` error when invoking a Lambda function with environment variables, the IAM role associated with the function may have been deleted and recreated after the function was created. You can fix the problem two ways: 1) updating the function's role to another role and then updating it back again to the recreated role, or 2) by using Terraform to `taint` the function and `apply` your configuration again to recreate the function. (When you create a function, Lambda grants permissions on the KMS key to the function's IAM role. If the IAM role is recreated, the grant is no longer valid. Changing the function's role or recreating the function causes Lambda to update the grant.)

## Example Usage

### Generate Sensitive Configuration

```terraform
variable "environment" {
  description = "The environment name (e.g., dev, prod)"
  type        = string
}

# Lambda function that generates API keys or secrets
ephemeral "aws_lambda_invocation" "secret_generator" {
  function_name = aws_lambda_function.secret_generator.function_name

  payload = jsonencode({
    service     = "api"
    environment = var.environment
    length      = 32
  })
}

# Use the generated secret without storing it in state
resource "aws_ssm_parameter" "api_key" {
  name  = "/app/${var.environment}/api-key"
  type  = "SecureString"
  value = jsondecode(ephemeral.aws_lambda_invocation.secret_generator.result).api_key

  tags = {
    Environment = var.environment
    Generated   = "ephemeral-lambda"
  }
}

# Output must be marked as ephemeral
output "key_generated" {
  value     = "API key generated and stored in Parameter Store"
  ephemeral = true
}
```

### Dynamic Resource Configuration

```terraform
# Function that calculates optimal resource sizing
ephemeral "aws_lambda_invocation" "resource_calculator" {
  function_name = "resource-optimizer"
  qualifier     = "production"

  payload = jsonencode({
    workload_type = var.workload_type
    expected_load = var.expected_requests_per_second
    region        = data.aws_region.current.name
  })
}

locals {
  sizing = jsondecode(ephemeral.aws_lambda_invocation.resource_calculator.result)
}

# Use calculated values for resource creation
resource "aws_autoscaling_group" "example" {
  name                = "optimized-asg"
  vpc_zone_identifier = var.subnet_ids
  target_group_arns   = [aws_lb_target_group.example.arn]
  health_check_type   = "ELB"

  min_size         = local.sizing.min_instances
  max_size         = local.sizing.max_instances
  desired_capacity = local.sizing.desired_instances

  launch_template {
    id      = aws_launch_template.example.id
    version = "$Latest"
  }

  tag {
    key                 = "OptimizedBy"
    value               = "ephemeral-lambda"
    propagate_at_launch = true
  }
}
```

### Validation and Compliance Checks

```terraform
variable "instance_type" {
  description = "The EC2 instance type to use"
  type        = string
}

# Function that validates configuration against compliance rules
ephemeral "aws_lambda_invocation" "compliance_validator" {
  function_name = "compliance-checker"
  log_type      = "Tail" # Include execution logs

  payload = jsonencode({
    resource_config = {
      instance_type     = var.instance_type
      storage_encrypted = var.encrypt_storage
      backup_enabled    = var.enable_backups
    }
    compliance_framework = "SOC2"
  })
}

locals {
  validation_result = jsondecode(ephemeral.aws_lambda_invocation.compliance_validator.result)
  is_compliant      = validation_result.compliant
  violations        = validation_result.violations
}

# Conditional resource creation based on compliance
resource "aws_instance" "example" {
  count = local.is_compliant ? 1 : 0

  ami           = data.aws_ami.example.id
  instance_type = var.instance_type

  root_block_device {
    encrypted = var.encrypt_storage
  }

  tags = {
    Environment     = var.environment
    ComplianceCheck = "passed"
  }
}

# Fail deployment if not compliant
resource "null_resource" "compliance_gate" {
  count = local.is_compliant ? 0 : 1

  provisioner "local-exec" {
    command = "echo 'Compliance violations: ${join(", ", local.violations)}' && exit 1"
  }
}
```

### External API Integration

```terraform
# Function that calls external APIs for configuration data
ephemeral "aws_lambda_invocation" "external_config" {
  function_name = "config-fetcher"
  client_context = base64encode(jsonencode({
    source  = "terraform"
    version = "1.0"
  }))

  payload = jsonencode({
    config_service_url = var.config_service_url
    environment        = var.environment
    service_name       = "web-app"
  })
}

locals {
  external_config = jsondecode(ephemeral.aws_lambda_invocation.external_config.result)
}

# Use external configuration
resource "aws_ecs_service" "example" {
  name            = "web-app"
  cluster         = aws_ecs_cluster.example.id
  task_definition = aws_ecs_task_definition.example.arn
  desired_count   = local.external_config.replica_count

  deployment_configuration {
    maximum_percent         = local.external_config.max_percent
    minimum_healthy_percent = local.external_config.min_healthy_percent
  }

  tags = {
    ConfigSource = "external-api"
    Environment  = var.environment
  }
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name or ARN of the Lambda function, version, or alias. You can append a version number or alias. If you specify only the function name, it is limited to 64 characters in length.
* `payload` - (Required) JSON that you want to provide to your Lambda function as input.

The following arguments are optional:

* `client_context` - (Optional) Up to 3583 bytes of base64-encoded data about the invoking client to pass to the function in the context object.
* `log_type` - (Optional) Set to `Tail` to include the execution log in the response. Valid values: `None` and `Tail`.
* `qualifier` - (Optional) Version or alias to invoke a published version of the function. Defaults to `$LATEST`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This ephemeral resource exports the following attributes in addition to the arguments above:

* `executed_version` - Version of the function that executed. When you invoke a function with an alias, this shows the version the alias resolved to.
* `function_error` - If present, indicates that an error occurred during function execution. Details about the error are included in `result`.
* `log_result` - Last 4 KB of the execution log, which is base64-encoded.
* `result` - String result of the Lambda function invocation.
* `status_code` - HTTP status code is in the 200 range for a successful request.

## Usage Notes

### Handling Sensitive Data

Since ephemeral resources are designed to not persist data in state, they are ideal for handling sensitive information:

```terraform
ephemeral "aws_lambda_invocation" "credentials" {
  function_name = "credential-generator"

  payload = jsonencode({
    service = "database"
    type    = "temporary"
  })
}

# Use credentials without storing them
resource "aws_secretsmanager_secret_version" "example" {
  secret_id     = aws_secretsmanager_secret.example.id
  secret_string = ephemeral.aws_lambda_invocation.credentials.result
}
```

### Error Handling

Always check for function errors in your configuration:

```terraform
locals {
  invocation_result = jsondecode(ephemeral.aws_lambda_invocation.example.result)
  has_error         = ephemeral.aws_lambda_invocation.example.function_error != null
}

# Fail if function returns an error
resource "null_resource" "validation" {
  count = local.has_error ? fail("Lambda function error: ${local.invocation_result.errorMessage}") : 0
}
```

### Logging

Enable detailed logging for debugging:

```terraform
ephemeral "aws_lambda_invocation" "example" {
  function_name = "my-function"
  log_type      = "Tail"

  payload = jsonencode({
    debug = true
  })
}

output "execution_logs" {
  value     = base64decode(ephemeral.aws_lambda_invocation.example.log_result)
  ephemeral = true
}
```
