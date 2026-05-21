resource "aws_internet_gateway" "test" {
{{- template "region" }}
  tags = {
    Name = var.rName
  }
}
