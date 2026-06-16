resource "aws_cleanrooms_membership" "test" {
{{- template "region" }}
  collaboration_id = aws_cleanrooms_collaboration.test.id
  query_log_status = "DISABLED"
{{- template "tags" . }}
}

resource "aws_cleanrooms_collaboration" "test" {
{{- template "region" }}
  name                     = var.rName
  creator_member_abilities = ["CAN_QUERY", "CAN_RECEIVE_RESULTS"]
  creator_display_name     = "Creator"
  description              = var.rName
  query_log_status         = "DISABLED"
  analytics_engine         = "SPARK"
}
