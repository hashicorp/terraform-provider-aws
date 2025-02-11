resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

{{- template "tags" . }}
}
