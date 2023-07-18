---
subcategory: "Batch"
layout: "aws"
page_title: "AWS: aws_batch_compute_environment"
description: |-
    Provides details about a batch compute environment
---

# Data Source: aws_batch_compute_environment

The Batch Compute Environment data source allows access to details of a specific
compute environment within AWS Batch.

## Example Usage

```terraform
data "aws_batch_compute_environment" "batch-mongo" {
  compute_environment_name = "batch-mongo-production"
}
```

## Argument Reference

This data source supports the following arguments:

* `compute_environment_name` - (Required) Name of the Batch Compute Environment

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the compute environment.
* `ecs_cluster_arn` - ARN of the underlying Amazon ECS cluster used by the compute environment.
* `service_role` - ARN of the IAM role that allows AWS Batch to make calls to other AWS services on your behalf.
* `type` - Type of the compute environment (for example, `MANAGED` or `UNMANAGED`).
* `status` - Current status of the compute environment (for example, `CREATING` or `VALID`).
* `status_reason` - Short, human-readable string to provide additional details about the current status of the compute environment.
* `state` - State of the compute environment (for example, `ENABLED` or `DISABLED`). If the state is `ENABLED`, then the compute environment accepts jobs from a queue and can scale out automatically based on queues.
* `tags` - Key-value map of resource tags
