---
subcategory: "Batch"
layout: "aws"
page_title: "AWS: aws_batch_compute_environment"
description: |-
  Creates a AWS Batch compute environment.
---

# Resource: aws_batch_compute_environment

Creates a AWS Batch compute environment. Compute environments contain the Amazon ECS container instances that are used to run containerized batch jobs.

For information about AWS Batch, see [What is AWS Batch?][1] .
For information about compute environment, see [Compute Environments][2] .

~> **Note:** To prevent a race condition during environment deletion, make sure to set `depends_on` to the related `aws_iam_role_policy_attachment`;
otherwise, the policy may be destroyed too soon and the compute environment will then get stuck in the `DELETING` state, see [Troubleshooting AWS Batch][3] .

## Example Usage

### EC2 Type

```terraform
data "aws_iam_policy_document" "ec2_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "ecs_instance_role" {
  name               = "ecs_instance_role"
  assume_role_policy = data.aws_iam_policy_document.ec2_assume_role.json
}

resource "aws_iam_role_policy_attachment" "ecs_instance_role" {
  role       = aws_iam_role.ecs_instance_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance_role" {
  name = "ecs_instance_role"
  role = aws_iam_role.ecs_instance_role.name
}

data "aws_iam_policy_document" "batch_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["batch.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "aws_batch_service_role" {
  name               = "aws_batch_service_role"
  assume_role_policy = data.aws_iam_policy_document.batch_assume_role.json
}

resource "aws_iam_role_policy_attachment" "aws_batch_service_role" {
  role       = aws_iam_role.aws_batch_service_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSBatchServiceRole"
}

resource "aws_security_group" "sample" {
  name = "aws_batch_compute_environment_security_group"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_vpc" "sample" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "sample" {
  vpc_id     = aws_vpc.sample.id
  cidr_block = "10.1.1.0/24"
}

resource "aws_placement_group" "sample" {
  name     = "sample"
  strategy = "cluster"
}

resource "aws_batch_compute_environment" "sample" {
  compute_environment_name = "sample"

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance_role.arn

    instance_type = [
      "c4.large",
    ]

    max_vcpus = 16
    min_vcpus = 0

    placement_group = aws_placement_group.sample.name

    security_group_ids = [
      aws_security_group.sample.id,
    ]

    subnets = [
      aws_subnet.sample.id,
    ]

    type = "EC2"
  }

  service_role = aws_iam_role.aws_batch_service_role.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.aws_batch_service_role]
}
```

### Fargate Type

```hcl
resource "aws_batch_compute_environment" "sample" {
  compute_environment_name = "sample"

  compute_resources {
    max_vcpus = 16

    security_group_ids = [
      aws_security_group.sample.id
    ]

    subnets = [
      aws_subnet.sample.id
    ]

    type = "FARGATE"
  }

  service_role = aws_iam_role.aws_batch_service_role.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.aws_batch_service_role]
}
```

### Setting Update Policy

```hcl
resource "aws_batch_compute_environment" "sample" {
  compute_environment_name = "sample"

  compute_resources {
    allocation_strategy = "BEST_FIT_PROGRESSIVE"
    instance_role       = aws_iam_instance_profile.ecs_instance.arn
    instance_type       = ["optimal"]
    max_vcpus           = 4
    min_vcpus           = 0
    security_group_ids = [
      aws_security_group.sample.id
    ]
    subnets = [
      aws_subnet.sample.id
    ]
    type = "EC2"
  }

  update_policy {
    job_execution_timeout_minutes = 30
    terminate_jobs_on_update      = false
  }

  type = "MANAGED"
}
```

## Argument Reference

* `compute_environment_name` - (Optional, Forces new resource) The name for your compute environment. Up to 128 letters (uppercase and lowercase), numbers, and underscores are allowed. If omitted, Terraform will assign a random, unique name.
* `compute_environment_name_prefix` - (Optional, Forces new resource) Creates a unique compute environment name beginning with the specified prefix. Conflicts with `compute_environment_name`.
* `compute_resources` - (Optional) Details of the compute resources managed by the compute environment. This parameter is required for managed compute environments. See details below.
* `eks_configuration` - (Optional) Details for the Amazon EKS cluster that supports the compute environment. See details below.
* `service_role` - (Optional) The full Amazon Resource Name (ARN) of the IAM role that allows AWS Batch to make calls to other AWS services on your behalf.
* `state` - (Optional) The state of the compute environment. If the state is `ENABLED`, then the compute environment accepts jobs from a queue and can scale out automatically based on queues. Valid items are `ENABLED` or `DISABLED`. Defaults to `ENABLED`.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `type` - (Required) The type of the compute environment. Valid items are `MANAGED` or `UNMANAGED`.
* `update_policy` - (Optional) Specifies the infrastructure update policy for the compute environment. See details below.

### compute_resources

