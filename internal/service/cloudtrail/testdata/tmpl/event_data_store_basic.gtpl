resource "aws_cloudtrail_event_data_store" "test" {
{{- template "region" }}
  name = var.rName

  termination_protection_enabled = false # For ease of deletion.
{{- template "tags" . }}
}
