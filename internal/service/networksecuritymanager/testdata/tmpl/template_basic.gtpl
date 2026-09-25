resource "aws_networksecuritymanager_template" "test" {
{{- template "region" }}
  name          = var.rName
  firewall_type = "WAF"
  rule_arns     = [aws_networksecuritymanager_rule.test.arn]
{{- template "tags" . }}
}

resource "aws_networksecuritymanager_rule" "test" {
{{- template "region" }}
  name          = var.rName
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}
