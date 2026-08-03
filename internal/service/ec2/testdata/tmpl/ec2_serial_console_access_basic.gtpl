resource "aws_ec2_serial_console_access" "test" {
{{- template "region" }}
  enabled = true
}
