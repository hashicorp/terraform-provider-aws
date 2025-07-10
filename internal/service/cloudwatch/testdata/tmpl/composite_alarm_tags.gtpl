resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name = var.rName
  alarm_rule = join(" OR ", formatlist("ALARM(%s)", aws_cloudwatch_metric_alarm.test[*].alarm_name))

{{- template "tags" . }}
}

resource "aws_cloudwatch_metric_alarm" "test" {
  count = 2

  alarm_name          = "${var.rName}-${count.index}"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 120
  statistic           = "Average"
  threshold           = 80

  dimensions = {
    InstanceId = "i-abcd1234"
  }
}
