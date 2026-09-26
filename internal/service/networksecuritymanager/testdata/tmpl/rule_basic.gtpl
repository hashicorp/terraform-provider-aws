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
{{- template "tags" . }}
}
