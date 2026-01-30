resource "aws_s3vectors_vector_bucket" "test" {
{{- template "region" }}
  vector_bucket_name = var.rName

{{- template "tags" . }}
}