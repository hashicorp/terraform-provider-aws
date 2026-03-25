resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 2
  max_capacity       = 15

{{- template "tags" . }}
}

resource "aws_dynamodb_table" "test" {
  name           = var.rName
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "TestKey"

  attribute {
    name = "TestKey"
    type = "S"
  }
}
