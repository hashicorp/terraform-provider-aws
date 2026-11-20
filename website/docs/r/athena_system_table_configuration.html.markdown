---
subcategory: "Athena"
layout: "aws"
page_title: "AWS: aws_athena_system_table_configuration"
description: |-
  Manages the Athena system table configuration for the current AWS account and Region.
---

# Resource: aws_athena_system_table_configuration

Manages the [Athena system table](https://docs.aws.amazon.com/athena/latest/ug/athena-system-tables.html) configuration for the current AWS account and Region. Creating it turns on Athena system tables. Athena then records query history in the S3 Tables table `query_executions`, in namespace `metadata` of the AWS managed table bucket `aws-athena`. An account has at most one configuration in each Region.

~> **Note:** Athena creates the `query_executions` table when it creates the configuration. The table keeps the encryption it was created with, so `kms_key_arn` cannot change while the table exists. A plan that adds or changes the key warns about this. While the table exists, that apply fails and changes nothing.

!> **Note:** To change `kms_key_arn`, delete the `query_executions` table first, then apply the new value. Deleting the table deletes the query history it holds. Removing `kms_key_arn` replaces the resource. If the table still exists, the create fails and system tables stay off until you delete the table and apply again.

~> **Note:** Destroying this resource turns off system tables and removes the tags. Athena keeps the `aws-athena` table bucket, the `query_executions` table, and its data. A later configuration in the same Region must use the same `kms_key_arn` unless you delete the table.

## Example Usage

### Basic Usage

```terraform
resource "aws_athena_system_table_configuration" "example" {}
```

### Customer Managed KMS Key

```terraform
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "example" {
  statement {
    sid       = "Root"
    actions   = ["kms:*"]
    resources = ["*"]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }

  statement {
    sid       = "AthenaSystemTables"
    actions   = ["kms:GenerateDataKey", "kms:Decrypt", "kms:DescribeKey"]
    resources = ["*"]

    principals {
      type        = "Service"
      identifiers = ["systemtables.athena.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }

  statement {
    sid       = "S3TablesMaintenance"
    actions   = ["kms:GenerateDataKey", "kms:Decrypt", "kms:DescribeKey"]
    resources = ["*"]

    principals {
      type        = "Service"
      identifiers = ["maintenance.s3tables.amazonaws.com"]
    }

    condition {
      test     = "StringLike"
      variable = "kms:EncryptionContext:aws:s3:arn"
      values   = ["arn:${data.aws_partition.current.partition}:s3tables:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:bucket/aws-athena/*"]
    }
  }
}

resource "aws_kms_key" "example" {
  description = "example-athena-system-tables"
  policy      = data.aws_iam_policy_document.example.json
}

resource "aws_athena_system_table_configuration" "example" {
  kms_key_arn = aws_kms_key.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `kms_key_arn` - (Optional) ARN of the KMS key that encrypts the system tables. Must be a key ARN, not an alias ARN. If not set, Athena encrypts the tables with S3 managed keys. Removing this argument forces a new resource to be created. See the notes above before changing it.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `table_name` - (Optional) Name of the system table configuration. Allowed values are: `ALL`. Default value: `ALL`. Changing this forces a new resource to be created.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the system table configuration.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_athena_system_table_configuration.example
  identity = {
    table_name = "ALL"
  }
}

resource "aws_athena_system_table_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `table_name` (String) Name of the system table configuration.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the Athena system table configuration using `table_name`. For example:

```terraform
import {
  to = aws_athena_system_table_configuration.example
  id = "ALL"
}
```

Using `terraform import`, import the Athena system table configuration using `table_name`. For example:

```console
% terraform import aws_athena_system_table_configuration.example ALL
```
