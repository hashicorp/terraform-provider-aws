resource "aws_observabilityadmin_telemetry_rule_for_organization" "test" {
  rule_name = var.rName

  rule {
    telemetry_type = "Metrics"
  }

  {{- template "region" }}
}