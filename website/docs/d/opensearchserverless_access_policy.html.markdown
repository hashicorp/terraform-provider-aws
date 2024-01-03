---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_access_policy"
description: |-
  Terraform data source for managing an AWS OpenSearch Serverless Access Policy.
---

# Data Source: aws_opensearchserverless_access_policy

Terraform data source for managing an AWS OpenSearch Serverless Access Policy.

## Example Usage

### Basic Usage

```terraform
data "aws_opensearchserverless_access_policy" "example" {
  name = aws_opensearchserverless_access_policy.example.name
  type = aws_opensearchserverless_access_policy.example.type
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the policy.
* `type` - (Required) Type of access policy. Must be `data`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` - Description of the policy. Typically used to store information about the permissions defined in the policy.
* `policy` - JSON policy document to use as the content for the new policy.
* `policy_version` - Version of the policy.
