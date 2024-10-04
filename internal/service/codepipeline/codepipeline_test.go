// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codepipeline_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
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

func TestAccCodePipeline_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var p types.PipelineDeclaration
	rName := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"
	codestarConnectionResourceName := "aws_codestarconnections_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.codepipeline_role", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "codepipeline", regexache.MustCompile(fmt.Sprintf("test-pipeline-%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.category", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.provider", "CodeStarSourceConnection"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.input_artifacts.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.FullRepositoryId", "lifesum-terraform/test"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.BranchName", "main"),
					resource.TestCheckResourceAttrPair(resourceName, "stage.0.action.0.configuration.ConnectionArn", codestarConnectionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.run_order", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.region", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.category", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.provider", "CodeBuild"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.output_artifacts.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.ProjectName", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.run_order", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCodePipelineConfig_basicUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "stage.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.input_artifacts.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.0", "artifacts"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.FullRepositoryId", "test-terraform/test-repo"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.BranchName", "stable"),
					resource.TestCheckResourceAttrPair(resourceName, "stage.0.action.0.configuration.ConnectionArn", codestarConnectionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.0", "artifacts"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.%", acctest.Ct1),
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
	var p types.PipelineDeclaration
	rName := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeStarConnectionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_basic(rName),
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
	var p types.PipelineDeclaration
	rName := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeStarConnectionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_emptyStageArtifacts(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "codepipeline", regexache.MustCompile(fmt.Sprintf("test-pipeline-%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.category", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.provider", "CodeBuild"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.output_artifacts.#", acctest.Ct0),
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
	var p types.PipelineDeclaration
	rName := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeStarConnectionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_deployServiceRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "stage.2.name", "Deploy"),
					resource.TestCheckResourceAttr(resourceName, "stage.2.action.0.category", "Deploy"),
					resource.TestCheckResourceAttrPair(resourceName, "stage.2.action.0.role_arn", "aws_iam_role.codepipeline_action_role", names.AttrARN),
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
	var p types.PipelineDeclaration
	rName := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeStarConnectionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCodePipelineConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
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
				Config: testAccCodePipelineConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccCodePipeline_MultiRegion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var p types.PipelineDeclaration
	rName := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheck(ctx, t, acctest.AlternateRegion())
			acctest.PreCheckPartitionHasService(t, names.CodeStarConnectionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_multiregion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.name", fmt.Sprintf("%s-Build", acctest.AlternateRegion())),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.region", acctest.AlternateRegion()),
				),
			},
			{
				Config:            testAccCodePipelineConfig_multiregion(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCodePipeline_MultiRegion_update(t *testing.T) {
	ctx := acctest.Context(t)
	var p types.PipelineDeclaration
	rName := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheck(ctx, t, acctest.AlternateRegion())
			acctest.PreCheckPartitionHasService(t, names.CodeStarConnectionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_multiregion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.name", fmt.Sprintf("%s-Build", acctest.AlternateRegion())),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.region", acctest.AlternateRegion()),
				),
			},
			{
				Config: testAccCodePipelineConfig_multiregionUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "BuildUpdated"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.name", fmt.Sprintf("%s-BuildUpdated", acctest.AlternateRegion())),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.region", acctest.AlternateRegion()),
				),
			},
			{
				Config:            testAccCodePipelineConfig_multiregionUpdated(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCodePipeline_MultiRegion_convertSingleRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var p types.PipelineDeclaration
	rName := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheck(ctx, t, acctest.AlternateRegion())
			acctest.PreCheckPartitionHasService(t, names.CodeStarConnectionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.0.region", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", ""),
				),
			},
			{
				Config: testAccCodePipelineConfig_multiregion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.0.region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.1.region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.name", fmt.Sprintf("%s-Build", acctest.AlternateRegion())),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.region", acctest.AlternateRegion()),
				),
			},
			{
				Config: testAccCodePipelineConfig_backToBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.0.region", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", acctest.Region()),
				),
			},
			{
				Config:            testAccCodePipelineConfig_backToBasic(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCodePipeline_withNamespace(t *testing.T) {
	ctx := acctest.Context(t)
	var p types.PipelineDeclaration
	rName := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeStarConnectionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_namespace(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "codepipeline", regexache.MustCompile(fmt.Sprintf("test-pipeline-%s", rName))),
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
	githubToken := acctest.SkipIfEnvVarNotSet(t, envvar.GithubToken)
	var p types.PipelineDeclaration
	rName := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_gitHubv1SourceAction(rName, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "stage.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.category", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.owner", "ThirdParty"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.provider", "GitHub"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", acctest.Ct4),
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
				Config: testAccCodePipelineConfig_gitHubv1SourceActionUpdated(rName, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "stage.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.category", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.owner", "ThirdParty"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.provider", "GitHub"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", acctest.Ct4),
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
	var p types.PipelineDeclaration
	rName := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeStarConnectionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_ecr(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "stage.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.category", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.provider", "ECR"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.input_artifacts.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.RepositoryName", "my-image-repo"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.ImageTag", "latest"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.run_order", acctest.Ct1),
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

func TestAccCodePipeline_pipelinetype(t *testing.T) {
	ctx := acctest.Context(t)
	var p types.PipelineDeclaration
	rName := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"
	codestarConnectionResourceName := "aws_codestarconnections_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeStarConnectionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_pipelinetype(rName, "V1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.codepipeline_role", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "codepipeline", regexache.MustCompile(fmt.Sprintf("test-pipeline-%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "execution_mode", string(types.ExecutionModeSuperseded)),
					resource.TestCheckResourceAttr(resourceName, "pipeline_type", string(types.PipelineTypeV1)),
					resource.TestCheckResourceAttr(resourceName, "stage.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.category", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.provider", "CodeStarSourceConnection"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.input_artifacts.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.FullRepositoryId", "lifesum-terraform/test"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.BranchName", "main"),
					resource.TestCheckResourceAttrPair(resourceName, "stage.0.action.0.configuration.ConnectionArn", codestarConnectionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.run_order", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.region", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.category", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.provider", "CodeBuild"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.output_artifacts.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.ProjectName", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.run_order", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", ""),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCodePipelineConfig_pipelinetypeUpdated1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "execution_mode", string(types.ExecutionModeQueued)),
					resource.TestCheckResourceAttr(resourceName, "pipeline_type", string(types.PipelineTypeV2)),
					resource.TestCheckResourceAttr(resourceName, "stage.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.input_artifacts.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.0", "artifacts"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.FullRepositoryId", "test-terraform/test-repo"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.BranchName", "stable"),
					resource.TestCheckResourceAttrPair(resourceName, "stage.0.action.0.configuration.ConnectionArn", codestarConnectionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.0", "artifacts"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.ProjectName", "test"),
					resource.TestCheckResourceAttr(resourceName, "trigger.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.provider_type", "CodeStarSourceConnection"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.source_action_name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.includes.#", "8"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.includes.0", "main"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.includes.1", "sub1"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.includes.2", "sub2"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.includes.3", "sub3"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.includes.4", "sub4"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.includes.5", "sub5"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.includes.6", "sub6"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.includes.7", "sub7"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.excludes.#", "8"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.excludes.0", "feature/test1"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.excludes.1", "feature/test2"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.excludes.2", "feature/test3"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.excludes.3", "feature/test4"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.excludes.4", "feature/test5"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.excludes.5", "feature/test6"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.excludes.6", "feature/test7"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.excludes.7", "feature/wildcard*"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.includes.#", "8"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.includes.0", "src/production1"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.includes.1", "src/production2"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.includes.2", "src/production3"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.includes.3", "src/production4"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.includes.4", "src/production5"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.includes.5", "src/production6"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.includes.6", "src/production7"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.includes.7", "src/production8"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.excludes.#", "8"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.excludes.0", "test/production1"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.excludes.1", "test/production2"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.excludes.2", "test/production3"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.excludes.3", "test/production4"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.excludes.4", "test/production5"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.excludes.5", "test/production6"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.excludes.6", "test/production7"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.excludes.7", "test/production8"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.includes.#", "8"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.includes.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.includes.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.includes.2", "tag3"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.includes.3", "tag4"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.includes.4", "tag5"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.includes.5", "tag6"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.includes.6", "tag7"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.includes.7", "tag8"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.excludes.#", "8"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.excludes.0", "tag11"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.excludes.1", "tag12"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.excludes.2", "tag13"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.excludes.3", "tag14"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.excludes.4", "tag15"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.excludes.5", "tag16"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.excludes.6", "tag17"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.excludes.7", "tag18"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.events.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.events.0", "OPEN"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.events.1", "UPDATED"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.events.2", "CLOSED"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.includes.#", "8"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.includes.0", "main1"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.includes.1", "sub11"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.includes.2", "sub12"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.includes.3", "sub13"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.includes.4", "sub14"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.includes.5", "sub15"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.includes.6", "sub16"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.includes.7", "sub17"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.excludes.#", "8"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.excludes.0", "feature/test11"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.excludes.1", "feature/test12"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.excludes.2", "feature/test13"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.excludes.3", "feature/test14"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.excludes.4", "feature/test15"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.excludes.5", "feature/test16"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.excludes.6", "feature/test17"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.excludes.7", "feature/wildcard1*"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.includes.#", "8"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.includes.0", "src/production11"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.includes.1", "src/production12"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.includes.2", "src/production13"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.includes.3", "src/production14"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.includes.4", "src/production15"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.includes.5", "src/production16"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.includes.6", "src/production17"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.includes.7", "src/production18"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.excludes.#", "8"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.excludes.0", "test/production11"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.excludes.1", "test/production12"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.excludes.2", "test/production13"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.excludes.3", "test/production14"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.excludes.4", "test/production15"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.excludes.5", "test/production16"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.excludes.6", "test/production17"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.excludes.7", "test/production18"),
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
				Config: testAccCodePipelineConfig_pipelinetypeUpdated2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "execution_mode", string(types.ExecutionModeQueued)),
					resource.TestCheckResourceAttr(resourceName, "pipeline_type", string(types.PipelineTypeV2)),
					resource.TestCheckResourceAttr(resourceName, "stage.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.input_artifacts.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.0", "artifacts"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.FullRepositoryId", "test-terraform/test-repo"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.BranchName", "stable"),
					resource.TestCheckResourceAttrPair(resourceName, "stage.0.action.0.configuration.ConnectionArn", codestarConnectionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.0", "artifacts"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.ProjectName", "test"),
					resource.TestCheckResourceAttr(resourceName, "variable.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "variable.0.name", "test_var1"),
					resource.TestCheckResourceAttr(resourceName, "variable.0.description", "This is test pipeline variable 1."),
					resource.TestCheckResourceAttr(resourceName, "variable.0.default_value", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "variable.1.name", "test_var2"),
					resource.TestCheckResourceAttr(resourceName, "variable.1.description", "This is test pipeline variable 2."),
					resource.TestCheckResourceAttr(resourceName, "variable.1.default_value", acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, "trigger.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.provider_type", "CodeStarSourceConnection"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.source_action_name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.includes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.includes.0", "main"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.excludes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.branches.0.excludes.0", "feature/test*"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.includes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.includes.0", "src/production1"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.excludes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.file_paths.0.excludes.0", "test/production1"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.includes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.includes.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.excludes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.push.0.tags.0.excludes.0", "tag11"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.events.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.events.0", "OPEN"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.includes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.includes.0", "main1"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.excludes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.branches.0.excludes.0", "feature/test1*"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.includes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.includes.0", "src/production11"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.excludes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.0.pull_request.0.file_paths.0.excludes.0", "test/production11"),
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
				Config: testAccCodePipelineConfig_pipelinetypeUpdated3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "execution_mode", string(types.ExecutionModeQueued)),
					resource.TestCheckResourceAttr(resourceName, "pipeline_type", string(types.PipelineTypeV2)),
					resource.TestCheckResourceAttr(resourceName, "stage.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.input_artifacts.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.0", "artifacts"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.FullRepositoryId", "test-terraform/test-repo"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.BranchName", "stable"),
					resource.TestCheckResourceAttrPair(resourceName, "stage.0.action.0.configuration.ConnectionArn", codestarConnectionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.0", "artifacts"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.ProjectName", "test"),
					resource.TestCheckResourceAttr(resourceName, "variable.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "variable.0.name", "test_var1"),
					resource.TestCheckResourceAttr(resourceName, "variable.0.description", "This is test pipeline variable 1."),
					resource.TestCheckResourceAttr(resourceName, "variable.0.default_value", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "variable.1.name", "test_var2"),
					resource.TestCheckResourceAttr(resourceName, "variable.1.description", "This is test pipeline variable 2."),
					resource.TestCheckResourceAttr(resourceName, "variable.1.default_value", acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.#", acctest.Ct1),
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
				Config: testAccCodePipelineConfig_pipelinetype(rName, "V2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.codepipeline_role", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "codepipeline", regexache.MustCompile(fmt.Sprintf("test-pipeline-%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "execution_mode", string(types.ExecutionModeSuperseded)),
					resource.TestCheckResourceAttr(resourceName, "pipeline_type", string(types.PipelineTypeV2)),
					resource.TestCheckResourceAttr(resourceName, "stage.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.category", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.provider", "CodeStarSourceConnection"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.input_artifacts.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.FullRepositoryId", "lifesum-terraform/test"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.BranchName", "main"),
					resource.TestCheckResourceAttrPair(resourceName, "stage.0.action.0.configuration.ConnectionArn", codestarConnectionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.run_order", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.region", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.category", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.provider", "CodeBuild"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.output_artifacts.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.ProjectName", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.run_order", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", ""),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.git_configuration.#", acctest.Ct1),
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

func TestAccCodePipeline_manualApprovalTimeoutInMinutes(t *testing.T) {
	ctx := acctest.Context(t)
	var p types.PipelineDeclaration
	rName := sdkacctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodePipelineConfig_manualApprovalTimeoutInMinutes(rName, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.codepipeline_role", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "codepipeline", regexache.MustCompile(fmt.Sprintf("test-pipeline-%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "stage.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.timeout_in_minutes", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Approval"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.timeout_in_minutes", "5"),
					resource.TestCheckResourceAttr(resourceName, "stage.2.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.2.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.2.action.0.timeout_in_minutes", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCodePipelineConfig_manualApprovalNoTimeoutInMinutes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &p),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.codepipeline_role", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "codepipeline", regexache.MustCompile(fmt.Sprintf("test-pipeline-%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "stage.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.timeout_in_minutes", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Approval"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.timeout_in_minutes", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage.2.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.2.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stage.2.action.0.timeout_in_minutes", acctest.Ct0),
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

func testAccCheckPipelineExists(ctx context.Context, n string, v *types.PipelineDeclaration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodePipelineClient(ctx)

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
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodePipelineClient(ctx)

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

			return fmt.Errorf("CodePipeline Pipeline %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T, regions ...string) {
	regions = append(regions, acctest.Region())
	for _, region := range regions {
		c := &conns.Config{
			Region: region,
		}
		client, diags := c.ConfigureProvider(ctx, acctest.Provider.Meta().(*conns.AWSClient))
		if diags.HasError() {
			t.Fatalf("error getting AWS client for region %s", region)
		}
		conn := client.CodePipelineClient(ctx)

		_, err := conn.ListPipelines(ctx, &codepipeline.ListPipelinesInput{})

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
  name     = "test-pipeline-%[1]s"
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

func testAccCodePipelineConfig_pipelinetype(rName, pipelineType string) string { // nosemgrep:ci.codepipeline-in-func-name
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

  pipeline_type = %[2]q
}

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}
`, rName, pipelineType))
}

func testAccCodePipelineConfig_pipelinetypeUpdated1(rName string) string { // nosemgrep:ci.codepipeline-in-func-name
	return acctest.ConfigCompose(
		testAccS3DefaultBucket(rName),
		testAccS3Bucket("updated", rName),
		testAccServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
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

  execution_mode = "QUEUED"

  pipeline_type = "V2"

  variable {
    name          = "test_var1"
    description   = "This is test pipeline variable 1."
    default_value = "value1"
  }

  variable {
    name          = "test_var2"
    description   = "This is test pipeline variable 2."
    default_value = "value2"
  }

  trigger {
    provider_type = "CodeStarSourceConnection"
    git_configuration {
      source_action_name = "Source"
      push {
        branches {
          includes = ["main", "sub1", "sub2", "sub3", "sub4", "sub5", "sub6", "sub7"]
          excludes = ["feature/test1", "feature/test2", "feature/test3", "feature/test4", "feature/test5", "feature/test6", "feature/test7", "feature/wildcard*"]
        }
        file_paths {
          includes = ["src/production1", "src/production2", "src/production3", "src/production4", "src/production5", "src/production6", "src/production7", "src/production8"]
          excludes = ["test/production1", "test/production2", "test/production3", "test/production4", "test/production5", "test/production6", "test/production7", "test/production8"]
        }
        tags {
          includes = ["tag1", "tag2", "tag3", "tag4", "tag5", "tag6", "tag7", "tag8"]
          excludes = ["tag11", "tag12", "tag13", "tag14", "tag15", "tag16", "tag17", "tag18"]
        }
      }
      pull_request {
        events = ["OPEN", "UPDATED", "CLOSED"]
        branches {
          includes = ["main1", "sub11", "sub12", "sub13", "sub14", "sub15", "sub16", "sub17"]
          excludes = ["feature/test11", "feature/test12", "feature/test13", "feature/test14", "feature/test15", "feature/test16", "feature/test17", "feature/wildcard1*"]
        }
        file_paths {
          includes = ["src/production11", "src/production12", "src/production13", "src/production14", "src/production15", "src/production16", "src/production17", "src/production18"]
          excludes = ["test/production11", "test/production12", "test/production13", "test/production14", "test/production15", "test/production16", "test/production17", "test/production18"]
        }
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

func testAccCodePipelineConfig_pipelinetypeUpdated2(rName string) string { // nosemgrep:ci.codepipeline-in-func-name
	return acctest.ConfigCompose(
		testAccS3DefaultBucket(rName),
		testAccS3Bucket("updated", rName),
		testAccServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
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

  execution_mode = "QUEUED"

  pipeline_type = "V2"

  variable {
    name          = "test_var1"
    description   = "This is test pipeline variable 1."
    default_value = "value1"
  }

  variable {
    name          = "test_var2"
    description   = "This is test pipeline variable 2."
    default_value = "value2"
  }

  trigger {
    provider_type = "CodeStarSourceConnection"
    git_configuration {
      source_action_name = "Source"
      push {
        branches {
          includes = ["main"]
          excludes = ["feature/test*"]
        }
        file_paths {
          includes = ["src/production1"]
          excludes = ["test/production1"]
        }
        tags {
          includes = ["tag1"]
          excludes = ["tag11"]
        }
      }
      pull_request {
        events = ["OPEN"]
        branches {
          includes = ["main1"]
          excludes = ["feature/test1*"]
        }
        file_paths {
          includes = ["src/production11"]
          excludes = ["test/production11"]
        }
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

func testAccCodePipelineConfig_pipelinetypeUpdated3(rName string) string { // nosemgrep:ci.codepipeline-in-func-name
	return acctest.ConfigCompose(
		testAccS3DefaultBucket(rName),
		testAccS3Bucket("updated", rName),
		testAccServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
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

  execution_mode = "QUEUED"

  pipeline_type = "V2"

  variable {
    name          = "test_var1"
    description   = "This is test pipeline variable 1."
    default_value = "value1"
  }

  variable {
    name          = "test_var2"
    description   = "This is test pipeline variable 2."
    default_value = "value2"
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
  name = "codepipeline-action-role-%[1]s"

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
  name     = "test-pipeline-%[1]s"
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

func testAccCodePipelineConfig_tags1(rName, tagKey1, tagValue1 string) string { // nosemgrep:ci.codepipeline-in-func-name
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
    %[2]q = %[3]q
  }
}

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}
`, rName, tagKey1, tagValue1))
}

func testAccCodePipelineConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string { // nosemgrep:ci.codepipeline-in-func-name
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
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
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

    region = %[2]q
  }

  artifact_store {
    location = aws_s3_bucket.alternate.bucket
    type     = "S3"

    encryption_key {
      id   = "5678"
      type = "KMS"
    }

    region = %[3]q
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
      region          = %[2]q
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
      region          = %[3]q
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

    region = %[2]q
  }

  artifact_store {
    location = aws_s3_bucket.alternate.bucket
    type     = "S3"

    encryption_key {
      id   = "8765"
      type = "KMS"
    }

    region = %[3]q
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
      region          = %[2]q
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
      region          = %[3]q
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

func testAccCodePipelineConfig_manualApprovalTimeoutInMinutes(rName string, timeoutInMinutes int) string { // nosemgrep:ci.codepipeline-in-func-name
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
    name = "Approval"

    action {
      name               = "Approval"
      category           = "Approval"
      owner              = "AWS"
      provider           = "Manual"
      version            = "1"
      timeout_in_minutes = %[2]d
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
`, rName, timeoutInMinutes))
}

func testAccCodePipelineConfig_manualApprovalNoTimeoutInMinutes(rName string) string { // nosemgrep:ci.codepipeline-in-func-name
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
    name = "Approval"

    action {
      name     = "Approval"
      category = "Approval"
      owner    = "AWS"
      provider = "Manual"
      version  = "1"
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
