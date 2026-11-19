---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_model_invocation_job"
description: |-
  Lists Bedrock Model Invocation Job resources.
---

# List Resource: aws_bedrock_model_invocation_job

Lists Bedrock Model Invocation Job resources.

## Example Usage

```terraform
list "aws_bedrock_model_invocation_job" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
