resource "aws_appconfig_deployment_strategy" "test" {
  name = var.rName

  deployment_duration_in_minutes = 3
  growth_factor                  = 10
  replicate_to                   = "NONE"

{{- template "tags" . }}
}
