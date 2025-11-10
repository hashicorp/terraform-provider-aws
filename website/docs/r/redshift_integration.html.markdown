---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_integration"
description: |-
  Terraform resource for managing a DynamoDB zero-ETL integration or S3 event integration with Amazon Redshift.
---

# Resource: aws_redshift_integration

Terraform resource for managing a DynamoDB zero-ETL integration or S3 event integration with Amazon Redshift. You can refer to the [User Guide](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/RedshiftforDynamoDB-zero-etl.html) for a DynamoDB zero-ETL integration or the [User Guide](https://docs.aws.amazon.com/redshift/latest/dg/loading-data-copy-job.html) for a S3 event integration.

## Example Usage

### Basic Usage

```terraform
resource "aws_dynamodb_table" "example" {
  name           = "dynamodb-table-example"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "example"

  attribute {
    name = "example"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }
}

resource "aws_redshiftserverless_namespace" "example" {
  namespace_name = "redshift-example"
}

resource "aws_redshiftserverless_workgroup" "example" {
  namespace_name      = aws_redshiftserverless_namespace.example.namespace_name
  workgroup_name      = "example-workgroup"
  base_capacity       = 8
  publicly_accessible = false

  subnet_ids = [aws_subnet.example1.id, aws_subnet.example2.id, aws_subnet.example3.id]

  config_parameter {
    parameter_key   = "enable_case_sensitive_identifier"
    parameter_value = "true"
  }
}

resource "aws_redshift_integration" "example" {
  integration_name = "example"
  source_arn       = aws_dynamodb_table.example.arn
  target_arn       = aws_redshiftserverless_namespace.example.arn
}
```

### Use own KMS key

```terraform
data "aws_caller_identity" "current" {}

resource "aws_kms_key" "example" {
  description             = "example"
  deletion_window_in_days = 10
}

resource "aws_kms_key_policy" "example" {
  key_id = aws_kms_key.example.id

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Effect = "Allow"
        Principal = {
          Service = "redshift.amazonaws.com"
        }
        Action = [
          "kms:Decrypt",
          "kms:CreateGrant"
        ]
        Resource = "*"
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
          ArnEquals = {
            "aws:SourceArn" = "arn:aws:redshift:*:${data.aws_caller_identity.current.account_id}:integration:*"
          }
        }
      }
    ]
  })
}

resource "aws_redshift_integration" "example" {
  integration_name = "example"
  source_arn       = aws_dynamodb_table.example.arn
  target_arn       = aws_redshiftserverless_namespace.example.arn
  kms_key_id       = aws_kms_key.example.arn

  additional_encryption_context = {
    "example" : "test",
  }
}
```

## Argument Reference

The following arguments are required:

* `integration_name` - (Required) Name of the integration.
* `source_arn` - (Required, Forces new resources) ARN of the database to use as the source for replication. You can specify a DynamoDB table or an S3 bucket.
* `target_arn` - (Required, Forces new resources) ARN of the Redshift data warehouse to use as the target for replication.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `additional_encryption_context` - (Optional, Forces new resources) Set of non-secret key–value pairs that contains additional contextual information about the data.
For more information, see the [User Guide](https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#encrypt_context).
You can only include this parameter if you specify the `kms_key_id` parameter.
* `description` - (Optional) Description of the integration.
* `kms_key_id` - (Optional, Forces new resources) KMS key identifier for the key to use to encrypt the integration.
If you don't specify an encryption key, Redshift uses a default AWS owned key.
You can only include this parameter if `source_arn` references a DynamoDB table.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

For more detailed documentation about each argument, refer to the [AWS official documentation](https://docs.aws.amazon.com/cli/latest/reference/redshift/create-integration.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Integration.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Integration using the `arn`. For example:

```terraform
import {
  to = aws_redshift_integration.example
  id = "arn:aws:redshift:us-west-2:123456789012:integration:abcdefgh-0000-1111-2222-123456789012"
}
```

Using `terraform import`, import Redshift Integration using the `arn`. For example:

```console
% terraform import aws_redshift_integration.example arn:aws:redshift:us-west-2:123456789012:integration:abcdefgh-0000-1111-2222-123456789012
```
