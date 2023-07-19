---
subcategory: "Batch"
layout: "aws"
page_title: "AWS: aws_batch_job_definition"
description: |-
  Provides a Batch Job Definition resource.
---

# Resource: aws_batch_job_definition

Provides a Batch Job Definition resource.

## Example Usage

```terraform
resource "aws_batch_job_definition" "test" {
  name = "tf_test_batch_job_definition"
  type = "container"
  container_properties = jsonencode({
    command = ["ls", "-la"],
    image   = "busybox"

    resourceRequirements = [
      {
        type  = "VCPU"
        value = "0.25"
      },
      {
        type  = "MEMORY"
        value = "512"
      }
    ]

    volumes = [
      {
        host = {
          sourcePath = "/tmp"
        }
        name = "tmp"
      }
    ]

    environment = [
      {
        name  = "VARNAME"
        value = "VARVAL"
      }
    ]

    mountPoints = [
      {
        sourceVolume  = "tmp"
        containerPath = "/tmp"
        readOnly      = false
      }
    ]

    ulimits = [
      {
        hardLimit = 1024
        name      = "nofile"
        softLimit = 1024
      }
    ]
  })
}
```

### Fargate Platform Capability

```terraform
resource "aws_iam_role" "ecs_task_execution_role" {
  name               = "tf_test_batch_exec_role"
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_batch_job_definition" "test" {
  name = "tf_test_batch_job_definition"
  type = "container"

  platform_capabilities = [
    "FARGATE",
  ]

  container_properties = jsonencode({
    command    = ["echo", "test"]
    image      = "busybox"
    jobRoleArn = "arn:aws:iam::123456789012:role/AWSBatchS3ReadOnly"

    fargatePlatformConfiguration = {
      platformVersion = "LATEST"
    }

    resourceRequirements = [
      {
        type  = "VCPU"
        value = "0.25"
      },
      {
        type  = "MEMORY"
        value = "512"
      }
    ]

    executionRoleArn = aws_iam_role.ecs_task_execution_role.arn
  })
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Specifies the name of the job definition.
* `type` - (Required) The type of job definition. Must be `container`.

The following arguments are optional:

* `container_properties` - (Optional) A valid [container properties](http://docs.aws.amazon.com/batch/latest/APIReference/API_RegisterJobDefinition.html)
    provided as a single valid JSON document. This parameter is required if the `type` parameter is `container`.
* `parameters` - (Optional) Specifies the parameter substitution placeholders to set in the job definition.
* `platform_capabilities` - (Optional) The platform capabilities required by the job definition. If no value is specified, it defaults to `EC2`. To run the job on Fargate resources, specify `FARGATE`.
* `propagate_tags` - (Optional) Specifies whether to propagate the tags from the job definition to the corresponding Amazon ECS task. Default is `false`.
* `retry_strategy` - (Optional) Specifies the retry strategy to use for failed jobs that are submitted with this job definition.
    Maximum number of `retry_strategy` is `1`.  Defined below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `timeout` - (Optional) Specifies the timeout for jobs so that if a job runs longer, AWS Batch terminates the job. Maximum number of `timeout` is `1`. Defined below.

### retry_strategy

* `attempts` - (Optional) The number of times to move a job to the `RUNNABLE` status. You may specify between `1` and `10` attempts.
* `evaluate_on_exit` - (Optional) The [evaluate on exit](#evaluate_on_exit) conditions under which the job should be retried or failed. If this parameter is specified, then the `attempts` parameter must also be specified. You may specify up to 5 configuration blocks.

#### evaluate_on_exit

* `action` - (Required) Specifies the action to take if all of the specified conditions are met. The values are not case sensitive. Valid values: `RETRY`, `EXIT`.
* `on_exit_code` - (Optional) A glob pattern to match against the decimal representation of the exit code returned for a job.
* `on_reason` - (Optional) A glob pattern to match against the reason returned for a job.
* `on_status_reason` - (Optional) A glob pattern to match against the status reason returned for a job.
  
### timeout

* `attempt_duration_seconds` - (Optional) The time duration in seconds after which AWS Batch terminates your jobs if they have not finished. The minimum value for the timeout is `60` seconds.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name of the job definition.
* `revision` - The revision of the job definition.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Batch Job Definition using the `arn`. For example:

```terraform
import {
  to = aws_batch_job_definition.test
  id = "arn:aws:batch:us-east-1:123456789012:job-definition/sample"
}
```

Using `terraform import`, import Batch Job Definition using the `arn`. For example:

```console
% terraform import aws_batch_job_definition.test arn:aws:batch:us-east-1:123456789012:job-definition/sample
```
