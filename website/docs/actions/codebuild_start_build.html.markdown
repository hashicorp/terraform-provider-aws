---
subcategory: "CodeBuild"
layout: "aws"
page_title: "AWS: aws_codebuild_start_build"
description: |-
  Starts a CodeBuild project build.
---

# Action: aws_codebuild_start_build

~> **Note:** `aws_codebuild_start_build` is in beta. Its interface and behavior may change as the feature evolves, and breaking changes are possible. It is offered as a technical preview without compatibility guarantees until Terraform 1.14 is generally available.

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
