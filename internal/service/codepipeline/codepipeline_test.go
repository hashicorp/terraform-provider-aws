// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codepipeline_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/envvar"
	tfcodepipeline "github.com/hashicorp/terraform-provider-aws/internal/service/codepipeline"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCodePipeline_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var p1, p2 codepipeline.PipelineDeclaration
	name := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"
	codestarConnectionResourceName := "aws_codestarconnections_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSupported(ctx, t)
			acctest.PreCheckPartitionHasService(t, codestarconnections.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p1),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.codepipeline_role", "arn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "codepipeline", regexp.MustCompile(fmt.Sprintf("test-pipeline-%s", name))),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "stage.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.category", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.provider", "CodeStarSourceConnection"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.input_artifacts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.FullRepositoryId", "lifesum-terraform/test"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.BranchName", "main"),
					resource.TestCheckResourceAttrPair(resourceName, "stage.0.action.0.configuration.ConnectionArn", codestarConnectionResourceName, "arn"),
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
				Config: testAccCodePipelineConfig_basicUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p2),

					resource.TestCheckResourceAttr(resourceName, "stage.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.input_artifacts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.0", "artifacts"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.FullRepositoryId", "test-terraform/test-repo"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.BranchName", "stable"),
					resource.TestCheckResourceAttrPair(resourceName, "stage.0.action.0.configuration.ConnectionArn", codestarConnectionResourceName, "arn"),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.0", "artifacts"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.ProjectName", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"stage.0.action.0.configuration.%",
					"stage.0.action.0.configuration.OAuthToken",
				},
			},
		},
	})
}

