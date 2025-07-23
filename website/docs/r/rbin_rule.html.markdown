---
subcategory: "Recycle Bin (RBin)"
layout: "aws"
page_title: "AWS: aws_rbin_rule"
description: |-
  Terraform resource for managing an AWS RBin Rule.
---

# Resource: aws_rbin_rule

Terraform resource for managing an AWS RBin Rule.

## Example Usage

### Basic Usage

```terraform
resource "aws_rbin_rule" "example" {
  description   = "Example tag-level retention rule"
  resource_type = "EBS_SNAPSHOT"

  resource_tags {
    resource_tag_key   = "tag_key"
    resource_tag_value = "tag_value"
  }

  retention_period {
    retention_period_value = 10
    retention_period_unit  = "DAYS"
  }

  tags = {
    "test_tag_key" = "test_tag_value"
  }
}
```

### Region-Level Retention Rule

```terraform
resource "aws_rbin_rule" "example" {
  description   = "Example region-level retention rule with exclusion tags"
  resource_type = "EC2_IMAGE"

  exclude_resource_tags {
    resource_tag_key   = "tag_key"
    resource_tag_value = "tag_value"
  }

  retention_period {
    retention_period_value = 10
    retention_period_unit  = "DAYS"
  }

  tags = {
    "test_tag_key" = "test_tag_value"
  }
}
```

## Argument Reference

The following arguments are required:

* `resource_type` - (Required) Resource type to be retained by the retention rule. Valid values are `EBS_SNAPSHOT` and `EC2_IMAGE`.
* `retention_period` - (Required) Information about the retention period for which the retention rule is to retain resources. See [`retention_period`](#retention_period) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Retention rule description.
* `exclude_resource_tags` - (Optional) Exclusion tags to use to identify resources that are to be excluded, or ignored, by a Region-level retention rule. See [`exclude_resource_tags`](#exclude_resource_tags) below.
* `lock_configuration` - (Optional) Information about the retention rule lock configuration. See [`lock_configuration`](#lock_configuration) below.
* `resource_tags` - (Optional) Resource tags to use to identify resources that are to be retained by a tag-level retention rule. See [`resource_tags`](#resource_tags) below.

### retention_period

The following arguments are required:

* `retention_period_unit` - (Required) Unit of time in which the retention period is measured. Currently, only DAYS is supported.
* `retention_period_value` - (Required) Period value for which the retention rule is to retain resources. The period is measured using the unit specified for RetentionPeriodUnit.

### exclude_resource_tags

The following argument is required:

* `resource_tag_key` - (Required) Tag key.

The following argument is optional:

* `resource_tag_value` - (Optional) Tag value.

### lock_configuration

The following argument is required:

* `unlock_delay` - (Required) Information about the retention rule unlock delay. See [`unlock_delay`](#unlock_delay) below.

### unlock_delay

The following arguments are required:

* `unlock_delay_unit` - (Required) Unit of time in which to measure the unlock delay. Currently, the unlock delay can be measure only in days.
* `unlock_delay_value` - (Required) Unlock delay period, measured in the unit specified for UnlockDelayUnit.

### resource_tags

The following argument is required:

* `resource_tag_key` - (Required) Tag key.

The following argument is optional:

* `resource_tag_value` - (Optional) Tag value.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - (String) ID of the Rule.
* `lock_end_time` - (Timestamp) Date and time at which the unlock delay is set to expire. Only returned for retention rules that have been unlocked and that are still within the unlock delay period.
* `lock_state` - (Optional) Lock state of the retention rules to list. Only retention rules with the specified lock state are returned. Valid values are `locked`, `pending_unlock`, `unlocked`.
* `status` - (String) State of the retention rule. Only retention rules that are in the `available` state retain resources. Valid values include `pending` and `available`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RBin Rule using the `id`. For example:

```terraform
import {
  to = aws_rbin_rule.example
  id = "examplerule"
}
```

Using `terraform import`, import RBin Rule using the `id`. For example:

```console
% terraform import aws_rbin_rule.example examplerule
```
