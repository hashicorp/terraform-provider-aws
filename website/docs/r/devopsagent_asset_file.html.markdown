---
subcategory: "DevOps Agent"
layout: "aws"
page_title: "AWS: aws_devopsagent_asset_file"
description: |-
  Manages a file within an AWS DevOps Agent Asset.
---

# Resource: aws_devopsagent_asset_file

Manages a file within an AWS DevOps Agent Asset. Asset files allow you to add individual files (e.g., documentation, configuration, scripts) to an existing asset without replacing the entire asset content.

## Example Usage

### Basic Usage

```terraform
resource "aws_devopsagent_agent_space" "example" {
  name = "example-space"
}

resource "aws_devopsagent_asset" "example" {
  agent_space_id = aws_devopsagent_agent_space.example.agent_space_id
  asset_type     = "skill"
  content_path   = "SKILL.md"
  content_body   = "# My Skill\n\nSkill description."

  metadata = jsonencode({
    name        = "example-skill"
    description = "An example skill"
    agent_types = ["GENERIC"]
  })
}

resource "aws_devopsagent_asset_file" "example" {
  agent_space_id = aws_devopsagent_agent_space.example.agent_space_id
  asset_id       = aws_devopsagent_asset.example.asset_id
  path           = "README.md"
  content_body   = "# Hello\n\nThis is a supplementary file for the skill."
}
```

### Multiple Files in an Asset

```terraform
resource "aws_devopsagent_asset_file" "readme" {
  agent_space_id = aws_devopsagent_agent_space.example.agent_space_id
  asset_id       = aws_devopsagent_asset.example.asset_id
  path           = "README.md"
  content_body   = "# My Skill\n\nDocumentation for this skill."
}

resource "aws_devopsagent_asset_file" "config" {
  agent_space_id = aws_devopsagent_agent_space.example.agent_space_id
  asset_id       = aws_devopsagent_asset.example.asset_id
  path           = "config.yaml"
  content_body   = "version: 1\nlog_level: info"
}
```

## Argument Reference

The following arguments are required:

* `agent_space_id` - (Required) Unique identifier of the agent space containing the asset. Forces new resource.
* `asset_id` - (Required) Unique identifier of the asset to add the file to. Forces new resource.
* `content_body` - (Required) Text content of the file.
* `path` - (Required) Path of the file within the asset (e.g., `README.md` or `docs/guide.md`). Forces new resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `asset_version` - Version of the asset after the file was created or last updated.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_devopsagent_asset_file.example
  identity = {
    agent_space_id = "space-12345678"
    asset_id       = "asset-87654321"
    path           = "README.md"
  }
}

resource "aws_devopsagent_asset_file" "example" {
  agent_space_id = "space-12345678"
  asset_id       = "asset-87654321"
  path           = "README.md"
  content_body   = "# Hello"
}
```

### Identity Schema

#### Required

* `agent_space_id` - Unique identifier of the agent space.
* `asset_id` - Unique identifier of the asset.
* `path` - Path of the file within the asset.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DevOps Agent Asset File using the `agent_space_id`, `asset_id`, and `path` separated by commas. For example:

```terraform
import {
  to = aws_devopsagent_asset_file.example
  id = "space-12345678,asset-87654321,README.md"
}
```

Using `terraform import`, import DevOps Agent Asset File using the `agent_space_id`, `asset_id`, and `path` separated by commas. For example:

```console
% terraform import aws_devopsagent_asset_file.example space-12345678,asset-87654321,README.md
```
