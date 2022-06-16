package codepipeline_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodepipeline "github.com/hashicorp/terraform-provider-aws/internal/service/codepipeline"
)

const envVarGithubTokenUsageWebhook = "token with GitHub permissions to repository for CodePipeline webhook creation"

func TestAccCodePipelineWebhook_basic(t *testing.T) {
	githubToken := conns.SkipIfEnvVarEmpty(t, conns.EnvVarGithubToken, envVarGithubTokenUsageWebhook)

	var v codepipeline.ListWebhookItem
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSupported(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_basic(rName, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "codepipeline", regexp.MustCompile(fmt.Sprintf("webhook:%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "authentication", "GITHUB_HMAC"),
					resource.TestCheckResourceAttr(resourceName, "target_action", "Source"),
					resource.TestCheckResourceAttrPair(resourceName, "target_pipeline", "aws_codepipeline.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter.*", map[string]string{
						"json_path":    "$.ref",
						"match_equals": "refs/head/{Branch}",
					}),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.0.secret_token", "super-secret"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWebhookConfig_filters(rName, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter.*", map[string]string{
						"json_path":    "$.ref",
						"match_equals": "refs/head/{Branch}",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter.*", map[string]string{
						"json_path":    "$.head_commit.modified",
						"match_equals": "^.*mypath.*$",
					}),
				),
			},
			{
				Config: testAccWebhookConfig_basic(rName, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter.*", map[string]string{
						"json_path":    "$.ref",
						"match_equals": "refs/head/{Branch}",
					}),
				),
			},
		},
	})
}

