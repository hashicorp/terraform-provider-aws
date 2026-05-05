<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Roadmap:  October 2025 - December 2025

Every few months, the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top Community Issues, [Core Services](https://hashicorp.github.io/terraform-provider-aws/core-services/), and internal priorities. Where community sourced contributions exist we will work with the authors to review and merge their work. Where this does not exist or the original contributors are not available we will create the resources and implementation ourselves.

Each weekly release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap.

This roadmap does not describe all the work that will be included within this timeframe, but it does describe our focus. We will include other work as events occur.

In the period spanning July to September 2025 the AWS Provider added support for the following (among many others):

- Oracle Database on AWS
- DynamoDB Warm Throughput
- Amazon Workspaces Web

From October - December 2025, we will be prioritizing the following areas of work (and more):

## New Services / Features

### Amazon Bedrock AgentCore

Issue: [#43424](https://github.com/hashicorp/terraform-provider-aws/issues/43424)

[Amazon Bedrock AgentCore](https://aws.amazon.com/about-aws/whats-new/2025/10/amazon-bedrock-agentcore-available/) Amazon Bedrock AgentCore is an agentic platform to build, deploy and operate highly capable agents securely at scale using any framework, model, or protocol. AgentCore lets you build agents faster, enable agents to take actions across tools and data, run agents securely with low-latency and extended runtimes, and monitor agents in production - all without any infrastructure management.

Support for Amazon Bedrock AgentCore includes the following new resources:

New Resource(s):

- `aws_bedrockagentcore_agent_runtime`
- `aws_bedrockagentcore_runtime_endpoint`
- `aws_bedrockagentcore_gateway`
- `aws_bedrockagentcore_browser`
- `aws_bedrockagentcore_code_interpreter`
- `aws_bedrockagentcore_gateway_target`
- `aws_bedrockagentcore_memory`
- `aws_bedrockagentcore_oauth2_credential_provider`
- `aws_bedrockagentcore_workload_provider`
- `aws_bedrockagentcore_apikey_credential_provider`

### AWS Transfer Family Web Apps

Issue: [#40996](https://github.com/hashicorp/terraform-provider-aws/issues/40996)

[AWS Transfer Family Web Apps](https://aws.amazon.com/aws-transfer-family/web-apps/) Transfer Family web apps offer a no-code, fully managed browser-based experience that enables secure file transfers to and from Amazon S3. Transfer Family web apps enable your authenticated users to perform essential file operations—including listing, uploading, downloading, and deleting—while maintaining security, reliability, and compliance.

Support for AWS Transfer Family Web Apps includes:

New Resource(s):

- `aws_transfer_web_app`
- `aws_transfer_web_app_customization`

### Amazon SaaS Manager for Amazon CloudFront

Issue: [#42409](https://github.com/hashicorp/terraform-provider-aws/issues/42409)

[Amazon SaaS Manager for Amazon CloudFront](https://aws.amazon.com/about-aws/whats-new/2025/04/saas-manager-amazon-cloudfront/) Amazon SaaS Manager for Amazon CloudFront is a new Amazon CloudFront feature designed to efficiently manage content delivery across multiple websites for Software-as-a-Service (SaaS) providers, web development platforms, and companies with multiple brands/websites. CloudFront SaaS Manager provides a unified experience, alleviating the operational burden of managing multiple websites at scale, including TLS certificate management, DDoS protection, and observability.

New Resource(s):

- `aws_cloudfront_distribution_tenant`
- `aws_cloudfront_connection_group`

Affected Resource:

- `aws_cloudfront_distribution`

## Enhancements to Existing Services

This quarter most of our efforts will be focused on enhancements and stability improvements of our core services, rather than adding brand new services to the provider. The following list comprises the items most important to the community.

- [Allow to set preferred_cache_cluster_azs for ElastiCache Redis Сluster](https://github.com/hashicorp/terraform-provider-aws/issues/37497)
- [AWS_Route53_zone: support attribute-only search](https://github.com/hashicorp/terraform-provider-aws/pull/39671)
- [Add support for concurrency cross channel behaviour to aws_connect_routing_profile](https://github.com/hashicorp/terraform-provider-aws/issues/35018)
- [Parameter to enable Certificate-based-authentication in the directory configuration of Appstream](https://github.com/hashicorp/terraform-provider-aws/issues/31766)
- [Resources for Custom Billing View](https://github.com/hashicorp/terraform-provider-aws/issues/40677)
- [Support CHALLENGE WAF actions and overrides on individual WAF Rule Group Rules](https://github.com/hashicorp/terraform-provider-aws/issues/27862)
- [Add required suffix when specifying log group ARN](https://github.com/hashicorp/terraform-provider-aws/pull/35941)

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
