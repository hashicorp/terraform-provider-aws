resource "aws_s3_bucket_metadata_configuration" "test" {
{{- template "region" }}
  bucket = aws_s3_bucket.test.bucket

  metadata_configuration {
    inventory_table_configuration {
      configuration_state = "DISABLED"
    }

    journal_table_configuration {
      record_expiration {
        days       = 7
        expiration = "ENABLED"
      }
    }
  }
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket = var.rName
}
