---
subcategory: "Elastic Beanstalk"
layout: "aws"
page_title: "AWS: aws_elastic_beanstalk_application_version"
description: |-
  Provides an Elastic Beanstalk Application Version Resource
---

# Resource: aws_elastic_beanstalk_application_version

Provides an Elastic Beanstalk Application Version Resource. Elastic Beanstalk allows
you to deploy and manage applications in the AWS cloud without worrying about
the infrastructure that runs those applications.

This resource creates a Beanstalk Application Version that can be deployed to a Beanstalk
Environment.

~> **NOTE on Application Version Resource:**  When using the Application Version resource with multiple
[Elastic Beanstalk Environments](elastic_beanstalk_environment.html) it is possible that an error may be returned
when attempting to delete an Application Version while it is still in use by a different environment.
To work around this you can either create each environment in a separate AWS account or create your `aws_elastic_beanstalk_application_version` resources with a unique names in your Elastic Beanstalk Application. For example &lt;revision&gt;-&lt;environment&gt;.

## Example Usage

```terraform
resource "aws_s3_bucket" "default" {
  bucket = "tftest.applicationversion.bucket"
}

resource "aws_s3_object" "default" {
  bucket = aws_s3_bucket.default.id
  key    = "beanstalk/go-v1.zip"
  source = "go-v1.zip"
}

resource "aws_elastic_beanstalk_application" "default" {
  name        = "tf-test-name"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_application_version" "default" {
  name        = "tf-test-version-label"
  application = "tf-test-name"
  description = "application version created by terraform"
  bucket      = aws_s3_bucket.default.id
  key         = aws_s3_object.default.key
}
```

### Container Image Usage

Application versions for the `Kubernetes` environment tier are backed by a container image instead of a source bundle. Specify a container image that you built and pushed to a registry yourself:

```terraform
resource "aws_elastic_beanstalk_application_version" "image" {
  name        = "tf-test-version-label"
  application = aws_elastic_beanstalk_application.default.name

  image_configuration {
    source {
      uri = "111122223333.dkr.ecr.us-east-1.amazonaws.com/my-app:latest"
    }
  }
}
```

Or have Elastic Beanstalk build the image from the Application Version source bundle:

```terraform
resource "aws_elastic_beanstalk_application_version" "build" {
  name        = "tf-test-version-label"
  application = aws_elastic_beanstalk_application.default.name
  bucket      = aws_s3_bucket.default.id
  key         = aws_s3_object.default.key

  image_configuration {
    build {
      type                    = "docker"
      dockerfile_location     = "Dockerfile"
      code_build_service_role = aws_iam_role.codebuild.arn
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `application` - (Required) Name of the Beanstalk Application the version is associated with.
* `name` - (Required) Unique name for the this Application Version.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `bucket` - (Optional) S3 bucket that contains the Application Version source bundle. Required unless `image_configuration.source` is specified.
* `description` - (Optional) Short description of the Application Version.
* `force_delete` - (Optional) On delete, force an Application Version to be deleted when it may be in use by multiple Elastic Beanstalk Environments.
* `image_configuration` - (Optional) Configuration block for the container image backing the Application Version. Conflicts with `build_configuration`. Detailed below.
* `key` - (Optional) S3 object that is the Application Version source bundle. Required unless `image_configuration.source` is specified.
* `process` - (Optional) Pre-processes and validates the environment manifest (env.yaml ) and configuration files (*.config files in the .ebextensions folder) in the source bundle. Validating configuration files can identify issues prior to deploying the application version to an environment. You must turn processing on for application versions that you create using AWS CodeBuild or AWS CodeCommit. For application versions built from a source bundle in Amazon S3, processing is optional. It validates Elastic Beanstalk configuration files. It doesn’t validate your application’s configuration files, like proxy server or Docker configuration.
* `tags` - (Optional) Key-value map of tags for the Elastic Beanstalk Application Version. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### image_configuration

Exactly one of `source` and `build` must be specified.

* `source` - (Optional) Configuration block for a container image that you built and pushed to a registry yourself. Elastic Beanstalk deploys the image without a build step. Conflicts with `bucket` and `key`. Detailed below.
* `build` - (Optional) Configuration block for the settings Elastic Beanstalk uses to build a container image from the Application Version source bundle. Requires `bucket` and `key`. Detailed below.

#### source

* `uri` - (Required) Location of the container image.

#### build

* `type` - (Required) How Elastic Beanstalk builds the container image. Valid values are `docker` and `buildpack`.
* `code_build_service_role` - (Required) ARN of the IAM role that AWS CodeBuild assumes to run the build.
* `architecture` - (Optional) Target CPU architecture for the build. Valid values are `amd64` and `arm64`.
* `buildpack` - (Optional) Buildpack to use when `type` is `buildpack`.
* `compute_type` - (Optional) Size of the compute resources that run the build. Valid values are `BUILD_GENERAL1_SMALL`, `BUILD_GENERAL1_MEDIUM`, and `BUILD_GENERAL1_LARGE`. Defaults to `BUILD_GENERAL1_MEDIUM`.
* `dockerfile_location` - (Optional) Location of the Dockerfile within the source bundle when `type` is `docker`.
* `timeout_in_minutes` - (Optional) How long, in minutes from 5 to 480, Elastic Beanstalk waits before stopping a build that has not completed. Defaults to 60.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN assigned by AWS for this Elastic Beanstalk Application.
* `build_arn` - Reference to the artifact from the AWS CodeBuild build. Only set for an Application Version that Elastic Beanstalk builds.
* `image_uri` - Location of the container image for the Application Version. For an image built by Elastic Beanstalk, this is the image it pushed after the build succeeded.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
