---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_policy"
description: |-
  Terraform resource for managing an AWS Resilience Hub V2 Policy.
---

# Resource: aws_resiliencehubv2_policy

Terraform resource for managing an AWS Resilience Hub V2 Policy.

A resilience policy defines your resilience expectations through modular, composable requirements. Rather than choosing a single rigid policy type, you construct policies by selecting the requirements that matter to your application: availability SLO, multi-AZ disaster recovery, multi-region disaster recovery, and data recovery objectives.

## Example Usage

### Basic Usage

```hcl
resource "aws_resiliencehubv2_policy" "example" {
  name = "example-policy"

  availability_slo {
    target = 99.9
  }
}
```

### Multi-AZ with Data Recovery

```hcl
resource "aws_resiliencehubv2_policy" "example" {
  name        = "example-policy"
  description = "Policy with multi-AZ and data recovery targets"

  availability_slo {
    target = 99.99
  }

  data_recovery {
    time_between_backups_in_minutes = 60
  }

  multi_az {
    disaster_recovery_approach = "ACTIVE_ACTIVE"
    rpo_in_minutes             = 5
    rto_in_minutes             = 10
  }

  tags = {
    Environment = "production"
  }
}
```

### Multi-Region

```hcl
resource "aws_resiliencehubv2_policy" "example" {
  name = "example-multi-region-policy"

  availability_slo {
    target = 99.95
  }

  multi_region {
    disaster_recovery_approach = "ACTIVE_PASSIVE"
    rpo_in_minutes             = 15
    rto_in_minutes             = 30
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the policy. Changing this value requires creating a new resource.

The following arguments are optional:

* `availability_slo` - (Optional) Availability SLO configuration. See [`availability_slo` Block](#availability_slo-block) below.
* `data_recovery` - (Optional) Data recovery configuration. See [`data_recovery` Block](#data_recovery-block) below.
* `description` - (Optional) Description of the policy.
* `multi_az` - (Optional) Multi-AZ disaster recovery configuration. See [`multi_az` Block](#multi_az-block) below.
* `multi_region` - (Optional) Multi-region disaster recovery configuration. See [`multi_region` Block](#multi_region-block) below.
* `region` - (Optional, **Deprecated**) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `availability_slo` Block

The `availability_slo` block supports:

* `target` - (Required) Availability target as a percentage (e.g., `99.9`).

### `data_recovery` Block

The `data_recovery` block supports:

* `time_between_backups_in_minutes` - (Required) Maximum time between backups in minutes.

### `multi_az` Block

The `multi_az` block supports:

* `disaster_recovery_approach` - (Required) Multi-AZ disaster recovery approach. Valid values: `ACTIVE_ACTIVE`, `HOT_STANDBY`, `WARM_STANDBY`, `PILOT_LIGHT`, `BACKUP_AND_RESTORE`.
* `rpo_in_minutes` - (Optional) Recovery point objective in minutes.
* `rto_in_minutes` - (Optional) Recovery time objective in minutes.

### `multi_region` Block

The `multi_region` block supports:

* `disaster_recovery_approach` - (Required) Multi-region disaster recovery approach. Valid values: `ACTIVE_ACTIVE`, `HOT_STANDBY`, `WARM_STANDBY`, `PILOT_LIGHT`, `BACKUP_AND_RESTORE`.
* `rpo_in_minutes` - (Optional) Recovery point objective in minutes.
* `rto_in_minutes` - (Optional) Recovery time objective in minutes.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the policy.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_resiliencehubv2_policy.example
  identity = {
    "arn" = "arn:aws:resiliencehub:us-west-2:123456789012:policy/example-policy:abc123"
  }
}

resource "aws_resiliencehubv2_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) Amazon Resource Name (ARN) of the Resilience Hub V2 Policy.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Resilience Hub V2 Policy using the `arn`. For example:

```terraform
import {
  to = aws_resiliencehubv2_policy.example
  id = "arn:aws:resiliencehub:us-west-2:123456789012:policy/example-policy:abc123"
}
```

Using `terraform import`, import Resilience Hub V2 Policy using the `arn`. For example:

```console
% terraform import aws_resiliencehubv2_policy.example arn:aws:resiliencehub:us-west-2:123456789012:policy/example-policy:abc123
```