func TestAccCodePipeline_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var p codepipeline.PipelineDeclaration
	name := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSupported(ctx, t)
			acctest.PreCheckPartitionHasService(t, codestarconnections.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodepipeline.ResourcePipeline(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodePipeline_emptyStageArtifacts(t *testing.T) {
	ctx := acctest.Context(t)
	var p codepipeline.PipelineDeclaration
	name := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSupported(ctx, t)
			acctest.PreCheckPartitionHasService(t, codestarconnections.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_emptyStageArtifacts(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "codepipeline", regexp.MustCompile(fmt.Sprintf("test-pipeline-%s$", name))),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.category", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.provider", "CodeBuild"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.output_artifacts.#", "0"),
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

func TestAccCodePipeline_deployWithServiceRole(t *testing.T) {
	ctx := acctest.Context(t)
	var p codepipeline.PipelineDeclaration
	name := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSupported(ctx, t)
			acctest.PreCheckPartitionHasService(t, codestarconnections.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_deployServiceRole(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
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

func TestAccCodePipeline_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var p1, p2, p3 codepipeline.PipelineDeclaration
	name := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSupported(ctx, t)
			acctest.PreCheckPartitionHasService(t, codestarconnections.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_tags(name, "tag1value", "tag2value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p1),
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
				Config: testAccCodePipelineConfig_tags(name, "tag1valueUpdate", "tag2valueUpdate"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p2),
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
				Config: testAccCodePipelineConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccCodePipeline_MultiRegion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var p codepipeline.PipelineDeclaration
	resourceName := "aws_codepipeline.test"

	name := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckSupported(ctx, t, acctest.AlternateRegion())
			acctest.PreCheckPartitionHasService(t, codestarconnections.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_multiregion(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.name", fmt.Sprintf("%s-Build", acctest.AlternateRegion())),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.region", acctest.AlternateRegion()),
				),
			},
			{
				Config:            testAccCodePipelineConfig_multiregion(name),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCodePipeline_MultiRegion_update(t *testing.T) {
	ctx := acctest.Context(t)
	var p1, p2 codepipeline.PipelineDeclaration
	resourceName := "aws_codepipeline.test"

	name := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckSupported(ctx, t, acctest.AlternateRegion())
			acctest.PreCheckPartitionHasService(t, codestarconnections.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_multiregion(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p1),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.name", fmt.Sprintf("%s-Build", acctest.AlternateRegion())),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.region", acctest.AlternateRegion()),
				),
			},
			{
				Config: testAccCodePipelineConfig_multiregionUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p2),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "BuildUpdated"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.name", fmt.Sprintf("%s-BuildUpdated", acctest.AlternateRegion())),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.region", acctest.AlternateRegion()),
				),
			},
			{
				Config:            testAccCodePipelineConfig_multiregionUpdated(name),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCodePipeline_MultiRegion_convertSingleRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var p1, p2 codepipeline.PipelineDeclaration
	resourceName := "aws_codepipeline.test"

	name := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckSupported(ctx, t, acctest.AlternateRegion())
			acctest.PreCheckPartitionHasService(t, codestarconnections.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p1),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", ""),
				),
			},
			{
				Config: testAccCodePipelineConfig_multiregion(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p2),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.name", fmt.Sprintf("%s-Build", acctest.AlternateRegion())),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.region", acctest.AlternateRegion()),
				),
			},
			{
				Config: testAccCodePipelineConfig_backToBasic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p1),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", acctest.Region()),
				),
			},
			{
				Config:            testAccCodePipelineConfig_backToBasic(name),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCodePipeline_withNamespace(t *testing.T) {
	ctx := acctest.Context(t)
	var p1 codepipeline.PipelineDeclaration
	name := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSupported(ctx, t)
			acctest.PreCheckPartitionHasService(t, codestarconnections.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_namespace(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "codepipeline", regexp.MustCompile(fmt.Sprintf("test-pipeline-%s", name))),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.namespace", "SourceVariables"),
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

func TestAccCodePipeline_withGitHubV1SourceAction(t *testing.T) {
	ctx := acctest.Context(t)
	githubToken := envvar.SkipIfEmpty(t, envvar.GithubToken, "token with GitHub permissions to repository for CodePipeline source configuration")

	var v codepipeline.PipelineDeclaration
	name := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSupported(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_gitHubv1SourceAction(name, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &v),

					resource.TestCheckResourceAttr(resourceName, "stage.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.category", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.owner", "ThirdParty"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.provider", "GitHub"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Owner", "lifesum-terraform"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Repo", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Branch", "main"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.OAuthToken", githubToken),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"stage.0.action.0.configuration.%",
					"stage.0.action.0.configuration.OAuthToken",
				},
			},
			{
				Config: testAccCodePipelineConfig_gitHubv1SourceActionUpdated(name, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &v),

					resource.TestCheckResourceAttr(resourceName, "stage.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.category", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.owner", "ThirdParty"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.provider", "GitHub"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Owner", "test-terraform"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Repo", "test-repo"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Branch", "stable"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.OAuthToken", githubToken),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"stage.0.action.0.configuration.%",
					"stage.0.action.0.configuration.OAuthToken",
				},
			},
		},
	})
}

func TestAccCodePipeline_ecr(t *testing.T) {
	ctx := acctest.Context(t)
	var p codepipeline.PipelineDeclaration
	name := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSupported(ctx, t)
			acctest.PreCheckPartitionHasService(t, codestarconnections.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_ecr(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "stage.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.category", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.provider", "ECR"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.input_artifacts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.RepositoryName", "my-image-repo"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.ImageTag", "latest"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.run_order", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.region", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
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

func testAccCheckPipelineExists(ctx context.Context, n string, v *codepipeline.PipelineDeclaration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CodePipeline ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodePipelineConn(ctx)

		output, err := tfcodepipeline.FindPipelineByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output.Pipeline

		return nil
	}
}

func testAccCheckPipelineDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodePipelineConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codepipeline" {
				continue
			}

			_, err := tfcodepipeline.FindPipelineByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CodePipeline %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreCheckSupported(ctx context.Context, t *testing.T, regions ...string) {
	regions = append(regions, acctest.Region())
	for _, region := range regions {
		conf := &conns.Config{
			Region: region,
		}
		client, diags := conf.ConfigureProvider(ctx, acctest.Provider.Meta().(*conns.AWSClient))

		if diags.HasError() {
			t.Fatalf("error getting AWS client for region %s", region)
		}
		conn := client.CodePipelineConn(ctx)

		input := &codepipeline.ListPipelinesInput{}
		_, err := conn.ListPipelinesWithContext(ctx, input)

		if acctest.PreCheckSkipError(err) {
			t.Skipf("skipping acceptance testing: %s", err)
		}

		if err != nil {
			t.Fatalf("unexpected PreCheck error: %s", err)
		}
	}
}

func testAccServiceIAMRole(rName string) string {
	return fmt.Sprintf(`
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
  role = aws_iam_role.codepipeline_role.id

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
`, rName)
}

func testAccServiceIAMRoleWithAssumeRole(rName string) string {
	return fmt.Sprintf(`
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
  role = aws_iam_role.codepipeline_role.id

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

func testAccCodePipelineConfig_basic(rName string) string { // nosemgrep:ci.codepipeline-in-func-name
	return acctest.ConfigCompose(
		testAccS3DefaultBucket(rName),
		testAccServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = aws_iam_role.codepipeline_role.arn

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
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.test.arn
        FullRepositoryId = "lifesum-terraform/test"
        BranchName       = "main"
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

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}
`, rName))
}

