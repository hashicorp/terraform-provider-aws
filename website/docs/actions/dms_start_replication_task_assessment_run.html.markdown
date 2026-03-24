---
subcategory: "DMS (Database Migration)"
layout: "aws"
page_title: "AWS: aws_dms_start_replication_task_assessment_run"
description: |-
  Starts an AWS DMS premigration assessment run for a replication task and waits for it to reach a terminal state.
---

# Action: aws_dms_start_replication_task_assessment_run

Starts an AWS DMS premigration assessment run for a replication task and waits for the assessment run to reach a terminal state.

For information about DMS pre-migration assessment, see the [Amazon Data Migration Tasks](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Tasks.html). For specific information about creating pre-migration assessment requests, see the [StartReplicationTaskAssessmentRun](https://docs.aws.amazon.com/dms/latest/APIReference/API_StartReplicationTaskAssessmentRun.html) page in the Amazon DMS API Reference.

~> **Note:** DMS premigration assessment requests can take several minutes to complete. This action will wait for the assessment to finish before continuing.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3_bucket" "assessment_results" {
  bucket = "example-dms-assessment-results"
}

resource "aws_iam_role" "dms_assessment" {
  name = "example-dms-assessment-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "dms.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "dms_assessment" {
  name = "example-dms-assessment-policy"
  role = aws_iam_role.dms_assessment.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "s3:GetBucketLocation",
        "s3:ListBucket",
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject"
      ]
      Resource = [
        aws_s3_bucket.assessment_results.arn,
        "${aws_s3_bucket.assessment_results.arn}/*"
      ]
    }]
  })
}

resource "aws_dms_replication_task" "example" {
  # ... replication task configuration
}

action "aws_dms_start_replication_task_assessment_run" "example" {
  config {
    assessment_run_name     = "example-assessment-run"
    replication_task_arn    = aws_dms_replication_task.example.replication_task_arn
    result_location_bucket  = aws_s3_bucket.assessment_results.bucket
    result_location_folder  = "premigration-assessments"
    service_access_role_arn = aws_iam_role.dms_assessment.arn
    timeout                 = 1800
  }
}

resource "terraform_data" "example" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aws_dms_start_replication_task_assessment_run.example]
    }
  }

  depends_on = [
    aws_dms_replication_task.example,
    aws_iam_role_policy.dms_assessment
  ]
}
```

## Argument Reference

This action supports the following arguments:

* `assessment_run_name` - (Required) Unique name for the DMS premigration assessment run.
* `exclude` - (Optional) List of individual assessment names to exclude from the run. Cannot be set with `include_only`.
* `include_only` - (Optional) List of individual assessment names to include in the run. Cannot be set with `exclude`.
* `replication_task_arn` - (Required) ARN of the DMS replication task to assess.
* `result_encryption_mode` - (Optional) Encryption mode for assessment results. Valid values are `SSE_KMS` and `SSE_S3`.
* `result_kms_key_arn` - (Optional) ARN of the KMS key used to encrypt assessment results when `result_encryption_mode` is `SSE_KMS`.
* `result_location_bucket` - (Required) Name of the S3 bucket where DMS stores assessment results.
* `result_location_folder` - (Optional) Folder prefix within the S3 bucket where DMS stores assessment results.
* `service_access_role_arn` - (Required) ARN of the IAM role that DMS assumes to write assessment results and start the run.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `timeout` - (Optional) Timeout in seconds to wait for the assessment run to complete. Defaults to 1800 seconds (30 minutes). Must be at least 60 seconds.

~> **Note:** `result_kms_key_arn` can only be specified when `result_encryption_mode` is `SSE_KMS`.
