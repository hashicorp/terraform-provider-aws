---
subcategory: "Athena"
layout: "aws"
page_title: "AWS: aws_athena_named_query"
description: |-
    Provides an Athena Named Query data source.
---

# Data Source: aws_athena_named_query

Provides an Athena Named Query data source.

## Example Usage

```terraform
data "aws_athena_named_query" "example" {
  name = "athenaQueryName"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The plain language name for the query. Maximum length of 128.
* `workgroup` - (Optional) The workgroup to which the query belongs. Defaults to `primary`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `database` - Database to which the query belongs.
* `description` - Brief explanation of the query.
* `id` - The unique ID of the query.
* `query` - Text of the query itself.
