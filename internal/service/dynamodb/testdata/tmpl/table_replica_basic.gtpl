resource "aws_dynamodb_table_replica" "test" {
  global_table_arn = aws_dynamodb_table.test.arn
{{- template "tags" . }}
}

resource "aws_dynamodb_table" "test" {
  provider         = "awsalternate"
  name             = var.rName
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  tags = {
    # Should not show up on `aws_dynamodb_table_replica`
    Name = var.rName
  }

  lifecycle {
    ignore_changes = [replica]
  }
}
