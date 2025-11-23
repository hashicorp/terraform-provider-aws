---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_outbound_web_identity_federation"
description: |-
  Manages an AWS IAM (Identity & Access Management) Outbound Web Identity Federation.
---

# Resource: aws_iam_outbound_web_identity_federation

Manages an AWS IAM (Identity & Access Management) Outbound Web Identity Federation.

~> **NOTE:** Removing this Terraform resource disables IAM Outbound Web Identity Federation.

## Example Usage

```terraform
resource "aws_iam_outbound_web_identity_federation" "example" {
  enabled = true
}
```

## Argument Reference

* `enabled` - (Optional) Whether or not Outbound Web Identity Federation is enabled. Valid values are `true` or `false`. Defaults to `true`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier for the account registration. Since registration is applied globaly, this will be the Account ID.
* `issuer_identifier` - A unique issuer URL for your AWS account that hosts the OpenID Connect (OIDC) discovery endpoints.


## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IAM Outbound Web Identity Federation resources using the `id`. For example:

```terraform
import {
  to = aws_iam_outbound_web_identity_federation.example
  id = "123456789012"
}
```

Using `terraform import`, import IAM Outbound Web Identity Federation  resources using the `id`. For example:

```console
% terraform import aws_iam_outbound_web_identity_federation.example 123456789012
```