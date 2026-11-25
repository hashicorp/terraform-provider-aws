resource "aws_ec2_public_ipv4_pool" "test" {
{{- template "region" }}
{{- template "tags" . }}
}
