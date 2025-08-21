resource "aws_s3_bucket_acl" "test" {
{{- template "region" }}
  depends_on = [aws_s3_bucket_ownership_controls.test]

  bucket = aws_s3_bucket.test.bucket
  acl    = "private"
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket = var.rName
}

resource "aws_s3_bucket_ownership_controls" "test" {
{{- template "region" }}
  bucket = aws_s3_bucket.test.bucket
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}
