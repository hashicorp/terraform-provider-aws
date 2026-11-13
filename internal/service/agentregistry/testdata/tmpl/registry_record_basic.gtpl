resource "aws_agentregistry_registry" "test" {
{{- template "region" }}
  name = var.rName
  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }
}

resource "aws_agentregistry_registry_record" "test" {
{{- template "region" }}
  registry_id = aws_agentregistry_registry.test.registry_id
  name        = var.rName
  record_type = "CUSTOM"

  descriptors {
    custom {
      data = jsonencode({
        name = var.rName
      })
    }
  }
{{- template "tags" . }}
}
