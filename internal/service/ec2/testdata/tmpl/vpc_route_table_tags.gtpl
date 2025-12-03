resource "aws_route_table" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id

{{- template "tags" . }}
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.1.0.0/16"
}
