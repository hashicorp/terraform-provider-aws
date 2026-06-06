---
subcategory: "CloudWatch Observability Admin"
layout: "aws"
page_title: "AWS: aws_observabilityadmin_s3_table_integration"
description: |-
  Manages a CloudWatch Observability Admin S3 Table Integration.
---

# Resource: aws_observabilityadmin_s3_table_integration

Manages a CloudWatch Observability Admin S3 Table Integration. This integration enables CloudWatch to duplicate telemetry data to Amazon S3 Tables, making it available for analysis by tools such as Amazon Athena and Amazon Redshift.

For more information, see the [CloudWatch Logs S3 Tables integration documentation](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/s3-tables-integration.html).

## Example Usage

### Basic Integration with AES256 Encryption

```terraform
resource "aws_iam_role" "example" {
  name = "example-s3-table-integration"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "logs.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "example" {
  role = aws_iam_role.example.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3tables:CreateTableBucket",
          "s3tables:ListTableBuckets",
          "s3tables:GetTableBucket",
          "s3tables:CreateNamespace",
          "s3tables:GetNamespace",
          "s3tables:ListNamespaces",
          "s3tables:CreateTable",
          "s3tables:GetTable",
          "s3tables:ListTables",
          "s3tables:PutTableData",
          "s3tables:GetTableData",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_observabilityadmin_s3_table_integration" "example" {
  role_arn = aws_iam_role.example.arn

  encryption {
    sse_algorithm = "AES256"
  }
}
```

### Integration with KMS Encryption

```terraform
resource "aws_iam_role" "example" {
  name = "example-s3-table-integration"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "logs.amazonaws.com"
      }
    }]
  })
}

resource "aws_kms_key" "example" {
  description             = "S3 Table Integration KMS key"
  deletion_window_in_days = 7
}

resource "aws_observabilityadmin_s3_table_integration" "example" {
  role_arn = aws_iam_role.example.arn

  encryption {
    sse_algorithm = "aws:kms"
    kms_key_arn   = aws_kms_key.example.arn
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `encryption` - (Required, Forces new resource) Encryption configuration block. [Documented below](#encryption-block).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `role_arn` - (Required, Forces new resource) Amazon Resource Name (ARN) of the IAM role that grants the S3 Table integration permissions to access necessary resources.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `encryption` Block

* `kms_key_arn` - (Optional, Forces new resource) ARN of the KMS key to use for encryption. Required when `sse_algorithm` is `aws:kms`.
* `sse_algorithm` - (Required, Forces new resource) Server-side encryption algorithm. Valid values: `AES256`, `aws:kms`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the S3 Table integration.
* `destination_table_bucket_arn` - ARN of the S3 Table bucket where CloudWatch data is stored. AWS automatically creates a bucket named `_aws-cloudwatch_` if one does not already exist.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_observabilityadmin_s3_table_integration.example
  identity = {
    "arn" = "arn:aws:observabilityadmin:us-east-1:123456789012:s3-table-integration/example-id"
  }
}

resource "aws_observabilityadmin_s3_table_integration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) ARN of the S3 Table integration.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Observability Admin S3 Table Integrations using the `arn`. For example:

```terraform
import {
  to = aws_observabilityadmin_s3_table_integration.example
  id = "arn:aws:observabilityadmin:us-east-1:123456789012:s3-table-integration/example-id"
}
```

Using `terraform import`, import CloudWatch Observability Admin S3 Table Integrations using the `arn`. For example:

```console
% terraform import aws_observabilityadmin_s3_table_integration.example arn:aws:observabilityadmin:us-east-1:123456789012:s3-table-integration/example-id
```
