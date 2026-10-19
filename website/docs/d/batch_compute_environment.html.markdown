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
  name = "batch-mongo-production"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the Batch Compute Environment
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the compute environment.
* `ecs_cluster_arn` - ARN of the underlying Amazon ECS cluster used by the compute environment.
* `service_role` - ARN of the IAM role that allows AWS Batch to make calls to other AWS services on your behalf.
* `state` - State of the compute environment (for example, `ENABLED` or `DISABLED`). If the state is `ENABLED`, then the compute environment accepts jobs from a queue and can scale out automatically based on queues.
* `status` - Current status of the compute environment (for example, `CREATING` or `VALID`).
* `status_reason` - Short, human-readable string to provide additional details about the current status of the compute environment.
* `tags` - Key-value map of resource tags
* `type` - Type of the compute environment (for example, `MANAGED` or `UNMANAGED`).
* `update_policy` - Infrastructure update policy for the compute environment.

### `update_policy` Block

* `job_execution_timeout_minutes` - Time, in minutes, that a job can run before the compute environment infrastructure is updated.
* `terminate_jobs_on_update` - Whether running jobs are terminated when the compute environment infrastructure is updated.
