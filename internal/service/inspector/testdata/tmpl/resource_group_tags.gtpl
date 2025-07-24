resource "aws_inspector_resource_group" "test" {
{{- template "region" }}
  tags = {
    Name = "foo"
  }
{{- template "tags" }}
}
