# Roadmap:  Nov 2024 - Jan 2025

Every few months, the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top Community Issues, [Core Services](https://hashicorp.github.io/terraform-provider-aws/core-services/), and internal priorities. Where community sourced contributions exist we will work with the authors to review and merge their work. Where this does not exist or the original contributors are not available we will create the resources and implementation ourselves.

Each weekly release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap.

This roadmap does not describe all the work that will be included within this timeframe, but it does describe our focus. We will include other work as events occur.

In the period spanning Nov to Jan 2025 the AWS Provider added support for the following (among many others):

- AWS Chatbot
- Amazon Bedrock
- Amazon Route 53 Profiles & Zones
- Amazon Bedrock Agents
- Completed the migration to Amazon GO SDK v2

From Nov - Jan 2025, we will be prioritizing the following areas of work:

## New Services

### Amazon S3 Tables

Issue: [#40407](https://github.com/hashicorp/terraform-provider-aws/issues/40407)

[Amazon S3 Tables](https://aws.amazon.com/about-aws/whats-new/2024/12/amazon-s3-tables-apache-iceberg-tables-analytics-workloads/) Amazon S3 Tables deliver the first cloud object store with built-in Apache Iceberg support and the easiest way to store tabular data at scale. S3 Tables are specifically optimized for analytics workloads, resulting in up to 3x faster query throughput and up to 10x higher transactions per second than self-managed tables.

Support for additional S3 resources may include:

New Resource(s):

- `aws_s3tables_table_bucket`
- `aws_s3tables_table_bucket_policy`
- `aws_s3tables_table`
- `aws_s3tables_table_policy`
- `aws_s3tables_namespace`

### Amazon S3 Express Bucket Lifecycle Configuration

Issue: [#40261](https://github.com/hashicorp/terraform-provider-aws/issues/40261)

[Amazon S3 Express Bucket Lifecycle Configuration](https://aws.amazon.com/about-aws/whats-new/2024/11/amazon-s3-express-one-zone-s3-lifecycle-expirations/) Amazon S3 Express One Zone, a high-performance S3 storage class for latency-sensitive applications, now supports object expiration using S3 Lifecycle. S3 Lifecycle can expire objects based on age to help you automatically optimize storage costs.

Support for Amazon S3 Express resources may include:

Affected Resource(s):

- `aws_s3_bucket_lifecycle_configuration`

### Amazon EKS: Auto Mode

Issue: [#40373](https://github.com/hashicorp/terraform-provider-aws/issues/40373)

[Amazon EKS: Auto Mode](https://aws.amazon.com/about-aws/whats-new/2024/12/amazon-eks-auto-mode/) a new feature that fully automates compute, storage, and networking management for Kubernetes clusters. Amazon EKS Auto Mode simplifies running Kubernetes by offloading cluster operations to AWS, improves the performance and security of your applications, and helps optimize compute costs.

### Amazon ECS: Availability Zone Rebalancing

Issue: [#40221](https://github.com/hashicorp/terraform-provider-aws/issues/40221)

[ECS: Availability Zone Rebalancing](https://aws.amazon.com/about-aws/whats-new/2024/11/amazon-ecs-az-rebalancing-speeds-mean-time-recovery-event/) a new feature that automatically redistributes containerized workloads across AZs. This capability helps reduce the mean time to recovery after infrastructure events, enabling applications to maintain high availability without requiring manual intervention.

## Enhancements to Existing Services

This quarter most of our efforts will be focused on enhancements and stability improvements of our core services, rather than adding brand new services to the provider. The following list comprises the items most important to the community.

- [Enable Deletion Protection for DynamoDB Table Replicas](https://github.com/hashicorp/terraform-provider-aws/issues/30213)
- [WAFv2 update rules shared with Firewall Manager](https://github.com/hashicorp/terraform-provider-aws/issues/36941)
- [Add support for enabling primary ipv6 address on EC2 instance](https://github.com/hashicorp/terraform-provider-aws/pull/36425)
- [Timestream Scheduled Query](https://github.com/hashicorp/terraform-provider-aws/issues/22507)
- [Log Anomaly Detector](https://github.com/hashicorp/terraform-provider-aws/issues/22507)

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
