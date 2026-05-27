resource "aws_s3_bucket_cors_configuration" "test" {
{{- template "region" }}
  bucket = aws_s3_bucket.test.id

  cors_rule {
    allowed_methods = ["PUT"]
    allowed_origins = ["https://www.example.com"]
  }
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket = var.rName
}

