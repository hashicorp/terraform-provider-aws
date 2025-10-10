resource "aws_s3_bucket_public_access_block" "test" {
{{- template "region" }}
  bucket = aws_s3_bucket.test.bucket

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket = var.rName
}
