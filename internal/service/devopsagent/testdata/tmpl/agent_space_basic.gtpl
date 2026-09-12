resource "aws_devopsagent_agent_space" "test" {
  name = "tf-acc-test-devopsagent"
{{- template "tags" . }}
}
