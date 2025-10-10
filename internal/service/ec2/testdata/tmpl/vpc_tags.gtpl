resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.1.0.0/16"

{{- template "tags" . }}
}
