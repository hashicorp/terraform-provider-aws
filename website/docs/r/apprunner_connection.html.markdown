---
subcategory: "App Runner"
layout: "aws"
page_title: "AWS: aws_apprunner_connection"
description: |-
  Manages an App Runner Connection.
---

# Resource: aws_apprunner_connection

Manages an App Runner Connection.

~> **NOTE:** After creation, you must complete the authentication handshake using the App Runner console.

## Example Usage

```terraform
resource "aws_apprunner_connection" "example" {
  connection_name = "example"
  provider_type   = "GITHUB"

  tags = {
    Name = "example-apprunner-connection"
  }
}
```

## Argument Reference

The following arguments supported:

* `connection_name` - (Required) Name of the connection.
* `provider_type` - (Required) Source repository provider. Valid values: `GITHUB`.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the connection.
* `status` - Current state of the App Runner connection. When the state is `AVAILABLE`, you can use the connection to create an [`aws_apprunner_service` resource](apprunner_service.html).
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import App Runner Connections using the `connection_name`. For example:

```terraform
import {
  to = aws_apprunner_connection.example
  id = "example"
}
```

Using `terraform import`, import App Runner Connections using the `connection_name`. For example:

```console
% terraform import aws_apprunner_connection.example example
```
