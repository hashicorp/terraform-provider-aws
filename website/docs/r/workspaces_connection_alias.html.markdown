---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_connection_alias"
description: |-
  Terraform resource for managing an AWS WorkSpaces Connection Alias.
---

# Resource: aws_workspaces_connection_alias

Terraform resource for managing an AWS WorkSpaces Connection Alias.

## Example Usage

### Basic Usage

```terraform
resource "aws_workspaces_connection_alias" "example" {
  connection_string = "testdomain.test"
}
```

## Argument Reference

The following arguments are required:

* `connection_string` - (Required) The connection string specified for the connection alias. The connection string must be in the form of a fully qualified domain name (FQDN), such as www.example.com.
* `tags` – (Optional) A map of tags assigned to the WorkSpaces Connection Alias. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The identifier of the connection alias.
* `owner_account_id` - The identifier of the Amazon Web Services account that owns the connection alias.
* `state` - The current state of the connection alias.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces Connection Alias using the connection alias ID. For example:

```terraform
import {
  to = aws_workspaces_connection_alias.example
  id = "rft-8012925589"
}
```

Using `terraform import`, import WorkSpaces Connection Alias using the connection alias ID. For example:

```console
% terraform import aws_workspaces_connection_alias.example rft-8012925589
```
