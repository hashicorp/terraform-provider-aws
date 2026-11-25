resource "aws_s3tables_table_bucket_metrics_configuration" "test" {
{{- template "region" }}
  table_bucket_arn = aws_s3tables_table_bucket.test.arn
}

resource "aws_s3tables_table_bucket" "test" {
{{- template "region" }}
  name = var.rName
}
