---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_instances"
description: |-
  Get information on SSM managed instances.
---

# Data Source: aws_ssm_instances

Use this data source to get the instance IDs of SSM managed instances.

## Example Usage

```terraform
data "aws_ssm_instances" "example" {
  filter {
    name   = "PlatformTypes"
    values = ["Linux"]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.

### filter Configuration Block

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter field. Valid values can be found in the [SSM InstanceInformationStringFilter API Reference](https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_InstanceInformationStringFilter.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - Set of instance IDs of the matched SSM managed instances.
