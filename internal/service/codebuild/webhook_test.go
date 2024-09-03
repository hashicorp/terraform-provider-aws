// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodebuild "github.com/hashicorp/terraform-provider-aws/internal/service/codebuild"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeBuildWebhook_bitbucket(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	sourceLocation := testAccBitbucketSourceLocationFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeBitbucket)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_bitbucket(rName, sourceLocation),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_filter", ""),
					resource.TestCheckResourceAttr(resourceName, "project_name", rName),
					resource.TestMatchResourceAttr(resourceName, "payload_url", regexache.MustCompile(`^https://`)),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secret", ""),
					resource.TestMatchResourceAttr(resourceName, names.AttrURL, regexache.MustCompile(`^https://`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

func TestAccCodeBuildWebhook_gitHub(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_gitHub(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_filter", ""),
					resource.TestCheckResourceAttr(resourceName, "project_name", rName),
					resource.TestMatchResourceAttr(resourceName, "payload_url", regexache.MustCompile(`^https://`)),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secret", ""),
					resource.TestMatchResourceAttr(resourceName, names.AttrURL, regexache.MustCompile(`^https://`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

func TestAccCodeBuildWebhook_gitHubEnterprise(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithubEnterprise)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_gitHubEnterprise(rName, "dev"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_filter", "dev"),
					resource.TestCheckResourceAttr(resourceName, "project_name", rName),
					resource.TestMatchResourceAttr(resourceName, "payload_url", regexache.MustCompile(`^https://`)),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.#", acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, "secret", regexache.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrURL, ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
			{
				Config: testAccWebhookConfig_gitHubEnterprise(rName, "master"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_filter", "master"),
					resource.TestCheckResourceAttr(resourceName, "project_name", rName),
					resource.TestMatchResourceAttr(resourceName, "payload_url", regexache.MustCompile(`^https://`)),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.#", acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, "secret", regexache.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrURL, ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

func TestAccCodeBuildWebhook_buildType(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_buildType(rName, "BUILD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "build_type", "BUILD"),
				),
			},
			{
				Config: testAccWebhookConfig_gitHub(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &webhook),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccWebhookConfig_buildType(rName, "BUILD_BATCH"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "build_type", "BUILD_BATCH"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

func TestAccCodeBuildWebhook_scopeConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_scopeConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.0.scope", "GITHUB_GLOBAL"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

func TestAccCodeBuildWebhook_branchFilter(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_branchFilter(rName, "master"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_filter", "master"),
				),
			},
			{
				Config: testAccWebhookConfig_branchFilter(rName, "dev"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_filter", "dev"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

func TestAccCodeBuildWebhook_filterGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_filterGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &webhook),
					testAccCheckWebhookFilter(&webhook, [][]types.WebhookFilter{
						{
							{
								Type:                  types.WebhookFilterTypeEvent,
								Pattern:               aws.String("PUSH"),
								ExcludeMatchedPattern: aws.Bool(false),
							},
							{
								Type:                  types.WebhookFilterTypeHeadRef,
								Pattern:               aws.String("refs/heads/master"),
								ExcludeMatchedPattern: aws.Bool(true),
							},
						},
						{
							{
								Type:                  types.WebhookFilterTypeEvent,
								Pattern:               aws.String("PULL_REQUEST_UPDATED"),
								ExcludeMatchedPattern: aws.Bool(false),
							},
						},
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

func TestAccCodeBuildWebhook_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_gitHub(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &webhook),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodebuild.ResourceWebhook(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodeBuildWebhook_Disappears_project(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"
	projectResourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_gitHub(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &webhook),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodebuild.ResourceProject(), projectResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckWebhookFilter(webhook *types.Webhook, expectedFilters [][]types.WebhookFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		got, want := webhook.FilterGroups, expectedFilters
		if diff := cmp.Diff(got, want, cmpopts.IgnoreUnexported(types.WebhookFilter{})); diff != "" {
			return fmt.Errorf("unexpected WebhookFilter diff (+wanted, -got): %s", diff)
		}

		return nil
	}
}

func testAccCheckWebhookDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codebuild_webhook" {
				continue
			}

			_, err := tfcodebuild.FindWebhookByProjectName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CodeBuild Webhook %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckWebhookExists(ctx context.Context, n string, v *types.Webhook) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildClient(ctx)

		output, err := tfcodebuild.FindWebhookByProjectName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccWebhookConfig_bitbucket(rName, sourceLocation string) string {
	return acctest.ConfigCompose(testAccProjectConfig_sourceTypeBitbucket(rName, sourceLocation), `
resource "aws_codebuild_webhook" "test" {
  project_name = aws_codebuild_project.test.name
}
`)
}

func testAccWebhookConfig_gitHub(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basic(rName), `
resource "aws_codebuild_webhook" "test" {
  project_name = aws_codebuild_project.test.name
}
`)
}

func testAccWebhookConfig_gitHubEnterprise(rName string, branchFilter string) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://example.com/terraform/acceptance-testing.git"
    type     = "GITHUB_ENTERPRISE"
  }
}

resource "aws_codebuild_webhook" "test" {
  project_name  = aws_codebuild_project.test.name
  branch_filter = %[2]q
}
`, rName, branchFilter))
}

func testAccWebhookConfig_buildType(rName, branchFilter string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basic(rName), fmt.Sprintf(`
resource "aws_codebuild_webhook" "test" {
  build_type   = %[1]q
  project_name = aws_codebuild_project.test.name
}
`, branchFilter))
}

func testAccWebhookConfig_branchFilter(rName, branchFilter string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basic(rName), fmt.Sprintf(`
resource "aws_codebuild_webhook" "test" {
  branch_filter = %[1]q
  project_name  = aws_codebuild_project.test.name
}
`, branchFilter))
}

func testAccWebhookConfig_filterGroup(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basic(rName), `
resource "aws_codebuild_webhook" "test" {
  project_name = aws_codebuild_project.test.name

  filter_group {
    filter {
      type    = "EVENT"
      pattern = "PUSH"
    }

    filter {
      type                    = "HEAD_REF"
      pattern                 = "refs/heads/master"
      exclude_matched_pattern = true
    }
  }

  filter_group {
    filter {
      type    = "EVENT"
      pattern = "PULL_REQUEST_UPDATED"
    }
  }
}
`)
}

func testAccWebhookConfig_scopeConfiguration(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "CODEBUILD_DEFAULT_WEBHOOK_SOURCE_LOCATION"
    type     = "GITHUB"
  }
}

resource "aws_codebuild_webhook" "test" {
  project_name = aws_codebuild_project.test.name
  scope_configuration {
    name  = %[1]q
    scope = "GITHUB_GLOBAL"
  }
}
`, rName))
}
