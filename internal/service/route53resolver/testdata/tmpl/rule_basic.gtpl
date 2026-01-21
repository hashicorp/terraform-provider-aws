resource "aws_route53_resolver_rule" "test" {
{{- template "region" }}
  domain_name = var.rName
  rule_type   = "SYSTEM"
{{- template "tags" }}
}
