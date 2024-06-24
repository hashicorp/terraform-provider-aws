---
subcategory: "CodeStar Connections"
layout: "aws"
page_title: "AWS: aws_codestarconnections_connection"
description: |-
  Provides details about CodeStar Connection
---

# Data Source: aws_codestarconnections_connection

Provides details about CodeStar Connection.

## Example Usage

### By ARN

```terraform
data "aws_codestarconnections_connection" "example" {
  arn = aws_codestarconnections_connection.example.arn
}
```

### By Name

```terraform
data "aws_codestarconnections_connection" "example" {
  name = aws_codestarconnections_connection.example.name
}
```

## Argument Reference

This data source supports the following arguments:

* `arn` - (Optional) CodeStar Connection ARN.
* `name` - (Optional) CodeStar Connection name.

~> **NOTE:** When both `arn` and `name` are specified, `arn` takes precedence.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `connection_status` - CodeStar Connection status. Possible values are `PENDING`, `AVAILABLE` and `ERROR`.
* `id` - CodeStar Connection ARN.
* `host_arn` - ARN of the host associated with the connection.
* `name` - Name of the CodeStar Connection. The name is unique in the calling AWS account.
* `provider_type` - Name of the external provider where your third-party code repository is configured. Possible values are `Bitbucket`, `GitHub` and `GitLab`. For connections to GitHub Enterprise Server or GitLab Self-Managed instances, you must create an [aws_codestarconnections_host](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/codestarconnections_host) resource and use `host_arn` instead.
* `tags` - Map of key-value resource tags to associate with the resource.
