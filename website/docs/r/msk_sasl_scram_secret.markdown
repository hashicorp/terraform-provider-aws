---
subcategory: "Managed Streaming for Kafka (MSK)"
layout: "aws"
page_title: "AWS: aws_msk_sasl_scram_secret"
description: |-
  Terraform resource for managing an AWS Managed Streaming for Kafka Sasl/Scram secrets
---

# Resource: aws_msk_sasl_scram_secret

Manages AWS Managed Streaming for Kafka Sasl/Scram secrets.  It is important to note that a policy is required for the secret resource in order for Kafka to be able to read it. This policy is attached automatically when the `aws_msk_sasl_scram_secret` is used, however, this policy will not be in terraform and as such, will present a diff on plan/apply. For that reason, you must use the secrets managed policy as shown below in order to ensure that the state is in a clean state after the creation of secret and the association to the cluster.

## Example Usage

```hcl
resource "aws_msk_cluster" "example" {
  cluster_name           = "example"
  kafka_version          = "2.4.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    instance_type   = "kafka.m5.large"
    ebs_volume_size = 1000
    client_subnets = [
      aws_subnet.subnet_az1.id,
      aws_subnet.subnet_az2.id,
      aws_subnet.subnet_az3.id,
    ]
    security_groups = [aws_security_group.sg.id]
  }
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

resource "aws_secretsmanager_secret_version" "example" {
  secret_id     = aws_secretsmanager_secret.example.id
  secret_string = jsonencode({ username = "example_user", password = "example_password" })
}

resource "aws_kms_key" "example" {
  description = "example"
}

resource "aws_msk_sasl_scram_secret" "msk" {
  cluster_arn     = aws_msk_cluster.example.arn
  secret_arn_list = [aws_secretsmanager_secret.example.arn]
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

* `cluster_arn` - (Required) Arn of the Kafka cluster.
* `secret_arn_list` - (Required) List of AWS Secret Manager secrets Arns to be associated.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `scram_secrets` - Amazon Resource Name (ARN) of the MSK cluster.


