resource "aws_s3_bucket_notification" "test" {
{{- template "region" }}
  bucket = aws_s3_bucket.test.id

  eventbridge = true
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket = var.rName
}

resource "aws_s3_bucket_public_access_block" "test" {
{{- template "region" }}
  bucket = aws_s3_bucket.test.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "test" {
{{- template "region" }}
  bucket = aws_s3_bucket.test.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "test" {
{{- template "region" }}
  depends_on = [
    aws_s3_bucket_public_access_block.test,
    aws_s3_bucket_ownership_controls.test,
  ]

  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}
