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
  description = "examplerule"
  resource_tags {
    resource_tag_key   = tag_key
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

* `resource_type` - (Required) The resource type to be retained by the retention rule.
* `retention_period` - (Required) Information about the retention period for which the retention rule is to retain resources.

The following arguments are optional:

* `description` - (Optional) The retention rule description.
* `resourceTags` - (Optional) Specifies the resource tags to use to identify resources that are to be retained by a tag-level retention rule.
* `retentionPeriod` - (Optional) Information about the retention period for which the retention rule is to retain resources.
* `lockConfiguration` - (Optional) Information about the retention rule lock configuration.
* `lockState` - (Optional) The lock state of the retention rules to list. Only retention rules with the specified lock state are returned. Valid values are `locked | pending_unlock | unlocked`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - (String) ID of the RBinRule.
* `lockEndTime` - (Timestamp) The date and time at which the unlock delay is set to expire. Only returned for retention rules that have been unlocked and that are still within the unlock delay period.
* `status` - (String) The state of the retention rule. Only retention rules that are in the `available` state retain resources. Valid values include `pending | available`.

## Import

RBin RBinRule can be imported using the `id`, e.g.,

```
$ terraform import aws_rbin_rule.example examplerule
```
