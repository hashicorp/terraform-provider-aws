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

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `instance_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the SSO Instance.
* `attribute` - (Required) See [AccessControlAttribute](#accesscontrolattribute) for more details.

### AccessControlAttribute

* `key` - (Required) The name of the attribute associated with your identities in your identity source. This is used to map a specified attribute in your identity source with an attribute in AWS SSO.
* `value` - (Required) The value used for mapping a specified attribute to an identity source. See [AccessControlAttributeValue](#accesscontrolattributevalue)

### AccessControlAttributeValue

* `source` - (Required) The identity source to use when mapping a specified attribute to AWS SSO.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The identifier of the Instance Access Control Attribute `instance_arn`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSO Account Assignments using the `instance_arn`. For example:

```terraform
import {
  to = aws_ssoadmin_instance_access_control_attributes.example
  id = "arn:aws:sso:::instance/ssoins-0123456789abcdef"
}
```

Using `terraform import`, import SSO Account Assignments using the `instance_arn`. For example:

```console
% terraform import aws_ssoadmin_instance_access_control_attributes.example arn:aws:sso:::instance/ssoins-0123456789abcdef
```
