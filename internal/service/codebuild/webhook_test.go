// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package codebuild_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcodebuild "github.com/hashicorp/terraform-provider-aws/internal/service/codebuild"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeBuildWebhook_bitbucket(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	sourceLocation := testAccBitbucketSourceLocationFromEnv()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeBitbucket)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_bitbucket(rName, sourceLocation),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_filter", ""),
					resource.TestCheckResourceAttr(resourceName, "manual_creation", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "project_name", rName),
					resource.TestMatchResourceAttr(resourceName, "payload_url", regexache.MustCompile(`^https://`)),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.#", "0"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_gitHub(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_filter", ""),
					resource.TestCheckResourceAttr(resourceName, "manual_creation", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "project_name", rName),
					resource.TestMatchResourceAttr(resourceName, "payload_url", regexache.MustCompile(`^https://`)),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "secret", ""),
					resource.TestMatchResourceAttr(resourceName, names.AttrURL, regexache.MustCompile(`^https://`)),
					resource.TestCheckResourceAttr(resourceName, "pull_request_build_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pull_request_build_policy.0.requires_comment_approval", string(types.PullRequestBuildCommentApprovalAllPullRequests)),
					resource.TestCheckResourceAttr(resourceName, "pull_request_build_policy.0.approver_roles.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "pull_request_build_policy.0.approver_roles.*", string(types.PullRequestBuildApproverRoleGithubWrite)),
					resource.TestCheckTypeSetElemAttr(resourceName, "pull_request_build_policy.0.approver_roles.*", string(types.PullRequestBuildApproverRoleGithubMaintain)),
					resource.TestCheckTypeSetElemAttr(resourceName, "pull_request_build_policy.0.approver_roles.*", string(types.PullRequestBuildApproverRoleGithubAdmin)),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithubEnterprise)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_gitHubEnterprise(rName, "dev"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_filter", "dev"),
					resource.TestCheckResourceAttr(resourceName, "manual_creation", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "project_name", rName),
					resource.TestMatchResourceAttr(resourceName, "payload_url", regexache.MustCompile(`^https://`)),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.#", "0"),
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
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_filter", "master"),
					resource.TestCheckResourceAttr(resourceName, "project_name", rName),
					resource.TestMatchResourceAttr(resourceName, "payload_url", regexache.MustCompile(`^https://`)),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.#", "0"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_buildType(rName, "BUILD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "build_type", "BUILD"),
				),
			},
			{
				Config: testAccWebhookConfig_gitHub(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccWebhookConfig_buildType(rName, "BUILD_BATCH"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_scopeConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.#", "1"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_branchFilter(rName, "master"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_filter", "master"),
				),
			},
			{
				Config: testAccWebhookConfig_branchFilter(rName, "dev"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_filterGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_gitHub(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcodebuild.ResourceWebhook(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodeBuildWebhook_Disappears_project(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"
	projectResourceName := "aws_codebuild_project.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_gitHub(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcodebuild.ResourceProject(), projectResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodeBuildWebhook_manualCreation(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_manualCreation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "manual_creation", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"manual_creation", "secret"},
			},
		},
	})
}

func TestAccCodeBuildWebhook_upgradeV5_94_1(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.CodeBuildServiceID),
		CheckDestroy: testAccCheckWebhookDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.94.1",
					},
				},
				Config: testAccWebhookConfig_gitHub(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccWebhookConfig_gitHub(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccCodeBuildWebhook_gitHubWithPullRequestBuildPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_gitHubWithPullRequestBuildPolicy(
					rName,
					string(types.PullRequestBuildCommentApprovalAllPullRequests),
					enum.Slice(
						types.PullRequestBuildApproverRoleGithubRead,
						types.PullRequestBuildApproverRoleGithubWrite,
						types.PullRequestBuildApproverRoleGithubMaintain,
						types.PullRequestBuildApproverRoleGithubAdmin,
					),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "pull_request_build_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pull_request_build_policy.0.requires_comment_approval", string(types.PullRequestBuildCommentApprovalAllPullRequests)),
					resource.TestCheckResourceAttr(resourceName, "pull_request_build_policy.0.approver_roles.#", "4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "pull_request_build_policy.0.approver_roles.*", string(types.PullRequestBuildApproverRoleGithubRead)),
					resource.TestCheckTypeSetElemAttr(resourceName, "pull_request_build_policy.0.approver_roles.*", string(types.PullRequestBuildApproverRoleGithubWrite)),
					resource.TestCheckTypeSetElemAttr(resourceName, "pull_request_build_policy.0.approver_roles.*", string(types.PullRequestBuildApproverRoleGithubMaintain)),
					resource.TestCheckTypeSetElemAttr(resourceName, "pull_request_build_policy.0.approver_roles.*", string(types.PullRequestBuildApproverRoleGithubAdmin)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
			{
				Config: testAccWebhookConfig_gitHubWithPullRequestBuildPolicy(
					rName,
					string(types.PullRequestBuildCommentApprovalForkPullRequests),
					enum.Slice(
						types.PullRequestBuildApproverRoleGithubMaintain,
						types.PullRequestBuildApproverRoleGithubAdmin,
					),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "pull_request_build_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pull_request_build_policy.0.requires_comment_approval", string(types.PullRequestBuildCommentApprovalForkPullRequests)),
					resource.TestCheckResourceAttr(resourceName, "pull_request_build_policy.0.approver_roles.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "pull_request_build_policy.0.approver_roles.*", string(types.PullRequestBuildApproverRoleGithubMaintain)),
					resource.TestCheckTypeSetElemAttr(resourceName, "pull_request_build_policy.0.approver_roles.*", string(types.PullRequestBuildApproverRoleGithubAdmin)),
				),
			},
			{
				Config: testAccWebhookConfig_gitHubWithPullRequestBuildPolicyNoApproverRoles(
					rName,
					string(types.PullRequestBuildCommentApprovalDisabled),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "pull_request_build_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pull_request_build_policy.0.requires_comment_approval", string(types.PullRequestBuildCommentApprovalDisabled)),
				),
			},
		},
	})
}

