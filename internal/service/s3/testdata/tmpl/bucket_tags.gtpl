resource "aws_s3_bucket" "test" {
  bucket = var.rName

{{- template "tags" . }}
}
