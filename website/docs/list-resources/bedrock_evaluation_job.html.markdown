---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_evaluation_job"
description: |-
  Lists Bedrock Evaluation Job resources.
---

# List Resource: aws_bedrock_evaluation_job

Lists Bedrock Evaluation Job resources.

## Example Usage

```terraform
list "aws_bedrock_evaluation_job" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
