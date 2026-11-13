resource "aws_securityhub_account" "test" {
{{- template "region" }}
}

resource "aws_securityhub_action_target" "test" {
{{- template "region" }}
  depends_on  = [aws_securityhub_account.test]
  description = "description1"
  identifier  = "testaction"
  name        = "Test action"
}
