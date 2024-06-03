---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_bootstrap_brokers"
description: |-
    Get a list of brokers that a client application can use to bootstrap.
---

# Data Source: aws_msk_bootstrap_brokers

Get a list of brokers that a client application can use to bootstrap.

## Example Usage

```terraform
data "aws_msk_bootstrap_brokers" "example" {
  cluster_arn = aws_msk_cluster.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `cluster_arn` - (Required) ARN of the cluster the nodes belong to.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `bootstrap_brokers` - Comma separated list of one or more hostname:port pairs of kafka brokers suitable to bootstrap connectivity to the kafka cluster.
* `bootstrap_brokers_public_sasl_iam` - One or more DNS names (or IP addresses) and SASL IAM port pairs.
* `bootstrap_brokers_public_sasl_scram` - One or more DNS names (or IP addresses) and SASL SCRAM port pairs.
* `bootstrap_brokers_public_tls` - One or more DNS names (or IP addresses) and TLS port pairs.
* `bootstrap_brokers_sasl_iam` - One or more DNS names (or IP addresses) and SASL IAM port pairs.
* `bootstrap_brokers_sasl_scram` - One or more DNS names (or IP addresses) and SASL SCRAM port pairs.
* `bootstrap_brokers_tls` - One or more DNS names (or IP addresses) and TLS port pairs.
* `bootstrap_brokers_vpc_connectivity_sasl_iam` - A string containing one or more DNS names (or IP addresses) and SASL IAM port pairs for VPC connectivity.
* `bootstrap_brokers_vpc_connectivity_sasl_scram` - A string containing one or more DNS names (or IP addresses) and SASL SCRAM port pairs for VPC connectivity.
* `bootstrap_brokers_vpc_connectivity_tls` - A string containing one or more DNS names (or IP addresses) and TLS port pairs for VPC connectivity.
