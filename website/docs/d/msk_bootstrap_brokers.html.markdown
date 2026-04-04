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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cluster_arn` - (Required) ARN of the cluster the nodes belong to.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `bootstrap_brokers` - Comma-separated list of `hostname:port` pairs of Kafka brokers for plaintext connectivity. Empty string if plaintext is not enabled.
* `bootstrap_brokers_public_sasl_iam` - Comma-separated list of `hostname:port` pairs for public SASL/IAM access. Empty string if not enabled.
* `bootstrap_brokers_public_sasl_scram` - Comma-separated list of `hostname:port` pairs for public SASL/SCRAM access. Empty string if not enabled.
* `bootstrap_brokers_public_tls` - Comma-separated list of `hostname:port` pairs for public TLS access. Empty string if not enabled.
* `bootstrap_brokers_sasl_iam` - Comma-separated list of `hostname:port` pairs for SASL/IAM access. Empty string if not enabled.
* `bootstrap_brokers_sasl_scram` - Comma-separated list of `hostname:port` pairs for SASL/SCRAM access. Empty string if not enabled.
* `bootstrap_brokers_tls` - Comma-separated list of `hostname:port` pairs for TLS access. Empty string if not enabled.
* `bootstrap_brokers_vpc_connectivity_sasl_iam` - Comma-separated list of `hostname:port` pairs for VPC connectivity with SASL/IAM. Empty string if not enabled.
* `bootstrap_brokers_vpc_connectivity_sasl_scram` - Comma-separated list of `hostname:port` pairs for VPC connectivity with SASL/SCRAM. Empty string if not enabled.
* `bootstrap_brokers_vpc_connectivity_tls` - Comma-separated list of `hostname:port` pairs for VPC connectivity with TLS. Empty string if not enabled.
