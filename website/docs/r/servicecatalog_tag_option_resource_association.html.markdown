---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_tag_option_resource_association"
description: |-
  Manages a Service Catalog Tag Option Resource Association
---

# Resource: aws_servicecatalog_tag_option_resource_association

Manages a Service Catalog Tag Option Resource Association.

-> **Tip:** A "resource" is either a Service Catalog portfolio or product.

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalog_tag_option_resource_association" "example" {
  resource_id   = "prod-dnigbtea24ste"
  tag_option_id = "tag-pjtvyakdlyo3m"
}
```

## Argument Reference

The following arguments are required:

* `resource_id` - (Required) Resource identifier.
* `tag_option_id` - (Required) Tag Option identifier.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Identifier of the association.
* `resource_arn` - ARN of the resource.
* `resource_created_time` - Creation time of the resource.
* `resource_description` - Description of the resource.
* `resource_name` - Description of the resource.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `3m`)
- `read` - (Default `10m`)
- `delete` - (Default `3m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_servicecatalog_tag_option_resource_association` using the tag option ID and resource ID. For example:

```terraform
import {
  to = aws_servicecatalog_tag_option_resource_association.example
  id = "tag-pjtvyakdlyo3m:prod-dnigbtea24ste"
}
```

Using `terraform import`, import `aws_servicecatalog_tag_option_resource_association` using the tag option ID and resource ID. For example:

```console
% terraform import aws_servicecatalog_tag_option_resource_association.example tag-pjtvyakdlyo3m:prod-dnigbtea24ste
```
