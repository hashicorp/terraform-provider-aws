resource "aws_networkfirewall_firewall_policy_v2" "test" {
{{- template "region" }}
  name = var.rName

  firewall_policy {}
{{- template "tags" . }}
}
