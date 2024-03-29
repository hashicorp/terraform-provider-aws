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

  compute_environment_order {
    order               = 1
    compute_environment = aws_batch_compute_environment.test_environment_1.arn
  }

  compute_environment_order {
    order               = 2
    compute_environment = aws_batch_compute_environment.test_environment_2.arn
  }
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

  compute_environment_order {
    order               = 1
    compute_environment = aws_batch_compute_environment.test_environment_1.arn
  }

  compute_environment_order {
    order               = 2
    compute_environment = aws_batch_compute_environment.test_environment_2.arn
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Specifies the name of the job queue.
* `compute_environments` - (Deprecated) (Optional) This parameter is deprecated, please use `compute_environment_order` instead. List of compute environment ARNs mapped to a job queue. The position of the compute environments in the list will dictate the order. When importing a AWS Batch Job Queue, the parameter `compute_environments` will always be used over `compute_environment_order`. Please adjust your HCL accordingly.
* `compute_environment_order` - (Optional) The set of compute environments mapped to a job queue and their order relative to each other. The job scheduler uses this parameter to determine which compute environment runs a specific job. Compute environments must be in the VALID state before you can associate them with a job queue. You can associate up to three compute environments with a job queue.  
* `priority` - (Required) The priority of the job queue. Job queues with a higher priority
    are evaluated first when associated with the same compute environment.
* `scheduling_policy_arn` - (Optional) The ARN of the fair share scheduling policy. If this parameter is specified, the job queue uses a fair share scheduling policy. If this parameter isn't specified, the job queue uses a first in, first out (FIFO) scheduling policy. After a job queue is created, you can replace but can't remove the fair share scheduling policy.
* `state` - (Required) The state of the job queue. Must be one of: `ENABLED` or `DISABLED`
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### compute_environment_order

* `compute_environment` - (Required) The Amazon Resource Name (ARN) of the compute environment.
* `order` - (Required) The order of the compute environment. Compute environments are tried in ascending order. For example, if two compute environments are associated with a job queue, the compute environment with a lower order integer value is tried for job placement first.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name of the job queue.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `update` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Batch Job Queue using the `arn`. For example:

```terraform
import {
  to = aws_batch_job_queue.test_queue
  id = "arn:aws:batch:us-east-1:123456789012:job-queue/sample"
}
```

Using `terraform import`, import Batch Job Queue using the `arn`. For example:

```console
% terraform import aws_batch_job_queue.test_queue arn:aws:batch:us-east-1:123456789012:job-queue/sample
```
