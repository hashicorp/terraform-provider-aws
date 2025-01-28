---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_integration"
description: |-
  Terraform resource for managing an AWS RDS (Relational Database) Integration.
---

# Resource: aws_rds_integration

Terraform resource for managing an AWS RDS (Relational Database) zero-ETL integration. You can refer to the [User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/zero-etl.setting-up.html).

## Example Usage

### Basic Usage

```terraform
resource "aws_redshiftserverless_namespace" "example" {
  namespace_name = "redshift-example"
}

resource "aws_redshiftserverless_workgroup" "example" {
  namespace_name      = aws_redshiftserverless_namespace.example.namespace_name
  workgroup_name      = "example-workspace"
  base_capacity       = 8
  publicly_accessible = false

  subnet_ids = [aws_subnet.example1.id, aws_subnet.example2.id, aws_subnet.example3.id]

  config_parameter {
    parameter_key   = "enable_case_sensitive_identifier"
    parameter_value = "true"
  }
}

resource "aws_rds_integration" "example" {
  integration_name = "example"
  source_arn       = aws_rds_cluster.example.arn
  target_arn       = aws_redshiftserverless_namespace.example.arn

  lifecycle {
    ignore_changes = [
      kms_key_id
    ]
  }
}
```

### Use own KMS key

```terraform
data "aws_caller_identity" "current" {}

resource "aws_kms_key" "example" {
  deletion_window_in_days = 10
  policy                  = data.aws_iam_policy_document.key_policy.json
}

data "aws_iam_policy_document" "key_policy" {
  statement {
    actions   = ["kms:*"]
    resources = ["*"]
    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }

  statement {
    actions   = ["kms:CreateGrant"]
    resources = ["*"]
    principals {
      type        = "Service"
      identifiers = ["redshift.amazonaws.com"]
    }
  }
}

resource "aws_rds_integration" "example" {
  integration_name = "example"
  source_arn       = aws_rds_cluster.example.arn
  target_arn       = aws_redshiftserverless_namespace.example.arn
  kms_key_id       = aws_kms_key.example.arn

  additional_encryption_context = {
    "example" : "test",
  }
}
```

## Argument Reference

For more detailed documentation about each argument, refer to the [AWS official documentation](https://docs.aws.amazon.com/cli/latest/reference/rds/create-integration.html).

The following arguments are required:

* `integration_name` - (Required, Forces new resources) Name of the integration.
* `source_arn` - (Required, Forces new resources) ARN of the database to use as the source for replication.
* `target_arn` - (Required, Forces new resources) ARN of the Redshift data warehouse to use as the target for replication.

The following arguments are optional:

* `additional_encryption_context` - (Optional, Forces new resources) Set of non-secret key–value pairs that contains additional contextual information about the data.
For more information, see the [User Guide](https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#encrypt_context).
You can only include this parameter if you specify the `kms_key_id` parameter.
* `data_filter` - (Optional, Forces new resources) Data filters for the integration.
These filters determine which tables from the source database are sent to the target Amazon Redshift data warehouse.
The value should match the syntax from the AWS CLI which includes an `include:` or `exclude:` prefix before a filter expression.
Multiple expressions are separated by a comma.
See the [Amazon RDS data filtering guide](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/zero-etl.filtering.html) for additional details.
* `kms_key_id` - (Optional, Forces new resources) KMS key identifier for the key to use to encrypt the integration.
If you don't specify an encryption key, RDS uses a default AWS owned key.
If you use the default AWS owned key, you should ignore `kms_key_id` parameter by using [`lifecycle` parameter](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#ignore_changes) to avoid unintended change after the first creation.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Integration.
* `id` - ID of the Integration.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `10m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RDS (Relational Database) Integration using the `arn`. For example:

```terraform
import {
  to = aws_rds_integration.example
  id = "arn:aws:rds:us-west-2:123456789012:integration:abcdefgh-0000-1111-2222-123456789012"
}
```

Using `terraform import`, import RDS (Relational Database) Integration using the `arn`. For example:

```console
% terraform import aws_rds_integration.example arn:aws:rds:us-west-2:123456789012:integration:abcdefgh-0000-1111-2222-123456789012
```
