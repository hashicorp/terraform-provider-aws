---
subcategory: "CodeBuild"
layout: "aws"
page_title: "AWS: aws_codebuild_project"
description: |-
  Lists CodeBuild Project resources.
---

# List Resource: aws_codebuild_project

Lists CodeBuild Project resources.

## Example Usage

```terraform
list "aws_codebuild_project" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
