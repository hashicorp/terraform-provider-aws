---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_instances"
description: |-
  Get information on SSO Instances.
---

# Data Source: aws_ssoadmin_instances

Use this data source to get ARNs and Identity Store IDs of Single Sign-On (SSO) Instances.

## Example Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

output "arn" {
  value = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}

output "identity_store_id" {
  value = tolist(data.aws_ssoadmin_instances.example.identity_store_ids)[0]
}
```

## Argument Reference

This data source does not support any arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - Set of Amazon Resource Names (ARNs) of the SSO Instances.
* `id` - AWS Region.
* `identity_store_ids` - Set of identifiers of the identity stores connected to the SSO Instances.