func TestAccCodeBuildWebhook_gitHubWithPullRequestBuildPolicyNoApproverRoles(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codebuild_webhook.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_gitHubWithPullRequestBuildPolicyNoApproverRoles(
					rName,
					string(types.PullRequestBuildCommentApprovalAllPullRequests),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(ctx, t, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "pull_request_build_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pull_request_build_policy.0.requires_comment_approval", string(types.PullRequestBuildCommentApprovalAllPullRequests)),
					resource.TestCheckResourceAttr(resourceName, "pull_request_build_policy.0.approver_roles.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "pull_request_build_policy.0.approver_roles.*", string(types.PullRequestBuildApproverRoleGithubWrite)),
					resource.TestCheckTypeSetElemAttr(resourceName, "pull_request_build_policy.0.approver_roles.*", string(types.PullRequestBuildApproverRoleGithubMaintain)),
					resource.TestCheckTypeSetElemAttr(resourceName, "pull_request_build_policy.0.approver_roles.*", string(types.PullRequestBuildApproverRoleGithubAdmin)),
				),
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

func testAccCheckWebhookDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CodeBuildClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codebuild_webhook" {
				continue
			}

			_, err := tfcodebuild.FindWebhookByProjectName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckWebhookExists(ctx context.Context, t *testing.T, n string, v *types.Webhook) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CodeBuildClient(ctx)

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
	return acctest.ConfigCompose(testAccProjectConfig_basicGitHub(rName), `
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
	return acctest.ConfigCompose(testAccProjectConfig_basicGitHub(rName), fmt.Sprintf(`
resource "aws_codebuild_webhook" "test" {
  build_type   = %[1]q
  project_name = aws_codebuild_project.test.name
}
`, branchFilter))
}

func testAccWebhookConfig_branchFilter(rName, branchFilter string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basicGitHub(rName), fmt.Sprintf(`
resource "aws_codebuild_webhook" "test" {
  branch_filter = %[1]q
  project_name  = aws_codebuild_project.test.name
}
`, branchFilter))
}

func testAccWebhookConfig_filterGroup(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basicGitHub(rName), `
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

func testAccWebhookConfig_manualCreation(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basicGitHub(rName), `
resource "aws_codebuild_webhook" "test" {
  project_name    = aws_codebuild_project.test.name
  manual_creation = true
}
`)
}

func testAccWebhookConfig_gitHubWithPullRequestBuildPolicy(rName, requiresCommentApproval string, approverRoles []string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basicGitHub(rName), fmt.Sprintf(`
resource "aws_codebuild_webhook" "test" {
  project_name = aws_codebuild_project.test.name
  pull_request_build_policy {
    requires_comment_approval = %[1]q
    approver_roles            = ["%[2]s"]
  }
}
`, requiresCommentApproval, strings.Join(approverRoles, "\", \"")))
}

func testAccWebhookConfig_gitHubWithPullRequestBuildPolicyNoApproverRoles(rName, requiresCommentApproval string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basicGitHub(rName), fmt.Sprintf(`
resource "aws_codebuild_webhook" "test" {
  project_name = aws_codebuild_project.test.name
  pull_request_build_policy {
    requires_comment_approval = %[1]q
  }
}
`, requiresCommentApproval))
}
