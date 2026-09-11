resource "aws_glue_catalog_table" "test" {
{{- template "region" }}
  name          = var.rName
  database_name = aws_glue_catalog_database.test.name
}

resource "aws_glue_catalog_database" "test" {
{{- template "region" }}
  name = var.rName
}
