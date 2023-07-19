---
subcategory: "EFS (Elastic File System)"
layout: "aws"
page_title: "AWS: aws_efs_file_system_policy"
description: |-
  Provides an Elastic File System (EFS) File System Policy resource.
---

# Resource: aws_efs_file_system_policy

Provides an Elastic File System (EFS) File System Policy resource.

## Example Usage

```terraform
resource "aws_efs_file_system" "fs" {
  creation_token = "my-product"
}

data "aws_iam_policy_document" "policy" {
  statement {
    sid    = "ExampleStatement01"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    actions = [
      "elasticfilesystem:ClientMount",
      "elasticfilesystem:ClientWrite",
    ]

    resources = [aws_efs_file_system.fs.arn]

    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["true"]
    }
  }
}

resource "aws_efs_file_system_policy" "policy" {
  file_system_id = aws_efs_file_system.fs.id
  policy         = data.aws_iam_policy_document.policy.json
}
```

## Argument Reference

The following arguments are required:

* `file_system_id` - (Required) The ID of the EFS file system.
* `policy` - (Required) The JSON formatted file system policy for the EFS file system. see [Docs](https://docs.aws.amazon.com/efs/latest/ug/access-control-overview.html#access-control-manage-access-intro-resource-policies) for more info.

The following arguments are optional:

* `bypass_policy_lockout_safety_check` - (Optional) A flag to indicate whether to bypass the `aws_efs_file_system_policy` lockout safety check. The policy lockout safety check determines whether the policy in the request will prevent the principal making the request will be locked out from making future `PutFileSystemPolicy` requests on the file system. Set `bypass_policy_lockout_safety_check` to `true` only when you intend to prevent the principal that is making the request from making a subsequent `PutFileSystemPolicy` request on the file system. The default value is `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID that identifies the file system (e.g., fs-ccfc0d65).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the EFS file system policies using the `id`. For example:

```terraform
import {
  to = aws_efs_file_system_policy.foo
  id = "fs-6fa144c6"
}
```

Using `terraform import`, import the EFS file system policies using the `id`. For example:

```console
% terraform import aws_efs_file_system_policy.foo fs-6fa144c6
```
