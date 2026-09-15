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
* `region` - (Optional, **Deprecated**) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` - Description of the plan.
* `execution_role` - Execution role ARN for the plan.
* `name` - Name of the plan.
* `owner` - Owner of the plan.
* `primary_region` - Primary region for the plan.
* `recovery_approach` - Recovery approach for the plan.
* `recovery_time_objective_minutes` - Recovery time objective in minutes.
* `regions` - List of regions included in the plan.
* `tags` - Map of tags assigned to the resource.
* `updated_at` - Timestamp when the plan was last updated (RFC3339 format).
* `version` - Version of the plan.
