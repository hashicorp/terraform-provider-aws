---
subcategory: "Detective"
layout: "aws"
page_title: "AWS: aws_detective_invitation_accepter"
description: |-
  Provides a resource to manage an Amazon Detective member invitation accepter.
---

# Resource: aws_detective_invitation_accepter

Provides a resource to manage an [Amazon Detective Invitation Accepter](https://docs.aws.amazon.com/detective/latest/APIReference/API_AcceptInvitation.html). Ensure that the accepter is configured to use the AWS account you wish to _accept_ the invitation from the primary graph owner account.

## Example Usage

```terraform
resource "aws_detective_graph" "primary" {}

resource "aws_detective_member" "primary" {
  account_id    = "ACCOUNT ID"
  email_address = "EMAIL"
  graph_arn     = aws_detective_graph.primary.id
  message       = "Message of the invite"
}

resource "aws_detective_invitation_accepter" "member" {
  provider  = "awsalternate"
  graph_arn = aws_detective_graph.primary.graph_arn

  depends_on = [aws_detective_member.primary]
}
```

## Argument Reference

This resource supports the following arguments:

* `graph_arn` - (Required) ARN of the behavior graph that the member account is accepting the invitation for.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier (ID) of the Detective invitation accepter.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_detective_invitation_accepter` using the graph ARN. For example:

```terraform
import {
  to = aws_detective_invitation_accepter.example
  id = "arn:aws:detective:us-east-1:123456789101:graph:231684d34gh74g4bae1dbc7bd807d02d"
}
```

Using `terraform import`, import `aws_detective_invitation_accepter` using the graph ARN. For example:

```console
% terraform import aws_detective_invitation_accepter.example arn:aws:detective:us-east-1:123456789101:graph:231684d34gh74g4bae1dbc7bd807d02d
```
