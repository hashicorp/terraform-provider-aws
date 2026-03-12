---
subcategory: "CodeBuild"
layout: "aws"
page_title: "AWS: aws_codebuild_start_build"
description: |-
  Starts a CodeBuild project build.
---

# Action: aws_codebuild_start_build

Starts a CodeBuild project build. This action will initiate a build and wait for it to complete, providing progress updates during execution.

For information about AWS CodeBuild, see the [AWS CodeBuild User Guide](https://docs.aws.amazon.com/codebuild/latest/userguide/). For specific information about starting builds, see the [StartBuild](https://docs.aws.amazon.com/codebuild/latest/APIReference/API_StartBuild.html) page in the AWS CodeBuild API Reference.

## Example Usage

### Basic Usage

```terraform
resource "aws_codebuild_project" "example" {
  name         = "example-project"
  service_role = aws_iam_role.example.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "aws/codebuild/amazonlinux2-x86_64-standard:3.0"
    type         = "LINUX_CONTAINER"
  }

  source {
    type      = "NO_SOURCE"
    buildspec = "version: 0.2\nphases:\n  build:\n    commands:\n      - echo 'Hello World'"
  }
}

action "aws_codebuild_start_build" "example" {
  config {
    project_name = aws_codebuild_project.example.name
  }
}

resource "terraform_data" "build_trigger" {
  input = "trigger-build"

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aws_codebuild_start_build.example]
    }
  }
}
```

### Build with Environment Variables

```terraform
action "aws_codebuild_start_build" "deploy" {
  config {
    project_name   = aws_codebuild_project.deploy.name
    source_version = "main"
    timeout        = 1800

    environment_variables_override {
      name  = "ENVIRONMENT"
      value = "production"
      type  = "PLAINTEXT"
    }

    environment_variables_override {
      name  = "API_KEY"
      value = "/prod/api-key"
      type  = "PARAMETER_STORE"
    }
  }
}
```

### Building and Deploying Container Images

```terraform
resource "aws_codebuild_project" "container_build" {
  name = "my-container-build"
  # ... configuration to build and push to ECR ...
}

action "aws_codebuild_start_build" "container" {
  config {
    project_name = aws_codebuild_project.container_build.name
    timeout      = 3600
  }
}

resource "terraform_data" "build_trigger" {
  input = filemd5("Dockerfile")

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_codebuild_start_build.container]
    }
  }
}

resource "aws_ecs_service" "app" {
  # ... configuration ...
  task_definition = aws_ecs_task_definition.app.arn

  depends_on = [terraform_data.build_trigger]
}

resource "aws_ecs_task_definition" "app" {
  container_definitions = jsonencode([{
    image = "${aws_ecr_repository.app.repository_url}:latest"
    # ... other config ...
  }])
}
```

## Dependency Management

The `aws_codebuild_start_build` action is synchronous and waits for the CodeBuild build to complete before returning. However, when using `action_trigger` with lifecycle events, the timing of when the trigger resource is marked complete affects Terraform's dependency graph.

### Using with `before_create` (Recommended for Build Artifacts)

When dependent resources need artifacts produced by the build (e.g., S3 objects, ECR images), use `before_create` to ensure the action completes before the trigger resource is marked complete:

```terraform
resource "terraform_data" "build_trigger" {
  input = filemd5("src/app.py") # Trigger on source changes

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_codebuild_start_build.example]
    }
  }
}

resource "aws_lambda_function" "app" {
  s3_bucket = "my-bucket"
  s3_key    = "artifact.zip" # Uploaded by CodeBuild

  depends_on = [terraform_data.build_trigger]
}
```

Execution order:

1. Action runs and waits for build completion
2. `terraform_data.build_trigger` is marked complete
3. `aws_lambda_function.app` starts creating (artifact exists)

### Using with `after_create`

With `after_create`, the trigger resource completes before the action runs. This breaks dependency chains:

```terraform
resource "terraform_data" "build_trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create] # ⚠️ Resource completes first
      actions = [action.aws_codebuild_start_build.example]
    }
  }
}

resource "aws_lambda_function" "app" {
  depends_on = [terraform_data.build_trigger] # ⚠️ Doesn't wait for action
}
```

Execution order:

1. `terraform_data.build_trigger` is marked complete
2. `aws_lambda_function.app` starts creating immediately
3. Action runs asynchronously (may fail due to missing artifact)

Use `after_create` only when dependent resources don't need to wait for the build to complete.

### Key Differences

| Event | Trigger Completes | Action Runs | Use Case |
|-------|-------------------|-------------|----------|
| `before_create` | After action completes | Before resource marked complete | Dependent resources need build artifacts |
| `after_create` | Before action runs | After resource marked complete | Fire-and-forget builds, no dependencies |

## Argument Reference

The following arguments are required:

* `project_name` - (Required) Name of the CodeBuild project to build.

The following arguments are optional:

* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `source_version` - (Optional) Version of the build input to be built. For GitHub, this can be a commit SHA, branch name, or tag name.
* `timeout` - (Optional) Timeout in seconds for the build operation. Defaults to 1800 seconds (30 minutes).
* `environment_variables_override` - (Optional) Environment variables to override for this build. See [Environment Variables Override](#environment-variables-override) below.

### Environment Variables Override

* `name` - (Required) Environment variable name.
* `value` - (Required) Environment variable value.
* `type` - (Optional) Environment variable type. Valid values are `PLAINTEXT`, `PARAMETER_STORE`, or `SECRETS_MANAGER`. Defaults to `PLAINTEXT`.
