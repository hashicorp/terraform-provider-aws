---
subcategory: "CodePipeline"
layout: "aws"
page_title: "AWS: aws_codepipeline_codepipeline"
description: |-
  Lists CodePipeline Pipeline resources.
---

# List Resource: aws_codepipeline_codepipeline

Lists CodePipeline Pipeline resources.

## Example Usage

```terraform
list "aws_codepipeline_codepipeline" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
