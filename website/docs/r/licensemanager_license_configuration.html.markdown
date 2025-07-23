---
subcategory: "License Manager"
layout: "aws"
page_title: "AWS: aws_licensemanager_license_configuration"
description: |-
  Provides a License Manager license configuration resource.
---

# Resource: aws_licensemanager_license_configuration

Provides a License Manager license configuration resource.

~> **Note:** Removing the `license_count` attribute is not supported by the License Manager API - use `terraform taint aws_licensemanager_license_configuration.<id>` to recreate the resource instead.

## Example Usage

```terraform
resource "aws_licensemanager_license_configuration" "example" {
  name                     = "Example"
  description              = "Example"
  license_count            = 10
  license_count_hard_limit = true
  license_counting_type    = "Socket"

  license_rules = [
    "#minimumSockets=2",
  ]

  tags = {
    foo = "barr"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the license configuration.
* `description` - (Optional) Description of the license configuration.
* `license_count` - (Optional) Number of licenses managed by the license configuration.
* `license_count_hard_limit` - (Optional) Sets the number of available licenses as a hard limit.
* `license_counting_type` - (Required) Dimension to use to track license inventory. Specify either `vCPU`, `Instance`, `Core` or `Socket`.
* `license_rules` - (Optional) Array of configured License Manager rules.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Rules

License rules should be in the format of `#RuleType=RuleValue`. Supported rule types:

* `minimumVcpus` - Resource must have minimum vCPU count in order to use the license. Default: 1
* `maximumVcpus` - Resource must have maximum vCPU count in order to use the license. Default: unbounded, limit: 10000
* `minimumCores` - Resource must have minimum core count in order to use the license. Default: 1
* `maximumCores` - Resource must have maximum core count in order to use the license. Default: unbounded, limit: 10000
* `minimumSockets` - Resource must have minimum socket count in order to use the license. Default: 1
* `maximumSockets` - Resource must have maximum socket count in order to use the license. Default: unbounded, limit: 10000
* `allowedTenancy` - Defines where the license can be used. If set, restricts license usage to selected tenancies. Specify a comma delimited list of `EC2-Default`, `EC2-DedicatedHost`, `EC2-DedicatedInstance`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The license configuration ARN.
* `id` - The license configuration ARN.
* `owner_account_id` - Account ID of the owner of the license configuration.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import license configurations using the `id`. For example:

```terraform
import {
  to = aws_licensemanager_license_configuration.example
  id = "arn:aws:license-manager:eu-west-1:123456789012:license-configuration:lic-0123456789abcdef0123456789abcdef"
}
```

Using `terraform import`, import license configurations using the `id`. For example:

```console
% terraform import aws_licensemanager_license_configuration.example arn:aws:license-manager:eu-west-1:123456789012:license-configuration:lic-0123456789abcdef0123456789abcdef
```
