---
subcategory: "Device Farm"
layout: "aws"
page_title: "AWS: aws_devicefarm_instance_profile"
description: |-
  Provides a Devicefarm instance profile
---

# Resource: aws_devicefarm_instance_profile

Provides a resource to manage AWS Device Farm Instance Profiles.
∂
~> **NOTE:** AWS currently has limited regional support for Device Farm (e.g., `us-west-2`). See [AWS Device Farm endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/devicefarm.html) for information on supported regions.

## Example Usage

```terraform
resource "aws_devicefarm_instance_profile" "example" {
  name = "example"
}
```

## Argument Reference

* `description` - (Optional) The description of the instance profile.
* `exclude_app_packages_from_cleanup` - (Optional) An array of strings that specifies the list of app packages that should not be cleaned up from the device after a test run.
* `name` - (Required) The name for the instance profile.
* `package_cleanup` - (Optional) When set to `true`, Device Farm removes app packages after a test run. The default value is `false` for private devices.
* `reboot_after_use` - (Optional) When set to `true`, Device Farm reboots the instance after a test run. The default value is `true`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name of this instance profile.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

DeviceFarm Instance Profiles can be imported by their arn:

```
$ terraform import aws_devicefarm_instance_profile.example arn:aws:devicefarm:us-west-2:123456789012:instanceprofile:4fa784c7-ccb4-4dbf-ba4f-02198320daa1
```
