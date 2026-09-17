resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
{{- template "region" }}
  bucket = aws_s3_bucket.test.bucket

  rule {
    # This is Amazon S3 bucket default encryption.
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket = var.rName
}
