---
subcategory: "CloudWatch Observability Access Manager"
layout: "aws"
page_title: "AWS: aws_oam_link"
description: |-
  Terraform resource for managing an AWS CloudWatch Observability Access Manager Link.
---

# Resource: aws_oam_link

Terraform resource for managing an AWS CloudWatch Observability Access Manager Link.

## Example Usage

### Basic Usage

```terraform
resource "aws_oam_link" "example" {
  label_template  = "$AccountName"
  resource_types  = ["AWS::CloudWatch::Metric"]
  sink_identifier = aws_oam_sink.test.id
  tags = {
    Env = "prod"
  }
}
```

## Argument Reference

The following arguments are required:

* `label_template` - (Required) Human-readable name to use to identify this source account when you are viewing data from it in the monitoring account.
* `resource_types` - (Required) Types of data that the source account shares with the monitoring account.
* `sink_identifier` - (Required) Identifier of the sink to use to create this link.

The following arguments are optional:

* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the link.
* `label` - Label that is assigned to this link.
* `link_id` - ID string that AWS generated as part of the link ARN.
* `sink_arn` - ARN of the sink that is used for this link.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `1m`)
* `update` - (Default `1m`)
* `delete` - (Default `1m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Observability Access Manager Link using the `arn`. For example:

```terraform
import {
  to = aws_oam_link.example
  id = "arn:aws:oam:us-west-2:123456789012:link/link-id"
}
```

Using `terraform import`, import CloudWatch Observability Access Manager Link using the `arn`. For example:

```console
% terraform import aws_oam_link.example arn:aws:oam:us-west-2:123456789012:link/link-id
```
