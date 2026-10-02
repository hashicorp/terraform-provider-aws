resource "aws_default_vpc" "test" {
{{- template "region" }}

{{- template "tags" . }}
}
