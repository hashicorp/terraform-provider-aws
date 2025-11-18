resource "aws_s3tables_table_replication" "test" {
{{- template "region" }}
  table_arn = aws_s3tables_table.test.arn
}

resource "aws_s3tables_table" "test" {
{{- template "region" }}
  name             = var.rName
  namespace        = aws_s3tables_namespace.test.namespace
  table_bucket_arn = aws_s3tables_namespace.test.table_bucket_arn
  format           = "ICEBERG"
}

resource "aws_s3tables_namespace" "test" {
{{- template "region" }}
  namespace        = var.rName
  table_bucket_arn = aws_s3tables_table_bucket.test.arn

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_s3tables_table_bucket" "test" {
{{- template "region" }}
  name = var.rName
}