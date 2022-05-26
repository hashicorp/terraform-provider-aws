---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_patch_baseline"
description: |-
  Provides an SSM Patch Baseline data source
---

# Data Source: aws_ssm_patch_baseline

Provides an SSM Patch Baseline data source. Useful if you wish to reuse the default baselines provided.

## Example Usage

To retrieve a baseline provided by AWS:

```terraform
data "aws_ssm_patch_baseline" "centos" {
  owner            = "AWS"
  name_prefix      = "AWS-"
  operating_system = "CENTOS"
}
```

To retrieve a baseline on your account:

```terraform
data "aws_ssm_patch_baseline" "default_custom" {
  owner            = "Self"
  name_prefix      = "MyCustomBaseline"
  default_baseline = true
  operating_system = "WINDOWS"
}
```

## Argument Reference

The following arguments are supported:

* `owner` - (Required) The owner of the baseline. Valid values: `All`, `AWS`, `Self` (the current account).
* `name_prefix` - (Optional) Filter results by the baseline name prefix.
* `default_baseline` - (Optional) Filters the results against the baselines default_baseline field.
* `operating_system` - (Optional) The specified OS for the baseline.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `approved_patches` - A list of explicitly approved patches for the baseline.
* `approved_patches_compliance_level` - The compliance level for approved patches.
* `approved_patches_enable_non_security` - Indicates whether the list of approved patches includes non-security updates that should be applied to the instances.
* `approval_rule` - A list of rules used to include patches in the baseline.
    * `approve_after_days` - The number of days after the release date of each patch matched by the rule the patch is marked as approved in the patch baseline.
    * `approve_until_date` - The cutoff date for auto approval of released patches. Any patches released on or before this date are installed automatically. Date is formatted as `YYYY-MM-DD`. Conflicts with `approve_after_days`
    * `compliance_level` - The compliance level for patches approved by this rule.
    * `enable_non_security` - Boolean enabling the application of non-security updates.
    * `patch_filter` - The patch filter group that defines the criteria for the rule.
        * `key` - The key for the filter.
        * `values` - The value for the filter.
* `global_filter` - A set of global filters used to exclude patches from the baseline.
    * `key` - The key for the filter.
    * `values` - The value for the filter.
* `id` - The id of the baseline.
* `name` - The name of the baseline.
* `description` - The description of the baseline.
* `rejected_patches` - A list of rejected patches.
* `rejected_patches_action` - The action specified to take on patches included in the `rejected_patches` list.
* `source` - Information about the patches to use to update the managed nodes, including target operating systems and source repositories.
    * `configuration` - The value of the yum repo configuration.
    * `name` - The name specified to identify the patch source.
    * `products` - The specific operating system versions a patch repository applies to.
