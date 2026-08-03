resource "aws_s3tables_table_bucket_policy" "test" {
{{- template "region" }}
  resource_policy  = data.aws_iam_policy_document.test.json
  table_bucket_arn = aws_s3tables_table_bucket.test.arn
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["s3tables:*"]
    principals {
      type        = "AWS"
      identifiers = [data.aws_caller_identity.current.account_id]
    }
    resources = ["${aws_s3tables_table_bucket.test.arn}/*"]
  }
}

resource "aws_s3tables_table_bucket" "test" {
{{- template "region" }}
  name = var.rName
}

data "aws_caller_identity" "current" {}
