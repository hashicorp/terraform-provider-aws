resource "aws_cloudwatch_log_metric_filter" "test" {
{{- template "region" }}
  name           = "${var.rName}-filter"
  pattern        = ""
  log_group_name = aws_cloudwatch_log_group.test.name

  metric_transformation {
    name      = "metric1"
    namespace = "ns1"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_group" "test" {
{{- template "region" }}
  name              = "${var.rName}-group"
  retention_in_days = 1
}
