resource "aws_agentregistry_registry" "test" {
{{- template "region" }}
  name = var.rName

  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }
{{- template "tags" . }}
}
