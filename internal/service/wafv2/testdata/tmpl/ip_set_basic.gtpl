resource "aws_wafv2_ip_set" "test" {
{{- template "region" }}
  name               = var.rName
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32", "5.6.7.8/32"]
}
