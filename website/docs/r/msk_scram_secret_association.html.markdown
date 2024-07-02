---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_scram_secret_association"
description: |-
  Associates SCRAM secrets with a Managed Streaming for Kafka (MSK) cluster.
---

# Resource: aws_msk_scram_secret_association

Associates SCRAM secrets stored in the Secrets Manager service with a Managed Streaming for Kafka (MSK) cluster.

-> **Note:** The following assumes the MSK cluster has SASL/SCRAM authentication enabled. See below for example usage or refer to the [Username/Password Authentication](https://docs.aws.amazon.com/msk/latest/developerguide/msk-password.html) section of the MSK Developer Guide for more details.

To set up username and password authentication for a cluster, create an [`aws_secretsmanager_secret` resource](/docs/providers/aws/r/secretsmanager_secret.html) and associate
a username and password with the secret with an [`aws_secretsmanager_secret_version` resource](/docs/providers/aws/r/secretsmanager_secret_version.html). When creating a secret for the cluster,
the `name` must have the prefix `AmazonMSK_` and you must either use an existing custom AWS KMS key or create a new
custom AWS KMS key for your secret with the [`aws_kms_key` resource](/docs/providers/aws/r/kms_key.html). It is important to note that a policy is required for the `aws_secretsmanager_secret`
resource in order for Kafka to be able to read it. This policy is attached automatically when the `aws_msk_scram_secret_association` is used,
however, this policy will not be in terraform and as such, will present a diff on plan/apply. For that reason, you must use the [`aws_secretsmanager_secret_policy`
resource](/docs/providers/aws/r/secretsmanager_secret_policy.html) as shown below in order to ensure that the state is in a clean state after the creation of secret and the association to the cluster.

## Example Usage

```terraform
resource "aws_msk_scram_secret_association" "example" {
  cluster_arn     = aws_msk_cluster.example.arn
  secret_arn_list = [aws_secretsmanager_secret.example.arn]

  depends_on = [aws_secretsmanager_secret_version.example]
}

resource "aws_msk_cluster" "example" {
  cluster_name = "example"
  # ... other configuration...
  client_authentication {
    sasl {
      scram = true
    }
  }
}

resource "aws_secretsmanager_secret" "example" {
  name       = "AmazonMSK_example"
  kms_key_id = aws_kms_key.example.key_id
}

resource "aws_kms_key" "example" {
  description = "Example Key for MSK Cluster Scram Secret Association"
}

resource "aws_secretsmanager_secret_version" "example" {
  secret_id     = aws_secretsmanager_secret.example.id
  secret_string = jsonencode({ username = "user", password = "pass" })
}

data "aws_iam_policy_document" "example" {
  statement {
    sid    = "AWSKafkaResourcePolicy"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["kafka.amazonaws.com"]
    }

    actions   = ["secretsmanager:getSecretValue"]
    resources = [aws_secretsmanager_secret.example.arn]
  }
}

resource "aws_secretsmanager_secret_policy" "example" {
  secret_arn = aws_secretsmanager_secret.example.arn
  policy     = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

This resource supports the following arguments:

* `cluster_arn` - (Required, Forces new resource) Amazon Resource Name (ARN) of the MSK cluster.
* `secret_arn_list` - (Optional) List of AWS Secrets Manager secret ARNs.
* `secret_arn` - (Optional) Single AWS Secrets Manager secret ARN.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Amazon Resource Name (ARN) of the MSK cluster.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MSK SCRAM Secret Associations using the `id`. For example:

```terraform
import {
  to = aws_msk_scram_secret_association.example
  id = "arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3"
}
```

Alternatively, import a single MSK SCRAM Secret Association using a combination of the `id` and Secrets Manager secret ARN delimited by `#`. For example:

```terraform
import {
  to = aws_msk_scram_secret_association.example
  id = "arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3#arn:aws:secretsmanager:us-west-2:123456789012:secret:/example"
}
```

Using `terraform import`, import MSK SCRAM Secret Associations using the `id`. For example:

```console
% terraform import aws_msk_scram_secret_association.example arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3
```

Alternatively, import a single MSK SCRAM Secret Association using a combination of the `id` and Secrets Manager secret ARN delimited by `#`. For example:

```console
% terraform import aws_msk_scram_secret_association.example arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3#arn:aws:secretsmanager:us-west-2:123456789012:secret:/example
```
