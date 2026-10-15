<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Roadmap: October 2026 - December 2026

As we head into Q4, we want to give a clearer view into some of the areas the AWS Provider team will be focused on over the coming months.

This roadmap highlights a subset of that work, informed by Top Community Issues, customer feedback and requests, [Core Services](https://hashicorp.github.io/terraform-provider-aws/core-services/), AWS service priorities, and other areas we see as important for the provider. Where community contributions already exist, we will work with contributors to review and move those forward. Where they do not, we will take on the implementation directly.

This is not intended to cover everything the team will work on. Weekly releases will continue to include community pull requests, bug fixes, enhancements, new features, and support for AWS service changes outside the areas called out below.

From October through December 2026, our areas of focus will include the following:

## New Services / Features

### AI and Agents

- **Amazon Bedrock Data Automation**: blueprints and projects ([#41643](https://github.com/hashicorp/terraform-provider-aws/issues/41643))
- **Amazon Bedrock Enforced Guardrails**: `aws_bedrock_enforced_guardrail_configuration` ([#47408](https://github.com/hashicorp/terraform-provider-aws/pull/47408))
- **AWS DevOps Agent**: agent spaces and asset management ([#48404](https://github.com/hashicorp/terraform-provider-aws/issues/48404))
- **Amazon SageMaker AI**: inference components ([#48894](https://github.com/hashicorp/terraform-provider-aws/pull/48894)) and JumpStart private hub coverage

### Compute and Content Delivery

- **Amazon EC2 Capacity Manager** ([#45827](https://github.com/hashicorp/terraform-provider-aws/issues/45827))
- **Amazon CloudFront flat-rate plans** ([#45450](https://github.com/hashicorp/terraform-provider-aws/issues/45450))

### Data, Identity, and Messaging

- **AWS Directory Service user and group management** ([#39469](https://github.com/hashicorp/terraform-provider-aws/issues/39469))
- **Oracle Database@AWS multitenant databases (CDB/PDB)**
- **Amazon FSx for NetApp ONTAP S3 access points** ([#45362](https://github.com/hashicorp/terraform-provider-aws/issues/45362))
- **AWS End User Messaging protect configurations** ([#39605](https://github.com/hashicorp/terraform-provider-aws/issues/39605))

## Enhancements to Existing Services

Alongside new service coverage, we will continue expanding and improving existing AWS resources, with particular focus on highly requested community and customer asks.

- [Round out Lambda MicroVMs with image hooks, logging, and resource settings, plus an image version data source](https://github.com/hashicorp/terraform-provider-aws/issues/48526)
- [Use an RDS cluster as the replication source for `aws_db_instance`](https://github.com/hashicorp/terraform-provider-aws/issues/39270)
- [Expose `publicly_accessible` on `aws_rds_cluster`](https://github.com/hashicorp/terraform-provider-aws/issues/39618)
- [Update `tunnel_bandwidth` on `aws_vpn_connection` without replacement](https://github.com/hashicorp/terraform-provider-aws/pull/47843)
- [Add a placement group to `aws_instance` without recreation](https://github.com/hashicorp/terraform-provider-aws/issues/47679)
- [Manage AWS Managed Microsoft AD directory settings](https://github.com/hashicorp/terraform-provider-aws/issues/46905)
- [Expanded QuickSight custom permissions](https://github.com/hashicorp/terraform-provider-aws/pull/47135), including deny-by-default and user-level permissions
- [Directory Service IP routes](https://github.com/hashicorp/terraform-provider-aws/issues/50129) (released in v6.67.0)
- [`platform_version` for `aws_bedrockagentcore_agent_runtime`](https://github.com/hashicorp/terraform-provider-aws/issues/50078) (merged; ships in v6.68.0)
- Stability and correctness improvements for tag plan behavior, KMS key deletion, Direct Connect MACsec keys, and CloudWatch Logs metric filter limits

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
