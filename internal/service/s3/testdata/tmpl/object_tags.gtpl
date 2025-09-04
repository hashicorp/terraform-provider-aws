resource "aws_s3_object" "test" {
{{- template "region" }}
  # Must have bucket versioning enabled first
  bucket = aws_s3_bucket_versioning.test.bucket
  key    = var.rName

{{- template "tags" . }}
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket = var.rName
}

resource "aws_s3_bucket_versioning" "test" {
{{- template "region" }}
  bucket = aws_s3_bucket.test.bucket
  versioning_configuration {
    status = "Enabled"
  }
}
