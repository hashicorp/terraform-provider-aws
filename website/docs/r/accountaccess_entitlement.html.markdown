---
subcategory: "Account Access"
layout: "aws"
page_title: "AWS: aws_accountaccess_entitlement"
description: |-
  Manages an AWS Account Access Entitlement granting an Identity Center principal access to an IAM role in a target account.
---

# Resource: aws_accountaccess_entitlement

Manages an AWS Account Access Entitlement. An Entitlement grants an IAM Identity Center principal (user or group) the ability to assume a specific IAM role in a target AWS account through an Account Access [Application](accountaccess_application.html.markdown).

~> **Note:** Entitlements are immutable. Changing `principal_id`, `principal_type`, `role_arn`, or `application_arn` triggers replacement.

~> **Note:** The IAM role referenced by `role_arn` must have a trust policy that allows the Account Access service to assume it. The role's `assume_role_policy` must grant `sts:AssumeRole`, `sts:SetContext`, and `sts:TagSession` to the `account-access.amazonaws.com` service principal. Without `sts:TagSession`, credential retrieval for the entitlement fails. See the [Complete Example](#complete-example) below.

## Example Usage

### User Principal

```terraform
resource "aws_accountaccess_application" "example" {
  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}

data "aws_ssoadmin_instances" "example" {}

resource "aws_accountaccess_entitlement" "developer" {
  application_arn = aws_accountaccess_application.example.arn
  principal_id    = "11111111-2222-3333-4444-555555555555"
  principal_type  = "USER"
  role_arn        = "arn:aws:iam::123456789012:role/Developer"
}
```

### Group Principal

```terraform
resource "aws_accountaccess_entitlement" "engineering" {
  application_arn = aws_accountaccess_application.example.arn
  principal_id    = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
  principal_type  = "GROUP"
  role_arn        = "arn:aws:iam::123456789012:role/Engineering"
}
```

### Complete Example

The target IAM role must trust the Account Access service. This example provisions a role with the required trust policy and grants an entitlement to it.

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_accountaccess_application" "example" {
  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}

# The target role must allow the Account Access service to assume it.
# sts:TagSession is required for credential retrieval to succeed.
resource "aws_iam_role" "target" {
  name = "account-access-developer"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "account-access.amazonaws.com"
        }
        Action = [
          "sts:AssumeRole",
          "sts:SetContext",
          "sts:TagSession",
        ]
      },
    ]
  })
}

resource "aws_accountaccess_entitlement" "example" {
  application_arn = aws_accountaccess_application.example.arn
  principal_id    = "11111111-2222-3333-4444-555555555555"
  principal_type  = "USER"
  role_arn        = aws_iam_role.target.arn
}
```

## Argument Reference

The following arguments are required:

* `application_arn` - (Required) ARN of the parent Account Access Application. Forces replacement when changed.
* `principal_id` - (Required) Identity Center user or group ID (a UUID). Forces replacement when changed.
* `principal_type` - (Required) Type of principal. Valid values: `USER`, `GROUP`. Forces replacement when changed.
* `role_arn` - (Required) ARN of the IAM role in the target AWS account that the principal is granted access to. Forces replacement when changed.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `account_id` - 12-digit AWS account ID extracted from the target role ARN.
* `account_name` - Best-effort human-readable name of the target account, when the service can resolve it.
* `created_at` - Date and time, in [RFC3339 format](https://datatracker.ietf.org/doc/html/rfc3339), when the Entitlement was created.
* `entitlement_id` - Service-assigned unique identifier for this Entitlement.
* `id` - Composite identifier in the form `<application_arn>,<entitlement_id>`. Used for `terraform import`.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_accountaccess_entitlement.example
  identity = {
    application_arn = "arn:aws:account-access:us-east-1:123456789012:application/aam-0123456789abcdef"
    entitlement_id  = "ent-0123456789abcdef"
  }
}
```

### Identity Schema

#### Required

* `application_arn` (String) ARN of the parent Account Access Application.
* `entitlement_id` (String) Service-assigned unique identifier for this Entitlement.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Account Access Entitlements using the composite `<application_arn>,<entitlement_id>` ID. For example:

```terraform
import {
  to = aws_accountaccess_entitlement.example
  id = "arn:aws:account-access:us-east-1:123456789012:application/aam-0123456789abcdef,ent-0123456789abcdef"
}
```

Using `terraform import`, import Account Access Entitlements using the composite ID. For example:

```console
% terraform import aws_accountaccess_entitlement.example arn:aws:account-access:us-east-1:123456789012:application/aam-0123456789abcdef,ent-0123456789abcdef
```
