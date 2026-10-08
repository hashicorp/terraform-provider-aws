resource "aws_eip" "test" {
{{- template "region" }}
  domain = "vpc"
{{- template "tags" . }}
}
