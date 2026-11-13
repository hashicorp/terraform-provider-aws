resource "aws_arczonalshift_autoshift_observer_notification_status" "test" {
{{- template "region" }}
  status = "ENABLED"
}
