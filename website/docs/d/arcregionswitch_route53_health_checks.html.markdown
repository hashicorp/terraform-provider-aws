---
subcategory: "ARC (Application Recovery Controller) Region Switch"
layout: "aws"
page_title: "AWS: aws_arcregionswitch_route53_health_checks"
description: |-
  Terraform data source for managing Amazon ARC Region Switch Route53 Health Checks.
---

# Data Source: aws_arcregionswitch_route53_health_checks

Terraform data source for managing Amazon ARC Region Switch Route53 Health Checks.

## Example Usage

### Basic Usage

```terraform
data "aws_arcregionswitch_route53_health_checks" "example" {
  plan_arn = "arn:aws:arc-region-switch::123456789012:plan/example-plan:abc123"
}
```

## Argument Reference

This data source supports the following arguments:

* `plan_arn` - (Required) ARN of the ARC Region Switch Plan.
* `region` - (Optional, **Deprecated**) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `health_checks` - List of Route53 health checks associated with the plan. Each health check contains:
    * `health_check_id` - ID of the Route53 health check.
    * `hosted_zone_id` - Hosted zone ID for the health check.
    * `record_name` - Record name for the health check.
    * `region` - Region for the health check.
    * `status` - Status of the health check. Valid values: `healthy`, `unhealthy`, `unknown`.
