---
subcategory: "X-Ray"
layout: "aws"
page_title: "AWS: aws_xray_resource_policy"
description: |-
  Terraform resource for managing an AWS X-Ray Resource Policy.
---

# Resource: aws_xray_resource_policy

Terraform resource for managing an AWS X-Ray Resource Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_xray_resource_policy" "test" {
  policy_name                 = "test"
  policy_document             = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Sid\":\"AllowXRayAccess\",\"Effect\":\"Allow\",\"Principal\":{\"AWS\":\"*\"},\"Action\":[\"xray:*\",\"xray:PutResourcePolicy\"],\"Resource\":\"*\"}]}"
  bypass_policy_lockout_check = true
}
```

## Argument Reference

The following arguments are required:

* `policy_name` - (Required) name of the resource policy. Must be unique within a specific Amazon Web Services account.
* `policy_document` - (Required) JSON string of the resource policy or resource policy document, which can be up to 5kb in size.

The following arguments are optional:

* `policy_revision_id` - (Optional) Specifies a specific policy revision, to ensure an atomic create operation. By default the resource policy is created if it does not exist, or updated with an incremented revision id. The revision id is unique to each policy in the account. If the policy revision id does not match the latest revision id, the operation will fail with an InvalidPolicyRevisionIdException exception. You can also provide a PolicyRevisionId of 0. In this case, the operation will fail with an InvalidPolicyRevisionIdException exception if a resource policy with the same name already exists.
* `bypass_policy_lockout_check` - (Optional) Flag to indicate whether to bypass the resource policy lockout safety check. Setting this value to true increases the risk that the policy becomes unmanageable. Do not set this value to true indiscriminately. Use this parameter only when you include a policy in the request and you intend to prevent the principal that is making the request from making a subsequent PutResourcePolicy request. The default value is `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `last_updated_time` - When the policy was last updated, in Unix time seconds.
* `policy_revision_id` - Returns the current policy revision id for this policy name.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import X-Ray Resource Policy using the `policy_name`. For example:

```terraform
import {
  to = aws_xray_resource_policy.example
  id = "resource_policy-name"
}
```

Using `terraform import`, import X-Ray Resource Policy using the `policy_name`. For example:

```console
% terraform import aws_xray_resource_policy.example resource_policy-name
```