func testAccCodePipelineConfig_basicUpdated(rName string) string { // nosemgrep:ci.codepipeline-in-func-name
	return acctest.ConfigCompose(
		testAccS3DefaultBucket(rName),
		testAccS3Bucket("updated", rName),
		testAccServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%s"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.updated.bucket
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
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["artifacts"]

      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.test.arn
        FullRepositoryId = "test-terraform/test-repo"
        BranchName       = "stable"
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

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}
`, rName))
}

func testAccCodePipelineConfig_emptyStageArtifacts(rName string) string { // nosemgrep:ci.codepipeline-in-func-name
	return acctest.ConfigCompose(
		testAccS3DefaultBucket(rName),
		testAccServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.test.arn
        FullRepositoryId = "lifesum-terraform/test"
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
      input_artifacts  = ["test", ""]
      output_artifacts = [""]
      version          = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }
}

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}
`, rName))
}

func testAccDeployActionIAMRole(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "codepipeline_action_role" {
  name = "codepipeline-action-role-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "codepipeline_action_policy" {
  name = "codepipeline_action_policy"
  role = aws_iam_role.codepipeline_action_role.id

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
    }
  ]
}
EOF
}
`, rName)
}

func testAccCodePipelineConfig_deployServiceRole(rName string) string { // nosemgrep:ci.codepipeline-in-func-name
	return acctest.ConfigCompose(
		testAccS3DefaultBucket(rName),
		testAccServiceIAMRoleWithAssumeRole(rName),
		testAccDeployActionIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%s"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.test.bucket
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
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["artifacts"]

      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.test.arn
        FullRepositoryId = "lifesum-terraform/test"
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
      role_arn        = aws_iam_role.codepipeline_action_role.arn
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

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}
`, rName))
}

func testAccCodePipelineConfig_tags(rName, tag1, tag2 string) string { // nosemgrep:ci.codepipeline-in-func-name
	return acctest.ConfigCompose(
		testAccS3DefaultBucket(rName),
		testAccServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = aws_iam_role.codepipeline_role.arn

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
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.test.arn
        FullRepositoryId = "lifesum-terraform/test"
        BranchName       = "main"
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

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}
`, rName, tag1, tag2))
}

func testAccCodePipelineConfig_multiregion(rName string) string { // nosemgrep:ci.codepipeline-in-func-name
	return acctest.ConfigCompose(
		acctest.ConfigAlternateRegionProvider(),
		testAccS3DefaultBucket(rName),
		testAccServiceIAMRole(rName),
		testAccS3BucketWithProvider("alternate", rName, "awsalternate"),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.test.bucket
    type     = "S3"

    encryption_key {
      id   = "1234"
      type = "KMS"
    }

    region = "%[2]s"
  }

  artifact_store {
    location = aws_s3_bucket.alternate.bucket
    type     = "S3"

    encryption_key {
      id   = "5678"
      type = "KMS"
    }

    region = "%[3]s"
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.test.arn
        FullRepositoryId = "lifesum-terraform/test"
        BranchName       = "main"
      }
    }
  }

  stage {
    name = "Build"

    action {
      region          = "%[2]s"
      name            = "Build"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["test"]
      version         = "1"

      configuration = {
        ProjectName = "Test"
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

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}
`, rName, acctest.Region(), acctest.AlternateRegion()))
}

