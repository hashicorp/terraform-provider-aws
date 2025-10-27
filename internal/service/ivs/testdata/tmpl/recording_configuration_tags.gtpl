resource "aws_ivs_recording_configuration" "test" {
{{- template "region" }}
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.test.id
    }
  }
{{- template "tags" }}
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket        = var.rName
  force_destroy = true
}
