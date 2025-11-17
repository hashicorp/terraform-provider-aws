resource "aws_s3_bucket_versioning" "test" {
{{- template "region" }}
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket = var.rName
}
