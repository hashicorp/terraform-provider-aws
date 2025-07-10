---
subcategory: "RAM (Resource Access Manager)"
layout: "aws"
page_title: "AWS: aws_ram_resource_share_accepter"
description: |-
  Manages accepting a Resource Access Manager (RAM) Resource Share invitation.
---

# Resource: aws_ram_resource_share_accepter

Manage accepting a Resource Access Manager (RAM) Resource Share invitation. From a _receiver_ AWS account, accept an invitation to share resources that were shared by a _sender_ AWS account. To create a resource share in the _sender_, see the [`aws_ram_resource_share` resource](/docs/providers/aws/r/ram_resource_share.html).

~> **Note:** If both AWS accounts are in the same Organization and [RAM Sharing with AWS Organizations is enabled](https://docs.aws.amazon.com/ram/latest/userguide/getting-started-sharing.html#getting-started-sharing-orgs), this resource is not necessary as RAM Resource Share invitations are not used.

## Example Usage

This configuration provides an example of using multiple Terraform AWS providers to configure two different AWS accounts. In the _sender_ account, the configuration creates a `aws_ram_resource_share` and uses a data source in the _receiver_ account to create a `aws_ram_principal_association` resource with the _receiver's_ account ID. In the _receiver_ account, the configuration accepts the invitation to share resources with the `aws_ram_resource_share_accepter`.

```terraform
provider "aws" {
  profile = "profile2"
}

provider "aws" {
  alias   = "alternate"
  profile = "profile1"
}

resource "aws_ram_resource_share" "sender_share" {
  provider = aws.alternate

  name                      = "tf-test-resource-share"
  allow_external_principals = true

  tags = {
    Name = "tf-test-resource-share"
  }
}

resource "aws_ram_principal_association" "sender_invite" {
  provider = aws.alternate

  principal          = data.aws_caller_identity.receiver.account_id
  resource_share_arn = aws_ram_resource_share.sender_share.arn
}

data "aws_caller_identity" "receiver" {}

resource "aws_ram_resource_share_accepter" "receiver_accept" {
  share_arn = aws_ram_principal_association.sender_invite.resource_share_arn
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `share_arn` - (Required) The ARN of the resource share.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `invitation_arn` - The ARN of the resource share invitation.
* `share_id` - The ID of the resource share as displayed in the console.
* `status` - The status of the resource share (ACTIVE, PENDING, FAILED, DELETING, DELETED).
* `receiver_account_id` - The account ID of the receiver account which accepts the invitation.
* `sender_account_id` - The account ID of the sender account which submits the invitation.
* `share_name` - The name of the resource share.
* `resources` - A list of the resource ARNs shared via the resource share.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import resource share accepters using the resource share ARN. For example:

```terraform
import {
  to = aws_ram_resource_share_accepter.example
  id = "arn:aws:ram:us-east-1:123456789012:resource-share/c4b56393-e8d9-89d9-6dc9-883752de4767"
}
```

Using `terraform import`, import resource share accepters using the resource share ARN. For example:

```console
% terraform import aws_ram_resource_share_accepter.example arn:aws:ram:us-east-1:123456789012:resource-share/c4b56393-e8d9-89d9-6dc9-883752de4767
```
