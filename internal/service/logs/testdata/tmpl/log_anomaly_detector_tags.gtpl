resource aws_cloudwatch_log_group "test" {
  count = 2

  name = "${var.rName}-${count.index}"
}

data aws_cloudwatch_log_groups "test" {
  log_group_name_prefix = var.rName

  depends_on = [aws_cloudwatch_log_group.test[0], aws_cloudwatch_log_group.test[1]]
}

resource "aws_logs_log_anomaly_detector" "test" {
  detector_name        = var.rName
  log_arn_group_list   = [aws_cloudwatch_log_group.test.arn]
  evaluation_frequency = "TEN_MIN"
  enabled              = "false"

{{- template "tags" . }}
}