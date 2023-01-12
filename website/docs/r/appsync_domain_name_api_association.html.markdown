---
subcategory: "AppSync"
layout: "aws"
page_title: "AWS: aws_appsync_domain_name_api_association"
description: |-
  Provides an AppSync API Association.
---

# Resource: aws_appsync_domain_name_api_association

Provides an AppSync API Association.

## Example Usage

```terraform
resource "aws_appsync_domain_name_api_association" "example" {
  api_id      = aws_appsync_graphql_api.example.id
  domain_name = aws_appsync_domain_name.example.domain_name
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) API ID.
* `domain_name` - (Required) Appsync domain name.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Appsync domain name.

## Import

`aws_appsync_domain_name_api_association` can be imported using the AppSync domain name, e.g.,

```
$ terraform import aws_appsync_domain_name_api_association.example example.com
```
