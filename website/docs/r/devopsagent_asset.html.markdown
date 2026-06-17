---
subcategory: "DevOps Agent"
layout: "aws"
page_title: "AWS: aws_devopsagent_asset"
description: |-
  Manages an AWS DevOps Agent Asset.
---

# Resource: aws_devopsagent_asset

Manages an AWS DevOps Agent Asset. Assets represent operational knowledge artifacts such as skills, AGENTS.md files, and attachments that shape how the DevOps Agent behaves within an Agent Space.

## Example Usage

### Simple Skill with Inline Content

```terraform
resource "aws_devopsagent_agent_space" "example" {
  name = "example-space"
}

resource "aws_devopsagent_asset" "example" {
  agent_space_id = aws_devopsagent_agent_space.example.agent_space_id
  asset_type     = "skill"
  content_path   = "SKILL.md"
  content_body   = "# RDS Performance Investigation\n\nUse this skill when investigating database latency or connection errors."

  metadata = jsonencode({
    name        = "rds-performance-investigation"
    description = "Investigation procedures for RDS performance issues."
    agent_types = ["GENERIC"]
  })
}
```

### Skill from a Local File

```terraform
resource "aws_devopsagent_agent_space" "example" {
  name = "example-space"
}

resource "aws_devopsagent_asset" "example" {
  agent_space_id = aws_devopsagent_agent_space.example.agent_space_id
  asset_type     = "skill"
  filename       = "${path.module}/skills/SKILL.md"
  content_path   = "SKILL.md"

  metadata = jsonencode({
    name        = "rds-performance-investigation"
    description = "Investigation procedures for RDS performance issues."
    agent_types = ["GENERIC"]
  })
}
```

### Complex Skill from a Zip Bundle

When a skill includes multiple files (e.g., a `SKILL.md` plus reference material in a `references/` directory), upload the content as a zip file. The `content_path` attribute is ignored for zip uploads — file paths are taken from the archive structure.

```terraform
resource "aws_devopsagent_agent_space" "example" {
  name = "example-space"
}

resource "aws_devopsagent_asset" "complex_skill" {
  agent_space_id = aws_devopsagent_agent_space.example.agent_space_id
  asset_type     = "skill"
  filename       = "${path.module}/skills/rds-investigation.zip"

  metadata = jsonencode({
    agent_types = ["GENERIC"]
  })
}
```

### AGENTS.md

```terraform
resource "aws_devopsagent_agent_space" "example" {
  name = "example-space"
}

resource "aws_devopsagent_asset" "example" {
  agent_space_id = aws_devopsagent_agent_space.example.agent_space_id
  asset_type     = "agents_md"
  content_path   = "AGENTS.md"
  content_body   = "# Triage Instructions\n\nFollow these steps for new incidents."

  metadata = jsonencode({
    agent_type = "INCIDENT_TRIAGE"
  })
}
```

### Attachment from a Local File

```terraform
resource "aws_devopsagent_agent_space" "example" {
  name = "example-space"
}

resource "aws_devopsagent_asset" "topology_diagram" {
  agent_space_id = aws_devopsagent_agent_space.example.agent_space_id
  asset_type     = "attachment"
  filename       = "${path.module}/diagrams/topology.png"
  content_path   = "topology.png"

  metadata = jsonencode({
    filename  = "topology.png"
    extension = "png"
    size      = 184320
  })
}
```

## Argument Reference

The following arguments are required:

* `agent_space_id` - (Required) Unique identifier of the agent space where the asset is created. Forces new resource.
* `asset_type` - (Required) Type of asset to create. Valid values: `skill`, `agents_md`, `attachment`. Forces new resource.

The following arguments are optional:

* `content_body` - (Optional) Inline text content of the asset file (UTF-8). Mutually exclusive with `filename`. Used for simple single-file assets where the content can be expressed as a string.
* `content_path` - (Optional) Path of the file within the asset (e.g., `SKILL.md`, `references/guide.md`). Used with both `content_body` and non-zip `filename` uploads. Ignored for `.zip` file uploads (paths come from the archive structure). Defaults to the base filename when using `filename`, or `SKILL.md` when using `content_body` without specifying this attribute.
* `filename` - (Optional) Path to a local file to upload as the asset content. Mutually exclusive with `content_body`. For `.zip` files, the content is uploaded as a zip bundle containing multiple files. For all other file types, the content is uploaded as a single file.
* `metadata` - (Optional) JSON-encoded metadata object describing the asset. The required keys depend on the `asset_type`. For zip bundle skills, `name` and `description` are extracted from the `SKILL.md` frontmatter and any values supplied here are ignored. See the [AWS documentation](https://docs.aws.amazon.com/devopsagent/latest/userguide/about-aws-devops-agent-managing-assets.html) for details on metadata keys for each asset type.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `asset_id` - Unique identifier of the Asset assigned by the service.
* `asset_version` - Current version number of the Asset. Incremented on each update.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_devopsagent_asset.example
  identity = {
    agent_space_id = "8f6187a7-0388-4926-8217-3a0fe32f757c"
    asset_id       = "a1b2c3d4-5678-90ab-cdef-example11111"
  }
}

resource "aws_devopsagent_asset" "example" {
  agent_space_id = "8f6187a7-0388-4926-8217-3a0fe32f757c"
  asset_type     = "skill"
  content_path   = "SKILL.md"
  content_body   = "# Example Skill"

  metadata = jsonencode({
    name        = "example-skill"
    description = "An example skill."
    agent_types = ["GENERIC"]
  })
}
```

### Identity Schema

#### Required

* `agent_space_id` - Unique identifier of the agent space.
* `asset_id` - Unique identifier of the asset.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DevOps Agent Asset using `agent_space_id` and `asset_id` separated by a comma. For example:

```terraform
import {
  to = aws_devopsagent_asset.example
  id = "8f6187a7-0388-4926-8217-3a0fe32f757c,a1b2c3d4-5678-90ab-cdef-example11111"
}
```

Using `terraform import`, import DevOps Agent Asset using `agent_space_id` and `asset_id` separated by a comma. For example:

```console
% terraform import aws_devopsagent_asset.example 8f6187a7-0388-4926-8217-3a0fe32f757c,a1b2c3d4-5678-90ab-cdef-example11111
```
