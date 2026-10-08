---
subcategory: "Batch"
layout: "aws"
page_title: "AWS: aws_batch_job_queue"
description: |-
  Lists Batch Job Queue resources.
---

# List Resource: aws_batch_job_queue

~> **Note:** The `aws_batch_job_queue` List Resource is in beta. Its interface and behavior may change as the feature evolves, and breaking changes are possible. It is offered as a technical preview without compatibility guarantees until Terraform 1.14 is generally available.

Lists Batch Job Queue resources.

## Example Usage

```terraform
list "aws_batch_job_queue" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