func testAccCodePipelineConfig_multiregionUpdated(rName string) string { // nosemgrep:ci.codepipeline-in-func-name
	return acctest.ConfigCompose(
		acctest.ConfigAlternateRegionProvider(),
		testAccS3DefaultBucket(rName),
		testAccServiceIAMRole(rName),
		testAccS3BucketWithProvider("alternate", rName, "awsalternate"),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.test.bucket
    type     = "S3"

    encryption_key {
      id   = "4321"
      type = "KMS"
    }

    region = "%[2]s"
  }

  artifact_store {
    location = aws_s3_bucket.alternate.bucket
    type     = "S3"

    encryption_key {
      id   = "8765"
      type = "KMS"
    }

    region = "%[3]s"
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.test.arn
        FullRepositoryId = "lifesum-terraform/test"
        BranchName       = "main"
      }
    }
  }

  stage {
    name = "Build"

    action {
      region          = "%[2]s"
      name            = "BuildUpdated"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["test"]
      version         = "1"

      configuration = {
        ProjectName = "Test"
      }
    }

    action {
      region          = "%[3]s"
      name            = "%[3]s-BuildUpdated"
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

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}
`, rName, acctest.Region(), acctest.AlternateRegion()))
}

func testAccCodePipelineConfig_backToBasic(rName string) string { // nosemgrep:ci.codepipeline-in-func-name
	return acctest.ConfigCompose(
		acctest.ConfigAlternateRegionProvider(),
		testAccCodePipelineConfig_basic(rName),
	)
}

func testAccS3DefaultBucket(rName string) string {
	return testAccS3Bucket("test", rName)
}

func testAccS3Bucket(bucket, rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "%[1]s" {
  bucket = "tf-test-pipeline-%[1]s-%[2]s"
}
`, bucket, rName)
}

func testAccS3BucketWithProvider(bucket, rName, provider string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "%[1]s" {
  bucket   = "tf-test-pipeline-%[1]s-%[2]s"
  provider = %[3]s
}
`, bucket, rName, provider)
}

func testAccCodePipelineConfig_namespace(rName string) string { // nosemgrep:ci.codepipeline-in-func-name
	return acctest.ConfigCompose(
		testAccS3DefaultBucket(rName),
		testAccServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.foo.bucket
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
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["test"]
      namespace        = "SourceVariables"

      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.test.arn
        FullRepositoryId = "lifesum-terraform/test"
        BranchName       = "main"
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

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}

resource "aws_s3_bucket" "foo" {
  bucket = "tf-test-pipeline-%[1]s"
}
`, rName))
}

func testAccCodePipelineConfig_gitHubv1SourceAction(rName, githubToken string) string { // nosemgrep:ci.codepipeline-in-func-name
	return acctest.ConfigCompose(
		testAccS3DefaultBucket(rName),
		testAccServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = aws_iam_role.codepipeline_role.arn

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
        Branch     = "main"
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
`, rName, githubToken))
}

func testAccCodePipelineConfig_gitHubv1SourceActionUpdated(rName, githubToken string) string { // nosemgrep:ci.codepipeline-in-func-name
	return acctest.ConfigCompose(
		testAccS3DefaultBucket(rName),
		testAccServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = aws_iam_role.codepipeline_role.arn

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
      output_artifacts = ["artifacts"]

      configuration = {
        Owner      = "test-terraform"
        Repo       = "test-repo"
        Branch     = "stable"
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
      input_artifacts = ["artifacts"]
      version         = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }
}
`, rName, githubToken))
}

func testAccCodePipelineConfig_ecr(rName string) string { // nosemgrep:ci.codepipeline-in-func-name
	return acctest.ConfigCompose(
		testAccS3DefaultBucket(rName),
		testAccServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = aws_iam_role.codepipeline_role.arn

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
      owner            = "AWS"
      provider         = "ECR"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        RepositoryName = "my-image-repo"
        ImageTag       = "latest"
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
