---
subcategory: "DevOps Agent"
layout: "aws"
page_title: "AWS: aws_devopsagent_asset"
description: |-
  Manages an AWS DevOps Agent Asset.
---

# Resource: aws_devopsagent_asset

Manages an AWS DevOps Agent Asset.

~> **Note:** `content` and `metadata` are inputs retained in Terraform state. Terraform does not detect out-of-band changes to these inputs. `metadata_all` exposes the current metadata returned by AWS. Import cannot recover the original content or metadata input.

~> **Note:** Updating ZIP content replaces all files in the asset, including files managed separately. Updating single-file content adds or replaces that file without deleting other files. Updating source URL content re-syncs from the original source; an unchanged URL does not trigger a re-sync.

## Example Usage

### Basic Usage

```terraform
resource "aws_devopsagent_agent_space" "example" {
  name = "example"
}

resource "aws_devopsagent_asset" "example" {
  agent_space_id = aws_devopsagent_agent_space.example.agent_space_id
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
```

## Argument Reference

The following arguments are required:

* `agent_space_id` - (Required) Identifier of the Agent Space containing the asset. Changing this value requires replacement.
* `asset_type` - (Required) Type of asset. Changing this value requires replacement. See [Managing assets](https://docs.aws.amazon.com/devopsagent/latest/userguide/about-aws-devops-agent-managing-assets.html) for supported types and their metadata requirements.
* `content` - (Required) Content to upload. Exactly one block is required. See [`content` Block](#content-block) below.

The following arguments are optional:

* `metadata` - (Optional) JSON document describing the asset. Required metadata fields depend on `asset_type`. Updates replace supplied top-level fields and retain omitted fields in AWS; removing a field from this document does not delete it from AWS. Changes to metadata alone do not re-upload content.
* `region` - (Optional) Region where this resource will be managed. Defaults to the Region set in the provider configuration.

### `content` Block

Exactly one of the following blocks must be configured:

* `file` - (Optional) Single file to add or replace. See [`file` Block](#file-block) below.
* `source_url` - (Optional) Source URL to import. See [`source_url` Block](#source_url-block) below.
* `zip` - (Optional) ZIP archive containing asset files. See [`zip` Block](#zip-block) below.

### `file` Block

* `body` - (Required) File content. See [`body` Block](#body-block) below.
* `metadata` - (Optional) JSON metadata for the file.
* `path` - (Required) Path of the file within the asset.

### `body` Block

Exactly one of the following arguments must be configured:

* `bytes` - (Optional) Base64-encoded binary file content.
* `text` - (Optional) Text file content.

### `source_url` Block

* `url` - (Required) Source URL from which to import asset content.

### `zip` Block

* `zip_file` - (Required) Base64-encoded ZIP archive. Use `filebase64()` to load a local ZIP file.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `asset_id` - Identifier of the asset.
* `created_at` - Creation timestamp in RFC 3339 format.
* `metadata_all` - Full JSON metadata returned by AWS, including retained fields and service-provided defaults.
* `updated_at` - Last update timestamp in RFC 3339 format.
* `version` - Current asset version.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_devopsagent_asset.example
  identity = {
    agent_space_id = "12345678-1234-1234-1234-123456789012"
    asset_id       = "87654321-4321-4321-4321-210987654321"
  }
}
```

### Identity Schema

#### Required

* `agent_space_id` (String) Identifier of the Agent Space containing the asset.
* `asset_id` (String) Identifier of the asset.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an asset using its Agent Space ID and asset ID separated by a comma. For example:

```terraform
import {
  to = aws_devopsagent_asset.example
  id = "12345678-1234-1234-1234-123456789012,87654321-4321-4321-4321-210987654321"
}
```

Using `terraform import`, import an asset using its Agent Space ID and asset ID separated by a comma. For example:

```console
% terraform import aws_devopsagent_asset.example 12345678-1234-1234-1234-123456789012,87654321-4321-4321-4321-210987654321
```
