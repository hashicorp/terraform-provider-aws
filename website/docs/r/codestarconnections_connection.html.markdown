---
subcategory: "CodeStar Connections"
layout: "aws"
page_title: "AWS: aws_codestarconnections_connection"
description: |-
  Provides a CodeStar Connection
---

# Resource: aws_codestarconnections_connection

Provides a CodeStar Connection.

~> **NOTE:** The `aws_codestarconnections_connection` resource is created in the state `PENDING`. Authentication with the connection provider must be completed in the AWS Console. See the [AWS documentation](https://docs.aws.amazon.com/dtconsole/latest/userguide/connections-update.html) for details.

## Example Usage

```terraform
resource "aws_codestarconnections_connection" "example" {
  name          = "example-connection"
  provider_type = "Bitbucket"
}

resource "aws_codepipeline" "example" {
  name     = "tf-test-pipeline"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    # ...
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
        FullRepositoryId = "my-organization/test"
        BranchName       = "main"
      }
    }
  }

  stage {
    name = "Build"
    action {
      # ...
    }
  }

  stage {
    name = "Deploy"
    action {
      # ...
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the connection to be created. The name must be unique in the calling AWS account. Changing `name` will create a new resource.
* `provider_type` - (Optional) The name of the external provider where your third-party code repository is configured. Valid values are `Bitbucket`, `GitHub`, `GitHubEnterpriseServer`, `GitLab` or `GitLabSelfManaged`. Changing `provider_type` will create a new resource. Conflicts with `host_arn`
* `host_arn` - (Optional) The Amazon Resource Name (ARN) of the host associated with the connection. Conflicts with `provider_type`
* `tags` - (Optional) Map of key-value resource tags to associate with the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The codestar connection ARN.
* `arn` - The codestar connection ARN.
* `connection_status` - The codestar connection status. Possible values are `PENDING`, `AVAILABLE` and `ERROR`.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodeStar connections using the ARN. For example:

```terraform
import {
  to = aws_codestarconnections_connection.test-connection
  id = "arn:aws:codestar-connections:us-west-1:0123456789:connection/79d4d357-a2ee-41e4-b350-2fe39ae59448"
}
```

Using `terraform import`, import CodeStar connections using the ARN. For example:

```console
% terraform import aws_codestarconnections_connection.test-connection arn:aws:codestar-connections:us-west-1:0123456789:connection/79d4d357-a2ee-41e4-b350-2fe39ae59448
```
