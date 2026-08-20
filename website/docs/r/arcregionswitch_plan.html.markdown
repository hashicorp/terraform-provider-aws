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
* `regions` - (Required) List of AWS regions involved in the plan. Must contain at least 2 regions.
* `workflow` - (Required) Workflows that define the steps to execute. See [`workflow` Block](#workflow-block) for details.

The following arguments are optional:

* `associated_alarms` - (Optional) CloudWatch alarms associated with the plan. See [`associated_alarms` Block](#associated_alarms-block) for details.
* `description` - (Optional) Description of the plan.
* `primary_region` - (Optional) Primary region for the plan.
* `recovery_time_objective_minutes` - (Optional) Recovery time objective in minutes.
* `region` - (Optional, **Deprecated**) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `report_configuration` - (Optional) Configuration for automated execution reports. See [`report_configuration` Block](#report_configuration-block) for details.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `triggers` - (Optional) Triggers that can initiate the plan execution. See [`triggers` Block](#triggers-block) for details.

### `workflow` Block

* `step` - (Optional) Steps in the workflow. See [`step` Block](#step-block) for details.
* `workflow_description` - (Optional) Description of the workflow.
* `workflow_target_action` - (Required) Action to perform. Valid values: `activate`, `deactivate`.
* `workflow_target_region` - (Optional) Target region for the workflow.

### `step` Block

* `arc_routing_control_config` - (Optional) Configuration for ARC routing control. See [`arc_routing_control_config` Block](#arc_routing_control_config-block) for details.
* `aurora_provisioned_scaling_config` - (Optional) Configuration for Aurora provisioned scaling. See [`aurora_provisioned_scaling_config` Block](#aurora_provisioned_scaling_config-block) for details.
* `aurora_serverless_scaling_config` - (Optional) Configuration for Aurora Serverless scaling. See [`aurora_serverless_scaling_config` Block](#aurora_serverless_scaling_config-block) for details.
* `custom_action_lambda_config` - (Optional) Configuration for Lambda function execution. See [`custom_action_lambda_config` Block](#custom_action_lambda_config-block) for details.
* `description` - (Optional) Description of the step.
* `document_db_config` - (Optional) Configuration for DocumentDB global cluster operations. See [`document_db_config` Block](#document_db_config-block) for details.
* `ec2_asg_capacity_increase_config` - (Optional) Configuration for EC2 Auto Scaling group capacity increase. See [`ec2_asg_capacity_increase_config` Block](#ec2_asg_capacity_increase_config-block) for details.
* `ecs_capacity_increase_config` - (Optional) Configuration for ECS service capacity increase. See [`ecs_capacity_increase_config` Block](#ecs_capacity_increase_config-block) for details.
* `eks_resource_scaling_config` - (Optional) Configuration for EKS resource scaling. See [`eks_resource_scaling_config` Block](#eks_resource_scaling_config-block) for details.
* `execution_approval_config` - (Optional) Configuration for manual approval steps. See [`execution_approval_config` Block](#execution_approval_config-block) for details.
* `execution_block_type` - (Required) Type of execution block. Valid values: `ARCRegionSwitchPlan`, `ARCRoutingControl`, `AuroraGlobalDatabase`, `CustomActionLambda`, `DocumentDb`, `EC2AutoScaling`, `ECSServiceScaling`, `EKSResourceScaling`, `ManualApproval`, `Parallel`, `RdsCreateCrossRegionReplica`, `RdsPromoteReadReplica`, `Route53HealthCheck`.
* `global_aurora_config` - (Optional) Configuration for Aurora Global Database operations. See [`global_aurora_config` Block](#global_aurora_config-block) for details.
* `lambda_event_source_mapping_config` - (Optional) Configuration for Lambda event source mapping operations. See [`lambda_event_source_mapping_config` Block](#lambda_event_source_mapping_config-block) for details.
* `name` - (Required) Name of the step.
* `neptune_global_database_config` - (Optional) Configuration for Neptune global database operations. See [`neptune_global_database_config` Block](#neptune_global_database_config-block) for details.
* `parallel_config` - (Optional) Configuration for parallel execution of multiple steps. See [`parallel_config` Block](#parallel_config-block) for details.
* `rds_create_cross_region_read_replica_config` - (Optional) Configuration for creating cross-region RDS read replicas. See [`rds_create_cross_region_read_replica_config` Block](#rds_create_cross_region_read_replica_config-block) for details.
* `rds_promote_read_replica_config` - (Optional) Configuration for promoting RDS read replicas. See [`rds_promote_read_replica_config` Block](#rds_promote_read_replica_config-block) for details.
* `region_switch_plan_config` - (Optional) Configuration for executing a nested region switch plan. See [`region_switch_plan_config` Block](#region_switch_plan_config-block) for details.
* `route53_health_check_config` - (Optional) Configuration for Route53 health check operations. See [`route53_health_check_config` Block](#route53_health_check_config-block) for details.

### `arc_routing_control_config` Block

* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.
* `region_and_routing_controls` - (Required) Regions and their routing controls. See [`region_and_routing_controls` Block](#region_and_routing_controls-block) for details.
* `timeout_minutes` - (Optional) Timeout in minutes.

### `region_and_routing_controls` Block

* `region` - (Required) AWS region.
* `routing_control` - (Required) Routing controls. See [`routing_control` Block](#routing_control-block) for details.

### `routing_control` Block

* `routing_control_arn` - (Required) ARN of the routing control.
* `state` - (Required) State of the routing control. Valid values: `On`, `Off`.

### `custom_action_lambda_config` Block

* `lambda` - (Required) Lambda function configuration. See [`lambda` Block](#lambda-block) for details.
* `region_to_run` - (Required) Region where the Lambda function should run. Valid values: `activatingRegion`, `deactivatingRegion`.
* `retry_interval_minutes` - (Required) Retry interval in minutes.
* `timeout_minutes` - (Optional) Timeout in minutes.
* `ungraceful` - (Optional) Ungraceful behavior configuration. See [`workflow.step.custom_action_lambda_config.ungraceful` Block](#workflowstepcustom_action_lambda_configungraceful-block) for details.

### `lambda` Block

* `arn` - (Required) ARN of the Lambda function.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.

### `workflow.step.custom_action_lambda_config.ungraceful` Block

* `behavior` - (Required) Ungraceful behavior. Valid values: `skip`.

### `document_db_config` Block

* `behavior` - (Required) Behavior for global cluster operations. Valid values: `switchoverOnly`, `failover`.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `database_cluster_arns` - (Required) List of DocumentDB cluster ARNs.
* `external_id` - (Optional) External ID for cross-account role assumption.
* `global_cluster_identifier` - (Required) Global cluster identifier.
* `timeout_minutes` - (Optional) Timeout in minutes.
* `ungraceful` - (Optional) Ungraceful behavior configuration. See [`workflow.step.document_db_config.ungraceful` Block](#workflowstepdocument_db_configungraceful-block) for details.

### `workflow.step.document_db_config.ungraceful` Block

* `ungraceful` - (Required) Ungraceful behavior. Valid values: `failover`.

### `ec2_asg_capacity_increase_config` Block

* `asg` - (Required) Auto Scaling group configuration. See [`asg` Block](#asg-block) for details.
* `capacity_monitoring_approach` - (Required) Capacity monitoring approach. Valid values: `sampledMaxInLast24Hours`, `autoscalingMaxInLast24Hours`.
* `target_percent` - (Optional) Target capacity percentage.
* `timeout_minutes` - (Optional) Timeout in minutes.
* `ungraceful` - (Optional) Ungraceful behavior configuration. See [`workflow.step.ec2_asg_capacity_increase_config.ungraceful` Block](#workflowstepec2_asg_capacity_increase_configungraceful-block) for details.

### `asg` Block

* `arn` - (Required) ARN of the Auto Scaling group.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.

### `workflow.step.ec2_asg_capacity_increase_config.ungraceful` Block

* `minimum_success_percentage` - (Required) Minimum success percentage required.

### `ecs_capacity_increase_config` Block

* `capacity_monitoring_approach` - (Required) Capacity monitoring approach. Valid values: `sampledMaxInLast24Hours`, `containerInsightsMaxInLast24Hours`.
* `service` - (Required) ECS service configuration. See [`service` Block](#service-block) for details.
* `target_percent` - (Optional) Target capacity percentage.
* `timeout_minutes` - (Optional) Timeout in minutes.
* `ungraceful` - (Optional) Ungraceful behavior configuration. See [`workflow.step.ecs_capacity_increase_config.ungraceful` Block](#workflowstepecs_capacity_increase_configungraceful-block) for details.

### `service` Block

* `cluster_arn` - (Required) ARN of the ECS cluster.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.
* `service_arn` - (Required) ARN of the ECS service.

### `workflow.step.ecs_capacity_increase_config.ungraceful` Block

* `minimum_success_percentage` - (Required) Minimum success percentage required.

### `eks_resource_scaling_config` Block

* `capacity_monitoring_approach` - (Required) Capacity monitoring approach. Valid values: `sampledMaxInLast24Hours`, `autoscalingMaxInLast24Hours`.
* `eks_clusters` - (Optional) EKS clusters. See [`eks_clusters` Block](#eks_clusters-block) for details.
* `kubernetes_resource_type` - (Required) Kubernetes resource type. See [`kubernetes_resource_type` Block](#kubernetes_resource_type-block) for details.
* `scaling_resources` - (Optional) Scaling resources. See [`scaling_resources` Block](#scaling_resources-block) for details.
* `target_percent` - (Required) Target capacity percentage.
* `timeout_minutes` - (Optional) Timeout in minutes.
* `ungraceful` - (Optional) Ungraceful behavior configuration. See [`workflow.step.eks_resource_scaling_config.ungraceful` Block](#workflowstepeks_resource_scaling_configungraceful-block) for details.

### `kubernetes_resource_type` Block

* `api_version` - (Required) Kubernetes API version.
* `kind` - (Required) Kubernetes resource kind.

### `eks_clusters` Block

* `cluster_arn` - (Required) ARN of the EKS cluster.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.

### `scaling_resources` Block

* `namespace` - (Required) Kubernetes namespace.
* `resources` - (Required) Resources to scale. See [`resources` Block](#resources-block) for details.

### `resources` Block

* `hpa_name` - (Optional) Name of the Horizontal Pod Autoscaler.
* `name` - (Required) Name of the Kubernetes object.
* `namespace` - (Required) Kubernetes namespace.
* `resource_name` - (Required) Name of the resource.

### `workflow.step.eks_resource_scaling_config.ungraceful` Block

* `minimum_success_percentage` - (Required) Minimum success percentage required.

### `execution_approval_config` Block

* `approval_role` - (Required) ARN of the IAM role for approval.
* `timeout_minutes` - (Optional) Timeout in minutes for the approval.

### `global_aurora_config` Block

* `behavior` - (Required) Behavior for Aurora operations. Valid values: `switchoverOnly`, `failover`.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `database_cluster_arns` - (Required) List of database cluster ARNs.
* `external_id` - (Optional) External ID for cross-account role assumption.
* `global_cluster_identifier` - (Required) Global cluster identifier.
* `timeout_minutes` - (Optional) Timeout in minutes.
* `ungraceful` - (Optional) Ungraceful behavior configuration. See [`workflow.step.global_aurora_config.ungraceful` Block](#workflowstepglobal_aurora_configungraceful-block) for details.

### `workflow.step.global_aurora_config.ungraceful` Block

* `ungraceful` - (Required) Ungraceful behavior. Valid values: `failover`.

### `aurora_provisioned_scaling_config` Block

* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.
* `global_cluster_identifier` - (Required) Global cluster identifier.
* `instance_arns` - (Required) Map of regions to Aurora instance ARNs.
* `region_database_cluster_arns` - (Required) Map of regions to database cluster ARNs.
* `timeout_minutes` - (Optional) Timeout in minutes.

### `aurora_serverless_scaling_config` Block

* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.
* `global_cluster_identifier` - (Required) Global cluster identifier.
* `region_database_cluster_arns` - (Required) Map of regions to database cluster ARNs.
* `target_percent` - (Optional) Target capacity percentage.
* `timeout_minutes` - (Optional) Timeout in minutes.

### `lambda_event_source_mapping_config` Block

* `action` - (Required) Action to perform on the event source mapping.
* `region_event_source_mapping` - (Optional) Event source mappings per region. See [`region_event_source_mapping` Block](#region_event_source_mapping-block) for details.
* `timeout_minutes` - (Optional) Timeout in minutes.
* `ungraceful` - (Optional) Ungraceful behavior configuration. See [`workflow.step.lambda_event_source_mapping_config.ungraceful` Block](#workflowsteplambda_event_source_mapping_configungraceful-block) for details.

### `region_event_source_mapping` Block

* `arn` - (Required) ARN of the event source mapping.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.
* `region` - (Required) AWS region.

### `workflow.step.lambda_event_source_mapping_config.ungraceful` Block

* `behavior` - (Required) Ungraceful behavior.

### `neptune_global_database_config` Block

* `behavior` - (Required) Behavior for global database operations.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.
* `global_cluster_identifier` - (Required) Global cluster identifier.
* `region_database_cluster_arns` - (Required) Map of regions to database cluster ARNs.
* `timeout_minutes` - (Optional) Timeout in minutes.
* `ungraceful` - (Optional) Ungraceful behavior configuration. See [`workflow.step.neptune_global_database_config.ungraceful` Block](#workflowstepneptune_global_database_configungraceful-block) for details.

### `workflow.step.neptune_global_database_config.ungraceful` Block

* `ungraceful` - (Required) Ungraceful behavior.

### `rds_create_cross_region_read_replica_config` Block

* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `db_instance_arn_map` - (Required) Map of source DB instance identifiers to target DB instance ARNs.
* `external_id` - (Optional) External ID for cross-account role assumption.
* `timeout_minutes` - (Optional) Timeout in minutes.

### `rds_promote_read_replica_config` Block

* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `db_instance_arn_map` - (Required) Map of source DB instance identifiers to target DB instance ARNs.
* `external_id` - (Optional) External ID for cross-account role assumption.
* `timeout_minutes` - (Optional) Timeout in minutes.

### `region_switch_plan_config` Block

* `arn` - (Required) ARN of the nested region switch plan.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.

### `parallel_config` Block

* `step` - (Required) Steps to execute in parallel. See [`step` Block](#step-block) for details. The parallel step schema matches [`step` Block](#step-block) but does not support `parallel_config` to prevent infinite nesting.

### `route53_health_check_config` Block

* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.
* `hosted_zone_id` - (Required) Route53 hosted zone ID.
* `record_name` - (Required) DNS record name.
* `record_set` - (Optional) Configuration block for record sets. See [`record_set` Block](#record_set-block) for details.
* `timeout_minutes` - (Optional) Timeout in minutes.

### `record_set` Block

* `record_set_identifier` - (Required) Record set identifier.
* `region` - (Required) AWS region.

### `associated_alarms` Block

* `alarm_type` - (Required) Type of alarm. Valid values: `applicationHealth`, `trigger`.
* `cross_account_role` - (Optional) ARN of the cross-account role to assume.
* `external_id` - (Optional) External ID for cross-account role assumption.
* `map_block_key` - (Required) Name of the alarm.
* `resource_identifier` - (Required) Resource identifier (ARN) of the CloudWatch alarm.

### `triggers` Block

* `action` - (Required) Action to trigger. Valid values: `activate`, `deactivate`.
* `conditions` - (Required) Conditions that must be met. See [`conditions` Block](#conditions-block) for details.
* `description` - (Optional) Description of the trigger.
* `min_delay_minutes_between_executions` - (Required) Minimum delay in minutes between executions.
* `target_region` - (Required) Target region for the trigger.

### `conditions` Block

* `associated_alarm_name` - (Required) Name of the associated alarm.
* `condition` - (Required) Condition to check. Valid values: `red`, `green`.

### `report_configuration` Block

* `report_output` - (Required) Output destination for the report. See [`report_output` Block](#report_output-block) for details.

### `report_output` Block

* `s3_configuration` - (Required) S3 output configuration. See [`s3_configuration` Block](#s3_configuration-block) for details.

### `s3_configuration` Block

* `bucket_owner` - (Required) Account ID of the S3 bucket owner.
* `bucket_path` - (Required) S3 bucket path where reports will be stored.

### `workflow.step.parallel_config.step.custom_action_lambda_config.ungraceful` Block

* `behavior` - (Required) Ungraceful behavior. Valid values: `skip`.

### `workflow.step.parallel_config.step.document_db_config.ungraceful` Block

* `ungraceful` - (Required) Ungraceful behavior. Valid values: `failover`.

### `workflow.step.parallel_config.step.ec2_asg_capacity_increase_config.ungraceful` Block

* `minimum_success_percentage` - (Required) Minimum success percentage required.

### `workflow.step.parallel_config.step.ecs_capacity_increase_config.ungraceful` Block

* `minimum_success_percentage` - (Required) Minimum success percentage required.

### `workflow.step.parallel_config.step.eks_resource_scaling_config.ungraceful` Block

* `minimum_success_percentage` - (Required) Minimum success percentage required.

### `workflow.step.parallel_config.step.global_aurora_config.ungraceful` Block

* `ungraceful` - (Required) Ungraceful behavior. Valid values: `failover`.

### `workflow.step.parallel_config.step.lambda_event_source_mapping_config.ungraceful` Block

* `behavior` - (Required) Ungraceful behavior.

### `workflow.step.parallel_config.step.neptune_global_database_config.ungraceful` Block

* `ungraceful` - (Required) Ungraceful behavior.

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

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_arcregionswitch_plan.example
  identity = {
    "arn" = "arn:aws:arcregionswitch:us-east-1:123456789012:plan/example-plan"
  }
}

resource "aws_arcregionswitch_plan" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) Amazon Resource Name (ARN) of the ARC Region Switch Plan.

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
