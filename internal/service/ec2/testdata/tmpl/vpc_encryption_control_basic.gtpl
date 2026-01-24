resource "aws_vpc_encryption_control" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id
  mode   = "monitor"

{{- template "tags" . }}
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.1.0.0/16"
}
