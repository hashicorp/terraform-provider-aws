resource "aws_devopsguru_service_integration" "test" {
{{- template "region" }}
  # Default to existing configured settings
  kms_server_side_encryption {}

  logs_anomaly_detection {
    opt_in_status = "DISABLED"
  }
  ops_center {
    opt_in_status = "DISABLED"
  }
}
