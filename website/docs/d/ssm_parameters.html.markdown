---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_parameters"
description: |-
  Provides a list of AWS SSM Parameters.
---

# Data Source: aws_ssm_parameters

Provides a list of AWS SSM Parameters.
Add filters

## Example Usage

### Default

```terraform
data "aws_ssm_parameters" "example" {}
```

### Filters

```terraform
data "aws_ssm_parameters" "example" {
  parameter_filter {
    key    = "Path"
    option = "Recursive"
    values = ["/param/example/"]
  }
}

data "aws_ssm_parameters" "example" {
  parameter_filter {
    key    = "Name"
    option = "Contains"
    values = ["example"]
  }
}

data "aws_ssm_parameters" "example" {
  parameter_filter {
    key    = "Name"
    option = "BeginsWith"
    values = ["/param/ex"]
  }
}
```

### Shared

```terraform
data "aws_ssm_parameters" "example" {
  shared = true
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `parameter_filter` - (Optional) Configuration block(s) for filtering. Detailed below.
* `shared` - (Optional) When true, only parameters shared using RAM are returned. Defaults to `false`.

### `parameter_filter` Configuration Block

The `parameter_filter` configuration block supports the following arguments:

* `key` - (Required) Name of the filter field.
* `option` - (Optional) Name of the filter field. Valid values can be found in the.
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any
  given value matches.

This datasource uses
the [DescribeParameters](https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_DescribeParameters.html#systemsmanager-DescribeParameters-request-ParameterFilters)
API.
Check out the documentation for valid filter argument
values [SSM ParameterStringFilter API Reference](https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_ParameterStringFilter.html)

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - Set of parameter ARNs of the matched SSM parameters.
