---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_instance_access_control_attributes"
description: |-
  Provides a Single Sign-On (SSO) ABAC Resource: https://docs.aws.amazon.com/singlesignon/latest/userguide/abac.html
---

# Resource: aws_ssoadmin_instance_access_control_attributes

Provides a Single Sign-On (SSO) ABAC Resource: https://docs.aws.amazon.com/singlesignon/latest/userguide/abac.html

## Example Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_ssoadmin_instance_access_control_attributes" "example" {
  instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
  attribute {
    key = "name"
    value {
      source = ["$${path:name.givenName}"]
    }
  }
  attribute {
    key = "last"
    value {
      source = ["$${path:name.familyName}"]
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `instance_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the SSO Instance.
* `attribute` - (Required) See [AccessControlAttribute](#accesscontrolattribute) for more details.

### AccessControlAttribute

* `key` - (Required) The name of the attribute associated with your identities in your identity source. This is used to map a specified attribute in your identity source with an attribute in AWS SSO.
* `value` - (Required) The value used for mapping a specified attribute to an identity source. See [AccessControlAttributeValue](#accesscontrolattributevalue)

### AccessControlAttributeValue

* `source` - (Required) The identity source to use when mapping a specified attribute to AWS SSO.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The identifier of the Instance Access Control Attribute `instance_arn`.

## Import

SSO Account Assignments can be imported using the `instance_arn`

```
$ terraform import aws_ssoadmin_instance_access_control_attributes.example arn:aws:sso:::instance/ssoins-0123456789abcdef
```
