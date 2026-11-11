---
subcategory: "DevOps Agent"
layout: "aws"
page_title: "AWS: aws_devopsagent_asset_file"
description: |-
  Lists AWS DevOps Agent Asset Files.
---

# List Resource: aws_devopsagent_asset_file

Lists AWS DevOps Agent Asset Files in an asset.

Each result includes the file's identity (`agent_space_id`, `asset_id`, `path`, `account_id`, and `region`) and uses the file path as its display name.

When `include_resource = true`, results include the attributes described in the [`aws_devopsagent_asset_file` resource](../r/devopsagent_asset_file.html.markdown#attribute-reference), including `content`. The `metadata` input is not returned by AWS and remains unset. The metadata returned by AWS is available in `metadata_actual`.

## Example Usage

### Basic Usage

```terraform
list "aws_devopsagent_asset_file" "example" {
  provider = aws

  config {
    agent_space_id = "12345678-1234-1234-1234-123456789012"
    asset_id       = "87654321-4321-4321-4321-210987654321"
  }
}
```

## Argument Reference

The following arguments are required:

* `agent_space_id` - (Required) Identifier of the Agent Space containing the asset.
* `asset_id` - (Required) Identifier of the asset to list files from.

The following arguments are optional:

* `region` - (Optional) Region to query. Defaults to provider region.
