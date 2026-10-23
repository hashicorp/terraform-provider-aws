---
subcategory: "Kinesis Analytics V2"
layout: "aws"
page_title: "AWS: aws_kinesisanalyticsv2_snapshot_import"
description: |-
  Imports an Apache Flink savepoint from Amazon S3 as a Managed Service for Apache Flink enhanced snapshot.
---

# Resource: aws_kinesisanalyticsv2_snapshot_import

Imports an Apache Flink savepoint from Amazon S3 as a Managed Service for Apache Flink enhanced snapshot. Any application in enhanced snapshot mode in the same account and Region can then restore it. For more information, see [Snapshots](https://docs.aws.amazon.com/managed-flink/latest/java/how-snapshots.html) in the *Managed Service for Apache Flink Developer Guide*.

The source must be a relocatable savepoint (not a checkpoint) with a `_metadata` file under the prefix, and every object must have a full-object `CRC64NVME` checksum. Savepoints written by an enhanced snapshot export already meet these requirements. Do not modify or delete the source objects during the import.

~> **NOTE:** Managed Service for Apache Flink does not record a runtime or source application for an imported snapshot, and does not check the snapshot against the application that restores it. Confirm Apache Flink version and operator compatibility before restoring it into an application.

~> **NOTE:** The imported snapshot is not tied to any application and keeps accruing storage charges until deleted.

## Example Usage

### Import a Savepoint and Start an Application From It

```terraform
resource "aws_kinesisanalyticsv2_snapshot_import" "example" {
  name = "example-imported"

  import_source_configuration {
    source_type     = "S3"
    import_role_arn = aws_iam_role.import.arn

    source_s3_prefix_location {
      bucket_arn = aws_s3_bucket.savepoints.arn
      key_prefix = "exports/example/s9T0u1V2w3X4/data/"
    }
  }

  destination_encryption_configuration {
    key_type = "CUSTOMER_MANAGED_KEY"
    key_id   = aws_kms_key.example.arn
  }
}

resource "aws_kinesisanalyticsv2_application" "example" {
  name                   = "example"
  runtime_environment    = "FLINK-1_20"
  service_execution_role = aws_iam_role.example.arn

  application_configuration {
    # ... other configuration ...

    application_snapshot_configuration {
      snapshots_enabled = true
      snapshot_mode     = "ENHANCED_SNAPSHOTS"
    }

    run_configuration {
      application_restore_configuration {
        application_restore_type = "RESTORE_FROM_CUSTOM_SNAPSHOT"
        snapshot_arn             = aws_kinesisanalyticsv2_snapshot_import.example.arn
      }
    }
  }

  start_application = true
}
```

### Import Role

Managed Service for Apache Flink assumes `import_role_arn` to read the source. The caller needs `iam:PassRole` on the role, plus `kinesisanalytics:ImportSnapshot` on the ARN of the snapshot being created.

```terraform
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_iam_role" "import" {
  name = "example-msf-snapshot-import"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "kinesisanalytics.amazonaws.com" }
      Action    = "sts:AssumeRole"
      Condition = {
        StringEquals = { "aws:SourceAccount" = data.aws_caller_identity.current.account_id }
        ArnEquals    = { "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:kinesisanalytics:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:snapshot/example-imported" }
      }
    }]
  })
}

resource "aws_iam_role_policy" "import" {
  role = aws_iam_role.import.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["s3:GetObject", "s3:GetObjectAttributes"]
        Resource = "${aws_s3_bucket.savepoints.arn}/exports/example/*"
      },
      {
        Effect   = "Allow"
        Action   = "s3:ListBucket"
        Resource = aws_s3_bucket.savepoints.arn
      },
    ]
  })
}
```

## Argument Reference

The following arguments are required:

* `import_source_configuration` - (Required) Location of the savepoint to import and the role used to read it. See [`import_source_configuration` Block](#import_source_configuration-block) below.
* `name` - (Required) Name of the snapshot. Must be unique among the snapshots in the account and Region. Names beginning with `SYSTEMSNAPSHOT` are reserved.

The following arguments are optional:

* `description` - (Optional) Summary description of the snapshot.
* `destination_encryption_configuration` - (Optional) Encryption of the imported snapshot. If omitted, the snapshot is encrypted with an AWS owned key. See [Encryption Configuration Blocks](#encryption-configuration-blocks) below.
* `format` - (Optional) Format of the savepoint. Valid values: `CANONICAL`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the snapshot. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `import_source_configuration` Block

* `import_role_arn` - (Required) ARN of the IAM role that Managed Service for Apache Flink assumes to read the source. It needs `s3:GetObject` and `s3:GetObjectAttributes` on the source prefix, and `s3:ListBucket` on the bucket.
* `source_encryption_configuration` - (Optional) Encryption of the source objects. See [Encryption Configuration Blocks](#encryption-configuration-blocks) below.
* `source_s3_prefix_location` - (Required) Amazon S3 location of the savepoint. See [`source_s3_prefix_location` Block](#source_s3_prefix_location-block) below.
* `source_type` - (Required) Type of the source location. Valid values: `S3`.

### `source_s3_prefix_location` Block

* `bucket_arn` - (Required) ARN of the bucket. The bucket must be in the same account and Region as the snapshot.
* `key_prefix` - (Required) Prefix that directly contains the savepoint's `_metadata` file. Must end with `/`. For a savepoint written by an export, use the export's `ExportedPrefix` (the `.../data/` directory), not the prefix the export was given.

### Encryption Configuration Blocks

The `destination_encryption_configuration` and `source_encryption_configuration` blocks support the following arguments:

* `key_id` - (Optional) KMS key. Required when `key_type` is `CUSTOMER_MANAGED_KEY`. For `destination_encryption_configuration`, use the key ARN.
* `key_type` - (Required) Type of key. Valid values: `AWS_OWNED_KEY`, `CUSTOMER_MANAGED_KEY`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the snapshot.
* `creation_time` - Time the snapshot was created, in [RFC3339 format](https://datatracker.ietf.org/doc/html/rfc3339#section-5.8).
* `size_in_bytes` - Size of the snapshot, in bytes.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_kinesisanalyticsv2_snapshot_import.example
  identity = {
    name = "example-imported"
  }
}

resource "aws_kinesisanalyticsv2_snapshot_import" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` - Name of the snapshot.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import imported enhanced snapshots using the `name`. For example:

```terraform
import {
  to = aws_kinesisanalyticsv2_snapshot_import.example
  id = "example-imported"
}
```

Using `terraform import`, import imported enhanced snapshots using the `name`. For example:

```console
% terraform import aws_kinesisanalyticsv2_snapshot_import.example example-imported
```

The API does not return the import role, source type, or encryption arguments, so `import_source_configuration` and `destination_encryption_configuration` are not populated on import. The first apply after import records their configured values in state without replacing the snapshot.
