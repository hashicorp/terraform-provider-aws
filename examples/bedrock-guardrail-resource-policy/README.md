<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Bedrock Guardrail Resource Policy example

This example demonstrates how to use `aws_bedrock_guardrail_resource_policy` to
attach resource-based policies (RBPs) to an Amazon Bedrock guardrail and,
optionally, to a system-defined guardrail profile used for Cross-Region Inference
(CRIS).

RBPs are required for organisation-level enforced guardrails
(`BEDROCK_POLICY` in AWS Organizations Service Control Policies): member accounts
must be granted `bedrock:ApplyGuardrail` on a guardrail that lives in the
management account before the SCP can enforce it.

## What this example creates

| Resource | Purpose |
|---|---|
| `aws_bedrock_guardrail.this` | Guardrail with content and PII filters, living in the management account |
| `aws_bedrock_guardrail_resource_policy.guardrail` | RBP on the guardrail, scoped to all principals in the AWS Organisation |
| `aws_bedrock_guardrail_resource_policy.profile` | *(optional)* RBP on a system-defined guardrail profile for CRIS |

## Prerequisites

- Terraform ≥ 1.5
- Go toolchain (to build the provider locally)
- AWS credentials for the **management account** of an AWS Organization
- The caller must have permissions to call `bedrock:CreateGuardrail`,
  `bedrock:PutResourcePolicy`, and `organizations:DescribeOrganization`

## Building the provider locally

This example uses `aws_bedrock_guardrail_resource_policy`, which is not yet in a
published release of the provider.  You must build and install the provider from
this repository before running the example.

From the **repository root**:

```bash
make build
```

This compiles the provider and installs the binary to `$GOPATH/bin/terraform-provider-aws`.

## Running the example

Because the provider is not published to the registry, Terraform must be told to
use the local binary.  The `dev.tfrc` file in this directory contains the
necessary `dev_overrides` block.  Update the path inside it to match your
`$GOPATH/bin` directory if it differs, then prefix every Terraform command with
`TF_CLI_CONFIG_FILE=dev.tfrc`.

> **Note:** `dev_overrides` skips `terraform init` — go straight to `plan`.

### Basic (guardrail RBP only)

```bash
# from this directory, after running `make build` in the repo root
TF_CLI_CONFIG_FILE=dev.tfrc terraform plan
TF_CLI_CONFIG_FILE=dev.tfrc terraform apply
```

After applying, reference the `guardrail_arn` output in your
`BEDROCK_POLICY` service control policy.

### With Cross-Region Inference profile

If your organisation policy also enforces guardrails on Cross-Region Inference
calls, supply the system-defined profile ARN:

```bash
TF_CLI_CONFIG_FILE=dev.tfrc terraform apply \
  -var 'guardrail_profile_arn=arn:aws:bedrock:us-east-1::guardrail-profile/<id>'
```

### Override defaults

```bash
TF_CLI_CONFIG_FILE=dev.tfrc terraform apply \
  -var 'region=eu-west-1' \
  -var 'guardrail_name=my-org-guardrail'
```

## Outputs

| Name | Description |
|---|---|
| `guardrail_arn` | Use this in the `BEDROCK_POLICY` SCP |
| `guardrail_id` | Guardrail identifier |
| `organization_id` | AWS Organizations ID used to scope the RBPs |

## Clean up

```bash
TF_CLI_CONFIG_FILE=dev.tfrc terraform destroy
```
