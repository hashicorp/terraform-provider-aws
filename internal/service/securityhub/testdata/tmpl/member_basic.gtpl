resource "aws_securityhub_member" "test" {
{{- template "region" }}
  depends_on = [aws_securityhub_account.test]
  account_id = "111111111111"
}

resource "aws_securityhub_account" "test" {
{{- template "region" }}
}
