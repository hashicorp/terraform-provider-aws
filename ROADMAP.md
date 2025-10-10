# Roadmap:  July 2025 - September 2025

Every few months, the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top Community Issues, [Core Services](https://hashicorp.github.io/terraform-provider-aws/core-services/), and internal priorities. Where community sourced contributions exist we will work with the authors to review and merge their work. Where this does not exist or the original contributors are not available we will create the resources and implementation ourselves.

Each weekly release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap.

This roadmap does not describe all the work that will be included within this timeframe, but it does describe our focus. We will include other work as events occur.

In the period spanning May to June 2025 the AWS Provider added support for the following (among many others):

- Major Release v6.0 - Multi Region Support
- AWS Workspaces Web
- AWS Notifications

From July - September 2025, we will be prioritizing the following areas of work:

## New Services / Features

### DynamoDB Warm Throughput

Issue: [#40141](https://github.com/hashicorp/terraform-provider-aws/issues/40141)

[DynamoDB Warm Throughput](https://aws.amazon.com/blogs/database/pre-warming-amazon-dynamodb-tables-with-warm-throughput/) Amazon DynamoDB now supports a new warm throughput value and the ability to easily pre-warm DynamoDB tables and indexes. The warm throughput value provides visibility into the number of read and write operations your DynamoDB tables can readily handle, while pre-warming lets you proactively increase the value to meet future traffic demands.

Support for DynamoDB Warm Throughput may include:

Affected Resource(s):

- `aws_dynamodb_table`

### Amazon Connect phone number and contact flow association support

Issue: [#26015](https://github.com/hashicorp/terraform-provider-aws/issues/26015)

[Amazon Connect phone number and contact flow association support](https://aws.amazon.com/about-aws/whats-new/2022/04/amazon-connect-api-claim-phone-numbers/) Amazon Connect launches API to claim new phone numbers and configure them programmatically. Using this API, you can programmatically search for and claim available phone numbers, associate phone numbers to your contact flows, or release phone numbers that are no longer needed.

Support for Amazon Connect resources may include:

Affected Resource(s):

- `aws_connect_phone_number`
- `aws_connect_phone_number_contact_flow_association`

### AWS Control Tower APIs to register Organizational Units

Issue: [#35849](https://github.com/hashicorp/terraform-provider-aws/issues/35849)

[AWS Control Tower APIs to register Organizational Units](https://aws.amazon.com/about-aws/whats-new/2024/02/aws-control-tower-apis-register-organizational-units/) AWS Control Tower customers can now programmatically extend governance to organizational units (OUs) via APIs. These new APIs enable the AWS Control Tower baseline which contains best practice configurations, controls, and resources required for AWS Control Tower governance. For example, when you enable a baseline on an OU, member accounts within the OU will receive resources including AWS IAM roles, AWS CloudTrail, AWS Config, AWS Identity Center, and come under AWS Control Tower governance.

### WAFv2 update rules shared with Firewall Manager

Issue: [#36941](https://github.com/hashicorp/terraform-provider-aws/issues/36941)

[WAFv2 update rules shared with Firewall Manager](https://docs.aws.amazon.com/waf/latest/developerguide/waf-policies.html#waf-policies-rule-groups) The Terraform AWS provider does use the UpdateWebACL API, but only for updating WAF ACLs that it manages and not quite in the way we need for dynamically managing shared Web ACLs within organizations using AWS Firewall Manager (FMS). This functionality is key as it allows different accounts to add their own rules to a shared Web ACL, promoting a flexible approach to security management.

## Enhancements to Existing Services

This quarter most of our efforts will be focused on enhancements and stability improvements of our core services, rather than adding brand new services to the provider. The following list comprises the items most important to the community.

- [Fixes in-place UpdateService stabilization](https://github.com/hashicorp/terraform-provider-aws/pull/43502)
- [Support for Cognito managed login branding](https://github.com/hashicorp/terraform-provider-aws/issues/42580)
- [Updating capacity provider configuration for ECS services](https://github.com/hashicorp/terraform-provider-aws/issues/43004)
- [aws_s3_bucket_lifecycle_configuration empty filter block produces a warning](https://github.com/hashicorp/terraform-provider-aws/issues/42714)

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
