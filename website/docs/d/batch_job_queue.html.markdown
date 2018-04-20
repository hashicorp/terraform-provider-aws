---
layout: "aws"
page_title: "AWS: aws_batch_job_queue
sidebar_current: "docs-aws-datasource-batch-job-queue
description: |-
    Provides details about a batch job queue
---

# Data Source: aws_batch_job_queue

The Batch Job Queue data source allows access to details of a specific
job queue within AWS Batch.

## Example Usage

```hcl
data "aws_batch_job_queue" "test-queue" {
  name = "tf-test-batch-job-queue"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the job queue.

## Attributes Reference

The following attributes are exported:

* `arn` - The ARN of the job queue.
* `status` - The current status of the job queue (for example, `CREATING` or `VALID`).
* `status_reason` - A short, human-readable string to provide additional details about the current status
    of the job queue.
* `state` - Describes the ability of the queue to accept new jobs (for example, `ENABLED` or `DISABLED`).
* `priority` - The priority of the job queue. Job queues with a higher priority are evaluated first when
    associated with the same compute environment.
