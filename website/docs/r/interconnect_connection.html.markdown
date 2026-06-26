---
subcategory: "Interconnect"
layout: "aws"
page_title: "AWS: aws_interconnect_connection"
description: |-
  Terraform resource for managing an AWS Interconnect Connection.
---

# Resource: aws_interconnect_connection

Terraform resource for managing an AWS Interconnect connection.

AWS Interconnect provides managed private connectivity between your AWS network resources and select partner network resources, such as other cloud service providers (multicloud) or last-mile providers.

~> **Note:** A newly created connection starts in the `requested` state. To finish provisioning, you must confirm the connection on the Interconnect partner's portal, console, or APIs using the `activation_key` exported by this resource. This confirmation happens outside of AWS. Terraform considers the resource successfully created once the connection reaches the `requested` state.

## Example Usage

### Basic Usage

```terraform
resource "aws_interconnect_connection" "example" {
  bandwidth      = "10Gbps"
  environment_id = "mce-aws-gcp-iad"
  description    = "AWS to Partner interconnect"

  attach_point {
    direct_connect_gateway = "abcdef12-3456-7890-abcd-ef1234567890"
  }

  remote_account {
    identifier = "remote-account-identifier"
  }

  tags = {
    Environment = "prod"
  }
}
```

## Selecting a Region and Environment

The AWS Region and the remote partner location are selected separately:

* The **AWS Region** is set by the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference), or overridden per resource with the `region` argument.
* The **`environment_id`** selects a supported pairing of that AWS Region with a remote partner location (for example, AWS US East (N. Virginia) with GCP `us-east4`). Check the [regional availability](https://docs.aws.amazon.com/interconnect/latest/userguide/region-availability.html) in the AWS documentation.

~> **Note:** When overriding the resource-level `region` argument, the `environment_id` must belong to that same Region or the connection fails to create. Prefer setting the Region in the provider configuration.

## Argument Reference

The following arguments are required:

* `attach_point` - (Required) Attach point to which the connection logically connects within your AWS network. Changing this forces a new resource to be created. [See below](#attach_point).
* `bandwidth` - (Required) Desired bandwidth of the connection, in the format `<number>Mbps` or `<number>Gbps` (for example, `10Gbps`).
* `environment_id` - (Required) Identifier of the Environment on which this connection is created. The Environment determines the AWS Region and remote partner location pairing, and is scoped to the AWS Region you are operating in. Changing this forces a new resource to be created.

The following arguments are optional:

* `description` - (Optional) Description to distinguish this connection.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `remote_account` - (Optional) Account or principal identifying information that can be verified by the partner of the Environment. The kind of identifier expected is indicated by the Environment's `remote_identifier_type`. Changing this forces a new resource to be created. [See below](#remote_account).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### attach_point

Exactly one of the following arguments must be set.

* `arn` - (Optional) ARN of the attach point.
* `direct_connect_gateway` - (Optional) Identifier of a Direct Connect Gateway attach point.

### remote_account

* `identifier` - (Required) ID of the account or project on the partner's environment. For example, for GCP this is the Google Cloud project ID.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `activation_key` - Activation key associated with this connection. Use this key on the Interconnect partner's portal, console, or APIs to confirm the connection.
* `arn` - ARN of the connection.
* `billing_tier` - Billing tier this connection is currently assigned.
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
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an Interconnect Connection using the `arn`. For example:

```terraform
import {
  to = aws_interconnect_connection.example
  id = "arn:aws:interconnect:us-east-1:123456789012:connection/mcc-abcd1234"
}
```

Using `terraform import`, import an Interconnect Connection using the `arn`. For example:

```console
% terraform import aws_interconnect_connection.example arn:aws:interconnect:us-east-1:123456789012:connection/mcc-abcd1234
```
