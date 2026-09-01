data "aws_ram_resource_share" "test" {
{{- template "region" }}
  name                  = aws_ram_resource_share.test.name
  resource_owner        = "SELF"
  resource_share_status = "ACTIVE"
}
