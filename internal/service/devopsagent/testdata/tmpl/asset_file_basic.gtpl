resource "aws_devopsagent_asset_file" "test" {
{{- template "region" }}
  agent_space_id = aws_devopsagent_asset.test.agent_space_id
  asset_id       = aws_devopsagent_asset.test.asset_id
  path           = "references/runbook.md"

  content {
    text = "Check the runbook first."
  }
}

resource "aws_devopsagent_asset" "test" {
{{- template "region" }}
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_type     = "skill"
  metadata = jsonencode({
    name        = var.rName
    description = "Terraform acceptance test skill."
    agent_types = ["GENERIC"]
  })

  content {
    file {
      path = "SKILL.md"
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
