---
subcategory: "Batch"
layout: "aws"
page_title: "AWS: aws_batch_job_queue"
description: |-
  Provides a Batch Job Queue resource.
---

# Resource: aws_batch_job_queue

Provides a Batch Job Queue resource.

## Example Usage

### Basic Job Queue

```terraform
resource "aws_batch_job_queue" "test_queue" {
  name     = "tf-test-batch-job-queue"
  state    = "ENABLED"
  priority = 1
  compute_environments = [
    aws_batch_compute_environment.test_environment_1.arn,
    aws_batch_compute_environment.test_environment_2.arn,
  ]
}
```

### Job Queue with a fair share scheduling policy

```terraform
resource "aws_batch_scheduling_policy" "example" {
  name = "example"

  fair_share_policy {
    compute_reservation = 1
    share_decay_seconds = 3600

    share_distribution {
      share_identifier = "A1*"
      weight_factor    = 0.1
    }
  }
}

resource "aws_batch_job_queue" "example" {
  name = "tf-test-batch-job-queue"

  scheduling_policy_arn = aws_batch_scheduling_policy.example.arn
  state                 = "ENABLED"
  priority              = 1

  compute_environments = [
    aws_batch_compute_environment.test_environment_1.arn,
    aws_batch_compute_environment.test_environment_2.arn,
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the job queue.
* `compute_environments` - (Required) Specifies the set of compute environments
    mapped to a job queue and their order.  The position of the compute environments
    in the list will dictate the order.
* `priority` - (Required) The priority of the job queue. Job queues with a higher priority
    are evaluated first when associated with the same compute environment.
* `scheduling_policy_arn` - (Optional) The ARN of the fair share scheduling policy. If this parameter is specified, the job queue uses a fair share scheduling policy. If this parameter isn't specified, the job queue uses a first in, first out (FIFO) scheduling policy. After a job queue is created, you can replace but can't remove the fair share scheduling policy.
* `state` - (Required) The state of the job queue. Must be one of: `ENABLED` or `DISABLED`
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name of the job queue.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

Batch Job Queue can be imported using the `arn`, e.g.,

```
$ terraform import aws_batch_job_queue.test_queue arn:aws:batch:us-east-1:123456789012:job-queue/sample
```
