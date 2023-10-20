---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_bootstrap_brokers"
description: |-
    Get A list of brokers that a client application can use to bootstrap.
---

# Data Source: aws_msk_bootstrap_brokers

Get A list of brokers that a client application can use to bootstrap.

## Example Usage

```terraform
data "aws_msk_serverless_cluster" "example" {
  cluster_arn = aws_msk_serverless_cluster.example.arn
}

data "aws_msk_cluster" "example" {
  cluster_arn = aws_msk_cluster.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `cluster_arn` - (Required) ARN of the cluster the nodes belong to.

## Attribute Reference

### Nodes

* `bootstrap_broker_string` - A string containing one or more hostname:port pairs.
* `bootstrap_broker_string_tls` - A string containing one or more DNS names (or IP) and TLS port pairs.
* `bootstrap_broker_string_sasl_scram` - A string containing one or more DNS names (or IP) and Sasl Scram port pairs.
* `bootstrap_broker_string_sasl_iam` - A string that contains one or more DNS names (or IP addresses) and SASL IAM port pairs.
* `bootstrap_broker_string_public_tls` - A string containing one or more DNS names (or IP) and TLS port pairs.
* `bootstrap_broker_string_public_sasl_scram` - A string containing one or more DNS names (or IP) and Sasl Scram port pairs.
* `bootstrap_broker_string_public_sasl_iam` - A string that contains one or more DNS names (or IP addresses) and SASL IAM port pairs.
* `bootstrap_broker_string_vpc_connectivity_tls` - A string containing one or more DNS names (or IP) and TLS port pairs for VPC connectivity.
* `bootstrap_broker_string_vpc_connectivity_sasl_scram` - A string containing one or more DNS names (or IP) and SASL/SCRAM port pairs for VPC connectivity.
* `bootstrap_broker_string_vpc_connectivity_sasl_iam` - A string containing one or more DNS names (or IP) and SASL/IAM port pairs for VPC connectivity.
