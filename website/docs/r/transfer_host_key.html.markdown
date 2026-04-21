---
subcategory: "Transfer Family"
layout: "aws"
page_title: "AWS: aws_transfer_host_key"
description: |-
  Manages a host key for a server.
---

# Resource: aws_transfer_host_key

Manages a host key for a server. This is an [_additional server host key_](https://docs.aws.amazon.com/transfer/latest/userguide/server-host-key-add.html).

## Example Usage

```terraform
resource "aws_transfer_host_key" "example" {
  server_id   = aws_transfer_server.example.id
  description = "example additional host key"

  host_key_body_wo = <<EOT
# Private key PEM.
EOT
}
```

## Argument Reference

This resource supports the following arguments:

* `description` - (Optional) Text description.
* `host_key_body` - (Optional) Private key portion of an SSH key pair.
* `host_key_body_wo` - (Optional) [Write-only](https://developer.hashicorp.com/terraform/language/manage-sensitive-data/ephemeral#write-only-arguments) private key portion of an SSH key pair, guaranteed not to be written to plan or state artifacts. One of `host_key_body` or `host_key_body_wo` must be configured.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `server_id` - (Required) Server ID.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of host key.
* `host_key_id`  - ID of the host key.
* `host_key_fingerprint` - Public key fingerprint.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import host keys using the `server_id` and `host_key_id` separated by `,`. For example:

```terraform
import {
  to = aws_transfer_host_key.example
  id = "s-12345678,key-12345"
}
```

Using `terraform import`, import host keys using the `server_id` and `host_key_id` separated by `,`. For example:

```console
% terraform import aws_transfer_host_key.example s-12345678,key-12345
```
