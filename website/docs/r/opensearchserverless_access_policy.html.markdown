---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_access_policy"
description: |-
  Terraform resource for managing an AWS OpenSearch Serverless Access Policy.
---

# Resource: aws_opensearchserverless_access_policy

Terraform resource for managing an AWS OpenSearch Serverless Access Policy.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_opensearchserverless_access_policy" "test" {
  name   = "example"
  type   = "data"
  policy = <<-EOT
  [
	{
	  "Rules": [
		{
		  "ResourceType": "index",
		  "Resource": [
			"index/books/*"
		  ],
		  "Permission": [
			"aoss:CreateIndex",
			"aoss:ReadDocument",
			"aoss:UpdateIndex",
			"aoss:DeleteIndex",
			"aoss:WriteDocument"
		  ]
		}
	  ],
	  "Principal": [
		"arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:user/admin"
	  ]
	}
  ]
  EOT
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the policy.
* `policy` - (Required) JSON policy document to use as the content for the new policy
* `type` - (Required) Type of access policy. Must be `data`.

The following arguments are optional:

* `description` - (Optional) Description of the policy. Typically used to store information about the permissions defined in the policy.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `policy_version` - Version of the policy.

## Import

OpenSearchServerless Access Policy can be imported using the `name` and `type` arguments separated by a slash (`/`), e.g.,

```
$ terraform import aws_opensearchserverless_access_policy.example example/data
```
