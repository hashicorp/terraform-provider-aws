---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_permission_set"
description: |-
  Manages a Single Sign-On (SSO) Permission Set
---

# Resource: aws_ssoadmin_permission_set

Provides a Single Sign-On (SSO) Permission Set resource

~> **NOTE:** Updating this resource will automatically [Provision the Permission Set](https://docs.aws.amazon.com/singlesignon/latest/APIReference/API_ProvisionPermissionSet.html) to apply the corresponding updates to all assigned accounts.

## Example Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_ssoadmin_permission_set" "example" {
  name             = "Example"
  description      = "An example"
  instance_arn     = tolist(data.aws_ssoadmin_instances.example.arns)[0]
  relay_state      = "https://s3.console.aws.amazon.com/s3/home?region=us-east-1#"
  session_duration = "PT2H"
}
```

## Argument Reference

This resource supports the following arguments:

* `description` - (Optional) The description of the Permission Set.
* `instance_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the SSO Instance under which the operation will be executed.
* `name` - (Required, Forces new resource) The name of the Permission Set.
* `relay_state` - (Optional) The relay state URL used to redirect users within the application during the federation authentication process.
* `session_duration` - (Optional) The length of time that the application user sessions are valid in the ISO-8601 standard. Default: `PT1H`.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the Permission Set.
* `id` - The Amazon Resource Names (ARNs) of the Permission Set and SSO Instance, separated by a comma (`,`).
* `created_date` - The date the Permission Set was created in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSO Permission Sets using the `arn` and `instance_arn` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_ssoadmin_permission_set.example
  id = "arn:aws:sso:::permissionSet/ssoins-2938j0x8920sbj72/ps-80383020jr9302rk,arn:aws:sso:::instance/ssoins-2938j0x8920sbj72"
}
```

Using `terraform import`, import SSO Permission Sets using the `arn` and `instance_arn` separated by a comma (`,`). For example:

```console
% terraform import aws_ssoadmin_permission_set.example arn:aws:sso:::permissionSet/ssoins-2938j0x8920sbj72/ps-80383020jr9302rk,arn:aws:sso:::instance/ssoins-2938j0x8920sbj72
```
