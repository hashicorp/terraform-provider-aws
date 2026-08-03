---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_invoke"
description: |-
  Invokes an AWS Lambda function with the specified payload.
---

# Action: aws_lambda_invoke

Invokes an AWS Lambda function with the specified payload. This action allows for imperative invocation of Lambda functions with full control over invocation parameters.

For information about AWS Lambda functions, see the [AWS Lambda Developer Guide](https://docs.aws.amazon.com/lambda/latest/dg/). For specific information about invoking Lambda functions, see the [Invoke](https://docs.aws.amazon.com/lambda/latest/api/API_Invoke.html) page in the AWS Lambda API Reference.

~> **Note:** Synchronous invocations will wait for the function to complete execution, while asynchronous invocations return immediately after the request is _accepted_.

## Example Usage

### Basic Usage

```terraform
resource "aws_lambda_function" "example" {
  # ... function configuration
}

action "aws_lambda_invoke" "example" {
  config {
    function_name = aws_lambda_function.example.function_name
    payload = jsonencode({
      key1 = "value1"
      key2 = "value2"
    })
  }
}

resource "terraform_data" "example" {
  input = "trigger-lambda"

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_lambda_invoke.example]
    }
  }
}
```

### Invoke with Function Version

```terraform
action "aws_lambda_invoke" "versioned" {
  config {
    function_name = aws_lambda_function.example.function_name
    qualifier     = aws_lambda_function.example.version
    payload = jsonencode({
      operation = "process"
      data      = var.processing_data
    })
  }
}
```

### Asynchronous Invocation

```terraform
action "aws_lambda_invoke" "async" {
  config {
    function_name   = aws_lambda_function.worker.function_name
    invocation_type = "Event"
    payload = jsonencode({
      task_id = "background-job-${random_uuid.job_id.result}"
      data    = local.background_task_data
    })
  }
}
```

### Dry Run Validation

```terraform
action "aws_lambda_invoke" "validate" {
  config {
    function_name   = aws_lambda_function.validator.function_name
    invocation_type = "DryRun"
    payload = jsonencode({
      config = var.validation_config
    })
  }
}
```

### With Log Capture

```terraform
action "aws_lambda_invoke" "debug" {
  config {
    function_name = aws_lambda_function.debug.function_name
    log_type      = "Tail"
    payload = jsonencode({
      debug_level = "verbose"
      component   = "api-gateway"
    })
  }
}
```

### Mobile Application Context

```terraform
action "aws_lambda_invoke" "mobile" {
  config {
    function_name = aws_lambda_function.mobile_backend.function_name
    client_context = base64encode(jsonencode({
      client = {
        client_id   = "mobile-app"
        app_version = "1.2.3"
      }
      env = {
        locale = "en_US"
      }
    }))
    payload = jsonencode({
      user_id = var.user_id
      action  = "sync_data"
    })
  }
}
```

### CI/CD Pipeline Integration

Use this action in your deployment pipeline to trigger post-deployment functions:

```terraform
# Trigger warmup after deployment
resource "terraform_data" "deploy_complete" {
  input = local.deployment_id

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_lambda_invoke.warmup]
    }
  }

  depends_on = [aws_lambda_function.api]
}

action "aws_lambda_invoke" "warmup" {
  config {
    function_name = aws_lambda_function.api.function_name
    payload = jsonencode({
      action = "warmup"
      source = "terraform-deployment"
    })
  }
}
```

### Environment-Specific Processing

```terraform
locals {
  processing_config = var.environment == "production" ? {
    batch_size = 100
    timeout    = 900
    } : {
    batch_size = 10
    timeout    = 60
  }
}

action "aws_lambda_invoke" "process_data" {
  config {
    function_name = aws_lambda_function.processor.function_name
    payload = jsonencode(merge(local.processing_config, {
      data_source = var.data_source
      environment = var.environment
    }))
  }
}
```

### Complex Payload with Dynamic Content

```terraform
action "aws_lambda_invoke" "complex" {
  config {
    function_name = aws_lambda_function.orchestrator.function_name
    payload = jsonencode({
      workflow = {
        id    = "workflow-${timestamp()}"
        steps = var.workflow_steps
      }
      resources = {
        s3_bucket = aws_s3_bucket.data.bucket
        dynamodb  = aws_dynamodb_table.state.name
        sns_topic = aws_sns_topic.notifications.arn
      }
      metadata = {
        created_by  = "terraform"
        environment = var.environment
        version     = var.app_version
      }
    })
  }
}
```

## Argument Reference

This action supports the following arguments:

* `client_context` - (Optional) Up to 3,583 bytes of base64-encoded data about the invoking client to pass to the function in the context object. This is only used for mobile applications and should contain information about the client application and device.
* `function_name` - (Required) Name, ARN, or partial ARN of the Lambda function to invoke. You can specify a function name (e.g., `my-function`), a qualified function name (e.g., `my-function:PROD`), or a partial ARN (e.g., `123456789012:function:my-function`).
* `invocation_type` - (Optional) Invocation type. Valid values are `RequestResponse` (default) for synchronous invocation that waits for the function to complete and returns the response, `Event` for asynchronous invocation that returns immediately after the request is accepted, and `DryRun` to validate parameters and verify permissions without actually executing the function.
* `log_type` - (Optional) Set to `Tail` to include the execution log in the response. Only applies to synchronous invocations (`RequestResponse` invocation type). Defaults to `None`. When set to `Tail`, the last 4 KB of the execution log is included in the response and output as part of the progress messages.
* `payload` - (Required) JSON payload to send to the Lambda function. This should be a valid JSON string that represents the event data for your function. The payload size limit is 6 MB for synchronous invocations and 256 KB for asynchronous invocations.
* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `qualifier` - (Optional) Version or alias of the Lambda function to invoke. If not specified, the `$LATEST` version will be invoked. Can be a version number (e.g., `1`) or an alias (e.g., `PROD`).
* `tenant_id` - (Optional)  Tenant Id to serve invocations from specified tenant.
