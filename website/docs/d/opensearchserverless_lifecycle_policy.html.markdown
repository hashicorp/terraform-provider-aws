---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_lifecycle_policy"
description: |-
  Terraform data source for managing an AWS OpenSearch Serverless Lifecycle Policy.
---

# Data Source: aws_opensearchserverless_lifecycle_policy

Terraform data source for managing an AWS OpenSearch Serverless Lifecycle Policy.

## Example Usage

### Basic Usage

```terraform
data "aws_opensearchserverless_lifecycle_policy" "example" {
  name = "example-lifecycle-policy"
  type = "retention"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the policy
* `type` - (Required) Type of lifecycle policy. Must be `retention`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `created_date` - The date the lifecycle policy was created.
* `description` - Description of the policy. Typically used to store information about the permissions defined in the policy.
* `last_modified_date` - The date the lifecycle policy was last modified.
* `policy` - JSON policy document to use as the content for the new policy.
* `policy_version` - Version of the policy.
