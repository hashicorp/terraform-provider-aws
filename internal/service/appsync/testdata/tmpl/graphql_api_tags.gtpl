resource "aws_appsync_graphql_api" "test" {
  {{- template "region" }}
  authentication_type = "API_KEY"
  name                = var.rName
  visibility          = var.rName

  {{- template "tags" . }}
}
