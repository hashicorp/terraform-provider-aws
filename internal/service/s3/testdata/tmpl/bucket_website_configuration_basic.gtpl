resource "aws_s3_bucket_website_configuration" "test" {
{{- template "region" }}
  bucket = aws_s3_bucket.test.id
  index_document {
    suffix = "index.html"
  }
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket = var.rName
}
