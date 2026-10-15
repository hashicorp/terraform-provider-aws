resource "aws_dynamodb_resource_policy" "test" {
{{- template "region" }}
  resource_arn = aws_dynamodb_table.test.arn
  policy       = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["dynamodb:*"]
    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
    resources = [
      aws_dynamodb_table.test.arn,
      "${aws_dynamodb_table.test.arn}/*",
    ]
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

# testAccTableConfig_basic

resource "aws_dynamodb_table" "test" {
{{- template "region" }}
  name           = var.rName
  read_capacity  = 1
  write_capacity = 1
  hash_key       = var.rName

  attribute {
    name = var.rName
    type = "S"
  }
}
