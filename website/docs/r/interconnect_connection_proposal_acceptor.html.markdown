---
subcategory: "Interconnect"
layout: "aws"
page_title: "AWS: aws_interconnect_connection_proposal_acceptor"
description: |-
  Terraform resource for accepting an AWS Interconnect connection proposal.
---

# Resource: aws_interconnect_connection_proposal_acceptor

Terraform resource for accepting an AWS Interconnect connection proposal.

Use this resource for the partner-initiated creation flow, where you first create a connection proposal on the Interconnect partner's portal, console, or APIs, which produces an activation key. This resource then accepts that proposal in AWS and manages the lifecycle of the resulting connection.

For the AWS-initiated creation flow, use the [`aws_interconnect_connection`](/docs/providers/aws/r/interconnect_connection.html) resource instead.

## Example Usage

### Basic Usage

```terraform
resource "aws_interconnect_connection_proposal_acceptor" "example" {
  activation_key = "activation-key-from-partner-portal"
  description    = "Accepted multicloud interconnect"

  attach_point {
    direct_connect_gateway = "abcdef12-3456-7890-abcd-ef1234567890"
  }

  tags = {
    Environment = "prod"
  }
}
```

## Argument Reference

The following arguments are required:

* `activation_key` - (Required) Activation key generated on the Interconnect partner's portal, console, or APIs that captures the desired parameters from the initial creation request. Changing this forces a new resource to be created.
* `attach_point` - (Required) Attach point to which the connection logically connects within your AWS network. Changing this forces a new resource to be created. [See below](#attach_point).

The following arguments are optional:

* `description` - (Optional) Description to distinguish this connection.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### attach_point

Exactly one of the following arguments must be set.

* `arn` - (Optional) ARN of the attach point.
* `direct_connect_gateway` - (Optional) Identifier of a Direct Connect Gateway attach point.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the connection.
* `bandwidth` - Bandwidth of the connection.
* `billing_tier` - Billing tier this connection is currently assigned.
* `environment_id` - Environment on which the connection is placed.
* `id` - Identifier of the connection.
* `interconnect_provider` - Name of the provider on the remote side of this connection.
* `location` - Provider-specific location on the remote side of this connection.
* `owner_account` - Account that owns this connection.
* `shared_id` - Identifier used by both AWS and the remote partner to identify the connection.
* `state` - State of the connection. One of `requested`, `pending`, `available`, `down`, `deleting`, `deleted`, `failed`, or `updating`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `type` - Specific product type of this connection.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the resource using the `arn`. For example:

```terraform
import {
  to = aws_interconnect_connection_proposal_acceptor.example
  id = "arn:aws:interconnect:us-east-1:123456789012:connection/mcc-abcd1234"
}
```

Using `terraform import`, import the resource using the `arn`. For example:

```console
% terraform import aws_interconnect_connection_proposal_acceptor.example arn:aws:interconnect:us-east-1:123456789012:connection/mcc-abcd1234
```
