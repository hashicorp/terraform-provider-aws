---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_serverless_cluster"
description: |-
  Terraform resource for managing an Amazon MSK Serverless cluster.
---

# Resource: aws_msk_serverless_cluster

Manages an Amazon MSK Serverless cluster.

-> **Note:** To manage a _provisioned_ Amazon MSK cluster, use the [`aws_msk_cluster`](/docs/providers/aws/r/msk_cluster.html) resource.

## Example Usage

```terraform
resource "aws_msk_serverless_cluster" "example" {
  cluster_name = "Example"

  vpc_config {
    subnet_ids         = aws_subnet.example[*].id
    security_group_ids = [aws_security_group.example.id]
  }

  client_authentication {
    sasl {
      iam {
        enabled = true
      }
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `client_authentication` - (Required) Specifies client authentication information for the serverless cluster. See below.
* `cluster_name` - (Required) The name of the serverless cluster.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_config` - (Required) VPC configuration information. See below.

### client_authentication Argument Reference

* `sasl` - (Required) Details for client authentication using SASL. See below.

### sasl Argument Reference

* `iam` - (Required) Details for client authentication using IAM. See below.

### iam Argument Reference

* `enabled` - (Required) Whether SASL/IAM authentication is enabled or not.

### vpc_config Argument Reference

* `security_group_ids` - (Optional) Specifies up to five security groups that control inbound and outbound traffic for the serverless cluster.
* `subnet_ids` - (Required) A list of subnets in at least two different Availability Zones that host your client applications.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the serverless cluster.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `120m`)
* `delete` - (Default `120m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MSK serverless clusters using the cluster `arn`. For example:

```terraform
import {
  to = aws_msk_serverless_cluster.example
  id = "arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3"
}
```

Using `terraform import`, import MSK serverless clusters using the cluster `arn`. For example:

```console
% terraform import aws_msk_serverless_cluster.example arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3
```
