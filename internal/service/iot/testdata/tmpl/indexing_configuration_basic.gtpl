resource "aws_iot_indexing_configuration" "test" {
{{- template "region" }}
  thing_group_indexing_configuration {
    thing_group_indexing_mode = "OFF"
  }

  thing_indexing_configuration {
    thing_indexing_mode = "OFF"
  }
}
