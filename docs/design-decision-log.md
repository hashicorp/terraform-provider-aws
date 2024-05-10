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
