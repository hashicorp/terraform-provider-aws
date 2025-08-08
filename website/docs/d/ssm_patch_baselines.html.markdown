---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_patch_baselines"
description: |-
  Terraform data source for retrieving AWS SSM (Systems Manager) Patch Baselines.
---

# Data Source: aws_ssm_patch_baselines

Terraform data source for retrieving AWS SSM (Systems Manager) Patch Baselines.

## Example Usage

### Basic Usage

```terraform
data "aws_ssm_patch_baselines" "example" {}
```

### With Filters

```terraform
data "aws_ssm_patch_baselines" "example" {
  filter {
    key    = "OWNER"
    values = ["AWS"]
  }
  filter {
    key    = "OPERATING_SYSTEM"
    values = ["WINDOWS"]
  }
}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `filter` - (Optional) Key-value pairs used to filter the results. See [`filter`](#filter-argument-reference) below.
* `default_baselines` - (Optional) Only return baseline identities where `default_baseline` is `true`.

### `filter` Argument Reference

* `key` - (Required) Filter key. See the [AWS SSM documentation](https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_DescribePatchBaselines.html) for valid values.
* `values` - (Required) Filter values. See the [AWS SSM documentation](https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_DescribePatchBaselines.html) for example values.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `baseline_identities` - List of baseline identities. See [`baseline_identities`](#baseline_identities-attribute-reference) below.

### `baseline_identities` Attribute Reference

* `baseline_description` - Description of the patch baseline.
* `baseline_id` - ID of the patch baseline.
* `baseline_name` - Name of the patch baseline.
* `default_baseline` - Indicates whether this is the default baseline. AWS Systems Manager supports creating multiple default patch baselines. For example, you can create a default patch baseline for each operating system.
* `operating_system` - Operating system the patch baseline applies to.
