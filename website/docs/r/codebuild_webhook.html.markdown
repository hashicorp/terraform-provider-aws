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

### GitHub

~> **Note:** For GitHub source projects, the AWS account that Terraform uses to create this resource *must* have authorized CodeBuild to access GitHub's OAuth API in each applicable region. This is a manual step that must be done *before* creating webhooks with this resource. If OAuth is not configured, AWS will return an error similar to `ResourceNotFoundException: Could not find access token for server type github`. More information can be found in the [CodeBuild User Guide](https://docs.aws.amazon.com/codebuild/latest/userguide/sample-github-pull-request.html).

```hcl
resource "aws_codebuild_webhook" "example" {
  name = "${aws_codebuild_project.example.name}"
}
```

### GitHub Enterprise

More information creating webhooks with GitHub Enterprise can be found in the [CodeBuild User Guide](https://docs.aws.amazon.com/codebuild/latest/userguide/sample-github-enterprise.html).

```hcl
resource "aws_codebuild_webhook" "example" {
  name = "${aws_codebuild_project.example.name}"
}

resource "github_repository_webhook" "example" {
  active     = true
  events     = ["push"]
  name       = "example"
  repository = "${github_repository.example.name}"

  configuration {
    url          = "${aws_codebuild_webhook.example.payload_url}"
    secret       = "${aws_codebulld_webhook.example.secret}"
    content_type = "json"
    insecure_ssl = false
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the build project.
* `branch_filter` - (Optional) A regular expression used to determine which branches get built. Default is all branches are built.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the build project.
* `payload_url` - The CodeBuild endpoint where webhook events are sent.
* `secret` - The secret token of the associated repository. Not returned for all source types.
* `url` - The URL to the webhook.

## Import

CodeBuild Webhooks can be imported using the CodeBuild Project name, e.g.

```
$ terraform import aws_codebuild_webhook.example MyProjectName
```
