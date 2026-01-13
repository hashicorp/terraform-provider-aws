---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_phone_number_contact_flow_association"
description: |-
  Associates a flow with a phone number claimed to an Amazon Connect instance.
---

# Resource: aws_connect_phone_number_contact_flow_association

Associates a flow with a phone number claimed to an Amazon Connect instance.

## Example Usage

```terraform
resource "aws_connect_phone_number_contact_flow_association" "example" {
  phone_number_id = aws_connect_phone_number.example.id
  instance_id     = aws_connect_instance.example.id
  contact_flow_id = aws_connect_contact_flow.example.contact_flow_id
}
```

## Argument Reference

This resource supports the following arguments:

* `contact_flow_id` - (Required) Contact flow ID.
* `instance_id` - (Required) Amazon Connect instance ID.
* `phone_number_id` - (Required) Phone number ID.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_connect_phone_number_contact_flow_association` using the `phone_number_id`, `instance_id` and `contact_flow_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_connect_phone_number_contact_flow_association.example
  id = "36727a4c-4683-4e49-880c-3347c61110a4,fa6c1691-e2eb-4487-bdb9-1aaed6268ebd,c4acdc79-395e-4280-a294-9062f56b07bb"
}
```

Using `terraform import`, import `aws_connect_phone_number_contact_flow_association` using the `phone_number_id`, `instance_id` and `contact_flow_id` separated by a comma (`,`). For example:

```console
% terraform import aws_connect_phone_number_contact_flow_association.example 36727a4c-4683-4e49-880c-3347c61110a4,fa6c1691-e2eb-4487-bdb9-1aaed6268ebd,c4acdc79-395e-4280-a294-9062f56b07bb
```
