resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
{{- template "region" }}
  name = var.rName

  credential_provider_vendor = "GithubOauth2"

  oauth2_provider_config {
    github_oauth2_provider_config {
      client_id     = "test-client-id"
      client_secret = "test-client-secret"
    }
  }

{{- template "tags" . }}
}
