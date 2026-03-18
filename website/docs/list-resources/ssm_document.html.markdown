---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_document"
description: |-
  Lists SSM (Systems Manager) Document resources.
---

# List Resource: aws_ssm_document

Lists SSM (Systems Manager) Document resources.

## Example Usage

```terraform
list "aws_ssm_document" "example" {
  provider = aws
}
```

### Filter by name prefix

```terraform
list "aws_ssm_document" "example" {
  provider = aws

  config {
    filter {
      key    = "Name"
      values = ["example-prefix-"]
    }
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
* `filter` - (Optional) One or more filters to apply to the search. If omitted, the list returns self-owned SSM documents by default.

### `filter` Block

The `filter` block supports the following arguments:

* `key` - (Required) Filter key to apply. Valid values are the keys accepted by the SSM `ListDocuments` API, such as `Name`, `Owner`, `PlatformTypes`, `DocumentType`, `TargetType`, or `tag:<tag-key>`.
* `values` - (Required) One or more values for the selected filter key. For the `Name` key, a prefix match is supported by the SSM API.
