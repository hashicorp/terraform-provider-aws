---
subcategory: "S3 Glacier"
layout: "aws"
page_title: "AWS: aws_glacier_vault"
description: |-
  Provides a Glacier Vault.
---

# Resource: aws_glacier_vault

Provides a Glacier Vault Resource. You can refer to the [Glacier Developer Guide](https://docs.aws.amazon.com/amazonglacier/latest/dev/working-with-vaults.html) for a full explanation of the Glacier Vault functionality

~> **NOTE:** When removing a Glacier Vault, the Vault must be empty.

## Example Usage

```terraform
resource "aws_sns_topic" "aws_sns_topic" {
  name = "glacier-sns-topic"
}

data "aws_iam_policy_document" "my_archive" {
  statement {
    sid    = "add-read-only-perm"
    effect = "Allow"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions = [
      "glacier:InitiateJob",
      "glacier:GetJobOutput",
    ]

    resources = ["arn:aws:glacier:eu-west-1:432981146916:vaults/MyArchive"]
  }
}

resource "aws_glacier_vault" "my_archive" {
  name = "MyArchive"

  notification {
    sns_topic = aws_sns_topic.aws_sns_topic.arn
    events    = ["ArchiveRetrievalCompleted", "InventoryRetrievalCompleted"]
  }

  access_policy = data.aws_iam_policy_document.my_archive.json

  tags = {
    Test = "MyArchive"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the Vault. Names can be between 1 and 255 characters long and the valid characters are a-z, A-Z, 0-9, '_' (underscore), '-' (hyphen), and '.' (period).
* `access_policy` - (Optional) The policy document. This is a JSON formatted string.
  The heredoc syntax or `file` function is helpful here. Use the [Glacier Developer Guide](https://docs.aws.amazon.com/amazonglacier/latest/dev/vault-access-policy.html) for more information on Glacier Vault Policy
* `notification` - (Optional) The notifications for the Vault. Fields documented below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

**notification** supports the following:

* `events` - (Required) You can configure a vault to publish a notification for `ArchiveRetrievalCompleted` and `InventoryRetrievalCompleted` events.
* `sns_topic` - (Required) The SNS Topic ARN.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `location` - The URI of the vault that was created.
* `arn` - The ARN of the vault.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Glacier Vaults using the `name`. For example:

```terraform
import {
  to = aws_glacier_vault.archive
  id = "my_archive"
}
```

Using `terraform import`, import Glacier Vaults using the `name`. For example:

```console
% terraform import aws_glacier_vault.archive my_archive
```
