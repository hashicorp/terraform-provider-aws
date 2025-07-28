resource "aws_dynamodb_table" "test" {
  hash_key       = "TestTableHashKey"
  name           = var.rName
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

{{- template "tags" . }}
}