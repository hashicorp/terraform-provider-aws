---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_instance"
description: |-
  Get information on SSO Instances.
---

# Data Source: aws_ssoadmin_instances

Use this data source to get ARNs and Identity Store IDs of Single Sign-On (SSO) Instances.

## Example Usage

```hcl
data "aws_ssoadmin_instances" "example" {}

output "arn" {
  value = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}

output "identity_store_id" {
  value = tolist(data.aws_ssoadmin_instances.example.identity_store_ids)[0]
}
```

## Argument Reference

There are no arguments available for this data source.

## Attributes Reference

* `arn` - Set of Amazon Resource Names (ARNs) of the SSO Instances.
* `id` - AWS Region.
* `identity_store_ids` - Set of identifier(s) of the identity store(s) connected to the SSO instance(s).
