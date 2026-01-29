---
subcategory: "Elemental MediaPackage VOD"
layout: "aws"
page_title: "AWS: aws_mediapackagevod_packaging_group"
description: |-
  Creates an AWS Elemental MediaPackage VOD Packaging Group.
---

# Resource: aws_mediapackagevod_packaging_group

Creates an AWS Elemental MediaPackage VOD Packaging Group.

## Example Usage

```terraform
resource "aws_mediapackagevod_packaging_group" "example" {
  name = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) A unique identifier naming the Packaging Group.
* `authorization` - (Optional) Defines the authorization configuration for the Packaging Group.
* `egress_access_logs` - (Optional) Defines the egress logging configuration for the Packaging Group
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `authorization` Configuration Block

* `cdn_identifier_secret` - (Optional) The ARN of the Secrets Manager Secret containing the `MediaPackageCDNIdentifier` to authorize requests against. `secrets_role_arn` must also be specified if this property is.
* `secrets_role_arn` - (Optional) The ARN of the IAM Role that MediaPackage VOD will assume to get the `cdn_identifier_secret` to authorize requests. `cdn_identifier_secret` must also be specified if this property is.

### `egress_access_logs` Configuration Block

* `log_group_name` - (Optional) The name of the AWS CloudWatch Log Group to log egress traffic to. If specified, `log_group_name` must start with `/aws/MediaPackage/`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the Packaging Group.
* `domain` - The egress domain of the Packaging Group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an Elemental MediaPackage VOD Packaging Group using the packaging group's name. For example:

```terraform
import {
  to = aws_mediapackagevod_packaging_group.example
  id = "example"
}
```

Using `terraform import`, import Elemental MediaPackage VOD Packaging Group using the packaging group's `name`. For example:

```console
% terraform import aws_mediapackagevod_packaging_group.example example
```
