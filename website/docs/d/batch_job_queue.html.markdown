---
subcategory: "Batch"
layout: "aws"
page_title: "AWS: aws_batch_job_queue"
description: |-
    Provides details about a batch job queue
---

# Data Source: aws_batch_job_queue

The Batch Job Queue data source allows access to details of a specific
job queue within AWS Batch.

## Example Usage

```terraform
data "aws_batch_job_queue" "test-queue" {
  name = "tf-test-batch-job-queue"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the job queue.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the job queue.
* `compute_environment_order` - Compute environments that are attached to the job queue and the order in which job placement is preferred. Compute environments are selected for job placement in ascending order.
    * `compute_environment` - ARN of the compute environment.
    * `order` - Order of the compute environment.
* `job_state_time_limit_action` - Action that AWS Batch takes after the job has remained at the head of the queue in the specified state for longer than the specified time.
    * `action` - Action to take when a job is at the head of the job queue in the specified state for the specified period of time.
    * `max_time_seconds` - Approximate amount of time, in seconds, that must pass with the job in the specified state before the action is taken.
    * `reason` - Reason to log for the action being taken.
    * `state` - State of the job needed to trigger the action.
* `priority` - Priority of the job queue. Job queues with a higher priority are evaluated first when associated with the same compute environment.
* `scheduling_policy_arn` - ARN of the fair share scheduling policy. If this attribute has a value, the job queue uses a fair share scheduling policy. If this attribute does not have a value, the job queue uses a first in, first out (FIFO) scheduling policy.
* `state` - Ability of the queue to accept new jobs (for example, `ENABLED` or `DISABLED`).
* `status` - Current status of the job queue (for example, `CREATING` or `VALID`).
* `status_reason` - Short, human-readable string to provide additional details about the current status of the job queue.
* `tags` - Key-value map of resource tags.
