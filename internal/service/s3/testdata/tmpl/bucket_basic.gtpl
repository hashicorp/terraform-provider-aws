resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket = var.rName

{{- template "tags" . }}
}
