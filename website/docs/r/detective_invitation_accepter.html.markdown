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
  account_id = "ACCOUNT ID"
  email      = "EMAIL"
  graph_arn  = aws_detective_graph.primary.id
  message    = "Message of the invite"
}

resource "aws_detective_invitation_accepter" "member" {
  provider  = "awsalternate"
  graph_arn = aws_detective_member.primary.graph_arn

  depends_on = [aws_detective_member.test]
}
```

## Argument Reference

The following arguments are supported:

* `graph_arn` - (Required) ARN of the behavior graph that the member account is accepting the invitation for.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique identifier (ID) of the Detective invitation accepter.

## Import

`aws_detective_invitation_accepter` can be imported using the graph ARN, e.g.

```
$ terraform import aws_detective_invitation_accepter.example arn:aws:detective:us-east-1:123456789101:graph:231684d34gh74g4bae1dbc7bd807d02d
```
