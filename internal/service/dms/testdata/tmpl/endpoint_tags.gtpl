resource "aws_dms_endpoint" "test" {
  database_name = "tf-test-dms-db"
  endpoint_id   = var.rName
  endpoint_type = "source"
  engine_name   = "aurora"
  password      = "tftest"
  port          = 3306
  server_name   = "tftest"
  ssl_mode      = "none"
  username      = "tftest"
{{- template "tags" . }}
}
