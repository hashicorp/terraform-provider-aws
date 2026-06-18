resource "aws_backup_report_plan" "test" {
  name        = var.rName
  description = var.rName

  report_delivery_channel {
    formats = [
      "CSV"
    ]
    s3_bucket_name = aws_s3_bucket.test.id
  }

  report_setting {
    report_template = "RESTORE_JOB_REPORT"
  }

{{- template "tags" . }}
}

# testAccReportPlanConfig_base

resource "aws_s3_bucket" "test" {
  bucket = local.bucket_name
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket                  = aws_s3_bucket.test.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

locals {
  bucket_name = replace(var.rName, "_", "-")
}
