---
subcategory: "Opensearch"
layout: "aws"
page_title: "AWS: aws_opensearch_engine_version"
description: |-
  Get information on an Opensearch engine version.
---

# Data Source: aws_opensearch_engine_version

Use this data source to get information about Opensearch engine version

## Example Usage

```terraform
data "aws_opensearch_engine_version" "engine_version" {
  version = "OpenSearch_2.3"
}
```

## Argument Reference

The following arguments are supported:

* `version` – (Optional) Version of the Opensearch engine. Conflict with `preferred_versions`.
* `preferred_versions` - (Optional) Ordered list of preffered version. The first match in this list will be returned.
 Conflict with `version`

## Attributes Reference

The following attributes are exported:

* `version` – Opensearch engine version.
