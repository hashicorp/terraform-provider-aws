---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_permissions_boundary_attachment"
description: |-
  Attaches a permissions boundary policy to a Single Sign-On (SSO) Permission Set resource
---

# Resource: aws_ssoadmin_permissions_boundary_attachment

Attaches a permissions boundary policy to a Single Sign-On (SSO) Permission Set resource

~> **NOTE:** A permission set can have at most one permissions boundary attached; using more than one `aws_ssoadmin_permissions_boundary_attachment` references the same permission set will show a permanent difference.

## Example Usage

### Attaching a customer-managed policy

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_ssoadmin_permission_set" "example" {
  name         = "Example"
  instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}

resource "aws_iam_policy" "example" {
  name        = "TestPolicy"
  description = "My test policy"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "ec2:Describe*",
        ]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}

resource "aws_ssoadmin_permissions_boundary_attachment" "example" {
  instance_arn       = aws_ssoadmin_permission_set.example.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.example.arn
  permissions_boundary {
    customer_managed_policy_reference {
      name = aws_iam_policy.example.name
      path = "/"
    }
  }
}

```

### Attaching an AWS-managed policy

```terraform
resource "aws_ssoadmin_permissions_boundary_attachment" "example" {
  instance_arn       = aws_ssoadmin_permission_set.example.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.example.arn
  permissions_boundary {
    managed_policy_arn = "arn:aws:iam::aws:policy/ReadOnlyAccess"
  }
}
```

## Argument Reference

The following arguments are required:

* `instance_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the SSO Instance under which the operation will be executed.
* `permission_set_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the Permission Set.
* `permissions_boundary` - (Required, Forces new resource) The permissions boundary policy

### permissions_boundary Configuration Block

The `permissions_boundary` config block describes the permissions boundary policy to attach. You can reference either an AWS-managed policy, or a customer-managed policy.

Only one of `managed_policy_arn` and `customer_managed_policy_reference` may be set on a given `permissions_boundary` resource.

* `managed_policy_arn` - (Optional) the ARN of an AWS-managed IAM policy to use as the permissions boundary
* `customer_managed_policy_reference` - (Optional) a reference to a customer-managed IAM policy to use as the permissions boundary.  Same as the [`customer_managed_policy_reference`](ssoadmin_customer_managed_policy_attachment.html#customer-managed-policy-reference) argument of the [`aws_ssoadmin_customer_managed_policy_attachment`](ssoadmin_customer_managed_policy_attachment.html) resource.  You must have an IAM policy that matches the name and path in each AWS account where you want to deploy your specified permission set.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Permission Set Amazon Resource Name (ARN) and SSO Instance ARN, separated by a comma (`,`).

## Import

SSO Admin Permissions Boundary Attachments can be imported using the `permission_set_arn` and `instance_arn`, separated by a comma (`,`) e.g.,

```
$ terraform import aws_ssoadmin_permissions_boundary_attachment.example arn:aws:sso:::permissionSet/ssoins-2938j0x8920sbj72/ps-80383020jr9302rk,arn:aws:sso:::instance/ssoins-2938j0x8920sbj72
```
