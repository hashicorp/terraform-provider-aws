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

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the job queue.
* `scheduling_policy_arn` - The ARN of the fair share scheduling policy. If this attribute has a value, the job queue uses a fair share scheduling policy. If this attribute does not have a value, the job queue uses a first in, first out (FIFO) scheduling policy.
* `status` - Current status of the job queue (for example, `CREATING` or `VALID`).
* `status_reason` - Short, human-readable string to provide additional details about the current status
    of the job queue.
* `state` - Describes the ability of the queue to accept new jobs (for example, `ENABLED` or `DISABLED`).
* `tags` - Key-value map of resource tags
* `priority` - Priority of the job queue. Job queues with a higher priority are evaluated first when
    associated with the same compute environment.
* `compute_environment_order` - The compute environments that are attached to the job queue and the order in
    which job placement is preferred. Compute environments are selected for job placement in ascending order.
    * `compute_environment_order.#.order` - The order of the compute environment.
    * `compute_environment_order.#.compute_environment` - The ARN of the compute environment.
* `job_state_time_limit_action` - Specifies an action that AWS Batch will take after the job has remained at the head of the queue in the specified state for longer than the specified time.
    * `job_state_time_limit_action.#.action` - The action to take when a job is at the head of the job queue in the specified state for the specified period of time.
    * `job_state_time_limit_action.#.max_time_seconds` - The approximate amount of time, in seconds, that must pass with the job in the specified state before the action is taken.
    * `job_state_time_limit_action.#.reason` - The reason to log for the action being taken.
    * `job_state_time_limit_action.#.state` - The state of the job needed to trigger the action.
