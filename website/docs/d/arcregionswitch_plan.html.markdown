---
subcategory: "Application Recovery Controller Region Switch"
layout: "aws"
page_title: "AWS: aws_arcregionswitch_plan"
description: |-
  Terraform data source for managing an AWS ARC Region Switch Plan.
---

# Data Source: aws_arcregionswitch_plan

Terraform data source for managing an AWS ARC Region Switch Plan.

## Example Usage

### Basic Usage

```terraform
data "aws_arcregionswitch_plan" "example" {
  arn = "arn:aws:arcregionswitch:us-west-2:123456789012:plan/example-plan"
}
```

### With Route53 Health Check Waiting

```terraform
data "aws_arcregionswitch_plan" "example" {
  arn                    = "arn:aws:arcregionswitch:us-west-2:123456789012:plan/example-plan"
  wait_for_health_checks = true
}
```

## Argument Reference

This data source supports the following arguments:

* `arn` - (Required) ARN of the ARC Region Switch Plan.
* `region` - (Optional) AWS region where the plan is located.
* `wait_for_health_checks` - (Optional) Wait for Route53 health check IDs to be populated (takes ~4 minutes). Default is `false`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `associated_alarms` - Set of associated alarms for the plan. Each alarm contains:
    * `alarm_type` - Type of the alarm.
    * `cross_account_role` - Cross-account role ARN for the alarm.
    * `external_id` - External ID for the alarm.
    * `name` - Name of the alarm.
    * `resource_identifier` - Resource identifier for the alarm.
* `description` - Description of the plan.
* `execution_role` - Execution role ARN for the plan.
* `name` - Name of the plan.
* `owner` - Owner of the plan.
* `primary_region` - Primary region for the plan.
* `recovery_approach` - Recovery approach for the plan.
* `recovery_time_objective_minutes` - Recovery time objective in minutes.
* `regions` - List of regions included in the plan.
* `route53_health_checks` - List of Route53 health checks associated with the plan. Each health check contains:
    * `health_check_id` - ID of the Route53 health check.
    * `hosted_zone_id` - Hosted zone ID for the health check.
    * `record_name` - Record name for the health check.
    * `region` - Region for the health check.
* `tags` - Map of tags assigned to the resource.
* `trigger` - List of trigger configurations for the plan. Each trigger contains:
    * `action` - Action to trigger.
    * `conditions` - List of conditions for the trigger. Each condition contains:
    * `associated_alarm_name` - Name of the associated alarm.
    * `condition` - Condition for the trigger.
    * `description` - Description of the trigger.
    * `min_delay_minutes_between_executions` - Minimum delay in minutes between executions.
    * `target_region` - Target region for the trigger.
* `updated_at` - Timestamp when the plan was last updated.
* `version` - Version of the plan.
* `workflow` - List of workflow configurations for the plan. Each workflow contains:
    * `step` - List of steps in the workflow.
    * `workflow_description` - Description of the workflow.
    * `workflow_target_action` - Target action for the workflow.
    * `workflow_target_region` - Target region for the workflow.
