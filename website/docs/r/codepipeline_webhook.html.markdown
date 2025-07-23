---
subcategory: "CodePipeline"
layout: "aws"
page_title: "AWS: aws_codepipeline_webhook"
description: |-
  Provides a CodePipeline Webhook
---

# Resource: aws_codepipeline_webhook

Provides a CodePipeline Webhook.

## Example Usage

```terraform
resource "aws_codepipeline" "bar" {
  name     = "tf-test-pipeline"
  role_arn = aws_iam_role.bar.arn

  artifact_store {
    location = aws_s3_bucket.bar.bucket
    type     = "S3"

    encryption_key {
      id   = data.aws_kms_alias.s3kmskey.arn
      type = "KMS"
    }
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "ThirdParty"
      provider         = "GitHub"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        Owner  = "my-organization"
        Repo   = "test"
        Branch = "master"
      }
    }
  }

  stage {
    name = "Build"

    action {
      name            = "Build"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["test"]
      version         = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }
}

# A shared secret between GitHub and AWS that allows AWS
# CodePipeline to authenticate the request came from GitHub.
# Would probably be better to pull this from the environment
# or something like SSM Parameter Store.
locals {
  webhook_secret = "super-secret"
}

resource "aws_codepipeline_webhook" "bar" {
  name            = "test-webhook-github-bar"
  authentication  = "GITHUB_HMAC"
  target_action   = "Source"
  target_pipeline = aws_codepipeline.bar.name

  authentication_configuration {
    secret_token = local.webhook_secret
  }

  filter {
    json_path    = "$.ref"
    match_equals = "refs/heads/{Branch}"
  }
}

# Wire the CodePipeline webhook into a GitHub repository.
resource "github_repository_webhook" "bar" {
  repository = github_repository.repo.name

  name = "web"

  configuration {
    url          = aws_codepipeline_webhook.bar.url
    content_type = "json"
    insecure_ssl = true
    secret       = local.webhook_secret
  }

  events = ["push"]
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The name of the webhook.
* `authentication` - (Required) The type of authentication  to use. One of `IP`, `GITHUB_HMAC`, or `UNAUTHENTICATED`.
* `authentication_configuration` - (Optional) An `auth` block. Required for `IP` and `GITHUB_HMAC`. Auth blocks are documented below.
* `filter` (Required) One or more `filter` blocks. Filter blocks are documented below.
* `target_action` - (Required) The name of the action in a pipeline you want to connect to the webhook. The action must be from the source (first) stage of the pipeline.
* `target_pipeline` - (Required) The name of the pipeline.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

An `authentication_configuration` block supports the following arguments:

* `secret_token` - (Optional) The shared secret for the GitHub repository webhook. Set this as `secret` in your `github_repository_webhook`'s `configuration` block. Required for `GITHUB_HMAC`.
* `allowed_ip_range` - (Optional) A valid CIDR block for `IP` filtering. Required for `IP`.

A `filter` block supports the following arguments:

* `json_path` - (Required) The [JSON path](https://github.com/json-path/JsonPath) to filter on.
* `match_equals` - (Required) The value to match on (e.g., `refs/heads/{Branch}`). See [AWS docs](https://docs.aws.amazon.com/codepipeline/latest/APIReference/API_WebhookFilterRule.html) for details.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The CodePipeline webhook's ARN.
* `id` - The CodePipeline webhook's ARN.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `url` - The CodePipeline webhook's URL. POST events to this endpoint to trigger the target.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodePipeline Webhooks using their ARN. For example:

```terraform
import {
  to = aws_codepipeline_webhook.example
  id = "arn:aws:codepipeline:us-west-2:123456789012:webhook:example"
}
```

Using `terraform import`, import CodePipeline Webhooks using their ARN. For example:

```console
% terraform import aws_codepipeline_webhook.example arn:aws:codepipeline:us-west-2:123456789012:webhook:example
```
