resource "aws_dynamodb_table_item" "test" {
{{- template "region" }}
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key

  item = <<ITEM
{
  "TestTableHashKey": {"S": "${var.rName}"},
  "one": {"N": "11111"}
}
ITEM
}

resource "aws_dynamodb_table" "test" {
{{- template "region" }}
  name           = var.rName
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}
