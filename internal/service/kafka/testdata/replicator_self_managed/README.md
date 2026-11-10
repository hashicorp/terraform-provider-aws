<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Self-managed replicator acceptance test infrastructure

`TestAccKafkaReplicator_selfManagedSASLSCRAM` verifies the self-managed source path of
`aws_msk_replicator` end-to-end using SASL/SCRAM. MSK Replicator needs a source cluster that
is reachable from the replicator's VPC and whose Kafka `cluster.id` matches the configured
value — neither of which can be produced inside a throwaway acceptance-test VPC. The test
therefore provisions only the replicator and reads the surrounding infrastructure from
environment variables.

`main.tf` provisions that infrastructure: a shared VPC, a source MSK cluster (impersonating a
self-managed Apache Kafka cluster) with SASL/SCRAM enabled and a SCRAM user, a target MSK
cluster, and a customer-managed KMS key + Secrets Manager secret holding the SCRAM
credentials. Using MSK for the source means its server certificate is signed by a public CA,
so the replicator trusts it without a custom root CA.

## Usage

```console
terraform init
terraform apply
```

Two MSK clusters are created; apply takes ~30-45 minutes and incurs cost. Destroy with
`terraform destroy` when finished.

### Export the environment variables

All but `MSK_ONPREM_KAFKA_CLUSTER_ID` come straight from the outputs:

```console
export MSK_ONPREM_KAFKA_ENABLED=1
export MSK_ONPREM_KAFKA_BOOTSTRAP_BROKERS=$(terraform output -raw MSK_ONPREM_KAFKA_BOOTSTRAP_BROKERS)
export MSK_ONPREM_KAFKA_SASL_SCRAM_SECRET_ARN=$(terraform output -raw MSK_ONPREM_KAFKA_SASL_SCRAM_SECRET_ARN)
export MSK_ONPREM_KAFKA_TARGET_CLUSTER_ARN=$(terraform output -raw MSK_ONPREM_KAFKA_TARGET_CLUSTER_ARN)
export MSK_ONPREM_KAFKA_SUBNET_IDS=$(terraform output -raw MSK_ONPREM_KAFKA_SUBNET_IDS)
export MSK_ONPREM_KAFKA_SECURITY_GROUP_IDS=$(terraform output -raw MSK_ONPREM_KAFKA_SECURITY_GROUP_IDS)
```

### Obtain the Kafka cluster ID (manual)

`MSK_ONPREM_KAFKA_CLUSTER_ID` is the Kafka `cluster.id` reported by the source brokers. It is
not exposed by any MSK control-plane API, so read it from the data plane from a host with
network access to the source cluster (for example an EC2 instance in the VPC created above),
using the source's SASL/SCRAM bootstrap brokers:

```console
kafka-metadata-quorum.sh --bootstrap-server <bootstrap> --command-config client.properties describe --status
# or, with an AdminClient: describeCluster().clusterId()
export MSK_ONPREM_KAFKA_CLUSTER_ID=<cluster.id>
```

### Run the test

```console
TF_ACC=1 go test ./internal/service/kafka/ -run 'TestAccKafkaReplicator_selfManagedSASLSCRAM' -v -timeout 180m
```
