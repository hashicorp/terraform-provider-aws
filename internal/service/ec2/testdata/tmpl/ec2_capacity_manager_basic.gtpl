resource "aws_ec2_capacity_manager" "test" {
{{- template "region" }}
  enabled = true
}
