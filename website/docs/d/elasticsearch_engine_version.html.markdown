---
subcategory: "Elasticsearch"
layout: "aws"
page_title: "AWS: aws_elasticsearch_engine_version"
description: |-
  Get information on an Elasticsearch engine version.
---

# Data Source: aws_elasticsearch_engine_version

Use this data source to get information about Elasticsearch engine version

## Example Usage

```terraform
data "aws_elasticsearch_engine_version" "engine_version" {
  version = "1.5"
}
```

## Argument Reference

The following arguments are supported:

* `version` – (Optional) Version of the Elasticsearch engine. Conflict with `preferred_versions`.
* `preferred_versions` - (Optional) Ordered list of preffered version. The first match in this list will be returned.
 Conflict with `version`

## Attributes Reference

The following attributes are exported:

* `version` – Elasticsearch engine version.
