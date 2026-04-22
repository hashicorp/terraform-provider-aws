<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Nested Virtualization Example

This example launches an EC2 instance with nested virtualization enabled.

Nested virtualization is supported on 8th generation Intel-based instance types (C8i, M8i, R8i, and their flex variants).

To run, configure your AWS provider as described in https://www.terraform.io/docs/providers/aws/index.html

## Running the example

```bash
terraform init
terraform apply
```

## Cleanup

```bash
terraform destroy
```
