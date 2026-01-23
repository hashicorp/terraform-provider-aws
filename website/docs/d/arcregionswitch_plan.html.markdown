---
subcategory: "ARC (Application Recovery Controller) Region Switch"
layout: "aws"
page_title: "AWS: aws_arcregionswitch_plan"
description: |-
  Terraform data source for managing an Amazon ARC Region Switch plan.
---

# Data Source: aws_arcregionswitch_plan

Terraform data source for managing an Amazon ARC Region Switch plan.

## Example Usage

### Basic Usage

```terraform
data "aws_arcregionswitch_plan" "example" {
  arn = "arn:aws:arcregionswitch:us-west-2:123456789012:plan/example-plan"
}
```

## Argument Reference

This data source supports the following arguments:

* `arn` - (Required) ARN of the ARC Region Switch Plan.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` - Description of the plan.
* `execution_role` - Execution role ARN for the plan.
* `name` - Name of the plan.
* `primary_region` - Primary region for the plan.
* `recovery_approach` - Recovery approach for the plan.
* `recovery_time_objective_minutes` - Recovery time objective in minutes.
* `regions` - List of regions included in the plan.
* `route53_health_checks` - List of Route53 health checks associated with the plan. Each health check contains:
    * `health_check_id` - ID of the Route53 health check.
    * `hosted_zone_id` - Hosted zone ID for the health check.
    * `record_name` - Record name for the health check.
    * `region` - Region for the health check.
