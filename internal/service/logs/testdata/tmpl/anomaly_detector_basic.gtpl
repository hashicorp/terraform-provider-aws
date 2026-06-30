resource "aws_cloudwatch_log_anomaly_detector" "test" {
{{- template "region" }}
  detector_name           = var.rName
  log_group_arn_list      = [aws_cloudwatch_log_group.test.arn]
  anomaly_visibility_time = 7
  evaluation_frequency    = "TEN_MIN"
  enabled                 = "false"

{{- template "tags" . }}
}

resource "aws_cloudwatch_log_group" "test" {
{{- template "region" }}
  name = var.rName
}
