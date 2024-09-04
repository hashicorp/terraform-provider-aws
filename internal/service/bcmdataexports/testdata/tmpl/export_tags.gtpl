resource "aws_bcmdataexports_export" "test" {
  export {
    name = var.rName
    data_query {
      query_statement = "SELECT identity_line_item_id, identity_time_interval, line_item_product_code,line_item_unblended_cost FROM COST_AND_USAGE_REPORT"
      table_configurations = {
        "COST_AND_USAGE_REPORT" = {
          "TIME_GRANULARITY"                      = "HOURLY",
          "INCLUDE_RESOURCES"                     = "FALSE",
          "INCLUDE_MANUAL_DISCOUNT_COMPATIBILITY" = "FALSE",
          "INCLUDE_SPLIT_COST_ALLOCATION_DATA"    = "FALSE",
        }
      }
    }
    destination_configurations {
      s3_destination {
        s3_bucket = aws_s3_bucket.test.bucket
        s3_prefix = aws_s3_bucket.test.bucket_prefix
        s3_region = aws_s3_bucket.test.region
        s3_output_configurations {
          overwrite   = "OVERWRITE_REPORT"
          format      = "TEXT_OR_CSV"
          compression = "GZIP"
          output_type = "CUSTOM"
        }
      }
    }
    refresh_cadence {
      frequency = "SYNCHRONOUS"
    }
  }
{{- template "tags" . }}
}

# testAccExportConfigBase

resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.test.bucket
  policy = jsonencode({
    Id = var.rName
    Statement = [{
      Action = [
        "s3:PutObject",
        "s3:GetBucketPolicy"
      ]
      Effect = "Allow"
      Sid    = "EnableAWSDataExportsToWriteToS3AndCheckPolicy"
      Principal = {
        Service = [
          "billingreports.amazonaws.com",
          "bcm-data-exports.amazonaws.com"
        ]
      }
      Resource = [
        aws_s3_bucket.test.arn,
        "${aws_s3_bucket.test.arn}/*",
      ]
    }]
    Version = "2012-10-17"
  })
}

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}
