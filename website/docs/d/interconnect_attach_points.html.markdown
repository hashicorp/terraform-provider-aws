---
subcategory: "Interconnect"
layout: "aws"
page_title: "AWS: aws_interconnect_attach_points"
description: |-
  Terraform data source for listing AWS Interconnect attach points.
---

# Data Source: aws_interconnect_attach_points

Terraform data source for listing AWS Interconnect attach points available for a connection.

## Example Usage

### Basic Usage

```terraform
data "aws_interconnect_attach_points" "example" {
  environment_id = "example-environment-id"
}
```

## Argument Reference

The following arguments are required:

* `environment_id` - (Required) Identifier of the Environment for which to list attach points.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `attach_points` - List of attach points. [See below](#attach_points).

### attach_points

* `identifier` - Identifier for the specific type of the attach point.
* `name` - Descriptive name of the attach point identifier.
* `type` - Type of the attach point, which dictates the syntax of the identifier.
