resource "aws_cloudwatch_log_account_policy" "test" {
{{- template "region" }}
  policy_name = var.rName
  policy_type = "FIELD_INDEX_POLICY"

  policy_document = jsonencode({
    Fields = ["RequestId"]
  })

  selection_criteria = "LogGroupNamePrefix = \"/${var.rName}\""
}
