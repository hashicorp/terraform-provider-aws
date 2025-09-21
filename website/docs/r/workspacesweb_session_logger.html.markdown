---
subcategory: "WorkSpaces Web"
layout: "aws"
page_title: "AWS: aws_workspacesweb_session_logger"
description: |-
  Terraform resource for managing an AWS WorkSpaces Web Session Logger.
---

# Resource: aws_workspacesweb_session_logger

Terraform resource for managing an AWS WorkSpaces Web Session Logger.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "example-session-logs"
}

data "aws_iam_policy_document" "example" {
  statement {
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["workspaces-web.amazonaws.com"]
    }
    actions = [
      "s3:PutObject"
    ]
    resources = ["${aws_s3_bucket.example.arn}/*"]
  }
}

resource "aws_s3_bucket_policy" "example" {
  bucket = aws_s3_bucket.example.id
  policy = data.aws_iam_policy_document.example.json
}

resource "aws_workspacesweb_session_logger" "example" {
  display_name = "example-session-logger"

  event_filter {
    all {}
  }

  log_configuration {
    s3 {
      bucket           = aws_s3_bucket.example.id
      folder_structure = "Flat"
      log_file_format  = "Json"
    }
  }

  depends_on = [aws_s3_bucket_policy.example]
}
```

### Complete Configuration with KMS Encryption

```terraform
resource "aws_s3_bucket" "example" {
  bucket        = "example-session-logs"
  force_destroy = true
}

data "aws_iam_policy_document" "example" {
  statement {
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["workspaces-web.amazonaws.com"]
    }
    actions = [
      "s3:PutObject"
    ]
    resources = [
      aws_s3_bucket.example.arn,
      "${aws_s3_bucket.example.arn}/*"
    ]
  }
}

resource "aws_s3_bucket_policy" "example" {
  bucket = aws_s3_bucket.example.id
  policy = data.aws_iam_policy_document.example.json
}

data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "kms_key_policy" {
  statement {
    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
    actions   = ["kms:*"]
    resources = ["*"]
  }

  statement {
    principals {
      type        = "Service"
      identifiers = ["workspaces-web.amazonaws.com"]
    }
    actions = [
      "kms:Encrypt",
      "kms:GenerateDataKey*",
      "kms:ReEncrypt*",
      "kms:Decrypt"
    ]
    resources = ["*"]
  }
}

resource "aws_kms_key" "example" {
  description = "KMS key for WorkSpaces Web Session Logger"
  policy      = data.aws_iam_policy_document.kms_key_policy.json
}

resource "aws_workspacesweb_session_logger" "example" {
  display_name         = "example-session-logger"
  customer_managed_key = aws_kms_key.example.arn
  additional_encryption_context = {
    Environment = "Production"
    Application = "WorkSpacesWeb"
  }

  event_filter {
    include = ["SessionStart", "SessionEnd"]
  }

  log_configuration {
    s3 {
      bucket           = aws_s3_bucket.example.id
      bucket_owner     = data.aws_caller_identity.current.account_id
      folder_structure = "NestedByDate"
      key_prefix       = "workspaces-web-logs/"
      log_file_format  = "JsonLines"
    }
  }

  tags = {
    Name        = "example-session-logger"
    Environment = "Production"
  }

  depends_on = [aws_s3_bucket_policy.example, aws_kms_key.example]
}
```

## Argument Reference

The following arguments are required:

* `event_filter` - (Required) Event filter that determines which events are logged. See [Event Filter](#event-filter) below.
* `log_configuration` - (Required) Configuration block for specifying where logs are delivered. See [Log Configuration](#log-configuration) below.

The following arguments are optional:

* `additional_encryption_context` - (Optional) Map of additional encryption context key-value pairs.
* `customer_managed_key` - (Optional) ARN of the customer managed KMS key used to encrypt sensitive information.
* `display_name` - (Optional) Human-readable display name for the session logger resource. Forces replacement if changed.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Log Configuration

* `s3` - (Required) Configuration block for S3 log delivery. See [S3 Configuration](#s3-configuration) below.

### Event Filter

Exactly one of the following must be specified:

* `all` - (Optional) Block that specifies to monitor all events. Set to `{}` to monitor all events.
* `include` - (Optional) List of specific events to monitor. Valid values include session events like `SessionStart`, `SessionEnd`, etc.

### S3 Configuration

* `bucket` - (Required) S3 bucket name where logs are delivered.
* `folder_structure` - (Required) Folder structure that defines the organizational structure for log files in S3. Valid values: `FlatStructure`, `DateBasedStructure`.
* `log_file_format` - (Required) Format of the log file written to S3. Valid values: `Json`, `Parquet`.
* `bucket_owner` - (Optional) Expected bucket owner of the target S3 bucket.
* `key_prefix` - (Optional) S3 path prefix that determines where log files are stored.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `associated_portal_arns` - List of ARNs of the web portals associated with the session logger.
* `session_logger_arn` - ARN of the session logger.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

~> **Note:** The `additional_encryption_context` and `customer_managed_key` attributes are computed when not specified and will be populated with values from the AWS API response.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces Web Session Logger using the `session_logger_arn`. For example:

```terraform
import {
  to = aws_workspacesweb_session_logger.example
  id = "arn:aws:workspaces-web:us-west-2:123456789012:sessionLogger/session_logger-id-12345678"
}
```

Using `terraform import`, import WorkSpaces Web Session Logger using the `session_logger_arn`. For example:

```console
% terraform import aws_workspacesweb_session_logger.example arn:aws:workspaces-web:us-west-2:123456789012:sessionLogger/session_logger-id-12345678
```
