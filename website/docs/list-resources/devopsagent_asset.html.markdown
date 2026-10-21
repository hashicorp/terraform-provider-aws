---
subcategory: "DevOps Agent"
layout: "aws"
page_title: "AWS: aws_devopsagent_asset"
description: |-
  Lists AWS DevOps Agent Assets.
---

# List Resource: aws_devopsagent_asset

Lists AWS DevOps Agent Assets in an Agent Space.

Each result includes the asset's identity (`agent_space_id`, `asset_id`, `account_id`, and `region`) and uses the asset ID as its display name.

When `include_resource = true`, results include the attributes described in the [`aws_devopsagent_asset` resource](../r/devopsagent_asset.html.markdown#attribute-reference). The `content` and `metadata` inputs are not returned by AWS and remain unset. The merged metadata returned by AWS is available in `metadata_all`.

## Example Usage

### Basic Usage

```terraform
list "aws_devopsagent_asset" "example" {
  provider = aws

  config {
    agent_space_id = "12345678-1234-1234-1234-123456789012"
  }
}
```

### Filter by Type and Update Time

```terraform
list "aws_devopsagent_asset" "example" {
  provider         = aws
  include_resource = true

  config {
    agent_space_id = "12345678-1234-1234-1234-123456789012"
    asset_type     = "skill"
    updated_after  = "2026-01-01T00:00:00Z"
    updated_before = "2026-02-01T00:00:00Z"
  }
}
```

## Argument Reference

The following arguments are required:

* `agent_space_id` - (Required) Identifier of the Agent Space to list assets from.

The following arguments are optional:

* `asset_type` - (Optional) Limits results to assets of this type.
* `region` - (Optional) Region to query. Defaults to provider region.
* `updated_after` - (Optional) Limits results to assets updated after this timestamp.
* `updated_before` - (Optional) Limits results to assets updated before this timestamp.
