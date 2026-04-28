resource "aws_ec2_managed_resource_visibility" "test" {
{{- template "region" }}
  default_visibility = "hidden"
}
