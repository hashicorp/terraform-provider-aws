resource "aws_config_connector" "test" {
{{- template "region" }}
  azure {
    client_identifier = "00000000-0000-0000-0000-000000000000"
    tenant_identifier = "11111111-1111-1111-1111-111111111111"
  }
{{- template "tags" . }}
}
