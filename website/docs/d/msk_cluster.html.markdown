---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_cluster"
description: |-
  Get information on an Amazon MSK Cluster
---

# Data Source: aws_msk_cluster

Get information on an Amazon MSK Cluster.

-> **Note:** This data sources returns information on _provisioned_ clusters.

## Example Usage

```terraform
data "aws_msk_cluster" "example" {
  cluster_name = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `cluster_name` - (Required) Name of the cluster.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the MSK cluster.
* `bootstrap_brokers` - Comma separated list of one or more hostname:port pairs of kafka brokers suitable to bootstrap connectivity to the kafka cluster. Contains a value if `encryption_info.0.encryption_in_transit.0.client_broker` is set to `PLAINTEXT` or `TLS_PLAINTEXT`. The resource sorts values alphabetically. AWS may not always return all endpoints so this value is not guaranteed to be stable across applies.
* `bootstrap_brokers_public_sasl_iam` - One or more DNS names (or IP addresses) and SASL IAM port pairs. For example, `b-1-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9198,b-2-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9198,b-3-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9198`. This attribute will have a value if `encryption_info.0.encryption_in_transit.0.client_broker` is set to `TLS_PLAINTEXT` or `TLS` and `client_authentication.0.sasl.0.iam` is set to `true` and `broker_node_group_info.0.connectivity_info.0.public_access.0.type` is set to `SERVICE_PROVIDED_EIPS` and the cluster fulfill all other requirements for public access. The resource sorts the list alphabetically. AWS may not always return all endpoints so the values may not be stable across applies.
* `bootstrap_brokers_public_sasl_scram` - One or more DNS names (or IP addresses) and SASL SCRAM port pairs. For example, `b-1-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9196,b-2-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9196,b-3-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9196`. This attribute will have a value if `encryption_info.0.encryption_in_transit.0.client_broker` is set to `TLS_PLAINTEXT` or `TLS` and `client_authentication.0.sasl.0.scram` is set to `true` and `broker_node_group_info.0.connectivity_info.0.public_access.0.type` is set to `SERVICE_PROVIDED_EIPS` and the cluster fulfill all other requirements for public access. The resource sorts the list alphabetically. AWS may not always return all endpoints so the values may not be stable across applies.
* `bootstrap_brokers_public_tls` - One or more DNS names (or IP addresses) and TLS port pairs. For example, `b-1-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9194,b-2-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9194,b-3-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9194`. This attribute will have a value if `encryption_info.0.encryption_in_transit.0.client_broker` is set to `TLS_PLAINTEXT` or `TLS` and `broker_node_group_info.0.connectivity_info.0.public_access.0.type` is set to `SERVICE_PROVIDED_EIPS` and the cluster fulfill all other requirements for public access. The resource sorts the list alphabetically. AWS may not always return all endpoints so the values may not be stable across applies.
* `bootstrap_brokers_sasl_iam` - One or more DNS names (or IP addresses) and SASL IAM port pairs. For example, `b-1.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9098,b-2.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9098,b-3.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9098`. This attribute will have a value if `encryption_info.0.encryption_in_transit.0.client_broker` is set to `TLS_PLAINTEXT` or `TLS` and `client_authentication.0.sasl.0.iam` is set to `true`. The resource sorts the list alphabetically. AWS may not always return all endpoints so the values may not be stable across applies.
* `bootstrap_brokers_sasl_scram` - One or more DNS names (or IP addresses) and SASL SCRAM port pairs. For example, `b-1.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9096,b-2.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9096,b-3.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9096`. This attribute will have a value if `encryption_info.0.encryption_in_transit.0.client_broker` is set to `TLS_PLAINTEXT` or `TLS` and `client_authentication.0.sasl.0.scram` is set to `true`. The resource sorts the list alphabetically. AWS may not always return all endpoints so the values may not be stable across applies.
* `bootstrap_brokers_tls` - One or more DNS names (or IP addresses) and TLS port pairs. For example, `b-1.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9094,b-2.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9094,b-3.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9094`. This attribute will have a value if `encryption_info.0.encryption_in_transit.0.client_broker` is set to `TLS_PLAINTEXT` or `TLS`. The resource sorts the list alphabetically. AWS may not always return all endpoints so the values may not be stable across applies.
* `broker_node_group_info` - Configuration block for the broker nodes of the Kafka cluster. See [`broker_node_group_info` Block](#broker_node_group_info-block) for details.
* `cluster_uuid` - UUID of the MSK cluster, for use in IAM policies.
* `customer_action_status` - Status indicating whether Amazon MSK requires customer action for the cluster. Valid values are `NONE`, `ACTION_RECOMMENDED`, and `CRITICAL_ACTION_REQUIRED`.
* `kafka_version` - Apache Kafka version.
* `number_of_broker_nodes` - Number of broker nodes in the cluster.
* `tags` - Map of key-value pairs assigned to the cluster.
* `zookeeper_connect_string` - Comma separated list of one or more hostname:port pairs to use to connect to the Apache Zookeeper cluster. The returned values are sorted alphabetically. The AWS API may not return all endpoints, so this value is not guaranteed to be stable across applies.
* `zookeeper_connect_string_tls` - Comma separated list of one or more hostname:port pairs to use to connect to the Apache Zookeeper cluster via TLS. The returned values are sorted alphabetically. The AWS API may not return all endpoints, so this value is not guaranteed to be stable across applies.

### `broker_node_group_info` Block

* `az_distribution` - Distribution of broker nodes across Availability Zones.
* `client_subnets` - List of subnets to connect to in the client VPC.
* `connectivity_info` - Information about the cluster access configuration. See [`connectivity_info` Block](#connectivity_info-block) for details.
* `instance_type` - Instance type used for the broker nodes.
* `security_groups` - List of security groups to associate with the elastic network interfaces to control who can communicate with the cluster.
* `storage_info` - Storage information for the broker nodes. See [`storage_info` Block](#storage_info-block) for details.

### `connectivity_info` Block

* `network_type` - Network type of the cluster.
* `public_access` - Public access configuration. See [`public_access` Block](#public_access-block) for details.
* `vpc_connectivity` - VPC connectivity configuration. See [`vpc_connectivity` Block](#vpc_connectivity-block) for details.

### `public_access` Block

* `type` - Public access type.

### `vpc_connectivity` Block

* `client_authentication` - Client authentication configuration for VPC connectivity. See [`client_authentication` Block](#client_authentication-block) for details.

### `client_authentication` Block

* `sasl` - SASL authentication configuration for VPC connectivity. See [`sasl` Block](#sasl-block) for details.
* `tls` - Whether TLS authentication is enabled for VPC connectivity.

### `sasl` Block

* `iam` - Whether SASL IAM authentication is enabled for VPC connectivity.
* `scram` - Whether SASL SCRAM authentication is enabled for VPC connectivity.

### `storage_info` Block

* `ebs_storage_info` - EBS storage information. See [`ebs_storage_info` Block](#ebs_storage_info-block) for details.

### `ebs_storage_info` Block

* `provisioned_throughput` - Provisioned throughput configuration. See [`provisioned_throughput` Block](#provisioned_throughput-block) for details.
* `volume_size` - Size in GiB of the EBS volume for the data drive on each broker node.

### `provisioned_throughput` Block

* `enabled` - Whether provisioned throughput is enabled.
* `volume_throughput` - Throughput value of the EBS volumes for the data drive on each kafka broker node in MiB per second.
