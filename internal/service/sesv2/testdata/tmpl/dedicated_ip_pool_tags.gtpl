resource "aws_sesv2_dedicated_ip_pool" "test" {
  pool_name = var.rName
{{- template "tags" . }}
}
