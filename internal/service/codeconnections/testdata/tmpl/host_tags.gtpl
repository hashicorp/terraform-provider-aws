resource "aws_codeconnections_host" "test" {
  name              = var.rName
  provider_endpoint = "https://example.com"
  provider_type     = "GitHubEnterpriseServer"
{{- template "tags" . }}
}
