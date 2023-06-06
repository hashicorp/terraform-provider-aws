---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_authentication_profile"
description: |-
  Creates a Redshift authentication profile
---

# Resource: aws_redshift_authentication_profile

Creates a Redshift authentication profile

## Example Usage

```terraform
resource "aws_redshift_authentication_profile" "example" {
  authentication_profile_name = "example"
  authentication_profile_content = jsonencode(
    {
      AllowDBUserOverride = "1"
      Client_ID           = "ExampleClientID"
      App_ID              = "example"
    }
  )
}
```

## Argument Reference

The following arguments are supported:

* `authentication_profile_name` - (Required, Forces new resource) The name of the authentication profile.
* `authentication_profile_content` - (Required) The content of the authentication profile in JSON format. The maximum length of the JSON string is determined by a quota for your account.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the authentication profile.

## Import

Redshift Authentication Profiles support import by `authentication_profile_name`, e.g.,

```console
$ terraform import aws_redshift_authentication_profile.test example
```
