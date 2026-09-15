resource "aws_s3tables_table" "test" {
{{- template "region" }}
  name             = replace(var.rName, "-", "_")
  namespace        = aws_s3tables_namespace.test.namespace
  table_bucket_arn = aws_s3tables_namespace.test.table_bucket_arn
  format           = "ICEBERG"

{{- template "tags" . }}
}

resource "aws_s3tables_namespace" "test" {
  namespace        = replace(var.rName, "-", "_")
  table_bucket_arn = aws_s3tables_table_bucket.test.arn

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_s3tables_table_bucket" "test" {
  name = var.rName
}