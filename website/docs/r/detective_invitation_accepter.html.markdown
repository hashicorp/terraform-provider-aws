---
subcategory: "Detective"
layout: "aws"
page_title: "AWS: aws_detective_invitation_accepter"
description: |-
  Provides a resource to manage an Amazon Detective Invitation Accepter.
---

# Resource: aws_detective_invitation_accepter

Provides a resource to manage an [Amazon Detective Invitation Accepter](https://docs.aws.amazon.com/detective/latest/APIReference/API_AcceptInvitation.html).

## Example Usage

```terraform
resource "aws_detective_graph" "primary" {
  provider = "awsalternate"
}

resource "aws_detective2_member" "primary" {
  provider   = "awsalternate"
  
  account_id = "ACCOUNT ID"
  email      = "EMAIL"
  graph_arn  = aws_detective_graph.admin.id
  message    = "Message of the invite"
}

resource "aws_detective_invitation_accepter" "member" {
  graph_arn = aws_detective_graph.primary.id
}
```

## Argument Reference

The following arguments are supported:

* `graph_arn` - (Required) ARN of the behavior graph that the member account is accepting the invitation for.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique identifier (ID) of the detective invitation accepter.

## Import

`aws_detective_invitation_accepter` can be imported using the admin account ID, e.g.

```
$ terraform import aws_detective_invitation_accepter.example arn:aws:detective:us-east1:graph:testing
```