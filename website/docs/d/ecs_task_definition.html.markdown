---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_task_definition"
description: |-
    Provides details about an ecs task definition
---

# Data Source: aws_ecs_task_definition

The ECS task definition data source allows access to details of
a specific AWS ECS task definition.

## Example Usage

```terraform
# Simply specify the family to find the latest ACTIVE revision in that family.
data "aws_ecs_task_definition" "mongo" {
  task_definition = aws_ecs_task_definition.mongo.family
}

resource "aws_ecs_cluster" "foo" {
  name = "foo"
}

resource "aws_ecs_task_definition" "mongo" {
  family = "mongodb"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "environment": [{
      "name": "SECRET",
      "value": "KEY"
    }],
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "memoryReservation": 64,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "mongo" {
  name          = "mongo"
  cluster       = aws_ecs_cluster.foo.id
  desired_count = 2

  # Track the latest ACTIVE revision
  task_definition = data.aws_ecs_task_definition.mongo.arn
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `task_definition` - (Required) Family for the latest ACTIVE revision, family and revision (family:revision) for a specific revision in the family, the ARN of the task definition to access to.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the task definition.
* `arn_without_revision` - ARN of the Task Definition with the trailing `revision` removed. This may be useful for situations where the latest task definition is always desired. If a revision isn't specified, the latest ACTIVE revision is used. See the [AWS documentation](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_StartTask.html#ECS-StartTask-request-taskDefinition) for details.
* `container_definitions` - A list of valid [container definitions](http://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_ContainerDefinition.html) provided as a single valid JSON document. Please note that you should only provide values that are part of the container definition document. For a detailed description of what parameters are available, see the [Task Definition Parameters](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html) section from the official [Developer Guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide).
* `cpu` - Number of cpu units used by the task. If the `requires_compatibilities` is `FARGATE` this field is required.
* `enable_fault_injection` - Enables fault injection and allows for fault injection requests to be accepted from the task's containers. Default is `false`.
* `ephemeral_storage` - The amount of ephemeral storage to allocate for the task. This parameter is used to expand the total amount of ephemeral storage available, beyond the default amount, for tasks hosted on AWS Fargate. See [Ephemeral Storage](#ephemeral_storage).
* `execution_role_arn` - ARN of the task execution role that the Amazon ECS container agent and the Docker daemon can assume.
* `family` - A unique name for your task definition.
The following arguments are optional:
* `ipc_mode` - IPC resource namespace to be used for the containers in the task The valid values are `host`, `task`, and `none`.
* `memory` - Amount (in MiB) of memory used by the task. If the `requires_compatibilities` is `FARGATE` this field is required.
* `network_mode` - Docker networking mode to use for the containers in the task. Valid values are `none`, `bridge`, `awsvpc`, and `host`.
* `pid_mode` - Process namespace to use for the containers in the task. The valid values are `host` and `task`.
* `placement_constraints` - Configuration block for rules that are taken into consideration during task placement. Maximum number of `placement_constraints` is `10`. [Detailed below](#placement_constraints).
* `proxy_configuration` - Configuration block for the App Mesh proxy. [Detailed below.](#proxy_configuration)
* `requires_compatibilities` - Set of launch types required by the task. The valid values are `EC2` and `FARGATE`.
* `revision` - Revision of the task in a particular family.
* `runtime_platform` - Configuration block for [runtime_platform](#runtime_platform) that containers in your task may use.
* `status` - Status of the task definition.
* `task_role_arn` - ARN of IAM role that allows your Amazon ECS container task to make calls to other AWS services.
* `volume` - Attributes corresponding to the `volume` argument of the [`aws_ecs_task_definition`](/docs/providers/aws/r/ecs_task_definition.html) resource.

### ephemeral_storage

* `size_in_gib` - The total amount, in GiB, of ephemeral storage to set for the task. The minimum supported value is `21` GiB and the maximum supported value is `200` GiB.

### placement_constraints

* `expression` -  Cluster Query Language expression to apply to the constraint. For more information, see [Cluster Query Language in the Amazon EC2 Container Service Developer Guide](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/cluster-query-language.html).
* `type` - Type of constraint. Use `memberOf` to restrict selection to a group of valid candidates. Note that `distinctInstance` is not supported in task definitions.

### proxy_configuration

* `container_name` - Name of the container that will serve as the App Mesh proxy.
* `properties` - Set of network configuration parameters to provide the Container Network Interface (CNI) plugin, specified a key-value mapping.
* `type` - Proxy type. The default value is `APPMESH`. The only supported value is `APPMESH`.

### runtime_platform

* `operating_system_family` - If the `requires_compatibilities` is `FARGATE` this field is required; must be set to a valid option from the [operating system family in the runtime platform](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html#runtime-platform) setting
* `cpu_architecture` - Must be set to either `X86_64` or `ARM64`; see [cpu architecture](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html#runtime-platform)
