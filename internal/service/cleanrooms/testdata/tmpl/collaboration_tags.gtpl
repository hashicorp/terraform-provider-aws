resource "aws_cleanrooms_collaboration" "test" {
{{- template "region" }}
  name                     = var.rName
  creator_member_abilities = ["CAN_QUERY", "CAN_RECEIVE_RESULTS"]
  creator_display_name     = var.rName
  description              = var.rName
  query_log_status         = "DISABLED"
  analytics_engine         = "SPARK"

  data_encryption_metadata {
    allow_clear_text                            = true
    allow_duplicates                            = true
    allow_joins_on_columns_with_different_names = true
    preserve_nulls                              = false
  }

  tags = {
    Project = var.rName
  }

{{- template "tags" }}
}