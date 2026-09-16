resource "aws_cloudwatch_log_storage_tier_policy" "test" {
{{- template "region" }}
  storage_tier = "INTELLIGENT_TIERING"
}
