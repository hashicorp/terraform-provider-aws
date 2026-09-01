---
subcategory: "Account Access"
layout: "aws"
page_title: "AWS: aws_accountaccess_entitlement"
description: |-
  Manages an AWS Account Access Entitlement granting an Identity Center principal access to an IAM role in a target account.
---

# Resource: aws_accountaccess_entitlement

Manages an AWS Account Access Entitlement. An Entitlement grants an IAM Identity Center principal the ability to assume a specific IAM role in a target AWS account through an Account Access [Application](accountaccess_application.html.markdown).

~> **Note:** Entitlements are immutable. Changing `application_arn` or `entitlement` triggers replacement.

~> **Note:** The IAM role referenced by `entitlement.principal_role.role_arn` must have a trust policy that allows the Account Access service to assume it. The role's `assume_role_policy` must grant `sts:AssumeRole`, `sts:SetContext`, and `sts:TagSession` to the `account-access.amazonaws.com` service principal. Without `sts:TagSession`, credential retrieval for the entitlement fails. See the [Complete Example](#complete-example) below.

## Example Usage

### User Principal

```terraform
resource "aws_accountaccess_application" "example" {
  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}

data "aws_ssoadmin_instances" "example" {}

resource "aws_accountaccess_entitlement" "example" {
  application_arn = aws_accountaccess_application.example.arn

  entitlement {
    principal_role {
      role_arn = "arn:aws:iam::123456789012:role/Developer"

      principal {
        identity_center {
          user_id = "11111111-2222-3333-4444-555555555555"
        }
      }
    }
  }
}
```

### Group Principal

```terraform
resource "aws_accountaccess_entitlement" "example" {
  application_arn = aws_accountaccess_application.example.arn

  entitlement {
    principal_role {
      role_arn = "arn:aws:iam::123456789012:role/Engineering"

      principal {
        identity_center {
          group_id = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
        }
      }
    }
  }
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
  name = "example-account-access-developer"

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

  entitlement {
    principal_role {
      role_arn = aws_iam_role.target.arn

      principal {
        identity_center {
          user_id = "11111111-2222-3333-4444-555555555555"
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `application_arn` - (Required) ARN of the parent Account Access Application. Forces replacement when changed.
* `entitlement` - (Required) Entitlement configuration. See [`entitlement` Block](#entitlement-block) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `entitlement` Block

The `entitlement` block supports:

* `principal_role` - (Required) Principal role entitlement configuration. See [`entitlement.principal_role` Block](#entitlementprincipal_role-block) below.

### `entitlement.principal_role` Block

The `entitlement.principal_role` block supports:

* `principal` - (Required) Principal configuration. See [`entitlement.principal_role.principal` Block](#entitlementprincipal_roleprincipal-block) below.
* `role_arn` - (Required) ARN of the IAM role in the target AWS account that the principal is granted access to.

### `entitlement.principal_role.principal` Block

The `entitlement.principal_role.principal` block supports:

* `identity_center` - (Required) IAM Identity Center principal configuration. See [`entitlement.principal_role.principal.identity_center` Block](#entitlementprincipal_roleprincipalidentity_center-block) below.

### `entitlement.principal_role.principal.identity_center` Block

The `entitlement.principal_role.principal.identity_center` block requires exactly one of the following arguments:

* `group_id` - (Optional) IAM Identity Center group ID.
* `user_id` - (Optional) IAM Identity Center user ID.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `entitlement_id` - Service-assigned unique identifier for this Entitlement.
* `entitlement.principal_role.account_id` - Target AWS account ID.
* `entitlement.principal_role.account_name` - Target AWS account name.

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
