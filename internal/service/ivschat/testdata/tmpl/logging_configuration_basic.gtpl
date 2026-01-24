resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket        = var.rName
  force_destroy = true
}

resource "aws_ivschat_logging_configuration" "test" {
{{- template "region" }}
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.test.id
    }
  }
{{- template "tags" }}
}
