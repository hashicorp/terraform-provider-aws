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
cluster with IAM authentication enabled, a customer-managed KMS key + Secrets Manager secret
holding the SCRAM credentials, and a bastion instance for reading the source cluster's Kafka
`cluster.id`. Using MSK for the source means its server certificate is signed by a public CA,
so the replicator trusts it without a custom root CA.

The target cluster must have IAM client authentication enabled. MSK Replicator authenticates
to its target over IAM, and `CreateReplicator` otherwise fails with
`InvalidInput.InvalidKafkaCluster` ("IAM Auth is not enabled for the Amazon MSK Cluster").

## Usage

```console
terraform init
terraform apply
```

Two MSK clusters are created; apply takes ~30-45 minutes and incurs cost (roughly $1.35/hour
for the clusters, endpoints, and bastion). Destroy with `terraform destroy` when finished.

> [!NOTE]
> The generated SCRAM password is stored in `terraform.tfstate` in this directory. The state
> file is gitignored; delete it along with the infrastructure.

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

### Obtain the Kafka cluster ID

`MSK_ONPREM_KAFKA_CLUSTER_ID` is the Kafka `cluster.id` reported by the source brokers. No MSK
control-plane API exposes it, so it has to be read from the data plane. The bastion created by
this config carries `/opt/cluster-id.sh`, which builds a SASL/SCRAM client config from the
Secrets Manager secret and asks the brokers via `kafka-cluster.sh cluster-id`. Run it with
Session Manager:

```console
command_id=$(aws ssm send-command \
  --instance-ids "$(terraform output -raw bastion_instance_id)" \
  --document-name AWS-RunShellScript \
  --parameters 'commands=["/opt/cluster-id.sh"]' \
  --query Command.CommandId --output text)

aws ssm get-command-invocation \
  --command-id "$command_id" \
  --instance-id "$(terraform output -raw bastion_instance_id)" \
  --query StandardOutputContent --output text
```

The output is `Cluster ID: <cluster.id>`. The bastion needs a minute or two after `apply` to
finish installing Java and the Kafka CLI via user data; `cloud-init status --wait` confirms it
is ready.

```console
export MSK_ONPREM_KAFKA_CLUSTER_ID=<cluster.id>
```

### Run the test

```console
TF_ACC=1 go test ./internal/service/kafka/ -run 'TestAccKafkaReplicator_selfManagedSASLSCRAM' -v -timeout 180m
```
