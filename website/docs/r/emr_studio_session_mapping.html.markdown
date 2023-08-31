---
subcategory: "EMR"
layout: "aws"
page_title: "AWS: aws_emr_studio_session_mapping"
description: |-
  Provides an Elastic MapReduce Studio
---

# Resource: aws_emr_studio_session_mapping

Provides an Elastic MapReduce Studio Session Mapping.

## Example Usage

```terraform
resource "aws_emr_studio_session_mapping" "example" {
  studio_id          = aws_emr_studio.example.id
  identity_type      = "USER"
  identity_id        = "example"
  session_policy_arn = aws_iam_policy.example.arn
}
```

## Argument Reference

The following arguments are required:

* `identity_id`- (Optional) The globally unique identifier (GUID) of the user or group from the Amazon Web Services SSO Identity Store.
* `identity_name` - (Optional) The name of the user or group from the Amazon Web Services SSO Identity Store.
* `identity_type` - (Required) Specifies whether the identity to map to the Amazon EMR Studio is a `USER` or a `GROUP`.
* `session_policy_arn` - (Required) The Amazon Resource Name (ARN) for the session policy that will be applied to the user or group. You should specify the ARN for the session policy that you want to apply, not the ARN of your user role.
* `studio_id` - (Required) The ID of the Amazon EMR Studio to which the user or group will be mapped.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id`- The id of the Elastic MapReduce Studio Session Mapping.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EMR studio session mappings using `studio-id:identity-type:identity-id`. For example:

```terraform
import {
  to = aws_emr_studio_session_mapping.example
  id = "es-xxxxx:USER:xxxxx-xxx-xxx"
}
```

Using `terraform import`, import EMR studio session mappings using `studio-id:identity-type:identity-id`. For example:

```console
% terraform import aws_emr_studio_session_mapping.example es-xxxxx:USER:xxxxx-xxx-xxx
```
