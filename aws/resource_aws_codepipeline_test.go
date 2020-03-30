package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSCodePipeline_basic(t *testing.T) {
	var p1, p2 codepipeline.PipelineDeclaration
	name := acctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodePipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p1),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.codepipeline_role", "arn"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codepipeline", regexp.MustCompile(fmt.Sprintf("test-pipeline-%s", name))),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.0.type", "S3"),
					resource.TestCheckResourceAttrPair(resourceName, "artifact_store.0.location", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.0.encryption_key.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.0.encryption_key.0.id", "1234"),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.0.encryption_key.0.type", "KMS"),
					resource.TestCheckResourceAttr(resourceName, "artifact_stores.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.category", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.owner", "ThirdParty"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.provider", "GitHub"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.input_artifacts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Owner", "lifesum-terraform"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Repo", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Branch", "master"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.run_order", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.region", ""),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.category", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.provider", "CodeBuild"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.output_artifacts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.ProjectName", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.run_order", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodePipelineConfig_basicUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p2),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.0.type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.0.encryption_key.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.0.encryption_key.0.id", "4567"),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.0.encryption_key.0.type", "KMS"),
					resource.TestCheckResourceAttr(resourceName, "artifact_stores.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.input_artifacts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.0", "artifacts"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Owner", "test-terraform"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Repo", "test-repo"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Branch", "stable"),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.0", "artifacts"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.ProjectName", "test"),
				),
			},
		},
	})
}

func TestAccAWSCodePipeline_emptyArtifacts(t *testing.T) {
	var p codepipeline.PipelineDeclaration
	name := acctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodePipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineConfig_emptyArtifacts(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codepipeline", regexp.MustCompile(fmt.Sprintf("test-pipeline-%s", name))),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.0.type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.0.encryption_key.#", "0"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodePipeline_deployWithServiceRole(t *testing.T) {
	var p codepipeline.PipelineDeclaration
	name := acctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodePipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineConfig_deployWithServiceRole(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "stage.2.name", "Deploy"),
					resource.TestCheckResourceAttr(resourceName, "stage.2.action.0.category", "Deploy"),
					resource.TestCheckResourceAttrPair(resourceName, "stage.2.action.0.role_arn", "aws_iam_role.codepipeline_action_role", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodePipeline_tags(t *testing.T) {
	var p1, p2, p3 codepipeline.PipelineDeclaration
	name := acctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodePipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineConfigWithTags(name, "tag1value", "tag2value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", fmt.Sprintf("test-pipeline-%s", name)),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodePipelineConfigWithTags(name, "tag1valueUpdate", "tag2valueUpdate"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", fmt.Sprintf("test-pipeline-%s", name)),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1valueUpdate"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2valueUpdate"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodePipelineConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSCodePipelineHashArtifactStore(region string) int {
	return hashcode.String(region)
}

func testAccCheckAWSCodePipelineHashArtifactStoreKey(region, key string) string {
	return fmt.Sprintf("artifact_stores.%d.%s", testAccCheckAWSCodePipelineHashArtifactStore(region), key)
}

func TestAccAWSCodePipeline_multiregion_basic(t *testing.T) {
	var p codepipeline.PipelineDeclaration
	resourceName := "aws_codepipeline.test"
	var providers []*schema.Provider

	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionsPreCheck(t)
			testAccAlternateRegionPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineConfig_multiregion(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "artifact_stores.#", "2"),

					resource.TestCheckResourceAttr(resourceName, testAccCheckAWSCodePipelineHashArtifactStoreKey(testAccGetRegion(), "type"), "S3"),
					resource.TestCheckResourceAttrPair(resourceName, testAccCheckAWSCodePipelineHashArtifactStoreKey(testAccGetRegion(), "location"), "aws_s3_bucket.local", "bucket"),
					resource.TestCheckResourceAttr(resourceName, testAccCheckAWSCodePipelineHashArtifactStoreKey(testAccGetRegion(), "region"), testAccGetRegion()),

					resource.TestCheckResourceAttr(resourceName, testAccCheckAWSCodePipelineHashArtifactStoreKey(testAccGetAlternateRegion(), "type"), "S3"),
					resource.TestCheckResourceAttrPair(resourceName, testAccCheckAWSCodePipelineHashArtifactStoreKey(testAccGetAlternateRegion(), "location"), "aws_s3_bucket.alternate", "bucket"),
					resource.TestCheckResourceAttr(resourceName, testAccCheckAWSCodePipelineHashArtifactStoreKey(testAccGetAlternateRegion(), "region"), testAccGetAlternateRegion()),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", fmt.Sprintf("%s-Build", testAccGetRegion())),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", testAccGetRegion()),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.name", fmt.Sprintf("%s-Build", testAccGetAlternateRegion())),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.region", testAccGetAlternateRegion()),
				),
			},
			{
				Config:            testAccAWSCodePipelineConfig_multiregion(name),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSCodePipelineExists(n string, pipeline *codepipeline.PipelineDeclaration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CodePipeline ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).codepipelineconn

		out, err := conn.GetPipeline(&codepipeline.GetPipelineInput{
			Name: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*pipeline = *out.Pipeline

		return nil
	}
}

func testAccCheckAWSCodePipelineDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codepipelineconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codepipeline" {
			continue
		}

		_, err := conn.GetPipeline(&codepipeline.GetPipelineInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err == nil {
			return fmt.Errorf("Expected AWS CodePipeline to be gone, but was still found")
		}
		if isAWSErr(err, "PipelineNotFoundException", "") {
			return nil
		}
		return err
	}

	return nil
}

func testAccPreCheckAWSCodePipeline(t *testing.T) {
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	conn := testAccProvider.Meta().(*AWSClient).codepipelineconn

	input := &codepipeline.ListPipelinesInput{}

	_, err := conn.ListPipelines(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSCodePipelineS3Bucket(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "foo" {
  bucket = "tf-test-pipeline-%s"
  acl    = "private"
}
`, rName)
}

func testAccAWSCodePipelineServiceIAMRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "codepipeline_role" {
  name = "codepipeline-role-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "codepipeline.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "codepipeline_policy" {
  name = "codepipeline_policy"
  role = "${aws_iam_role.codepipeline_role.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect":"Allow",
      "Action": [
        "s3:GetObject",
        "s3:GetObjectVersion",
        "s3:GetBucketVersioning"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "codebuild:BatchGetBuilds",
        "codebuild:StartBuild"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}
`, rName)
}

func testAccAWSCodePipelineServiceIAMRoleWithAssumeRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "codepipeline_role" {
  name = "codepipeline-role-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "codepipeline.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "codepipeline_policy" {
  name = "codepipeline_policy"
  role = "${aws_iam_role.codepipeline_role.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect":"Allow",
      "Action": [
        "s3:GetObject",
        "s3:GetObjectVersion",
        "s3:GetBucketVersioning"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "codebuild:BatchGetBuilds",
        "codebuild:StartBuild"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "sts:AssumeRole"
      ],
      "Resource": "${aws_iam_role.codepipeline_action_role.arn}"
    }
  ]
}
EOF
}
`, rName)
}

func testAccAWSCodePipelineConfig_basic(rName string) string {
	return composeConfig(
		testAccAWSCodePipelineS3Bucket(rName),
		testAccAWSCodePipelineServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%s"
  role_arn = "${aws_iam_role.codepipeline_role.arn}"

  artifact_store {
    location = "${aws_s3_bucket.test.bucket}"
    type     = "S3"

    encryption_key {
      id   = "1234"
      type = "KMS"
    }
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "ThirdParty"
      provider         = "GitHub"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        Owner  = "lifesum-terraform"
        Repo   = "test"
        Branch = "master"
      }
    }
  }

  stage {
    name = "Build"

    action {
      name            = "Build"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["test"]
      version         = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }
}
`, rName))
}

