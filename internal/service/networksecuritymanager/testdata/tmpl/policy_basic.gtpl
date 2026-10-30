resource "aws_networksecuritymanager_policy" "test" {
{{- template "region" }}
  name          = var.rName
  firewall_type = "SHIELD_ADVANCED"
  priority      = parseint(substr(sha1(var.rName), 0, 6), 16) + 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = true
  }

{{- template "tags" . }}
}
