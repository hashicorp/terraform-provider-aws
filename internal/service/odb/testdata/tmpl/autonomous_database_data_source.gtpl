data "aws_odb_autonomous_database" "test" {
{{- template "region" }}
  id = aws_odb_autonomous_database.test.id
}
