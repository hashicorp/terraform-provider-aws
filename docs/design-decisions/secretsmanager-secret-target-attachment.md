# SecretsManager Secret Target Attachment

**Summary:** Assess the feasibility of replicating the `AWS::SecretsManager::SecretTargetAttachment` CloudFormation function with Terraform.
**Created**: 2023-10-25

---

The AWS Terraform provider has a [prioritized issue](https://github.com/hashicorp/terraform-provider-aws/issues/9183) requesting a Terraform AWS provider implementation of the CloudFormation [`AWS::SecretsManager::SecretTargetAttachment`](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-secretsmanager-secrettargetattachment.html) function.
This document will assess the feasibility of replicating this functionality in the AWS Terraform provider and document the alternative options available using existing SecretsManager resources.

## Background

The `AWS::SecretsManager::SecretTargetAttachment` function is a convenience helper to supplement an existing SecretsManager secret with database connection information from services like Amazon RDS or Amazon Redshift.
This function has no API equivalent (see the SecretsManager [API documentation](https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_Operations.html)), and appears to operate as an orchestration job working across AWS services.
In the absence of public APIs, the AWS provider cannot easily implement proper Terraform lifecycle handling on top of this workflow.
However, the existing SecretsManager resources and configuration options on RDS database resources provide options for practitioners to replicate most of this functionality.

### Manual Secret with Supplemental Connection Information

With this approach the database (RDS Postgres in this example) is initially created with a random password.
An empty secret is created at the same time.
Once database creation is complete, the username, password, and database connection information are all written to a new version of the existing secret.

```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}

provider "aws" {}
provider "random" {}

locals {
  username = "foo"
}

resource "random_password" "password" {
  length  = 20
  special = false
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = "postgres"
  engine_version = "15.3"

  preferred_instance_classes = ["db.t3.micro", "db.t3.small"]
}

resource "aws_db_instance" "test" {
  allocated_storage    = 20
  db_name              = "testdb"
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  username             = local.username
  password             = random_password.password.result
  parameter_group_name = "default.postgres15"
  skip_final_snapshot  = true
}

resource "aws_secretsmanager_secret" "test" {
  name_prefix = "jb-test"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({
    engine   = aws_db_instance.test.engine
    host     = aws_db_instance.test.address
    username = local.username
    password = random_password.password.result
    dbname   = aws_db_instance.test.db_name
    port     = aws_db_instance.test.port
  })
}
```

Manually created secrets require maintaining the lambda function executing the rotation.
Specifically, the [RotateSecret](https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_RotateSecret.html) API requires the `LambdaFunctionArn` argument to be provided when setting a rotation for a manually created secret (it is technically optional, but can only be omitted for “managed” secrets created by AWS).
AWS provides Lambda function [templates](https://docs.aws.amazon.com/secretsmanager/latest/userguide/reference_available-rotation-templates.html) for the most common secret rotation use cases, and the [AWS Serverless Application Repository](https://serverlessrepo.aws.amazon.com/application) contains many pre-built secret rotation functions.
Once a rotation Lambda function is deployed, rotation can be managed with the `aws_secretsmanager_secret_rotation` resource.

```hcl
resource "aws_secretsmanager_secret_rotation" "test" {
  secret_id = aws_secretsmanager_secret.test.id

  # Function templates available in the SecretsManager documentation,
  # and pre-built functions available in the Serverless Application
  # Repository.
  rotation_lambda_arn = aws_lambda_function.secret_rotation.arn

  rotation_rules {
    automatically_after_days = 30
  }
}
```

### Managed Secret

With this approach the database (RDS Postgres in this example) is initially created with the `manage_master_user_password` argument set to `true` (no password required).
RDS will create a new "managed" secret to store credentials as part of database creation.

```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {}

locals {
  username = "foo"
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = "postgres"
  engine_version = "15.3"

  preferred_instance_classes = ["db.t3.micro", "db.t3.small"]
}

resource "aws_db_instance" "test" {
  allocated_storage    = 20
  db_name              = "testdb"
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  username             = local.username
  parameter_group_name = "default.postgres15"
  skip_final_snapshot  = true

  manage_master_user_password = true
}

# Optionally fetch the secret data if attributes need to be used as inputs
# elsewhere.
data "aws_secretsmanager_secret" "test" {
  arn = aws_db_instance.test.master_user_secret[0].secret_arn
}
```

If information about the managed secret is required as an input to other resources, the `aws_secretsmanager_secret` data source can be used.
Because this secret is managed by Amazon RDS, the secret value cannot be modified to include supplemental connection information.
This may be a limitation the `SecretTargetAttachment` CloudFormation is able to work around via some internal AWS process, but there is currently no mechanism to implement this with publicly documented APIs.

As an alternative, supplemental connection information could be applied as `tags` if it’s absolutely necessary for the information to exist on the secret itself.

```hcl
# Optionally import the managed secret and modify attributes via Terraform
# (secret value cannot be modified).
import {
  to = aws_secretsmanager_secret.test
  id = aws_db_instance.test.master_user_secret[0].secret_arn
}

resource "aws_secretsmanager_secret" "test" {
  # customize tags, description, etc.
}
```

## Proposal

Given the currently available AWS APIs, there isn’t a path to implement the workflow from the `SecretTargetAttachment` CloudFormation function in the AWS Terraform Provider.
Existing resources provide options to either fully customize a manual secret with both database credentials and connection information, or to fully offload secret management to AWS.
These options cover the core use case of storing and rotating database secrets via AWS SecretsManager.

With existing options already available in the provider, and no clear path forward the proposal is to close this issue.

## Consequences/Future Work

No work will be done to implement the requested functionality from the original issue. However, the investigation phase did uncover other potential enhancements to improve managed credential workflows in the AWS provider.

- The previous implementation of the `aws_secretsmanager_secret_rotation` resource did not allow for managed secret rotations to be modified (`rotation_lambda_arn` is required, but managed secrets omit this value on update). An [enhancement issue](https://github.com/hashicorp/terraform-provider-aws/issues/34108) was opened to support this use case, and the implementation was completed with [#34180](https://github.com/hashicorp/terraform-provider-aws/pull/34180).

- The `aws_redshift_cluster` resource did not implement support for managing master passwords. An [enhancement issue](https://github.com/hashicorp/terraform-provider-aws/issues/34169) was opened to support this use case, and the implementation was completed with [#34182](https://github.com/hashicorp/terraform-provider-aws/pull/34182).

## References

AWS Documentation

- [SecretsManager API Reference](https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_Operations.html)
- [RDS Managed Secrets Guide](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/rds-secrets-manager.html#rds-secrets-manager-db-instance)
- [`AWS::SecretsManager::SecretTargetAttachment`](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-secretsmanager-secrettargetattachment.html)

Terraform Resources

- [`aws_secretsmanager_secret`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret)
- [`aws_secretsmanager_secret_version`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret_version)
- [`aws_secretsmanager_secret_rotation`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret_rotation)
- [`aws_db_instance`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/db_instance)
