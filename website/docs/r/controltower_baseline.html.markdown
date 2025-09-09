---
subcategory: "Control Tower"
layout: "aws"
page_title: "AWS: aws_controltower_baseline"
description: |-
  Terraform resource for managing an AWS Control Tower Baseline.
---

# Resource: aws_controltower_baseline

Terraform resource for managing an AWS Control Tower Baseline.

## Example Usage

### Basic Usage

```terraform
resource "aws_controltower_baseline" "example" {
  baseline_identifier = "arn:aws:controltower:us-east-1::baseline/17BSJV3IGJ2QSGA2"
  baseline_version    = "4.0"
  target_identifier   = aws_organizations_organizational_unit.test.arn
  parameters {
    key   = "IdentityCenterEnabledBaselineArn"
    value = "arn:aws:controltower:us-east-1:664418989480:enabledbaseline/XALULM96QHI525UOC"
  }
}
```

## Argument Reference

The following arguments are required:

* `baseline_identifier` - (Required) The ARN of the baseline to be enabled.
* `baseline_version` - (Required) The version of the baseline to be enabled.
* `target_identifier` - (Required) The ARN of the target on which the baseline will be enabled. Only OUs are supported as targets.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `parameters` - (Optional) A list of key-value objects that specify enablement parameters, where key is a string and value is a document of any type. See [Parameter](#parameters) below for details.
* `tags` - (Optional) Tags to apply to the landing zone. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### parameters

* `key` - (Required) The key of the parameter.
* `value` - (Required) The value of the parameter.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Baseline.
* `operaton_identifier` - The ID (in UUID format) of the asynchronous operation.
* `tags_all` - A map of tags assigned to the landing zone, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Control Tower Baseline using the `arn`. For example:

```terraform
import {
  to = aws_controltower_baseline.example
  id = "arn:aws:controltower:us-east-1:012345678912:enabledbaseline/XALULM96QHI525UOC"
}
```

Using `terraform import`, import Control Tower Baseline using the `arn`. For example:

```console
% terraform import aws_controltower_baseline.example arn:aws:controltower:us-east-1:012345678912:enabledbaseline/XALULM96QHI525UOC
```
