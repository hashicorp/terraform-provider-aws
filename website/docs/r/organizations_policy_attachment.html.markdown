---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_policy_attachment"
description: |-
  Provides a resource to attach an AWS Organizations policy to an organization account, root, or unit.
---

# Resource: aws_organizations_policy_attachment

Provides a resource to attach an AWS Organizations policy to an organization account, root, or unit.

## Example Usage

### Organization Account

```terraform
resource "aws_organizations_policy_attachment" "account" {
  policy_id = aws_organizations_policy.example.id
  target_id = "123456789012"
}
```

### Organization Root

```terraform
resource "aws_organizations_policy_attachment" "root" {
  policy_id = aws_organizations_policy.example.id
  target_id = aws_organizations_organization.example.roots[0].id
}
```

### Organization Unit

```terraform
resource "aws_organizations_policy_attachment" "unit" {
  policy_id = aws_organizations_policy.example.id
  target_id = aws_organizations_organizational_unit.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `policy_id` - (Required) The unique identifier (ID) of the policy that you want to attach to the target.
* `target_id` - (Required) The unique identifier (ID) of the root, organizational unit, or account number that you want to attach the policy to.
* `skip_destroy` - (Optional) If set to `true`, destroy will **not** detach the policy and instead just remove the resource from state. This can be useful in situations where the attachment must be preserved to meet the AWS minimum requirement of 1 attached policy.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_organizations_policy_attachment` using the target ID and policy ID. For example:

With an account target:

```terraform
import {
  to = aws_organizations_policy_attachment.account
  id = "123456789012:p-12345678"
}
```

Using `terraform import`, import `aws_organizations_policy_attachment` using the target ID and policy ID. For example:

With an account target:

```console
% terraform import aws_organizations_policy_attachment.account 123456789012:p-12345678
```
