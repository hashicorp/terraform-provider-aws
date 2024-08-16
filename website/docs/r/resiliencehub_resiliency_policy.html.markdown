---
subcategory: "Resilience Hub"
layout: "aws"
page_title: "AWS: aws_resiliencehub_resiliency_policy"
description: |-
  Terraform resource for managing an AWS Resilience Hub Resiliency Policy.
---

# Resource: aws_resiliencehub_resiliency_policy

Terraform resource for managing an AWS Resilience Hub Resiliency Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_resiliencehub_resiliency_policy" "example" {
  
  policy_name        = "testexample"
  policy_description = "testexample"

  tier = "NonCritical"

  data_location_constraint = "AnyLocation"

  policy {
    region {
      rpo_in_secs = 86400
      rto_in_secs = 86400
    }
    az {
      rpo_in_secs = 86400
      rto_in_secs = 86400
    }
    hardware {
      rpo_in_secs = 86400
      rto_in_secs = 86400
    }
    software {
      rpo_in_secs = 86400
      rto_in_secs = 86400
    }
  }

}
```

## Argument Reference

The following arguments are required:

* `policy_name` (String) Name of Resiliency Policy.
* `tier` (String) Resiliency Policy Tier.
* `policy` (Attributes) The type of resiliency policy to be created, including the recovery time objective (RTO) and recovery point objective (RPO) in seconds. See [`policy`](#policy).

The following arguments are optional:

* `policy_description` (String) Description of Resiliency Policy.
* `data_location_constraint` (String) Data Location Constraint of the Policy.
* `tags` - (Map Of String) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `policy`

The following arguments are required:

* `az` - (Attributes) Specifies Availability Zone failure policy. See [`policy.az`](#policy.az)
* `hardware` - (Attributes) Specifies Infrastructure failure policy. See [`policy.hardware`](#policy.hardware)
* `software` - (Attributes) Specifies Application failure policy. See [`policy.software`](#policy.software)

The following arguments are optional:

* `region` - (Attributes) Specifies Region failure policy. [`policy.region`](#policy.region)

### `policy.az`

The following arguments are required:

* `rpo_in_secs` - (Number) RPO in seconds.
* `rto_in_secs` - (Number) RTO in seconds.

### `policy.hardware`

The following arguments are required:

* `rpo_in_secs` - (Number) RPO in seconds.
* `rto_in_secs` - (Number) RTO in seconds.

### `policy.software`

The following arguments are required:

* `rpo_in_secs` - (Number) RPO in seconds.
* `rto_in_secs` - (Number) RTO in seconds.

### `policy.region`

The following arguments are required:

* `rpo_in_secs` - (Number) RPO in seconds.
* `rto_in_secs` - (Number) RTO in seconds.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Resiliency Policy.
* `id` - ID of the Resiliency Policy.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Resilience Hub Resiliency Policy using the `arn`. For example:

```terraform
import {
  to = aws_resiliencehub_resiliency_policy.example
  id = "arn:aws:resiliencehub:us-east-1:123456789012:resiliency-policy/8c1cfa29-d1dd-4421-aa68-c9f64cced4c2"
}
```

Using `terraform import`, import Resilience Hub Resiliency Policy using the `arn`. For example:

```console
% terraform import aws_resiliencehub_resiliency_policy.example arn:aws:resiliencehub:us-east-1:123456789012:resiliency-policy/8c1cfa29-d1dd-4421-aa68-c9f64cced4c2
```
