---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_security_policy"
description: |-
  Get information on an OpenSearch Serverless Security Policy.
---

# Data Source: aws_opensearchserverless_security_policy

Use this data source to get information about an AWS OpenSearch Serverless Security Policy.

## Example Usage

```terraform
data "aws_opensearchserverless_security_policy" "example" {
  name = "example-security-policy"
  type = "encryption"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the policy
* `type` - (Required) Type of security policy. One of `encryption` or `network`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `created_date` - The date the security policy was created.
* `description` - Description of the security policy.
* `last_modified_date` - The date the security policy was last modified.
* `policy` - The JSON policy document without any whitespaces.
* `policy_version` - Version of the policy.
