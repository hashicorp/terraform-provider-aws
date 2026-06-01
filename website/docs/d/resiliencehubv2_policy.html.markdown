---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_policy"
description: |-
  Terraform data source for reading an AWS Resilience Hub V2 Policy.
---

# Data Source: aws_resiliencehubv2_policy

Terraform data source for reading an AWS Resilience Hub V2 Policy.

## Example Usage

### Basic Usage

```hcl
data "aws_resiliencehubv2_policy" "example" {
  arn = "arn:aws:resiliencehub:us-west-2:123456789012:policy/example-policy:abc123"
}
```

## Argument Reference

This data source supports the following arguments:

* `arn` - (Required) ARN of the policy.
* `region` - (Optional, **Deprecated**) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `availability_slo` - Availability SLO configuration. See [`availability_slo` Block](#availability_slo-block) below.
* `data_recovery` - Data recovery configuration. See [`data_recovery` Block](#data_recovery-block) below.
* `description` - Description of the policy.
* `multi_az` - Multi-AZ disaster recovery configuration. See [`multi_az` Block](#multi_az-block) below.
* `multi_region` - Multi-region disaster recovery configuration. See [`multi_region` Block](#multi_region-block) below.
* `name` - Name of the policy.
* `tags` - Map of tags assigned to the resource.

### `availability_slo` Block

* `target` - Availability target as a percentage.

### `data_recovery` Block

* `time_between_backups_in_minutes` - Maximum time between backups in minutes.

### `multi_az` Block

* `disaster_recovery_approach` - Disaster recovery approach.
* `rpo_in_minutes` - Recovery point objective in minutes.
* `rto_in_minutes` - Recovery time objective in minutes.

### `multi_region` Block

* `disaster_recovery_approach` - Disaster recovery approach.
* `rpo_in_minutes` - Recovery point objective in minutes.
* `rto_in_minutes` - Recovery time objective in minutes.
