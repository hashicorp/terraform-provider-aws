---
layout: "aws"
page_title: "AWS: aws_codebuild_webhook"
sidebar_current: "docs-aws-resource-codebuild-webhook"
description: |-
  Provides a CodeBuild Webhook resource.
---

# aws_codebuild_webhook

Provides a CodeBuild Webhook resource.

## Example Usage

```hcl
resource "aws_codebuild_webhook" "github" {
  name = "${aws_codebuild_project.my_project.name}"
}

resource "github_repository_webhook" "aws_codebuild" {
  repository = "${github_repository.repo.name}"
  name       = "web"

  configuration {
    url          = "${aws_codebuild_webhook.github.url}"
    content_type = "json"
  }

  events = ["pull_request", "push"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the build project.

## Attributes Reference

The following attributes are exported:

* `id` - The name of the build project.
* `url` - The URL to the webhook.
