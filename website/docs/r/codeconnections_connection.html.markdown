---
subcategory: "CodeConnections"
layout: "aws"
page_title: "AWS: aws_codeconnections_connection"
description: |-
  Terraform resource for managing an AWS CodeConnections Connection.
---

# Resource: aws_codeconnections_connection

Terraform resource for managing an AWS CodeConnections Connection.

~> **NOTE:** The `aws_codeconnections_connection` resource is created in the state `PENDING`. Authentication with the connection provider must be completed in the AWS Console. See the [AWS documentation](https://docs.aws.amazon.com/dtconsole/latest/userguide/connections-update.html) for details.

## Example Usage

### Basic Usage

```terraform
resource "aws_codeconnections_connection" "example" {
  name          = "example-connection"
  provider_type = "Bitbucket"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The name of the connection to be created. The name must be unique in the calling AWS account. Changing `name` will create a new resource.
* `provider_type` - (Optional) The name of the external provider where your third-party code repository is configured. Changing `provider_type` will create a new resource. Conflicts with `host_arn`.
* `host_arn` - (Optional) The Amazon Resource Name (ARN) of the host associated with the connection. Conflicts with `provider_type`
* `tags` - (Optional) Map of key-value resource tags to associate with the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The codeconnections connection ARN.
* `connection_status` - The codeconnections connection status. Possible values are `PENDING`, `AVAILABLE` and `ERROR`.
* `id` - (**Deprecated**) The codeconnections connection ARN.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodeConnections connection using the ARN. For example:

```terraform
import {
  to = aws_codeconnections_connection.test-connection
  id = "arn:aws:codeconnections:us-west-1:0123456789:connection/79d4d357-a2ee-41e4-b350-2fe39ae59448"
}
```

Using `terraform import`, import CodeConnections connection using the ARN. For example:

```console
% terraform import aws_codeconnections_connection.test-connection arn:aws:codeconnections:us-west-1:0123456789:connection/79d4d357-a2ee-41e4-b350-2fe39ae59448
```
