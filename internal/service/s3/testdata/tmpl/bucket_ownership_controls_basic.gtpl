resource "aws_s3_bucket_ownership_controls" "test" {
{{- template "region" }}
  bucket = aws_s3_bucket.test.bucket
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket = var.rName
}
