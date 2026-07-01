resource "aws_wafv2_web_acl_rule" "test" {
{{- template "region" }}
  name        = var.rName
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    block {}
  }

  statement {
    geo_match_statement {
      country_codes = ["US", "CA"]
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = var.rName
    sampled_requests_enabled   = false
  }
}

resource "aws_wafv2_web_acl" "test" {
{{- template "region" }}
  name  = var.rName
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = var.rName
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}
