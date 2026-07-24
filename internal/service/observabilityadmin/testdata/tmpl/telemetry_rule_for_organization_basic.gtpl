resource "aws_observabilityadmin_telemetry_rule_for_organization" "test" {
{{- template "region" }}
  rule_name = var.rName

  rule {
    resource_type  = "AWS::EC2::VPC"
    telemetry_type = "Logs"
  }

{{- template "tags" . }}

  depends_on = [aws_observabilityadmin_telemetry_evaluation_for_organization.test]
}

resource "aws_observabilityadmin_telemetry_evaluation_for_organization" "test" {
{{- template "region" }}
}
