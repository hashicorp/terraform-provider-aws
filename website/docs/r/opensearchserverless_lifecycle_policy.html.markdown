---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_lifecycle_policy"
description: |-
  Terraform resource for managing an AWS OpenSearch Serverless Lifecycle Policy.
---

# Resource: aws_opensearchserverless_lifecycle_policy

Terraform resource for managing an AWS OpenSearch Serverless Lifecycle Policy. See AWS documentation for [lifecycle policies](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/serverless-lifecycle.html).

## Example Usage

### Basic Usage

```terraform
resource "aws_opensearchserverless_lifecycle_policy" "example" {
  name = "example"
  type = "retention"
  policy = jsonencode({
    "Rules" : [
      {
        "ResourceType" : "index",
        "Resource" : ["index/autoparts-inventory/*"],
        "MinIndexRetention" : "81d"
      },
      {
        "ResourceType" : "index",
        "Resource" : ["index/sales/orders*"],
        "NoMinIndexRetention" : true
      }
    ]
  })
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the policy.
* `policy` - (Required) JSON policy document to use as the content for the new policy.
* `type` - (Required) Type of lifecycle policy. Must be `retention`.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description of the policy.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `policy_version` - Version of the policy.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearch Serverless Lifecycle Policy using the `name` and `type` arguments separated by a slash (`/`). For example:

```terraform
import {
  to = aws_opensearchserverless_lifecycle_policy.example
  id = "example/retention"
}
```

Using `terraform import`, import OpenSearch Serverless Lifecycle Policy using the `name` and `type` arguments separated by a slash (`/`). For example:

```console
% terraform import aws_opensearchserverless_lifecycle_policy.example example/retention
```
