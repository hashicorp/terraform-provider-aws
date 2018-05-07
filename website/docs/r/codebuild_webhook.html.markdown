---
layout: "aws"
page_title: "AWS: aws_codebuild_webhook"
sidebar_current: "docs-aws-resource-codebuild-webhook"
description: |-
  Provides a CodeBuild Webhook resource.
---

# aws_codebuild_webhook

Provides a CodeBuild Webhook resource.

~> **Note:** The AWS account that Terraform uses to create this resource *must* have authorized CodeBuild to access GitHub's OAuth API. This is a manual step that must be done *before* creating webhooks with this resource. If OAuth is not configured, AWS will return an error similar to `ResourceNotFoundException: Could not find access token for server type github`.

## Example Usage

```hcl
resource "aws_codebuild_webhook" "github" {
  name = "${aws_codebuild_project.my_project.name}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the build project.
* `branch_filter` - (Optional) A regular expression used to determine which branches get built. Default is all branches are built.

## Attributes Reference

The following attributes are exported:

* `id` - The name of the build project.
* `url` - The URL to the webhook.
