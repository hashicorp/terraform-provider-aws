// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codepipeline_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/envvar"
	tfcodepipeline "github.com/hashicorp/terraform-provider-aws/internal/service/codepipeline"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodePipelineWebhook_basic(t *testing.T) {
	ctx := acctest.Context(t)
	ghToken := acctest.SkipIfEnvVarNotSet(t, envvar.GithubToken)
	var v types.ListWebhookItem
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_basic(rName, ghToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "codepipeline", regexache.MustCompile(fmt.Sprintf("webhook:%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "authentication", "GITHUB_HMAC"),
					resource.TestCheckResourceAttr(resourceName, "target_action", "Source"),
					resource.TestCheckResourceAttrPair(resourceName, "target_pipeline", "aws_codepipeline.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter.*", map[string]string{
						"json_path":    "$.ref",
						"match_equals": "refs/head/{Branch}",
					}),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrURL),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.0.secret_token", "super-secret"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWebhookConfig_filters(rName, ghToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct2),
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
				Config: testAccWebhookConfig_basic(rName, ghToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	ghToken := acctest.SkipIfEnvVarNotSet(t, envvar.GithubToken)
	var v types.ListWebhookItem
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_ipAuth(rName, ghToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrURL),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	ghToken := acctest.SkipIfEnvVarNotSet(t, envvar.GithubToken)
	var v types.ListWebhookItem
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_unauthenticated(rName, ghToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrURL),
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
	ctx := acctest.Context(t)
	ghToken := acctest.SkipIfEnvVarNotSet(t, envvar.GithubToken)
	var v types.ListWebhookItem
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_tags1(rName, ghToken, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccWebhookConfig_tags2(rName, ghToken, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWebhookConfig_tags1(rName, ghToken, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccCodePipelineWebhook_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	ghToken := acctest.SkipIfEnvVarNotSet(t, envvar.GithubToken)
	var v types.ListWebhookItem
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_basic(rName, ghToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodepipeline.ResourceWebhook(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodePipelineWebhook_UpdateAuthentication_secretToken(t *testing.T) {
	ctx := acctest.Context(t)
	ghToken := acctest.SkipIfEnvVarNotSet(t, envvar.GithubToken)
	var v1, v2 types.ListWebhookItem
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_basic(rName, ghToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrURL),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.0.secret_token", "super-secret"),
				),
			},
			{
				Config: testAccWebhookConfig_secretTokenUpdated(rName, ghToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrURL),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.0.secret_token", "even-more-secret"),
					func(s *terraform.State) error {
						if aws.ToString(v2.Url) == aws.ToString(v1.Url) {
							return fmt.Errorf("Codepipeline webhook not recreated when updating authentication_configuration.secret_token")
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccCheckWebhookExists(ctx context.Context, n string, v *types.ListWebhookItem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodePipelineClient(ctx)

		output, err := tfcodepipeline.FindWebhookByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckWebhookDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodePipelineClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codepipeline_webhook" {
				continue
			}

			_, err := tfcodepipeline.FindWebhookByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CodePipeline Webhook %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccWebhookConfig_base(rName, githubToken string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
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

func testAccWebhookConfig_basic(rName, githubToken string) string {
	return acctest.ConfigCompose(testAccWebhookConfig_base(rName, githubToken), fmt.Sprintf(`
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
`, rName))
}

func testAccWebhookConfig_filters(rName, githubToken string) string {
	return acctest.ConfigCompose(testAccWebhookConfig_base(rName, githubToken), fmt.Sprintf(`
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
`, rName))
}

func testAccWebhookConfig_ipAuth(rName, githubToken string) string {
	return acctest.ConfigCompose(testAccWebhookConfig_base(rName, githubToken), fmt.Sprintf(`
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
`, rName))
}

func testAccWebhookConfig_unauthenticated(rName, githubToken string) string {
	return acctest.ConfigCompose(testAccWebhookConfig_base(rName, githubToken), fmt.Sprintf(`
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
`, rName))
}

func testAccWebhookConfig_tags1(rName, githubToken, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccWebhookConfig_base(rName, githubToken), fmt.Sprintf(`
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
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccWebhookConfig_tags2(rName, githubToken, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccWebhookConfig_base(rName, githubToken), fmt.Sprintf(`
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
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccWebhookConfig_secretTokenUpdated(rName, githubToken string) string {
	return acctest.ConfigCompose(testAccWebhookConfig_base(rName, githubToken), fmt.Sprintf(`
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
`, rName))
}
