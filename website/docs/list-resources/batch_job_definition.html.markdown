---
subcategory: "Batch"
layout: "aws"
page_title: "AWS: aws_batch_job_definition"
description: |-
  Lists Batch Job Definition resources.
---

# List Resource: aws_batch_job_definition

Lists Batch Job Definition resources.

## Example Usage

```terraform
list "aws_batch_job_definition" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
