---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_security_policy"
description: |-
  Terraform resource for managing an AWS OpenSearch Serverless Security Policy.
---

# Resource: aws_opensearchserverless_security_policy

Terraform resource for managing an AWS OpenSearch Serverless Security Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_opensearchserverless_security_policy" "example" {
  name   = "example"
  type   = "encryption"
  policy = <<-EOT
  {
	  "Rules": [
		  {
		  	"Resource": [
		  		"collection/example"
		  	],
		  	"ResourceType": "collection"
		  }
	  ],
	  "AWSOwnedKey": true
  }
  EOT
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the policy.
* `policy` - (Required) JSON policy document to use as the content for the new policy
* `type` - (Required) Type of security policy. One of `encryption` or `network`.

The following arguments are optional:

* `description` - (Optional) Description of the policy. Typically used to store information about the permissions defined in the policy.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `policy_version` - Version of the policy.

## Import

OpenSearchServerless Security Policy can be imported using the `name` and `type` arguments separated by a slash (`/`), e.g.,

```
$ terraform import aws_opensearchserverless_security_policy.example example/encryption
```
