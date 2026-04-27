---
subcategory: "ARC (Application Recovery Controller) Region Switch"
layout: "aws"
page_title: "AWS: aws_arcregionswitch_plan"
description: |-
  Terraform resource for managing an Amazon ARC Region Switch Plan.
---

# Resource: aws_arcregionswitch_plan

Terraform resource for managing an Amazon ARC Region Switch plan.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_role" "example" {
  name = "arc-region-switch-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "arc-region-switch.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_arcregionswitch_plan" "example" {
  name              = "example-plan"
  execution_role    = aws_iam_role.example.arn
  recovery_approach = "activePassive"
  regions           = ["us-east-1", "us-west-2"]
  primary_region    = "us-east-1"

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = "us-west-2"

    step {
      name                 = "manual-approval"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role   = aws_iam_role.example.arn
        timeout_minutes = 60
      }
    }
  }

  workflow {
    workflow_target_action = "deactivate"
    workflow_target_region = "us-east-1"

    step {
      name                 = "manual-approval"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role   = aws_iam_role.example.arn
        timeout_minutes = 60
      }
    }
  }
}
```

### Complex Usage with Multiple Step Types

```terraform
resource "aws_arcregionswitch_plan" "complex" {
  name                            = "complex-plan"
  execution_role                  = aws_iam_role.example.arn
  recovery_approach               = "activeActive"
  regions                         = ["us-east-1", "us-west-2"]
  description                     = "Complex plan with multiple execution block types"
  recovery_time_objective_minutes = 60

  associated_alarms {
    name                = "application-health-alarm"
    alarm_type          = "applicationHealth"
    resource_identifier = "arn:aws:cloudwatch:us-east-1:123456789012:alarm:MyAlarm"
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = "us-west-2"

    step {
      name                 = "lambda-step"
      execution_block_type = "CustomActionLambda"

      custom_action_lambda_config {
        region_to_run          = "activatingRegion"
        retry_interval_minutes = 5.0
        timeout_minutes        = 30

        lambda {
          arn = aws_lambda_function.example.arn
        }
      }
    }

    step {
      name                 = "parallel-step"
      execution_block_type = "Parallel"

      parallel_config {
        step {
          name                 = "asg-scaling"
          execution_block_type = "EC2AutoScaling"

          ec2_asg_capacity_increase_config {
            asg {
              arn = aws_autoscaling_group.example.arn
            }
            target_percent = 150
          }
        }

        step {
          name                 = "ecs-scaling"
          execution_block_type = "ECSServiceScaling"

          ecs_capacity_increase_config {
            service {
              cluster_arn = aws_ecs_cluster.example.arn
              service_arn = aws_ecs_service.example.arn
            }
            target_percent = 200
          }
        }
      }
    }
  }

  workflow {
    workflow_target_action = "deactivate"
    workflow_target_region = "us-east-1"

    step {
      name                 = "route53-health-check"
      execution_block_type = "Route53HealthCheck"

      route53_health_check_config {
        hosted_zone_id = aws_route53_zone.example.zone_id
        record_name    = "api.example.com"
      }
    }
  }

  triggers {
    action                               = "activate"
    target_region                        = "us-west-2"
    min_delay_minutes_between_executions = 30

    conditions {
      associated_alarm_name = "application-health-alarm"
      condition             = "red"
    }
  }

  tags = {
    Environment = "production"
  }
}
```

## Argument Reference

The following arguments are required:

* `execution_role` - (Required) ARN of the IAM role that ARC Region Switch will assume to execute the plan.
* `name` - (Required) Name of the plan. Must be unique within the account.
* `recovery_approach` - (Required) Recovery approach for the plan. Valid values: `activeActive`, `activePassive`.
* `regions` - (Required) List of AWS regions involved in the plan.
* `workflow` - (Required) List of workflows that define the steps to execute. See [Workflow](#workflow) below.

The following arguments are optional:

* `associated_alarms` - (Optional) Set of CloudWatch alarms associated with the plan. See [Associated Alarms](#associated-alarms) below.
* `description` - (Optional) Description of the plan.
* `primary_region` - (Optional) Primary region for the plan.
* `recovery_time_objective_minutes` - (Optional) Recovery time objective in minutes.
* `region` - (Optional, **Deprecated**) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `triggers` - (Optional) Set of triggers that can initiate the plan execution. See [Triggers](#triggers) below.

### Workflow

* `step` - (Optional) List of steps in the workflow. See [Step](#step) below.
* `workflow_description` - (Optional) Description of the workflow.
* `workflow_target_action` - (Required) Action to perform. Valid values: `activate`, `deactivate`.
* `workflow_target_region` - (Optional) Target region for the workflow.

### Step

* `arc_routing_control_config` - Configuration for ARC routing control. See [ARC Routing Control Config](#arc-routing-control-config) below.
* `custom_action_lambda_config` - Configuration for Lambda function execution. See [Custom Action Lambda Config](#custom-action-lambda-config) below.
* `description` - (Optional) Description of the step.
* `document_db_config` - Configuration for DocumentDB global cluster operations. See [DocumentDB Config](#documentdb-config) below.
* `ec2_asg_capacity_increase_config` - Configuration for EC2 Auto Scaling group capacity increase. See [EC2 ASG Capacity Increase Config](#ec2-asg-capacity-increase-config) below.
* `ecs_capacity_increase_config` - Configuration for ECS service capacity increase. See [ECS Capacity Increase Config](#ecs-capacity-increase-config) below.
* `eks_resource_scaling_config` - Configuration for EKS resource scaling. See [EKS Resource Scaling Config](#eks-resource-scaling-config) below.
* `execution_approval_config` - Configuration for manual approval steps. See [Execution Approval Config](#execution-approval-config) below.
* `execution_block_type` - (Required) Type of execution block. Valid values: `ARCRegionSwitchPlan`, `ARCRoutingControl`, `AuroraGlobalDatabase`, `CustomActionLambda`, `DocumentDb`, `EC2AutoScaling`, `ECSServiceScaling`, `EKSResourceScaling`, `ManualApproval`, `Parallel`, `Route53HealthCheck`.
* `global_aurora_config` - Configuration for Aurora Global Database operations. See [Global Aurora Config](#global-aurora-config) below.
* `name` - (Required) Name of the step.
* `parallel_config` - Configuration for parallel execution of multiple steps. See [Parallel Config](#parallel-config) below.
* `route53_health_check_config` - Configuration for Route53 health check operations. See [Route53 Health Check Config](#route53-health-check-config) below.

### ARC Routing Control Config

* `region_and_routing_controls` - (Required) List of regions and their routing controls. See [Region and Routing Controls](#region-and-routing-controls) below.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.
* `timeout_minutes` - (Optional) Timeout in minutes.

### Region and Routing Controls

* `region` - (Required) AWS region.
* `routing_control` - (Required) List of routing controls. See [Routing Control](#routing-control) below.

### Routing Control

* `routing_control_arn` - (Required) ARN of the routing control.
* `state` - (Required) State of the routing control. Valid values: `On`, `Off`.

### Custom Action Lambda Config

* `region_to_run` - (Required) Region where the Lambda function should run. Valid values: `activatingRegion`, `deactivatingRegion`.
* `retry_interval_minutes` - (Required) Retry interval in minutes.
* `timeout_minutes` - (Optional) Timeout in minutes.
* `lambda` - (Required) Lambda function configuration. See [Lambda](#lambda) below.
* `ungraceful` - (Optional) Ungraceful behavior configuration. See [Ungraceful](#ungraceful) below.

### Lambda

* `arn` - (Required) ARN of the Lambda function.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.

### Ungraceful

* `behavior` - (Required) Behavior when ungraceful. Valid values: `skip`.

### DocumentDB Config

* `behavior` - (Required) Behavior for global cluster operations. Valid values: `switchoverOnly`, `failover`.
* `database_cluster_arns` - (Required) List of DocumentDB cluster ARNs.
* `global_cluster_identifier` - (Required) Global cluster identifier.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.
* `timeout_minutes` - (Optional) Timeout in minutes.
* `ungraceful` - (Optional) Ungraceful behavior configuration. See [DocumentDB Ungraceful](#documentdb-ungraceful) below.

### DocumentDB Ungraceful

* `ungraceful` - (Required) Ungraceful behavior. Valid values: `failover`.

### EC2 ASG Capacity Increase Config

* `asg` - (Required) Auto Scaling group configuration. See [ASG](#asg) below.
* `capacity_monitoring_approach` - (Required) Capacity monitoring approach. Valid values: `sampledMaxInLast24Hours`, `autoscalingMaxInLast24Hours`.
* `target_percent` - (Optional) Target capacity percentage.
* `timeout_minutes` - (Optional) Timeout in minutes.
* `ungraceful` - (Optional) Ungraceful behavior configuration. See [Ungraceful](#ungraceful) below.

### ASG

* `arn` - (Required) ARN of the Auto Scaling group.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.

### Ungraceful

* `minimum_success_percentage` - (Required) Minimum success percentage required.

### Execution Approval Config

* `approval_role` - (Required) ARN of the IAM role for approval.
* `timeout_minutes` - (Optional) Timeout in minutes for the approval.

### Associated Alarms

* `alarm_type` - (Required) Type of alarm. Valid values: `applicationHealth`, `trigger`.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.
* `map_block_key` - (Required) Name of the alarm.
* `resource_identifier` - (Required) Resource identifier (ARN) of the CloudWatch alarm.

### Triggers

* `action` - (Required) Action to trigger. Valid values: `activate`, `deactivate`.
* `conditions` - (Required) List of conditions that must be met. See [Conditions](#conditions) below.
* `min_delay_minutes_between_executions` - (Required) Minimum delay in minutes between executions.
* `target_region` - (Required) Target region for the trigger.
* `description` - (Optional) Description of the trigger.

### Conditions

* `associated_alarm_name` - (Required) Name of the associated alarm.
* `condition` - (Required) Condition to check. Valid values: `red`, `green`.

### Global Aurora Config

* `behavior` - (Required) Behavior for Aurora operations. Valid values: `switchoverOnly`, `failover`.
* `global_cluster_identifier` - (Required) Global cluster identifier.
* `database_cluster_arns` - (Required) List of database cluster ARNs.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.
* `timeout_minutes` - (Optional) Timeout in minutes.
* `ungraceful` - (Optional) Ungraceful behavior configuration. See [Ungraceful Aurora](#ungraceful-aurora) below.

### Ungraceful Aurora

* `ungraceful` - (Required) Ungraceful behavior. Valid values: `failover`.

### ECS Capacity Increase Config

* `service` - (Required) ECS service configuration. See [ECS Service](#ecs-service) below.
* `capacity_monitoring_approach` - (Required) Capacity monitoring approach. Valid values: `sampledMaxInLast24Hours`, `containerInsightsMaxInLast24Hours`.
* `target_percent` - (Optional) Target capacity percentage.
* `timeout_minutes` - (Optional) Timeout in minutes.
* `ungraceful` - (Optional) Ungraceful behavior configuration. See [Ungraceful Capacity](#ungraceful-capacity) below.

### ECS Service

* `cluster_arn` - (Required) ARN of the ECS cluster.
* `service_arn` - (Required) ARN of the ECS service.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.

### EKS Resource Scaling Config

* `capacity_monitoring_approach` - (Required) Capacity monitoring approach. Valid values: `sampledMaxInLast24Hours`, `autoscalingMaxInLast24Hours`.
* `eks_clusters` - (Optional) List of EKS clusters. See [EKS Clusters](#eks-clusters) below.
* `kubernetes_resource_type` - (Required) Kubernetes resource type. See [Kubernetes Resource Type](#kubernetes-resource-type) below.
* `scaling_resources` - (Optional) List of scaling resources. See [Scaling Resources](#scaling-resources) below.
* `target_percent` - (Required) Target capacity percentage.
* `timeout_minutes` - (Optional) Timeout in minutes.
* `ungraceful` - (Optional) Ungraceful behavior configuration. See [Ungraceful Capacity](#ungraceful-capacity) below.

### Kubernetes Resource Type

* `api_version` - (Required) Kubernetes API version.
* `kind` - (Required) Kubernetes resource kind.

### EKS Clusters

* `cluster_arn` - (Required) ARN of the EKS cluster.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.

### Scaling Resources

* `namespace` - (Required) Kubernetes namespace.
* `resources` - (Required) Set of resources to scale. See [Resources](#resources) below.

### Resources

* `resource_name` - (Required) Name of the resource.
* `name` - (Required) Name of the Kubernetes object.
* `namespace` - (Required) Kubernetes namespace.
* `hpa_name` - (Optional) Name of the Horizontal Pod Autoscaler.

### Route53 Health Check Config

* `hosted_zone_id` - (Required) Route53 hosted zone ID.
* `record_name` - (Required) DNS record name.
* `record_set` - (Optional) Configuration block for record sets. See [Record Set](#record-set) below.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.
* `timeout_minutes` - (Optional) Timeout in minutes.

### Record Set

* `record_set_identifier` - (Required) Record set identifier.
* `region` - (Required) AWS region.

### Region Switch Plan Config

* `arn` - (Required) ARN of the nested region switch plan.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.

### Parallel Config

* `step` - (Required) List of steps to execute in parallel. Uses the same schema as [Step](#step) but without `parallel_config` to prevent infinite nesting.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the plan.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `update` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Application Recovery Controller Region Switch Plan using the `arn`. For example:

```terraform
import {
  to = aws_arcregionswitch_plan.example
  id = "arn:aws:arcregionswitch:us-east-1:123456789012:plan/example-plan"
}
```

Using `terraform import`, import Application Recovery Controller Region Switch Plan using the `arn`. For example:

```console
% terraform import aws_arcregionswitch_plan.example arn:aws:arcregionswitch:us-east-1:123456789012:plan/example-plan
```
