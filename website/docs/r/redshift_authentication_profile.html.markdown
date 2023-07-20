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

This resource supports the following arguments:

* `authentication_profile_name` - (Required, Forces new resource) The name of the authentication profile.
* `authentication_profile_content` - (Required) The content of the authentication profile in JSON format. The maximum length of the JSON string is determined by a quota for your account.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the authentication profile.

## Import

Import Redshift Authentication by `authentication_profile_name`. For example:

```console
$ terraform import aws_redshift_authentication_profile.test example
```
