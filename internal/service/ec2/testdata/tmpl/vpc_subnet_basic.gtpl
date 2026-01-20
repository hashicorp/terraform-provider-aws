resource "aws_subnet" "test" {
{{- template "region" }}
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

{{- template "tags" . }}
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.1.0.0/16"
}
