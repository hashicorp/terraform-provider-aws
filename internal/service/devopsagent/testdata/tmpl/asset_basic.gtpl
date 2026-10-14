resource "aws_devopsagent_asset" "test" {
{{- template "region" }}
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_type     = "agents_md"
  metadata       = jsonencode({ agent_type = "GENERIC" })

  content {
    file {
      path = "AGENTS.md"
      body {
        text = "Follow the operational runbook."
      }
    }
  }
}

resource "aws_devopsagent_agent_space" "test" {
{{- template "region" }}
  name = var.rName
}
