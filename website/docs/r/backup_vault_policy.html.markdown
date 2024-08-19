---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_vault_policy"
description: |-
  Provides an AWS Backup vault policy resource.
---

# Resource: aws_backup_vault_policy

Provides an AWS Backup vault policy resource.

## Example Usage

```terraform
resource "aws_backup_vault" "example" {
  name = "example"
}

data "aws_iam_policy_document" "example" {
  statement {
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    actions = [
      "backup:DescribeBackupVault",
      "backup:DeleteBackupVault",
      "backup:PutBackupVaultAccessPolicy",
      "backup:DeleteBackupVaultAccessPolicy",
      "backup:GetBackupVaultAccessPolicy",
      "backup:StartBackupJob",
      "backup:GetBackupVaultNotifications",
      "backup:PutBackupVaultNotifications",
    ]

    resources = [aws_backup_vault.example.arn]
  }
}

resource "aws_backup_vault_policy" "example" {
  backup_vault_name = aws_backup_vault.example.name
  policy            = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

This resource supports the following arguments:

* `backup_vault_name` - (Required) Name of the backup vault to add policy for.
* `policy` - (Required) The backup vault access policy document in JSON format.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the vault.
* `backup_vault_arn` - The ARN of the vault.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Backup vault policy using the `name`. For example:

```terraform
import {
  to = aws_backup_vault_policy.test
  id = "TestVault"
}
```

Using `terraform import`, import Backup vault policy using the `name`. For example:

```console
% terraform import aws_backup_vault_policy.test TestVault
```