func TestAccCodePipelineWebhook_ipAuth(t *testing.T) {
	githubToken := conns.SkipIfEnvVarEmpty(t, conns.EnvVarGithubToken, envVarGithubTokenUsageWebhook)

	var v codepipeline.ListWebhookItem
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSupported(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_ipAuth(rName, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.0.allowed_ip_range", "0.0.0.0/0"),
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

func TestAccCodePipelineWebhook_unauthenticated(t *testing.T) {
	githubToken := conns.SkipIfEnvVarEmpty(t, conns.EnvVarGithubToken, envVarGithubTokenUsageWebhook)

	var v codepipeline.ListWebhookItem
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSupported(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_unauthenticated(rName, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
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

func TestAccCodePipelineWebhook_tags(t *testing.T) {
	githubToken := conns.SkipIfEnvVarEmpty(t, conns.EnvVarGithubToken, envVarGithubTokenUsageWebhook)

	var v1, v2, v3 codepipeline.ListWebhookItem
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSupported(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_tags(rName, "tag1value", "tag2value", githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2value"),
				),
			},
			{
				Config: testAccWebhookConfig_tags(rName, "tag1valueUpdate", "tag2valueUpdate", githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1valueUpdate"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2valueUpdate"),
					func(s *terraform.State) error {
						if aws.StringValue(v2.Url) != aws.StringValue(v1.Url) {
							return fmt.Errorf("Codepipeline webhook recreated when changing tags")
						}
						return nil
					},
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWebhookConfig_basic(rName, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					func(s *terraform.State) error {
						if aws.StringValue(v3.Url) != aws.StringValue(v2.Url) {
							return fmt.Errorf("Codepipeline webhook recreated when deleting tags")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccCodePipelineWebhook_disappears(t *testing.T) {
	githubToken := conns.SkipIfEnvVarEmpty(t, conns.EnvVarGithubToken, envVarGithubTokenUsageWebhook)

	var v codepipeline.ListWebhookItem
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSupported(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_basic(rName, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodepipeline.ResourceWebhook(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodepipeline.ResourceWebhook(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodePipelineWebhook_UpdateAuthentication_secretToken(t *testing.T) {
	githubToken := conns.SkipIfEnvVarEmpty(t, conns.EnvVarGithubToken, envVarGithubTokenUsageWebhook)

	var v1, v2 codepipeline.ListWebhookItem
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSupported(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_basic(rName, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(resourceName, &v1),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.0.secret_token", "super-secret"),
				),
			},
			{
				Config: testAccWebhookConfig_secretTokenUpdated(rName, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(resourceName, &v2),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.0.secret_token", "even-more-secret"),
					func(s *terraform.State) error {
						if aws.StringValue(v2.Url) == aws.StringValue(v1.Url) {
							return fmt.Errorf("Codepipeline webhook not recreated when updating authentication_configuration.secret_token")
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccCheckWebhookExists(n string, webhook *codepipeline.ListWebhookItem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No webhook ARN is set as ID")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodePipelineConn

		resp, err := tfcodepipeline.GetWebhook(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*webhook = *resp

		return nil
	}
}

func testAccWebhookConfig_basic(rName, githubToken string) string {
	return testAccWebhookConfig_base(rName, githubToken) + fmt.Sprintf(`
resource "aws_codepipeline_webhook" "test" {
  name            = %[1]q
  authentication  = "GITHUB_HMAC"
  target_action   = "Source"
  target_pipeline = aws_codepipeline.test.name

  authentication_configuration {
    secret_token = "super-secret"
  }

  filter {
    json_path    = "$.ref"
    match_equals = "refs/head/{Branch}"
  }
}
`, rName)
}

func testAccWebhookConfig_filters(rName, githubToken string) string {
	return testAccWebhookConfig_base(rName, githubToken) + fmt.Sprintf(`
resource "aws_codepipeline_webhook" "test" {
  name            = %[1]q
  authentication  = "GITHUB_HMAC"
  target_action   = "Source"
  target_pipeline = aws_codepipeline.test.name

  authentication_configuration {
    secret_token = "super-secret"
  }

  filter {
    json_path    = "$.ref"
    match_equals = "refs/head/{Branch}"
  }

  filter {
    json_path    = "$.head_commit.modified"
    match_equals = "^.*mypath.*$"
  }
}
`, rName)
}

func testAccWebhookConfig_ipAuth(rName, githubToken string) string {
	return testAccWebhookConfig_base(rName, githubToken) + fmt.Sprintf(`
resource "aws_codepipeline_webhook" "test" {
  name            = %[1]q
  authentication  = "IP"
  target_action   = "Source"
  target_pipeline = aws_codepipeline.test.name

  authentication_configuration {
    allowed_ip_range = "0.0.0.0/0"
  }

  filter {
    json_path    = "$.ref"
    match_equals = "refs/head/{Branch}"
  }
}
`, rName)
}

func testAccWebhookConfig_unauthenticated(rName, githubToken string) string {
	return testAccWebhookConfig_base(rName, githubToken) + fmt.Sprintf(`
resource "aws_codepipeline_webhook" "test" {
  name            = %[1]q
  authentication  = "UNAUTHENTICATED"
  target_action   = "Source"
  target_pipeline = aws_codepipeline.test.name

  filter {
    json_path    = "$.ref"
    match_equals = "refs/head/{Branch}"
  }
}
`, rName)
}

func testAccWebhookConfig_tags(rName, tag1, tag2, githubToken string) string {
	return testAccWebhookConfig_base(rName, githubToken) + fmt.Sprintf(`
resource "aws_codepipeline_webhook" "test" {
  name            = %[1]q
  authentication  = "GITHUB_HMAC"
  target_action   = "Source"
  target_pipeline = aws_codepipeline.test.name

  authentication_configuration {
    secret_token = "super-secret"
  }

  filter {
    json_path    = "$.ref"
    match_equals = "refs/head/{Branch}"
  }

  tags = {
    Name = %[1]q
    tag1 = %[2]q
    tag2 = %[3]q
  }
}
`, rName, tag1, tag2)
}

func testAccWebhookConfig_secretTokenUpdated(rName, githubToken string) string {
	return testAccWebhookConfig_base(rName, githubToken) + fmt.Sprintf(`
resource "aws_codepipeline_webhook" "test" {
  name            = %[1]q
  authentication  = "GITHUB_HMAC"
  target_action   = "Source"
  target_pipeline = aws_codepipeline.test.name

  authentication_configuration {
    secret_token = "even-more-secret"
  }

  filter {
    json_path    = "$.ref"
    match_equals = "refs/head/{Branch}"
  }
}
`, rName)
}

func testAccWebhookConfig_base(rName, githubToken string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_iam_role" "test" {
  name = %[1]q

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

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
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

resource "aws_codepipeline" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  artifact_store {
    location = aws_s3_bucket.test.bucket
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
        Owner      = "lifesum-terraform"
        Repo       = "test"
        Branch     = "master"
        OAuthToken = %[2]q
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
`, rName, githubToken)
}
