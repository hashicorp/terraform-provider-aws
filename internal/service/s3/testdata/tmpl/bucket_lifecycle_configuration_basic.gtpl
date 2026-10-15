resource "aws_s3_bucket_lifecycle_configuration" "test" {
{{- template "region" }}
  bucket = aws_s3_bucket.test.bucket
  rule {
    id     = var.rName
    status = "Enabled"

    expiration {
      days = 365
    }

    filter {
      prefix = "prefix/"
    }
  }
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket = var.rName
}
