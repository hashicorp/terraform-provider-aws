---
subcategory: "Device Farm"
layout: "aws"
page_title: "AWS: aws_devicefarm_vpce_configuration"
description: |-
  Provides a Devicefarm vpce configuration
---

# Resource: aws_devicefarm_vpce_configuration

Provides a resource to manage AWS Device Farm VPCE Configurations.

~> **NOTE:** AWS currently has limited regional support for Device Farm (e.g., `us-west-2`). See [AWS Device Farm endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/devicefarm.html) for information on supported regions.

## Example Usage


```terraform
resource "aws_devicefarm_vpce_configuration" "example" {
  service_dns_name        = aws_vpc_endpoint_service.example.service_name
  vpce_configuration_name = "example"
  vpce_service_name       = "devicefarm.com"
}
```

## Argument Reference

* `vpce_configuration_description` - (Optional) The description of the vpce configuration.
* `service_dns_name` - (Required) The description of the vpce configuration.
* `vpce_configuration_name` - (Required) The description of the vpce configuration.
* `vpce_service_name` - (Required) The description of the vpce configuration.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name of this vpce configuration.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

DeviceFarm VPCE Configurations can be imported by their arn:

```
$ terraform import aws_devicefarm_vpce_configuration.example arn:aws:devicefarm:us-west-2:123456789012:vpceconfiguration:4fa784c7-ccb4-4dbf-ba4f-02198320daa1
```
