---
subcategory: "DevOps Agent"
layout: "aws"
page_title: "AWS: aws_devopsagent_asset_file"
description: |-
  Manages a file within an AWS DevOps Agent Asset.
---

# Resource: aws_devopsagent_asset_file

Manages a file within an AWS DevOps Agent Asset.

~> **Note:** `metadata` is an input retained in Terraform state. Terraform does not detect out-of-band changes to it. `metadata_actual` exposes the current metadata returned by AWS. Import cannot recover the original metadata input.

~> **Note:** Terraform serializes its own writes to files in the same asset. Concurrent writes to the same asset from outside a single Terraform run can cause AWS to drop a file from the asset's latest version.

## Example Usage

### Basic Usage

```terraform
resource "aws_devopsagent_agent_space" "example" {
  name = "example"
}

resource "aws_devopsagent_asset" "example" {
  agent_space_id = aws_devopsagent_agent_space.example.agent_space_id
  asset_type     = "skill"
  metadata = jsonencode({
    name        = "example"
    description = "Investigation procedures for example service issues."
    agent_types = ["GENERIC"]
  })

  content {
    file {
      path = "SKILL.md"
      body {
        text = "Follow the operational runbook in references/runbook.md."
      }
    }
  }
}

resource "aws_devopsagent_asset_file" "example" {
  agent_space_id = aws_devopsagent_asset.example.agent_space_id
  asset_id       = aws_devopsagent_asset.example.asset_id
  path           = "references/runbook.md"

  content {
    text = file("${path.module}/runbook.md")
  }
}
```

### Binary Content

```terraform
resource "aws_devopsagent_asset_file" "example" {
  agent_space_id = aws_devopsagent_asset.example.agent_space_id
  asset_id       = aws_devopsagent_asset.example.asset_id
  path           = "assets/topology.png"

  content {
    bytes = filebase64("${path.module}/topology.png")
  }
}
```

## Argument Reference

The following arguments are required:

* `agent_space_id` - (Required) Identifier of the Agent Space containing the asset. Changing this value requires replacement.
* `asset_id` - (Required) Identifier of the asset containing the file. Changing this value requires replacement.
* `content` - (Required) File content. Exactly one block is required. See [`content` Block](#content-block) below.
* `path` - (Required) Path of the file within the asset. Changing this value requires replacement.

The following arguments are optional:

* `metadata` - (Optional) JSON document describing the file. Updates replace supplied top-level fields and retain omitted fields in AWS; removing a field from this document does not delete it from AWS.
* `region` - (Optional) Region where this resource will be managed. Defaults to the Region set in the provider configuration.

### `content` Block

Exactly one of the following arguments must be configured:

* `bytes` - (Optional) Base64-encoded binary file content. Use `filebase64()` to load a local file.
* `text` - (Optional) Text file content.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_at` - Creation timestamp in RFC 3339 format.
* `metadata_actual` - Full JSON metadata returned by AWS, including retained fields and service-provided defaults.
* `updated_at` - Last update timestamp in RFC 3339 format.
* `version` - Asset version the file belongs to.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_devopsagent_asset_file.example
  identity = {
    agent_space_id = "12345678-1234-1234-1234-123456789012"
    asset_id       = "87654321-4321-4321-4321-210987654321"
    path           = "references/runbook.md"
  }
}
```

### Identity Schema

#### Required

* `agent_space_id` (String) Identifier of the Agent Space containing the asset.
* `asset_id` (String) Identifier of the asset containing the file.
* `path` (String) Path of the file within the asset.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an asset file using its Agent Space ID, asset ID, and path separated by commas. For example:

```terraform
import {
  to = aws_devopsagent_asset_file.example
  id = "12345678-1234-1234-1234-123456789012,87654321-4321-4321-4321-210987654321,references/runbook.md"
}
```

Using `terraform import`, import an asset file using its Agent Space ID, asset ID, and path separated by commas. For example:

```console
% terraform import aws_devopsagent_asset_file.example 12345678-1234-1234-1234-123456789012,87654321-4321-4321-4321-210987654321,references/runbook.md
```
