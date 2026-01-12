---
subcategory: "SFN (Step Functions)"
layout: "aws"
page_title: "AWS: aws_sfn_start_execution"
description: |-
  Starts a Step Functions state machine execution with the specified input data.
---

# Action: aws_sfn_start_execution

Starts a Step Functions state machine execution with the specified input data. This action allows for imperative execution of state machines with full control over execution parameters.

For information about AWS Step Functions, see the [AWS Step Functions Developer Guide](https://docs.aws.amazon.com/step-functions/latest/dg/). For specific information about starting executions, see the [StartExecution](https://docs.aws.amazon.com/step-functions/latest/apireference/API_StartExecution.html) page in the AWS Step Functions API Reference.

~> **Note:** For `STANDARD` workflows, executions with the same name and input are idempotent. For `EXPRESS` workflows, each execution is unique regardless of name and input.

## Example Usage

### Basic Usage

```terraform
resource "aws_sfn_state_machine" "example" {
  name     = "example-state-machine"
  role_arn = aws_iam_role.sfn.arn

  definition = jsonencode({
    Comment = "A simple minimal example"
    StartAt = "Hello"
    States = {
      Hello = {
        Type   = "Pass"
        Result = "Hello World!"
        End    = true
      }
    }
  })
}

action "aws_sfn_start_execution" "example" {
  config {
    state_machine_arn = aws_sfn_state_machine.example.arn
    input = jsonencode({
      user_id = "12345"
      action  = "process"
    })
  }
}

resource "terraform_data" "example" {
  input = "trigger-execution"

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_sfn_start_execution.example]
    }
  }
}
```

### Named Execution

```terraform
action "aws_sfn_start_execution" "named" {
  config {
    state_machine_arn = aws_sfn_state_machine.processor.arn
    name              = "deployment-${var.deployment_id}"
    input = jsonencode({
      deployment_id = var.deployment_id
      environment   = var.environment
    })
  }
}
```

### Execution with Version

```terraform
action "aws_sfn_start_execution" "versioned" {
  config {
    state_machine_arn = "${aws_sfn_state_machine.example.arn}:${aws_sfn_state_machine.example.version_number}"
    input = jsonencode({
      version = "v2"
      config  = var.processing_config
    })
  }
}
```

### Execution with Alias

```terraform
resource "aws_sfn_alias" "prod" {
  name              = "PROD"
  state_machine_arn = aws_sfn_state_machine.example.arn
  routing_configuration {
    state_machine_version_weight {
      state_machine_version_arn = aws_sfn_state_machine.example.arn
      weight                    = 100
    }
  }
}

action "aws_sfn_start_execution" "production" {
  config {
    state_machine_arn = aws_sfn_alias.prod.arn
    input = jsonencode({
      environment = "production"
      batch_size  = 1000
    })
  }
}
```

### X-Ray Tracing

```terraform
action "aws_sfn_start_execution" "traced" {
  config {
    state_machine_arn = aws_sfn_state_machine.example.arn
    trace_header      = "Root=1-${formatdate("YYYYMMDD", timestamp())}-${substr(uuid(), 0, 24)}"
    input = jsonencode({
      trace_id = "custom-trace-${timestamp()}"
      data     = var.processing_data
    })
  }
}
```

### CI/CD Pipeline Integration

Use this action in your deployment pipeline to trigger post-deployment workflows:

```terraform
resource "terraform_data" "deploy_complete" {
  input = local.deployment_id

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_sfn_start_execution.post_deploy]
    }
  }

  depends_on = [aws_lambda_function.processors]
}

action "aws_sfn_start_execution" "post_deploy" {
  config {
    state_machine_arn = aws_sfn_state_machine.data_pipeline.arn
    name              = "post-deploy-${local.deployment_id}"
    input = jsonencode({
      deployment_id = local.deployment_id
      environment   = var.environment
      resources = {
        lambda_functions = [for f in aws_lambda_function.processors : f.arn]
        s3_bucket        = aws_s3_bucket.data.bucket
      }
    })
  }
}
```

### Environment-Specific Processing

```terraform
locals {
  execution_config = var.environment == "production" ? {
    batch_size    = 1000
    max_retries   = 3
    timeout_hours = 24
    } : {
    batch_size    = 100
    max_retries   = 1
    timeout_hours = 2
  }
}

action "aws_sfn_start_execution" "batch_process" {
  config {
    state_machine_arn = aws_sfn_state_machine.batch_processor.arn
    input = jsonencode(merge(local.execution_config, {
      data_source = var.data_source
      output_path = var.output_path
    }))
  }
}
```

### Complex Workflow Orchestration

```terraform
action "aws_sfn_start_execution" "orchestrator" {
  config {
    state_machine_arn = aws_sfn_state_machine.orchestrator.arn
    input = jsonencode({
      workflow = {
        id    = "workflow-${timestamp()}"
        type  = "data-processing"
        steps = var.workflow_steps
      }
      resources = {
        compute = {
          lambda_functions = [for f in aws_lambda_function.workers : f.arn]
          ecs_cluster      = aws_ecs_cluster.processing.arn
        }
        storage = {
          input_bucket  = aws_s3_bucket.input.bucket
          output_bucket = aws_s3_bucket.output.bucket
          temp_bucket   = aws_s3_bucket.temp.bucket
        }
        messaging = {
          success_topic = aws_sns_topic.success.arn
          error_topic   = aws_sns_topic.errors.arn
        }
      }
      metadata = {
        created_by  = "terraform"
        environment = var.environment
        version     = var.app_version
        tags        = var.execution_tags
      }
    })
  }
}
```

## Argument Reference

This action supports the following arguments:

* `input` - (Optional) JSON input data for the execution. Must be valid JSON. Defaults to `{}` if not specified. The input size limit is 256 KB.
* `name` - (Optional) Name of the execution. Must be unique within the account/region/state machine for 90 days. If not provided, Step Functions automatically generates a UUID. Names must not contain whitespace, brackets, wildcards, or special characters.
* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `state_machine_arn` - (Required) ARN of the state machine to execute. Can be an unqualified ARN, version-qualified ARN (e.g., `arn:aws:states:region:account:stateMachine:name:version`), or alias-qualified ARN (e.g., `arn:aws:states:region:account:stateMachine:name:alias`).
* `trace_header` - (Optional) AWS X-Ray trace header for distributed tracing. Used to correlate execution traces across services.