func testAccAWSCodePipelineConfig_basicUpdated(rName string) string {
	return composeConfig(
		testAccAWSCodePipelineS3Bucket(rName),
		testAccAWSCodePipelineServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%s"
  role_arn = "${aws_iam_role.codepipeline_role.arn}"

  artifact_store {
    location = "${aws_s3_bucket.updated.bucket}"
    type     = "S3"

    encryption_key {
      id   = "4567"
      type = "KMS"
    }
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "ThirdParty"
      provider         = "GitHub"
      version          = "1"
      output_artifacts = ["artifacts"]

      configuration = {
        Owner  = "test-terraform"
        Repo   = "test-repo"
        Branch = "stable"
      }
    }
  }

  stage {
    name = "Build"

    action {
      name            = "Build"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["artifacts"]
      version         = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }
}
`, rName))
}

func testAccAWSCodePipelineConfig_emptyArtifacts(rName string) string {
	return composeConfig(
		testAccAWSCodePipelineS3Bucket(rName),
		testAccAWSCodePipelineServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%s"
  role_arn = "${aws_iam_role.codepipeline_role.arn}"

  artifact_store {
    location = "${aws_s3_bucket.test.bucket}"
    type     = "S3"
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "ThirdParty"
      provider         = "GitHub"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        Owner  = "lifesum-terraform"
        Repo   = "test"
        Branch = "master"
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
      input_artifacts  = ["test", ""]
      output_artifacts = [""]
      version          = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }
}
`, rName))
}

