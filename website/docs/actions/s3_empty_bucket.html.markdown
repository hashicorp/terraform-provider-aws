---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_empty_bucket"
description: |-
  Empties an S3 bucket by deleting all objects and versions
---

# Action: aws_s3_empty_bucket

~> **NOTE:** This action is in beta and its interface may change in future versions.

Empties an S3 bucket by deleting all objects and versions. This action is useful for preparing buckets for deletion or performing cleanup operations during Terraform workflows.

For more information about S3 buckets, see the [Amazon S3 User Guide](https://docs.aws.amazon.com/s3/latest/userguide/).

## Example Usage

### Basic Usage

```terraform
action "aws_s3_empty_bucket" "example" {
  config {
    bucket_name = "my-bucket"
  }
}

resource "terraform_data" "cleanup" {
  lifecycle {
    action_trigger {
      events  = [before_destroy]
      actions = [action.aws_s3_empty_bucket.example]
    }
  }
}
```

### Empty with Prefix Filter

```terraform
action "aws_s3_empty_bucket" "logs_cleanup" {
  config {
    bucket_name = "my-logs-bucket"
    prefix      = "old-logs/"
    batch_size  = 500
  }
}
```

### Versioned Bucket Cleanup

```terraform
action "aws_s3_empty_bucket" "versioned" {
  config {
    bucket_name = "my-versioned-bucket"
    timeout     = 3600  # 1 hour for large buckets
  }
}
```

### CI/CD Pipeline Cleanup

```terraform
action "aws_s3_empty_bucket" "artifacts_cleanup" {
  config {
    bucket_name = aws_s3_bucket.build_artifacts.bucket
    prefix      = "builds/${var.build_id}/"
  }
}

resource "terraform_data" "post_deploy" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aws_s3_empty_bucket.artifacts_cleanup]
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `bucket_name` - (Required) Name of the S3 bucket to empty.

The following arguments are optional:

* `prefix` - (Optional) Only delete objects whose keys begin with this prefix. If not specified, all objects will be deleted.
* `batch_size` - (Optional) Number of objects to delete per batch operation. Must be between 1 and 1000. Defaults to 1000.
* `timeout` - (Optional) Timeout in seconds for the empty operation. Must be between 60 and 7200. Defaults to 1800 (30 minutes).

## Notes

* This action deletes all object versions and delete markers when versioning is enabled on the bucket.
* The action uses batch delete operations for efficiency, processing up to 1000 objects per API call.
* Progress updates are sent during the deletion process showing the number of objects deleted.
* The bucket itself is not deleted, only its contents.
* Large buckets may take significant time to empty. Consider adjusting the timeout for buckets with millions of objects.

## Required IAM Permissions

The action requires the following IAM permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListBucket",
        "s3:ListBucketVersions",
        "s3:DeleteObject",
        "s3:DeleteObjectVersion",
        "s3:GetBucketVersioning"
      ],
      "Resource": [
        "arn:aws:s3:::bucket-name",
        "arn:aws:s3:::bucket-name/*"
      ]
    }
  ]
}
```
