---
subcategory: "CodePipeline"
layout: "aws"
page_title: "AWS: aws_codepipeline"
description: |-
  Provides a CodePipeline
---

# Resource: aws_codepipeline

Provides a CodePipeline.

## Example Usage

```terraform
resource "aws_codepipeline" "codepipeline" {
  name     = "tf-test-pipeline"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.codepipeline_bucket.bucket
    type     = "S3"

    encryption_key {
      id   = data.aws_kms_alias.s3kmskey.arn
      type = "KMS"
    }
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["source_output"]

      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.example.arn
        FullRepositoryId = "my-organization/example"
        BranchName       = "main"
      }
    }
  }

  stage {
    name = "Build"

    action {
      name             = "Build"
      category         = "Build"
      owner            = "AWS"
      provider         = "CodeBuild"
      input_artifacts  = ["source_output"]
      output_artifacts = ["build_output"]
      version          = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }

  stage {
    name = "Deploy"

    action {
      name            = "Deploy"
      category        = "Deploy"
      owner           = "AWS"
      provider        = "CloudFormation"
      input_artifacts = ["build_output"]
      version         = "1"

      configuration = {
        ActionMode     = "REPLACE_ON_FAILURE"
        Capabilities   = "CAPABILITY_AUTO_EXPAND,CAPABILITY_IAM"
        OutputFileName = "CreateStackOutput.json"
        StackName      = "MyStack"
        TemplatePath   = "build_output::sam-templated.yaml"
      }
    }
  }
}

resource "aws_codestarconnections_connection" "example" {
  name          = "example-connection"
  provider_type = "GitHub"
}

resource "aws_s3_bucket" "codepipeline_bucket" {
  bucket = "test-bucket"
}

resource "aws_s3_bucket_public_access_block" "codepipeline_bucket_pab" {
  bucket = aws_s3_bucket.codepipeline_bucket.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["codepipeline.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "codepipeline_role" {
  name               = "test-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "codepipeline_policy" {
  statement {
    effect = "Allow"

    actions = [
      "s3:GetObject",
      "s3:GetObjectVersion",
      "s3:GetBucketVersioning",
      "s3:PutObjectAcl",
      "s3:PutObject",
    ]

    resources = [
      aws_s3_bucket.codepipeline_bucket.arn,
      "${aws_s3_bucket.codepipeline_bucket.arn}/*"
    ]
  }

  statement {
    effect    = "Allow"
    actions   = ["codestar-connections:UseConnection"]
    resources = [aws_codestarconnections_connection.example.arn]
  }

  statement {
    effect = "Allow"

    actions = [
      "codebuild:BatchGetBuilds",
      "codebuild:StartBuild",
    ]

    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "codepipeline_policy" {
  name   = "codepipeline_policy"
  role   = aws_iam_role.codepipeline_role.id
  policy = data.aws_iam_policy_document.codepipeline_policy.json
}

data "aws_kms_alias" "s3kmskey" {
  name = "alias/myKmsKey"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the pipeline.
* `pipeline_type` - (Optional) Type of the pipeline. Possible values are: `V1` and `V2`. Default value is `V1`.
* `role_arn` - (Required) A service role Amazon Resource Name (ARN) that grants AWS CodePipeline permission to make calls to AWS services on your behalf.
* `artifact_store` (Required) One or more artifact_store blocks. Artifact stores are documented below.
* `execution_mode` (Optional) The method that the pipeline will use to handle multiple executions. The default mode is `SUPERSEDED`. For value values, refer to the [AWS documentation](https://docs.aws.amazon.com/codepipeline/latest/APIReference/API_PipelineDeclaration.html#CodePipeline-Type-PipelineDeclaration-executionMode).

  **Note:** `QUEUED` or `PARALLEL` mode can only be used with V2 pipelines.
* `stage` (Minimum of at least two `stage` blocks is required) A stage block. Stages are documented below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `trigger` - (Optional) A trigger block. Valid only when `pipeline_type` is `V2`. Triggers are documented below.
* `variable` - (Optional) A pipeline-level variable block. Valid only when `pipeline_type` is `V2`. Variable are documented below.

### `artifact_store`

An `artifact_store` block supports the following arguments:

* `location` - (Required) The location where AWS CodePipeline stores artifacts for a pipeline; currently only `S3` is supported.
* `type` - (Required) The type of the artifact store, such as Amazon S3
* `encryption_key` - (Optional) The encryption key block AWS CodePipeline uses to encrypt the data in the artifact store, such as an AWS Key Management Service (AWS KMS) key. If you don't specify a key, AWS CodePipeline uses the default key for Amazon Simple Storage Service (Amazon S3). An `encryption_key` block is documented below.
* `region` - (Optional) The region where the artifact store is located. Required for a cross-region CodePipeline, do not provide for a single-region CodePipeline.

#### `encryption_key`

An `encryption_key` block supports the following arguments:

* `id` - (Required) The KMS key ARN or ID
* `type` - (Required) The type of key; currently only `KMS` is supported

### `stage`

A `stage` block supports the following arguments:

* `name` - (Required) The name of the stage.
* `action` - (Required) The action(s) to include in the stage. Defined as an `action` block below
* `before_entry` - (Optional) The method to use when a stage allows entry. For example, configuring this field for conditions will allow entry to the stage when the conditions are met.
* `on_success` - (Optional) The method to use when a stage has succeeded. For example, configuring this field for conditions will allow the stage to succeed when the conditions are met.
* `on_failure` - (Optional) The method to use when a stage has not completed successfully. For example, configuring this field for rollback will roll back a failed stage automatically to the last successful pipeline execution in the stage.

#### `action`

~> The input artifact of an action must exactly match the output artifact declared in a preceding action, but the input artifact does not have to be the next action in strict sequence from the action that provided the output artifact. Actions in parallel can declare different output artifacts, which are in turn consumed by different following actions.

An `action` block supports the following arguments:

* `category` - (Required) A category defines what kind of action can be taken in the stage, and constrains the provider type for the action. Possible values are `Approval`, `Build`, `Deploy`, `Invoke`, `Source` and `Test`.
* `owner` - (Required) The creator of the action being called. Possible values are `AWS`, `Custom` and `ThirdParty`.
* `name` - (Required) The action declaration's name.
* `provider` - (Required) The provider of the service being called by the action. Valid providers are determined by the action category. Provider names are listed in the [Action Structure Reference](https://docs.aws.amazon.com/codepipeline/latest/userguide/action-reference.html) documentation.
* `version` - (Required) A string that identifies the action type.
* `configuration` - (Optional) A map of the action declaration's configuration. Configurations options for action types and providers can be found in the [Pipeline Structure Reference](http://docs.aws.amazon.com/codepipeline/latest/userguide/reference-pipeline-structure.html#action-requirements) and [Action Structure Reference](https://docs.aws.amazon.com/codepipeline/latest/userguide/action-reference.html) documentation. Note: The `DetectChanges` parameter (optional, default value is true) in the `configuration` section causes CodePipeline to automatically start your pipeline upon new commits. Please refer to AWS Documentation for more details: https://docs.aws.amazon.com/codepipeline/latest/userguide/action-reference-CodestarConnectionSource.html#action-reference-CodestarConnectionSource-config.
* `input_artifacts` - (Optional) A list of artifact names to be worked on.
* `output_artifacts` - (Optional) A list of artifact names to output. Output artifact names must be unique within a pipeline.
* `role_arn` - (Optional) The ARN of the IAM service role that will perform the declared action. This is assumed through the roleArn for the pipeline.
* `run_order` - (Optional) The order in which actions are run.
* `region` - (Optional) The region in which to run the action.
* `namespace` - (Optional) The namespace all output variables will be accessed from.

#### `before_entry`

A `before_entry` block supports the following arguments:

* `condition` - (Required) The conditions that are configured as entry condition. Defined as a `condition` block below.

#### `on_success`

A `on_success` block supports the following arguments:

* `condition` - (Required) The conditions that are success conditions. Defined as a `condition` block below.

#### `on_failure`

A `on_failure` block supports the following arguments:

* `condition` - (Optional) The conditions that are failure conditions. Defined as a `condition` block below.
* `result` - (Optional) The conditions that are configured as failure conditions. Possible values are `ROLLBACK`,  `FAIL`, `RETRY` and `SKIP`.
* `retry_configuration` - (Optional) The retry configuration specifies automatic retry for a failed stage, along with the configured retry mode. Defined as a `retry_configuration` block below.

##### `condition`

A `condition` block supports the following arguments:

* `result` - (Optional) The action to be done when the condition is met. For example, rolling back an execution for a failure condition. Possible values are `ROLLBACK`, `FAIL`, `RETRY` and `SKIP`.
* `rule` - (Optional) The rules that make up the condition. Defined as a `rule` block below.

##### `rule`

A `rule` block supports the following arguments:

* `name` - (Required) The name of the rule that is created for the condition, such as `VariableCheck`.
* `rule_type_id` - (Required) The ID for the rule type, which is made up of the combined values for `category`, `owner`, `provider`, and `version`. Defined as an `rule_type_id` block below.
* `commands` - (Optional) The shell commands to run with your commands rule in CodePipeline. All commands are supported except multi-line formats.
* `configuration` - (Optional) The action configuration fields for the rule. Configurations options for rule types and providers can be found in the [Rule structure reference](https://docs.aws.amazon.com/codepipeline/latest/userguide/rule-reference.html).
* `input_artifacts` - (Optional) The list of the input artifacts fields for the rule, such as specifying an input file for the rule.
* `region` - (Optional) The Region for the condition associated with the rule.
* `role_arn` - (Optional) The pipeline role ARN associated with the rule.
* `timeout_in_minutes` - (Optional) The action timeout for the rule.

##### `rule_type_id`

A `rule_type_id` block supports the following arguments:

* `category` - (Required) A category defines what kind of rule can be run in the stage, and constrains the provider type for the rule. The valid category is `Rule`.
* `provider` - (Required) The rule provider, such as the DeploymentWindow rule. For a list of rule provider names, see the rules listed in the [AWS CodePipeline rule reference](https://docs.aws.amazon.com/codepipeline/latest/userguide/rule-reference.html).
* `owner` - (Optional) The creator of the rule being called. The valid value for the Owner field in the rule category is `AWS`.
* `version` - (Optional) A string that describes the rule version.

##### `retry_configuration`

A `retry_configuration` block supports the following arguments:

* `retry_mode` - (Optional) The method that you want to configure for automatic stage retry on stage failure. You can specify to retry only failed action in the stage or all actions in the stage. Possible values are `FAILED_ACTIONS` and `ALL_ACTIONS`.

### `trigger`

A `trigger` block supports the following arguments:

* `provider_type` - (Required) The source provider for the event. Possible value is `CodeStarSourceConnection`.
* `git_configuration` - (Required) Provides the filter criteria and the source stage for the repository event that starts the pipeline. For more information, refer to the [AWS documentation](https://docs.aws.amazon.com/codepipeline/latest/userguide/pipelines-filter.html). A `git_configuration` block is documented below.

#### `git_configuration`

A `git_configuration` block supports the following arguments:

* `source_action_name` - (Required) The name of the pipeline source action where the trigger configuration, such as Git tags, is specified. The trigger configuration will start the pipeline upon the specified change only.
* `pull_request` - (Optional) The field where the repository event that will start the pipeline is specified as pull requests. A `pull_request` block is documented below.
* `push` - (Optional) The field where the repository event that will start the pipeline, such as pushing Git tags, is specified with details. A `push` block is documented below.

##### `pull_request`

A `pull_request` block supports the following arguments:

* `events` - (Optional) A list that specifies which pull request events to filter on (opened, updated, closed) for the trigger configuration. Possible values are `OPEN`, `UPDATED ` and `CLOSED`.
* `branches` - (Optional) The field that specifies to filter on branches for the pull request trigger configuration. A `branches` block is documented below.
* `file_paths` - (Optional) The field that specifies to filter on file paths for the pull request trigger configuration. A `file_paths` block is documented below.

##### `push`

A `push` block supports the following arguments:

* `branches` - (Optional) The field that specifies to filter on branches for the push trigger configuration. A `branches` block is documented below.
* `file_paths` - (Optional) The field that specifies to filter on file paths for the push trigger configuration. A `file_paths` block is documented below.
* `tags` - (Optional) The field that contains the details for the Git tags trigger configuration. A `tags` block is documented below.

##### `branches`

A `branches` block supports the following arguments:

* `includes` - (Optional) A list of patterns of Git branches that, when a commit is pushed, are to be included as criteria that starts the pipeline.
* `excludes` - (Optional) A list of patterns of Git branches that, when a commit is pushed, are to be excluded from starting the pipeline.

##### `file_paths`

A `file_paths` block supports the following arguments:

* `includes` - (Optional) A list of patterns of Git repository file paths that, when a commit is pushed, are to be included as criteria that starts the pipeline.
* `excludes` - (Optional) A list of patterns of Git repository file paths that, when a commit is pushed, are to be excluded from starting the pipeline.

##### `tags`

A `tags` block supports the following arguments:

* `includes` - (Optional) A list of patterns of Git tags that, when pushed, are to be included as criteria that starts the pipeline.
* `excludes` - (Optional) A list of patterns of Git tags that, when pushed, are to be excluded from starting the pipeline.

### `variable`

A `variable` block supports the following arguments:

* `name` - (Required) The name of a pipeline-level variable.
* `default_value` - (Optional) The default value of a pipeline-level variable.
* `description` - (Optional) The description of a pipeline-level variable.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Codepipeline ID.
* `arn` - Codepipeline ARN.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `trigger_all` - A list of all triggers present on the pipeline, including default triggers added by AWS for `V2` pipelines which omit an explicit `trigger` definition.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodePipelines using the `name`. For example:

```terraform
import {
  to = aws_codepipeline.example
  id = "example-pipeline"
}
```

Using `terraform import`, import CodePipelines using the `name`. For example:

```console
% terraform import aws_codepipeline.example example-pipeline
```
