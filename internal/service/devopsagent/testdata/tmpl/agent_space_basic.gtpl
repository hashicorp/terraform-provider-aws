resource "aws_devopsagent_agent_space" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" . }}
}
