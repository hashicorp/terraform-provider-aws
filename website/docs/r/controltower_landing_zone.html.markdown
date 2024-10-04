---
subcategory: "Control Tower"
layout: "aws"
page_title: "AWS: aws_controltower_landing_zone"
description: |-
  Creates a new landing zone using Control Tower.
---

# Resource: aws_controltower_landing_zone

Creates a new landing zone using Control Tower. For more information on usage, please see the
[AWS Control Tower Landing Zone User Guide](https://docs.aws.amazon.com/controltower/latest/userguide/how-control-tower-works.html).

## Example Usage

```terraform
resource "aws_controltower_landing_zone" "example" {
  manifest_json = file("${path.module}/LandingZoneManifest.json")
  version       = "3.2"
}
```

## Argument Reference

This resource supports the following arguments:

* `manifest_json` - (Required) The manifest JSON file is a text file that describes your AWS resources. For examples, review [Launch your landing zone](https://docs.aws.amazon.com/controltower/latest/userguide/lz-api-launch).
* `version` - (Required) The landing zone version.
* `tags` - (Optional) Tags to apply to the landing zone. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The identifier of the landing zone.
* `arn` - The ARN of the landing zone.
* `drift_status` - The drift status summary of the landing zone.
    * `status` - The drift status of the landing zone.
* `latest_available_version` - The latest available version of the landing zone.
* `tags_all` - A map of tags assigned to the landing zone, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `120m`)
- `update` - (Default `120m`)
- `delete` - (Default `120m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Control Tower Landing Zone using the `id`. For example:

```terraform
import {
  to = aws_controltower_landing_zone.example
  id = "1A2B3C4D5E6F7G8H"
}
```

Using `terraform import`, import a Control Tower Landing Zone using the `id`. For example:

```console
% terraform import aws_controltower_landing_zone.example 1A2B3C4D5E6F7G8H
```
