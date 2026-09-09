data "aws_agentregistry_registry_record" "test" {
{{- template "region" }}
  registry_id = aws_agentregistry_registry.test.registry_id
  record_id   = aws_agentregistry_registry_record.test.record_id
}
