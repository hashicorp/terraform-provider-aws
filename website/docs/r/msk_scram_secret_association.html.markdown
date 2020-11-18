---
subcategory: "Managed Streaming for Kafka (MSK)"
layout: "aws"
page_title: "AWS: aws_msk_scram_secret_association"
description: |-
  Provides resource for managing an AWS Managed Streaming for Kafka (MSK) Scram secret association
---

# Resource: aws_msk_scram_secret_association

Manages AWS Managed Streaming for Kafka Scram secrets.  It is important to note that a policy is required for the secret resource in order for Kafka to be able to read it. This policy is attached automatically when the `aws_msk_scram_secret_association` is used, however, this policy will not be in terraform and as such, will present a diff on plan/apply. For that reason, you must use the secrets managed policy as shown below in order to ensure that the state is in a clean state after the creation of secret and the association to the cluster.

## Example Usage

```hcl
resource "aws_msk_scram_secret_association" "example" {
  cluster_arn     = aws_msk_cluster.example.arn
  secret_arn_list = [aws_secretsmanager_secret.example.arn]
}

resource "aws_msk_cluster" "example" {
  cluster_name  = "example"
  kafka_version = "2.4.1"
  # ... other configuration...
  client_authentication {
    sasl {
      scram = true
    }
  }
}

resource "aws_secretsmanager_secret" "example" {
  name = "AmazonMSK_example"
}

resource "aws_secretsmanager_secret_policy" "msk" {
  secret_arn = aws_secretsmanager_secret.msk.arn
  policy     = <<POLICY
{
  "Version" : "2012-10-17",
  "Statement" : [ {
    "Sid": "AWSKafkaResourcePolicy",
    "Effect" : "Allow",
    "Principal" : {
      "Service" : "kafka.amazonaws.com"
    },
    "Action" : "secretsmanager:getSecretValue",
    "Resource" : "${aws_secretsmanager_secret.msk.arn}"
  } ]
}
POLICY
}
```

## Argument Reference

The following arguments are supported:

* `cluster_arn` - (Required, Forces new resource) Amazon Resource Name (ARN) of the MSK cluster.
* `secret_arn_list` - (Required) List of AWS Secrets Manager secret Amazon Resource Names (ARNs) to be associated.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the MSK cluster.


## Import

MSK Scram Secret Associations can be imported using the `id` e.g.

```
$ terraform import aws_msk_scram_secret_association.example arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3
```
