---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_permissions_boundary_attachment"
description: |-
  Attaches a permissions boundary policy to a Single Sign-On (SSO) Permission Set resource.
---

# Resource: aws_ssoadmin_permissions_boundary_attachment

Attaches a permissions boundary policy to a Single Sign-On (SSO) Permission Set resource.

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
* `permissions_boundary` - (Required, Forces new resource) The permissions boundary policy. See below.

### Permissions Boundary

The `permissions_boundary` config block describes the permissions boundary policy to attach. You can reference either an AWS-managed policy, or a customer managed policy, but only one may be set.

* `managed_policy_arn` - (Optional) AWS-managed IAM policy ARN to use as the permissions boundary.
* `customer_managed_policy_reference` - (Optional) Specifies the name and path of a customer managed policy. See below.

### Customer Managed Policy Reference

The `customer_managed_policy_reference` config block describes a customer managed IAM policy. You must have an IAM policy that matches the name and path in each AWS account where you want to deploy your specified permission set.

* `name` - (Required, Forces new resource) Name of the customer managed IAM Policy to be attached.
* `path` - (Optional, Forces new resource) The path to the IAM policy to be attached. The default is `/`. See [IAM Identifiers](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html#identifiers-friendly-names) for more information.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Permission Set Amazon Resource Name (ARN) and SSO Instance ARN, separated by a comma (`,`).

## Import

SSO Admin Permissions Boundary Attachments can be imported using the `permission_set_arn` and `instance_arn`, separated by a comma (`,`) e.g.,

```
$ terraform import aws_ssoadmin_permissions_boundary_attachment.example arn:aws:sso:::permissionSet/ssoins-2938j0x8920sbj72/ps-80383020jr9302rk,arn:aws:sso:::instance/ssoins-2938j0x8920sbj72
```