* `allocation_strategy` - (Optional) The allocation strategy to use for the compute resource in case not enough instances of the best fitting instance type can be allocated. For valid values, refer to the [AWS documentation](https://docs.aws.amazon.com/batch/latest/APIReference/API_ComputeResource.html#Batch-Type-ComputeResource-allocationStrategy). Defaults to `BEST_FIT`. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified.
* `bid_percentage` - (Optional) Integer of maximum percentage that a Spot Instance price can be when compared with the On-Demand price for that instance type before instances are launched. For example, if your bid percentage is 20% (`20`), then the Spot price must be below 20% of the current On-Demand price for that EC2 instance. If you leave this field empty, the default value is 100% of the On-Demand price. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified.
* `desired_vcpus` - (Optional) The desired number of EC2 vCPUS in the compute environment. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified.
* `ec2_configuration` - (Optional) Provides information used to select Amazon Machine Images (AMIs) for EC2 instances in the compute environment. If Ec2Configuration isn't specified, the default is ECS_AL2. This parameter isn't applicable to jobs that are running on Fargate resources, and shouldn't be specified.
* `ec2_key_pair` - (Optional) The EC2 key pair that is used for instances launched in the compute environment. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified.
* `image_id` - (Optional) The Amazon Machine Image (AMI) ID used for instances launched in the compute environment. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified. (Deprecated, use [`ec2_configuration`](#ec2_configuration) `image_id_override` instead)
* `instance_role` - (Optional) The Amazon ECS instance role applied to Amazon EC2 instances in a compute environment. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified.
* `instance_type` - (Optional) A list of instance types that may be launched. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified.
* `launch_template` - (Optional) The launch template to use for your compute resources. See details below. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified.
* `max_vcpus` - (Required) The maximum number of EC2 vCPUs that an environment can reach.
* `min_vcpus` - (Optional) The minimum number of EC2 vCPUs that an environment should maintain. For `EC2` or `SPOT` compute environments, if the parameter is not explicitly defined, a `0` default value will be set. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified.
* `placement_group` - (Optional) The Amazon EC2 placement group to associate with your compute resources.
* `security_group_ids` - (Optional) A list of EC2 security group that are associated with instances launched in the compute environment. This parameter is required for Fargate compute environments.
* `spot_iam_fleet_role` - (Optional) The Amazon Resource Name (ARN) of the Amazon EC2 Spot Fleet IAM role applied to a SPOT compute environment. This parameter is required for SPOT compute environments. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified.
* `subnets` - (Required) A list of VPC subnets into which the compute resources are launched.
* `tags` - (Optional) Key-value pair tags to be applied to resources that are launched in the compute environment. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified.
* `type` - (Required) The type of compute environment. Valid items are `EC2`, `SPOT`, `FARGATE` or `FARGATE_SPOT`.

### ec2_configuration

`ec2_configuration` supports the following:

* `image_id_override` - (Optional) The AMI ID used for instances launched in the compute environment that match the image type. This setting overrides the `image_id` argument in the [`compute_resources`](#compute_resources) block.
* `image_type` - (Optional) The image type to match with the instance type to select an AMI. If the `image_id_override` parameter isn't specified, then a recent [Amazon ECS-optimized Amazon Linux 2 AMI](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-optimized_AMI.html#al2ami) (`ECS_AL2`) is used.

### launch_template

`launch_template` supports the following:

* `launch_template_id` - (Optional) ID of the launch template. You must specify either the launch template ID or launch template name in the request, but not both.
* `launch_template_name` - (Optional) Name of the launch template.
* `version` - (Optional) The version number of the launch template. Default: The default version of the launch template.

### eks_configuration

`eks_configuration` supports the following:

* `eks_cluster_arn` - (Required) The Amazon Resource Name (ARN) of the Amazon EKS cluster.
* `kubernetes_namespace` - (Required) The namespace of the Amazon EKS cluster. AWS Batch manages pods in this namespace.

### update_policy

`update_policy` supports the following:

* `job_execution_timeout_minutes` - (Required) Specifies the job timeout (in minutes) when the compute environment infrastructure is updated.
* `terminate_jobs_on_update` - (Required) Specifies whether jobs are automatically terminated when the computer environment infrastructure is updated.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the compute environment.
* `ecs_cluster_arn` - The Amazon Resource Name (ARN) of the underlying Amazon ECS cluster used by the compute environment.
* `status` - The current status of the compute environment (for example, CREATING or VALID).
* `status_reason` - A short, human-readable string to provide additional details about the current status of the compute environment.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS Batch compute using the `compute_environment_name`. For example:

```terraform
import {
  to = aws_batch_compute_environment.sample
  id = "sample"
}
```

Using `terraform import`, import AWS Batch compute using the `compute_environment_name`. For example:

```console
% terraform import aws_batch_compute_environment.sample sample
```

[1]: http://docs.aws.amazon.com/batch/latest/userguide/what-is-batch.html
[2]: http://docs.aws.amazon.com/batch/latest/userguide/compute_environments.html
[3]: http://docs.aws.amazon.com/batch/latest/userguide/troubleshooting.html
