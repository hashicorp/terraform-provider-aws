---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_permission_set"
description: |-
  Get information on a Single Sign-On (SSO) Permission Set.
---

# Data Source: aws_ssoadmin_permission_set

Use this data source to get a Single Sign-On (SSO) Permission Set.

## Example Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

data "aws_ssoadmin_permission_set" "example" {
  instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
  name         = "Example"
}

output "arn" {
  value = data.aws_ssoadmin_permission_set.example.arn
}
```

## Argument Reference

This data source supports the following arguments:

~> **NOTE:** Either `arn` or `name` must be configured.

* `arn` - (Optional) ARN of the permission set.
* `instance_arn` - (Required) ARN of the SSO Instance associated with the permission set.
* `name` - (Optional) Name of the SSO Permission Set.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ARN of the Permission Set.
* `description` - Description of the Permission Set.
* `relay_state` - Relay state URL used to redirect users within the application during the federation authentication process.
* `session_duration` - Length of time that the application user sessions are valid in the ISO-8601 standard.
* `tags` - Key-value map of resource tags.
