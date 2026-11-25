resource "aws_iot_topic_rule" "test" {
{{- template "region" }}
  name        = var.rName
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"
{{- template "tags" . }}
}
