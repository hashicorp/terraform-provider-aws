resource "aws_arczonalshift_zonal_autoshift_configuration" "test" {
{{- template "region" }}
  resource_arn           = aws_lb.test.arn
  zonal_autoshift_status = "ENABLED"

  outcome_alarms {
    alarm_identifier = aws_cloudwatch_metric_alarm.outcome.arn
    type             = "CLOUDWATCH"
  }
}

resource "aws_lb" "test" {
{{- template "region" }}
  name               = var.rName
  internal           = true
  load_balancer_type = "application"
  subnets            = aws_subnet.test[*].id

  enable_deletion_protection = false
  enable_zonal_shift         = true

  tags = {
    Name = var.rName
  }
}

resource "aws_cloudwatch_metric_alarm" "outcome" {
{{- template "region" }}
  alarm_name          = "${var.rName}-outcome"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = "TargetResponseTime"
  namespace           = "AWS/ApplicationELB"
  period              = 60
  statistic           = "Average"
  threshold           = 1
  alarm_description   = "Outcome alarm for zonal autoshift practice run"

  dimensions = {
    LoadBalancer = aws_lb.test.arn_suffix
  }
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}
