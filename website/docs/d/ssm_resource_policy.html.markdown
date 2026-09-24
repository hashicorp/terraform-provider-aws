---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_resource_policy"
description: |-
  Terraform data source for managing an AWS Systems Manager Resource Policy.
---

# Data Source: aws_ssm_resource_policy

Terraform data source for retrieving an AWS Systems Manager resource policy attached to a Systems Manager resource (e.g. an advanced-tier `aws_ssm_parameter` or an `OpsItemGroup`).

## Example Usage

### Basic Usage

```terraform
data "aws_ssm_resource_policy" "example" {
  resource_arn = aws_ssm_parameter.example.arn
}
```

### Lookup by Policy ID

When more than one policy is attached to the resource, specify `policy_id` to select a particular policy:

```terraform
data "aws_ssm_resource_policy" "example" {
  resource_arn = aws_ssm_parameter.example.arn
  policy_id    = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
}
```

## Argument Reference

This data source supports the following arguments:

* `resource_arn` - (Required) Amazon Resource Name (ARN) of the resource whose policy will be retrieved.
* `policy_id` - (Optional) Unique identifier of the policy within the resource's policies. Required when multiple policies are attached to the resource.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Composite identifier in the format `<resource_arn>,<policy_id>`.
* `policy` - JSON-encoded resource policy attached to the resource.
* `policy_hash` - ID of the current policy version.
* `policy_id` - Unique identifier of the policy.
