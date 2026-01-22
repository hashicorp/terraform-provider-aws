resource "aws_dynamodb_table_export" "test" {
{{- template "region" }}
  s3_bucket = aws_s3_bucket.test.id
  table_arn = aws_dynamodb_table.test.arn
}

# testAccTableExportConfig_baseConfig

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket        = var.rName
  force_destroy = true
}

resource "aws_dynamodb_table" "test" {
{{- template "region" }}
  name           = var.rName
  read_capacity  = 2
  write_capacity = 2
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }
}
