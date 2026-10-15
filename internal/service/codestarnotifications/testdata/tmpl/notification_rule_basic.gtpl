resource "aws_codestarnotifications_notification_rule" "test" {
{{- template "region" }}
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = var.rName
  resource       = aws_codecommit_repository.test.arn
  status         = "ENABLED"

  target {
    address = aws_sns_topic.test.arn
  }
{{- template "tags" . }}
}

# testAccNotificationRuleConfig_base

resource "aws_codecommit_repository" "test" {
{{- template "region" }}
  repository_name = var.rName
}

resource "aws_sns_topic" "test" {
{{- template "region" }}
  name = var.rName
}
