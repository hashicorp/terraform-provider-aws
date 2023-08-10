---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_export_task"
description: |-
  Terraform resource for managing an AWS RDS (Relational Database) Export Task.
---

# Resource: aws_rds_export_task

Terraform resource for managing an AWS RDS (Relational Database) Export Task.

## Example Usage

### Basic Usage

```terraform
resource "aws_rds_export_task" "example" {
  export_task_identifier = "example"
  source_arn             = aws_db_snapshot.example.db_snapshot_arn
  s3_bucket_name         = aws_s3_bucket.example.id
  iam_role_arn           = aws_iam_role.example.arn
  kms_key_id             = aws_kms_key.example.arn
}
```

### Complete Usage

```terraform
resource "aws_s3_bucket" "example" {
  bucket        = "example"
  force_destroy = true
}

resource "aws_s3_bucket_acl" "example" {
  bucket = aws_s3_bucket.example.id
  acl    = "private"
}

resource "aws_iam_role" "example" {
  name = "example"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "export.rds.amazonaws.com"
        }
      },
    ]
  })
}

data "aws_iam_policy_document" "example" {
  statement {
    actions = [
      "s3:ListAllMyBuckets",
    ]
    resources = [
      "*"
    ]
  }
  statement {
    actions = [
      "s3:GetBucketLocation",
      "s3:ListBucket",
    ]
    resources = [
      aws_s3_bucket.example.arn,
    ]
  }
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject",
    ]
    resources = [
      "${aws_s3_bucket.example.arn}/*"
    ]
  }
}

resource "aws_iam_policy" "example" {
  name   = "example"
  policy = data.aws_iam_policy_document.example.json
}

resource "aws_iam_role_policy_attachment" "example" {
  role       = aws_iam_role.example.name
  policy_arn = aws_iam_policy.example.arn
}

resource "aws_kms_key" "example" {
  deletion_window_in_days = 10
}

resource "aws_db_instance" "example" {
  identifier           = "example"
  allocated_storage    = 10
  db_name              = "test"
  engine               = "mysql"
  engine_version       = "5.7"
  instance_class       = "db.t3.micro"
  username             = "foo"
  password             = "foobarbaz"
  parameter_group_name = "default.mysql5.7"
  skip_final_snapshot  = true
}

resource "aws_db_snapshot" "example" {
  db_instance_identifier = aws_db_instance.example.identifier
  db_snapshot_identifier = "example"
}

resource "aws_rds_export_task" "example" {
  export_task_identifier = "example"
  source_arn             = aws_db_snapshot.example.db_snapshot_arn
  s3_bucket_name         = aws_s3_bucket.example.id
  iam_role_arn           = aws_iam_role.example.arn
  kms_key_id             = aws_kms_key.example.arn

  export_only = ["database"]
  s3_prefix   = "my_prefix/example"
}
```

## Argument Reference

The following arguments are required:

* `export_task_identifier` - (Required) Unique identifier for the snapshot export task.
* `iam_role_arn` - (Required) ARN of the IAM role to use for writing to the Amazon S3 bucket.
* `kms_key_id` - (Required) ID of the Amazon Web Services KMS key to use to encrypt the snapshot.
* `s3_bucket_name` - (Required) Name of the Amazon S3 bucket to export the snapshot to.
* `source_arn` - (Required) Amazon Resource Name (ARN) of the snapshot to export.

The following arguments are optional:

* `export_only` - (Optional) Data to be exported from the snapshot. If this parameter is not provided, all the snapshot data is exported. Valid values are documented in the [AWS StartExportTask API documentation](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_StartExportTask.html#API_StartExportTask_RequestParameters).
* `s3_prefix` - (Optional) Amazon S3 bucket prefix to use as the file name and path of the exported snapshot.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `failure_cause` - Reason the export failed, if it failed.
* `id` - Unique identifier for the snapshot export task (same value as `export_task_identifier`).
* `percent_progress` - Progress of the snapshot export task as a percentage.
* `snapshot_time` - Time that the snapshot was created.
* `source_type` - Type of source for the export.
* `status` - Status of the export task.
* `task_end_time` - Time that the snapshot export task completed.
* `task_start_time` - Time that the snapshot export task started.
* `warning_message` - Warning about the snapshot export task, if any.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a RDS (Relational Database) Export Task using the `export_task_identifier`. For example:

```terraform
import {
  to = aws_rds_export_task.example
  id = "example"
}
```

Using `terraform import`, import a RDS (Relational Database) Export Task using the `export_task_identifier`. For example:

```console
% terraform import aws_rds_export_task.example example
```
