---
subcategory: "CodeBuild"
layout: "aws"
page_title: "AWS: aws_codebuild_project"
description: |-
  Lists CodeBuild Project resources.
---

# List Resource: aws_codebuild_project

~> **Note:** The `aws_codebuild_project` List Resource is in beta and may change in future versions of the provider.

Lists CodeBuild Project resources.

## Example Usage

```terraform
list "aws_codebuild_project" "example" {
  provider = aws
}
```

## Argument Reference

* `region` - (Optional) Region to query. Defaults to provider region.

## Attribute Reference

This list resource exports the same attributes as the [`aws_codebuild_project`](/docs/resources/codebuild_project.html) resource.
