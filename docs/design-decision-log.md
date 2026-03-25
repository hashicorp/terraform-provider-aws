<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Design Decision Log

This serves as an index over the various decisions we make as a maintainer team over what is considered best practice, and
what we should encourage/require as a design standard. These are not necessarily fixed, and are likely to evolve and be
replaced as new decisions are made. This is an evolution of an internal process and will pivot to take place in public
as much as possible to allow for external feedback from the community and core contributors.

| Decision                                                                                                 | Description                                                                                                                     | Issue Link                                                                 |
|----------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------|
| [Relationship Resource Design Standards](./design-decisions/relationship-resource-design-standards.md)   | Align on design standards for relationship management resources in the Terraform AWS Provider.                                  | [#9901](https://github.com/hashicorp/terraform-provider-aws/issues/9901)   |
| [SecretsManager Secret Target Attachment](./design-decisions/secretsmanager-secret-target-attachment.md) | Assess the feasibility of replicating the `AWS::SecretsManager::SecretTargetAttachment` CloudFormation function with Terraform. | [#9183](https://github.com/hashicorp/terraform-provider-aws/issues/9183)   |
| [RDS Blue Green Deployments](./design-decisions/rds-bluegreen-deployments.md)                            | Assess the feasibility extending blue green deployment functionality found in `aws_rds_instance` to `aws_rds_cluster`.          | [#28956](https://github.com/hashicorp/terraform-provider-aws/issues/28956) |
| [Exclusive Relationship Management Resources](./design-decisions/exclusive-relationship-management-resources.md)| A proposal describing the use case for "exclusive relationship management" resources and their function within the Terraform AWS provider.                                 | [#39203](https://github.com/hashicorp/terraform-provider-aws/pull/39203)   |
| [Standardize Use of the `id` Attribute](./design-decisions/standardize-use-of-the-id-attribute.md)| Define a standard for use of the "id" attribute given improvements to provider development and testing libraries have removed its requirement.                                 | [#37628](https://github.com/hashicorp/terraform-provider-aws/pull/37628) |
| [Use `plancheck.ExpectResourceAction` with disappears acceptance tests](./design-decisions/expect-resource-action-with-disappears-tests.md)| Acceptance tests exercising out of band deletion (colloquially named "disappears" tests) should utilize the terraform-plugin-testing library's plancheck package to assert expected post apply actions. | N/A |
| [Supporting module-scoped User-Agents with `provider_meta`](./design-decisions/provider_meta.md)| A proposal for implementing `provider_meta` support and aligning the standard and AWSCC provider implementations. | [#45464](https://github.com/hashicorp/terraform-provider-aws/pull/45464) |
| [Migration from `go-changelog` to Changie](./design-decisions/changie-migration.md)| Migrate CHANGELOG generation from `go-changelog` to Changie for improved automation, better developer experience, and support for complex release workflows including beta/GA cycles. | N/A |
