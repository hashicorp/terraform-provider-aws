resource "aws_s3_bucket_logging" "test" {
{{- template "region" }}
  bucket = aws_s3_bucket.test.id

  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = "log/"
}

resource "aws_s3_bucket" "log_bucket" {
{{- template "region" }}
  bucket = "${var.rName}-log"
}

resource "aws_s3_bucket_ownership_controls" "log_bucket_ownership" {
{{- template "region" }}
  bucket = aws_s3_bucket.log_bucket.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "log_bucket_acl" {
{{- template "region" }}
  depends_on = [aws_s3_bucket_ownership_controls.log_bucket_ownership]

  bucket = aws_s3_bucket.log_bucket.id
  acl    = "log-delivery-write"
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  depends_on = [aws_s3_bucket_acl.log_bucket_acl]

  bucket = var.rName
}