func testAccAWSCodePipelineDeployActionIAMRole(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_iam_role" "codepipeline_action_role" {
  name = "codepipeline-action-role-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "codepipeline_action_policy" {
  name = "codepipeline_action_policy"
  role = "${aws_iam_role.codepipeline_action_role.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect":"Allow",
      "Action": [
        "s3:GetObject",
        "s3:GetObjectVersion",
        "s3:GetBucketVersioning"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccAWSCodePipelineConfig_deployWithServiceRole(rName string) string {
	return composeConfig(
		testAccAWSCodePipelineS3Bucket(rName),
		testAccAWSCodePipelineServiceIAMRoleWithAssumeRole(rName),
		testAccAWSCodePipelineDeployActionIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%s"
  role_arn = "${aws_iam_role.codepipeline_role.arn}"

  artifact_store {
    location = "${aws_s3_bucket.test.bucket}"
    type     = "S3"

    encryption_key {
      id   = "4567"
      type = "KMS"
    }
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "ThirdParty"
      provider         = "GitHub"
      version          = "1"
      output_artifacts = ["artifacts"]

      configuration = {
        Owner  = "test-terraform"
        Repo   = "test-repo"
        Branch = "stable"
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
      input_artifacts  = ["artifacts"]
      output_artifacts = ["artifacts2"]
      version          = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }

  stage {
    name = "Deploy"

    action {
      name            = "CreateChangeSet"
      category        = "Deploy"
      owner           = "AWS"
      provider        = "CloudFormation"
      input_artifacts = ["artifacts2"]
      role_arn        = "${aws_iam_role.codepipeline_action_role.arn}"
      version         = "1"

      configuration = {
        ActionMode    = "CHANGE_SET_REPLACE"
        ChangeSetName = "changeset"
        StackName     = "stack"
        TemplatePath  = "artifacts2::template.yaml"
      }
    }
  }
}
`, rName))
}

func testAccAWSCodePipelineConfigWithTags(rName, tag1, tag2 string) string {
	return composeConfig(
		testAccAWSCodePipelineS3Bucket(rName),
		testAccAWSCodePipelineServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = "${aws_iam_role.codepipeline_role.arn}"

  artifact_store {
    location = "${aws_s3_bucket.test.bucket}"
    type     = "S3"

    encryption_key {
      id   = "1234"
      type = "KMS"
    }
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "ThirdParty"
      provider         = "GitHub"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        Owner  = "lifesum-terraform"
        Repo   = "test"
        Branch = "master"
      }
    }
  }

  stage {
    name = "Build"

    action {
      name            = "Build"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["test"]
      version         = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }

  tags = {
    Name = "test-pipeline-%[1]s"
    tag1 = %[2]q
    tag2 = %[3]q
  }
}
`, rName, tag1, tag2))
}

func testAccAWSCodePipelineConfig_multiregion(rName string) string {
	return composeConfig(
		testAccAlternateRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_s3_bucket" "local" {
  bucket = "tf-test-pipeline-local-%[1]s"
  acl    = "private"
}

resource "aws_s3_bucket" "alternate" {
  bucket = "tf-test-pipeline-alternate-%[1]s"
  acl    = "private"
  provider = "aws.alternate"
}

resource "aws_iam_role" "codepipeline_role" {
  name = "codepipeline-role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "codepipeline.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "codepipeline_policy" {
  name = "codepipeline_policy"
  role = "${aws_iam_role.codepipeline_role.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect":"Allow",
      "Action": [
        "s3:GetObject",
        "s3:GetObjectVersion",
        "s3:GetBucketVersioning"
      ],
      "Resource": [
        "${aws_s3_bucket.local.arn}",
        "${aws_s3_bucket.local.arn}/*",
        "${aws_s3_bucket.alternate.arn}",
        "${aws_s3_bucket.alternate.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "codebuild:BatchGetBuilds",
        "codebuild:StartBuild"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = "${aws_iam_role.codepipeline_role.arn}"

  artifact_stores {
			location = "${aws_s3_bucket.local.bucket}"
			type     = "S3"    
      region   = "%[2]s"
	}
  artifact_stores {
			location = "${aws_s3_bucket.alternate.bucket}"
			type     = "S3"  
      region   = "%[3]s"
	}

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "ThirdParty"
      provider         = "GitHub"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        Owner  = "lifesum-terraform"
        Repo   = "test"
        Branch = "master"
      }
    }
  }

  stage {
    name = "Build"

    action {
		  region          = "%[2]s"
      name            = "%[2]s-Build"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["test"]
      version         = "1"

      configuration = {
        ProjectName = "%[2]s-Test"
      }
    }

    action {
		  region          = "%[3]s"
      name            = "%[3]s-Build"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["test"]
      version         = "1"

      configuration = {
        ProjectName = "%[3]s-Test"
      }
    }
  }
}
`, rName, testAccGetRegion(), testAccGetAlternateRegion()))
}
