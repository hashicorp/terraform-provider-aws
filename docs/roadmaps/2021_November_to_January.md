# Roadmap:  November 2021 - January 2022

Every few months, the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top Community Issues, [Core Services](../contributing/core-services.md), and internal priorities. Where community sourced contributions exist we will work with the authors to review and merge their work. Where this does not exist or the original contributors are not available we will create the resources and implementation ourselves.

Each weekly release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap. To view all the items we've prioritized for this quarter, please see the [Roadmap milestone](https://github.com/hashicorp/terraform-provider-aws/milestone/138).

This roadmap does not describe all the work that will be included within this timeframe, but it does describe our focus. We will include other work as events occur .

In the period spanning August to October 2021, 573 Pull Requests were opened in the provider and 465 were merged, adding support for:

- Amazon Chime
- Amazon Connect
- Amazon AppStream 2.0
- Route 53 Recovery Control
- Graviton2 support for Lambda
- S3 Replication Time Control

We also launched a fully generated provider, the AWS Cloud Control (AWSCC) provider for Terraform. The AWSCC provider is currently in Technical Preview. Please check it out and let us know what you think.

- [HashiCorp Blog Announcement](https://www.hashicorp.com/blog/announcing-terraform-aws-cloud-control-provider-tech-preview)
- [GitHub Repository](https://github.com/hashicorp/terraform-provider-awscc)
- [AWS Cloud Control on the Terraform Registry](https://registry.terraform.io/providers/hashicorp/awscc/latest)

From November ‘21- January ‘22, we will be prioritizing the following areas of work:

## Provider Version v4.0.0

Issue: [#20433](https://github.com/hashicorp/terraform-provider-aws/issues/20433)

The next major release of the provider will include the adoption of the AWS Go SDK v2.0 as well as a refactor of one of our oldest and most used resources: S3.

There will also be the usual deprecations and sometimes breaking changes to existing resources which are necessary to maintain consistency of behavior across resources. Our goal is to focus on standardization to reduce technical debt and lay a strong foundation for future enhancement initiatives within the provider.

For details of the changes in full please refer to #20433. We would love to hear your feedback.

## Enhancements to Existing Services

- [Support for Managing Amazon CloudSearch Domains](https://github.com/hashicorp/terraform-provider-aws/issues/7833)
- [aws_config_remediation_configuration: No support for "automatic" remediation](https://github.com/hashicorp/terraform-provider-aws/issues/15491)
- [S3 Intelligent-Tiering Archive configuration](https://github.com/hashicorp/terraform-provider-aws/issues/16123)
- [IoT Thing Group](https://github.com/hashicorp/terraform-provider-aws/issues/8801)
- [Add resource for CodeCommit approval rule templates](https://github.com/hashicorp/terraform-provider-aws/issues/11461)
- [aws_dlm_lifecycle_policy - Implement support for "Cross Region copy"](https://github.com/hashicorp/terraform-provider-aws/issues/12204)
- [Add a data source for aws_key_pair](https://github.com/hashicorp/terraform-provider-aws/issues/15590)
- [Support ECS TaskSet](https://github.com/hashicorp/terraform-provider-aws/issues/8124)
- [Support for AthenaEngineVersion option in Athena work groups](https://github.com/hashicorp/terraform-provider-aws/issues/17456)
- [ECS Service can't update desired replicas when Blue Green deployment is enabled](https://github.com/hashicorp/terraform-provider-aws/issues/13658)
- [Add connection termination control to AWS LB target group](https://github.com/hashicorp/terraform-provider-aws/issues/17227)
- [WAFv2: Added support for custom response bodies](https://github.com/hashicorp/terraform-provider-aws/pull/19764)
- [New Resource aws_route53domains_domain](https://github.com/hashicorp/terraform-provider-aws/pull/12711)
- [Add aws_cognito_user resource](https://github.com/hashicorp/terraform-provider-aws/pull/19919)
- [AWS dynamodb table: restore from point in time](https://github.com/hashicorp/terraform-provider-aws/pull/19292)
- [Added `retain` parameter to `aws_lambda_layer_version` resource](https://github.com/hashicorp/terraform-provider-aws/pull/11997)
- [New Resource: aws_lambda_layer_version_permission](https://github.com/hashicorp/terraform-provider-aws/pull/11941)
- [resoure/aws_lb: Support WAF fial open](https://github.com/hashicorp/terraform-provider-aws/pull/16393)
- [aws_elb & aws_lb: Add desync_mitigation_mode](https://github.com/hashicorp/terraform-provider-aws/pull/14764)
- [Implement object lambda access points for S3](https://github.com/hashicorp/terraform-provider-aws/pull/19294)
- [WAFv2: Added support for label_match_statement and rule_label](https://github.com/hashicorp/terraform-provider-aws/pull/19576)
- [Cloudtrail: Exclude Management Event Sources](https://github.com/hashicorp/terraform-provider-aws/pull/17203)
- [Retry S3 OperationAborted errors](https://github.com/hashicorp/terraform-provider-aws/pull/12949)
- [aws_dms_endpoint: support for secrets id for oracle and postgres](https://github.com/hashicorp/terraform-provider-aws/pull/19040)
- [Add support for private_ip_list](https://github.com/hashicorp/terraform-provider-aws/pull/17846)
- [aws_emr_cluster: Fix aws_emr_security_configuration destroy issues](https://github.com/hashicorp/terraform-provider-aws/pull/12578)
- [New Resource: aws_ecrpublic_repository_policy](https://github.com/hashicorp/terraform-provider-aws/pull/16901)
- [aws_ecs_task_definition overwrites previous revision](https://github.com/hashicorp/terraform-provider-aws/issues/258)
- [Order is lost for data aws_iam_policy_document when applied to S3 buckets, iam roles, kms keys](https://github.com/hashicorp/terraform-provider-aws/issues/11801)
- [aws_ecs_cluster with capacity_providers cannot be destroyed](https://github.com/hashicorp/terraform-provider-aws/issues/11409)
- [Support for Account Settings Flags](https://github.com/hashicorp/terraform-provider-aws/issues/10168)
- [Execute AWS Lambda Only Once](https://github.com/hashicorp/terraform-provider-aws/issues/4746)

## Research Topics

Research topics include features, architectural changes, and ideas that we are pursuing in the longer term that may significantly impact the core user experience of the AWS provider. Research topics are discovery only and are not guaranteed to be included in a future release.

### Scaffolding for new resources, datasources and associated tests

Adding resources, datasources and test files to the provider is a repetitive task which should be automated to ensure consistency and speed up contributor and maintainer workflow. A simple cli tool should be able to generate these files in place, and ensure that any code reference additions required (ie adding to `provider.go`) are performed as part of the process.

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
