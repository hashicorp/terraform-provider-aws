---
subcategory: "OpenSearch"
layout: "aws"
page_title: "AWS: aws_opensearch_domain_replication"
description: |-
  Manages replication between two Amazon OpenSearch Service domains.
---

# Resource: aws_opensearch_domain_replication

Manages a replication relationship between two Amazon OpenSearch Service domains. Each resource represents exactly one primary-to-secondary pair. The domain named in `domain_name` becomes the primary and accepts writes. The domain in `peer_domain_arn` becomes the secondary and continuously copies the primary's indices.

Switchover and failover are not managed by this resource. Run them with the AWS CLI or SDK (`SwitchoverReplication`, `FailoverReplication`). These change only computed attributes, so neither causes a diff and the resource stays in state.

* After a switchover, the next refresh shows `role = SECONDARY`, `peer_role = PRIMARY` and `status = ACTIVE`. Replication continues in the reverse direction.
* After a failover, the next refresh shows `role = SECONDARY`, `peer_role = PRIMARY` and `status = EJECTED`. Replication has stopped. `terraform destroy` still removes the relationship. To protect the data again, destroy this resource and create a new one with the new primary as `domain_name`. The new secondary must hold no non-system indices.

~> **Note:** The secondary must hold no non-system indices when the relationship is created, because it is re-seeded from the primary. Both domains must have fine-grained access control, node-to-node encryption, encryption at rest and HTTPS enforced, and run the same engine version, OpenSearch 3.7 or later. Both domains must be in the same account. The caller needs `es:AddReplicationPeer` on both domains.

## Example Usage

### Basic Usage

```terraform
resource "aws_opensearch_domain_replication" "example" {
  domain_name     = aws_opensearch_domain.primary.domain_name
  peer_domain_arn = aws_opensearch_domain.secondary.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `domain_name` - (Required, Forces new resource) Name of the domain to replicate. It becomes the primary. Must be in the provider's Region.
* `peer_domain_arn` - (Required, Forces new resource) ARN of the domain to replicate to. It becomes the secondary and must hold no non-system indices.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `domain_arn` - ARN of the domain named in `domain_name`.
* `peer_role` - Current role of the domain in `peer_domain_arn`.
* `role` - Current role of the domain named in `domain_name`.
* `status` - Status of the relationship between the two domains. `ACTIVE` means the relationship is established. It does not mean the first copy of the primary's data has finished. After a failover, the status is `EJECTED` and replication has stopped.
* `status_message` - Additional detail about `status`, when the service provides it.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearch domain replication using the `domain_name` and `peer_domain_arn` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_opensearch_domain_replication.example
  id = "my-primary-domain,arn:aws:es:us-east-1:123456789012:domain/my-secondary-domain"
}
```

Using `terraform import`, import OpenSearch domain replication using the `domain_name` and `peer_domain_arn` separated by a comma (`,`). For example:

```console
% terraform import aws_opensearch_domain_replication.example my-primary-domain,arn:aws:es:us-east-1:123456789012:domain/my-secondary-domain
```
