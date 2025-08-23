resource "aws_codestarconnections_host" "test" {
{{- template "region" }}
  name              = var.rName
  provider_endpoint = "https://example.com"
  provider_type     = "GitHubEnterpriseServer"
}
