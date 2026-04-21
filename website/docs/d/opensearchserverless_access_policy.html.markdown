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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the policy.
* `type` - (Required) Type of access policy. Must be `data`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` - Description of the policy. Typically used to store information about the permissions defined in the policy.
* `policy` - JSON policy document to use as the content for the new policy.
* `policy_version` - Version of the policy.
