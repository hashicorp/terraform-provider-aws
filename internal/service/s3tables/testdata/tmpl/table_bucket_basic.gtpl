resource "aws_s3tables_table_bucket" "test" {
{{- template "region" }}
  name = var.rName
}
