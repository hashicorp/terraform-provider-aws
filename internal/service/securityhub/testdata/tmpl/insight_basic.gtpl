resource "aws_securityhub_insight" "test" {
{{- template "region" }}
  filters {
    aws_account_id {
      comparison = "EQUALS"
      value      = "1234567890"
    }
  }

  group_by_attribute = "AwsAccountId"

  name = var.rName

  depends_on = [aws_securityhub_account.test]
}

resource "aws_securityhub_account" "test" {
{{- template "region" }}
}