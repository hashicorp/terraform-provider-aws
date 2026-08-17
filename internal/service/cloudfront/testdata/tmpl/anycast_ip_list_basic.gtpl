resource "aws_cloudfront_anycast_ip_list" "test" {
  name     = var.rName
  ip_count = 3
{{- template "tags" . }}
}
