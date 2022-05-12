---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_security_profile"
description: |-
  Provides details about a specific Amazon Connect Security Profile.
---

# Data Source: aws_connect_security_profile

Provides details about a specific Amazon Connect Security Profile.

## Example Usage

By `name`

```hcl
data "aws_connect_security_profile" "example" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name        = "Example"
}
```

By `security_profile_id`

```hcl
data "aws_connect_security_profile" "example" {
  instance_id         = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  security_profile_id = "cccccccc-bbbb-cccc-dddd-111111111111"
}
```

## Argument Reference

~> **NOTE:** `instance_id` and one of either `name` or `security_profile_id` is required.

The following arguments are supported:

* `security_profile_id` - (Optional) Returns information on a specific Security Profile by Security Profile id
* `instance_id` - (Required) Reference to the hosting Amazon Connect Instance
* `name` - (Optional) Returns information on a specific Security Profile by name

## Attributes Reference

In addition to all of the arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the Security Profile.
* `description` - Specifies the description of the Security Profile.
* `id` - The identifier of the hosting Amazon Connect Instance and identifier of the Security Profile separated by a colon (`:`).
* `organization_resource_id` - The organization resource identifier for the security profile.
* `permissions` - Specifies a list of permissions assigned to the security profile.
* `tags` - A map of tags to assign to the Security Profile.
