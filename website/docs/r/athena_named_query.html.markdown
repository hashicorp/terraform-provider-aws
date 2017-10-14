---
layout: "aws"
page_title: "AWS: aws_athena_named_query"
sidebar_current: "docs-aws-resource-athena-named-query"
description: |-
  Provides an Athena Named Query resource.
---

# aws_athena_named_query

Provides an Athena Named Query resource.

~> **NOTE on Athena Named Query**: AWS CLI for Athena haven't provided a feature to create database yet while named query api requires database name. So you have to create database in advance.
[the AWS Docs](https://docs.aws.amazon.com/ja_jp/athena/latest/APIReference/API_CreateNamedQuery.html).

## Example Usage

```hcl
resource "aws_athena_named_query" "foo" {
  name = "bar"
	database = "users"
	query = "SELECT * FROM users limit 10;"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The plain language name for the query. Maximum length of 128.
* `database` - (Required) The database to which the query belongs.
* `query` - (Required) The text of the query itself. In other words, all query statements. Maximum length of 262144.
* `description` - (Optional) A brief explanation of the query. Maximum length of 1024.

## Attributes Reference

The following attributes are exported:

* `id` - The unique ID of the query.
