---
layout: "aws"
page_title: "AWS: aws_licensemanager_license_configuration"
sidebar_current: "docs-aws-resource-licensemanager-license-configuration"
description: |-
  Provides a License Manager license configuration resource.
---

# aws_licensemanager_license_configuration

Provides a License Manager license configuration resource.

## Example Usage

```hcl
resource "aws_licensemanager_license_configuration" "example" {
  name                     = "Example"
  description              = "Example"
  license_count            = 10
  license_count_hard_limit = true
  license_counting_type    = "Socket"

  license_rules = [
    "#minimumSockets=2"
  ]

  tags {
    foo = "barr"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the license configuration.
* `description` - (Optional) Description of the license configuration.
* `license_count` - (Optional) Number of licenses managed by the license configuration.
* `license_count_hard_limit` - (Optional) Sets the number of available licenses as a hard limit.
* `license_counting_type` - (Required) Dimension to use to track license inventory. Specify either `vCPU`, `Instance`, `Core` or `Socket`.
* `license_rules` - (Optional) Array of configured License Manager rules.
* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The license configuration ARN.

## Import

License configurations can be imported using the `id`, e.g.

```
$ terraform import aws_licensemanager_license_configuration.example arn:aws:license-manager:eu-west-1:123456789012:license-configuration:lic-0123456789abcdef0123456789abcdef
```
