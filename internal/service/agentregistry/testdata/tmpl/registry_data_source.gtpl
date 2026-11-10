data "aws_agentregistry_registry" "test" {
{{- template "region" }}
  registry_id = aws_agentregistry_registry.test.registry_id
}
