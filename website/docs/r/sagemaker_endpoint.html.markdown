---
layout: "aws"
page_title: "AWS: sagemaker_endpoint"
sidebar_current: "docs-aws-resource-sagemaker-endpoint"
description: |-
  Provides a Sagemaker endpoint resource.
---

# aws\_sagemaker\_endpoint

Provides a Sagemaker endpoint resource.

## Example Usage

Basic usage:

```hcl
resource "aws_sagemaker_endpoint" "e" {
    name = "my-endpoint"
    configuration_name = "my-endpoint-config"

    tags {
      Name = "main"
    }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of the endpoint. If omitted, Terraform will assign a random, unique name.
* `configuration_name` - (Required) The name of the endpoint configuration to use.
* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

The following attributes are exported:

* `name` - The name of the endpoint.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this endpoint.
* `creation_timestamp` - The creation timestamp of this endpoint.
* `last_modified_time` - Last time the endpoint has been modified.

## Import

Endpoints can be imported using the `name`, e.g.

```
$ terraform import aws_sagemaker_endpoint.test_endpoint my-endpoint
```
