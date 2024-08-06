// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodebuild "github.com/hashicorp/terraform-provider-aws/internal/service/codebuild"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.CodeBuildServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"InvalidInputException: Region",
	)
}

// This is used for testing aws_codebuild_webhook as well as aws_codebuild_project.
// The Terraform AWS user must have done the manual Bitbucket OAuth dance for this
// functionality to work. Additionally, the Bitbucket user that the Terraform AWS
// user logs in as must have access to the Bitbucket repository.
func testAccBitbucketSourceLocationFromEnv() string {
	sourceLocation := os.Getenv("AWS_CODEBUILD_BITBUCKET_SOURCE_LOCATION")
	if sourceLocation == "" {
		return "https://terraform@bitbucket.org/terraform/aws-test.git" // nosemgrep:ci.email-address
	}
	return sourceLocation
}

// This is used for testing aws_codebuild_webhook as well as aws_codebuild_project.
// The Terraform AWS user must have done the manual GitHub OAuth dance for this
// functionality to work. Additionally, the GitHub user that the Terraform AWS
// user logs in as must have access to the GitHub repository.
func testAccGitHubSourceLocationFromEnv() string {
	sourceLocation := os.Getenv("AWS_CODEBUILD_GITHUB_SOURCE_LOCATION")
	if sourceLocation == "" {
		return "https://github.com/hashibot-test/aws-test.git"
	}
	return sourceLocation
}

func TestProject_nameValidation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{Value: "_test", ErrCount: 1},
		{Value: "test", ErrCount: 0},
		{Value: "1_test", ErrCount: 0},
		{Value: "test**1", ErrCount: 1},
		{Value: sdkacctest.RandString(256), ErrCount: 1},
	}

	for _, tc := range cases {
		_, errors := tfcodebuild.ValidProjectName(tc.Value, "aws_codebuild_project")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the AWS CodeBuild project name to trigger a validation error - %s", errors)
		}
	}
}

func TestAccCodeBuildProject_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_codebuild_project.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "codebuild", fmt.Sprintf("project/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "badge_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "build_timeout", "60"),
					resource.TestCheckResourceAttr(resourceName, "queued_timeout", "480"),
					resource.TestCheckResourceAttr(resourceName, "cache.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", string(types.CacheTypeNoCache)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					acctest.CheckResourceAttrRegionalARN(resourceName, "encryption_key", "kms", "alias/aws/s3"),
					resource.TestCheckResourceAttr(resourceName, "environment.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "environment.0.compute_type", string(types.ComputeTypeBuildGeneral1Small)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "environment.0.image", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "environment.0.privileged_mode", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "environment.0.type", string(types.EnvironmentTypeLinuxContainer)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.image_pull_credentials_type", string(types.ImagePullCredentialsTypeCodebuild)),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.status", string(types.LogsConfigStatusTypeEnabled)),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.status", string(types.LogsConfigStatusTypeDisabled)),
					resource.TestCheckResourceAttr(resourceName, "project_visibility", "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secondary_source_version.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.git_clone_depth", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source.0.insecure_ssl", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "source.0.location", testAccGitHubSourceLocationFromEnv()),
					resource.TestCheckResourceAttr(resourceName, "source.0.report_build_status", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "source.0.type", "GITHUB"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct0),
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

func TestAccCodeBuildProject_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodebuild.ResourceProject(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodeBuildProject_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
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
				Config: testAccProjectConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccProjectConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_publicVisibility(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_codebuild_project.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_visibility(rName, "PUBLIC_READ"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "project_visibility", "PUBLIC_READ"),
					resource.TestCheckResourceAttrSet(resourceName, "public_project_alias"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_visibility(rName, "PRIVATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "project_visibility", "PRIVATE"),
				),
			},
			{
				Config: testAccProjectConfig_visibilityResourceRole(rName, "PRIVATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "project_visibility", "PRIVATE"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_access_role", roleResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_badgeEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_badgeEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "badge_enabled", acctest.CtTrue),
					resource.TestMatchResourceAttr(resourceName, "badge_url", regexache.MustCompile(`\b(https?).*\b`)),
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

func TestAccCodeBuildProject_buildTimeout(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_buildTimeout(rName, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "build_timeout", "120"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_buildTimeout(rName, 240),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "build_timeout", "240"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_queuedTimeout(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_queuedTimeout(rName, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "queued_timeout", "120"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_queuedTimeout(rName, 240),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "queued_timeout", "240"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_cache(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"
	s3Location1 := rName + "-1"
	s3Location2 := rName + "-2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectConfig_cache(rName, "", "S3"),
				ExpectError: regexache.MustCompile(`cache location is required when cache type is "S3"`),
			},
			{
				Config: testAccProjectConfig_cache(rName, "", string(types.CacheTypeNoCache)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", string(types.CacheTypeNoCache)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", string(types.CacheTypeNoCache)),
				),
			},
			{
				Config: testAccProjectConfig_cache(rName, s3Location1, "S3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cache.0.location", s3Location1),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "S3"),
				),
			},
			{
				Config: testAccProjectConfig_cache(rName, s3Location2, "S3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cache.0.location", s3Location2),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "S3"),
				),
			},
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", string(types.CacheTypeNoCache)),
				),
			},
			{
				Config: testAccProjectConfig_localCache(rName, "LOCAL_DOCKER_LAYER_CACHE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cache.0.modes.0", "LOCAL_DOCKER_LAYER_CACHE"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "LOCAL"),
				),
			},
			{
				Config: testAccProjectConfig_s3ComputedLocation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", string(types.CacheTypeS3)),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_description(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_fileSystemLocations(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID, "efs"), //using efs.EndpointsID will import efs and make linters sad
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_fileSystemLocations(rName, "/mount1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "environment.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "environment.0.compute_type", string(types.ComputeTypeBuildGeneral1Small)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "environment.0.image", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "environment.0.privileged_mode", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "environment.0.type", string(types.EnvironmentTypeLinuxContainer)),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.0.identifier", "test"),
					resource.TestMatchResourceAttr(resourceName, "file_system_locations.0.location", regexache.MustCompile(`/directory-path$`)),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.0.mount_options", "nfsvers=4.1,rsize=1048576,wsize=1048576,hard,timeo=450,retrans=3"),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.0.mount_point", "/mount1"),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.0.type", string(types.FileSystemTypeEfs)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_fileSystemLocations(rName, "/mount2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.0.identifier", "test"),
					resource.TestMatchResourceAttr(resourceName, "file_system_locations.0.location", regexache.MustCompile(`/directory-path$`)),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.0.mount_options", "nfsvers=4.1,rsize=1048576,wsize=1048576,hard,timeo=450,retrans=3"),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.0.mount_point", "/mount2"),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.0.type", string(types.FileSystemTypeEfs)),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_sourceVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_sourceVersion(rName, "master"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source_version", "master"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_encryptionKey(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_encryptionKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_key", "aws_kms_key.test", names.AttrARN),
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

func TestAccCodeBuildProject_Environment_environmentVariable(t *testing.T) {
	ctx := acctest.Context(t)
	var project1, project2, project3 types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_environmentVariableOne(rName, "KEY1", "VALUE1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_environmentVariableTwo(rName, "KEY1", "VALUE1UPDATED", "KEY2", "VALUE2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_environmentVariableZero(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project3),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.#", acctest.Ct0),
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

func TestAccCodeBuildProject_EnvironmentEnvironmentVariable_type(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_environmentVariableType(rName, string(types.EnvironmentVariableTypePlaintext)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.0.type", string(types.EnvironmentVariableTypePlaintext)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.1.type", string(types.EnvironmentVariableTypePlaintext)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_environmentVariableType(rName, string(types.EnvironmentVariableTypeParameterStore)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.0.type", string(types.EnvironmentVariableTypePlaintext)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.1.type", string(types.EnvironmentVariableTypeParameterStore)),
				),
			},
			{
				Config: testAccProjectConfig_environmentVariableType(rName, string(types.EnvironmentVariableTypeSecretsManager)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.0.type", string(types.EnvironmentVariableTypePlaintext)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.1.type", string(types.EnvironmentVariableTypeSecretsManager)),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_EnvironmentEnvironmentVariable_value(t *testing.T) {
	ctx := acctest.Context(t)
	var project1, project2, project3 types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_environmentVariableOne(rName, "KEY1", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_environmentVariableOne(rName, "KEY1", "VALUE1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_environmentVariableOne(rName, "KEY1", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project3),
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

func TestAccCodeBuildProject_Environment_certificate(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	oName := "certificate.pem"
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_environmentCertificate(rName, oName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					testAccCheckProjectCertificate(&project, fmt.Sprintf("%s/%s", rName, oName)),
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

func TestAccCodeBuildProject_Environment_registryCredential(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_environmentRegistryCredential1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_environmentRegistryCredential2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Logs_cloudWatchLogs(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_cloudWatchLogs(rName, string(types.LogsConfigStatusTypeEnabled), "group-name", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.status", string(types.LogsConfigStatusTypeEnabled)),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.group_name", "group-name"),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.stream_name", ""),
				),
			},
			{
				Config: testAccProjectConfig_cloudWatchLogs(rName, string(types.LogsConfigStatusTypeEnabled), "group-name", "stream-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.status", string(types.LogsConfigStatusTypeEnabled)),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.group_name", "group-name"),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.stream_name", "stream-name"),
				),
			},
			{
				Config: testAccProjectConfig_cloudWatchLogs(rName, string(types.LogsConfigStatusTypeDisabled), "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.status", string(types.LogsConfigStatusTypeDisabled)),
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

func TestAccCodeBuildProject_Logs_s3Logs(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_s3Logs(rName, string(types.LogsConfigStatusTypeEnabled), rName+"/build-log", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.status", string(types.LogsConfigStatusTypeEnabled)),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.location", rName+"/build-log"),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.encryption_disabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccProjectConfig_s3Logs(rName, string(types.LogsConfigStatusTypeEnabled), rName+"/build-log", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.status", string(types.LogsConfigStatusTypeEnabled)),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.location", rName+"/build-log"),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.encryption_disabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccProjectConfig_s3Logs(rName, string(types.LogsConfigStatusTypeDisabled), "", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.status", string(types.LogsConfigStatusTypeDisabled)),
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

func TestAccCodeBuildProject_buildBatch(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	if acctest.Partition() == "aws-us-gov" {
		t.Skip("CodeBuild Project build batch config is not supported in GovCloud partition")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_buildBatch(rName, true, "BUILD_GENERAL1_SMALL", 10, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.combine_artifacts", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.0.compute_types_allowed.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.0.compute_types_allowed.0", "BUILD_GENERAL1_SMALL"),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.0.maximum_builds_allowed", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.timeout_in_mins", "5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_buildBatch(rName, false, "BUILD_GENERAL1_MEDIUM", 20, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.combine_artifacts", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.0.compute_types_allowed.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.0.compute_types_allowed.0", "BUILD_GENERAL1_MEDIUM"),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.0.maximum_builds_allowed", "20"),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.timeout_in_mins", acctest.Ct10),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_buildBatchConfigDelete(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_buildBatchConfigDelete(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.combine_artifacts", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.0.compute_types_allowed.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.0.maximum_builds_allowed", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.timeout_in_mins", "2160"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_buildBatchConfigDelete(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckNoResourceAttr(resourceName, "build_batch_config.%"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Source_gitCloneDepth(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_sourceGitCloneDepth(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.git_clone_depth", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_sourceGitCloneDepth(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.git_clone_depth", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SourceGitSubmodules_codeCommit(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_sourceGitSubmodulesCodeCommit(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.git_submodules_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.git_submodules_config.0.fetch_submodules", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_sourceGitSubmodulesCodeCommit(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.git_submodules_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.git_submodules_config.0.fetch_submodules", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SourceGitSubmodules_gitHub(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_sourceGitSubmodulesGitHub(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_sourceGitSubmodulesGitHub(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SourceGitSubmodules_gitHubEnterprise(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithubEnterprise)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_sourceGitSubmodulesGitHubEnterprise(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_sourceGitSubmodulesGitHubEnterprise(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondarySourcesGitSubmodules_codeCommit(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_secondarySourcesGitSubmodulesCodeCommit(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier":                        "secondarySource1",
						"git_submodules_config.#":                  acctest.Ct1,
						"git_submodules_config.0.fetch_submodules": acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier":                        "secondarySource2",
						"git_submodules_config.#":                  acctest.Ct1,
						"git_submodules_config.0.fetch_submodules": acctest.CtTrue,
					}),
				),
			},
			{
				Config: testAccProjectConfig_secondarySourcesGitSubmodulesCodeCommit(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier":                        "secondarySource1",
						"git_submodules_config.#":                  acctest.Ct1,
						"git_submodules_config.0.fetch_submodules": acctest.CtFalse,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier":                        "secondarySource2",
						"git_submodules_config.#":                  acctest.Ct1,
						"git_submodules_config.0.fetch_submodules": acctest.CtFalse,
					}),
				),
			},
			{
				Config: testAccProjectConfig_secondarySourcesNone(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondarySourcesGitSubmodules_gitHub(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_secondarySourcesGitSubmodulesGitHub(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
				),
			},
			{
				Config: testAccProjectConfig_secondarySourcesGitSubmodulesGitHub(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondarySourcesGitSubmodules_gitHubEnterprise(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithubEnterprise)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_secondarySourcesGitSubmodulesGitHubEnterprise(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
				),
			},
			{
				Config: testAccProjectConfig_secondarySourcesGitSubmodulesGitHubEnterprise(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondarySourcesVersions(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_secondarySourceVersionsCodeCommit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "secondary_source_version.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_source_version.*", map[string]string{
						"source_identifier": "secondarySource1",
						"source_version":    "master",
					}),
				),
			},
			{
				Config: testAccProjectConfig_secondarySourceVersionsCodeCommitUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "secondary_source_version.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_source_version.*", map[string]string{
						"source_identifier": "secondarySource1",
						"source_version":    "master",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_source_version.*", map[string]string{
						"source_identifier": "secondarySource2",
						"source_version":    "master",
					}),
				),
			},
			{
				Config: testAccProjectConfig_secondarySourceVersionsCodeCommit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "secondary_source_version.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_source_version.*", map[string]string{
						"source_identifier": "secondarySource1",
						"source_version":    "master",
					}),
				),
			},
			{
				Config: testAccProjectConfig_secondarySourcesGitSubmodulesCodeCommit(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "secondary_source_version.#", acctest.Ct0),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource2",
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Source_insecureSSL(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_sourceInsecureSSL(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.insecure_ssl", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_sourceInsecureSSL(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.insecure_ssl", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SourceBuildStatus_gitHubEnterprise(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithubEnterprise)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_sourceBuildStatusGitHubEnterprise(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
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

func TestAccCodeBuildProject_SourceReportBuildStatus_gitHubEnterprise(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithubEnterprise)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_sourceReportBuildStatusGitHubEnterprise(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.report_build_status", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_sourceReportBuildStatusGitHubEnterprise(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.report_build_status", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SourceReportBuildStatus_bitbucket(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	sourceLocation := testAccBitbucketSourceLocationFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeBitbucket)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_sourceReportBuildStatusBitbucket(rName, sourceLocation, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.report_build_status", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_sourceReportBuildStatusBitbucket(rName, sourceLocation, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.report_build_status", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SourceReportBuildStatus_gitHub(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_sourceReportBuildStatusGitHub(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.report_build_status", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_sourceReportBuildStatusGitHub(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.report_build_status", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SourceType_bitbucket(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	sourceLocation := testAccBitbucketSourceLocationFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeBitbucket)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_sourceTypeBitbucket(rName, sourceLocation),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.type", "BITBUCKET"),
					resource.TestCheckResourceAttr(resourceName, "source.0.location", sourceLocation),
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

func TestAccCodeBuildProject_SourceType_codeCommit(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_sourceTypeCodeCommit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.type", "CODECOMMIT"),
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

func TestAccCodeBuildProject_SourceType_codePipeline(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_sourceTypeCodePipeline(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.type", "CODEPIPELINE"),
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

func TestAccCodeBuildProject_SourceType_gitHubEnterprise(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithubEnterprise)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_sourceTypeGitHubEnterprise(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.type", "GITHUB_ENTERPRISE"),
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

func TestAccCodeBuildProject_SourceType_s3(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_sourceTypeS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
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

func TestAccCodeBuildProject_SourceType_noSource(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"
	rBuildspec := `
version: 0.2
phases:
  build:
    commands:
      - rspec hello_world_spec.rb
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_sourceTypeNoSource(rName, "", rBuildspec),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.type", "NO_SOURCE"),
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

func TestAccCodeBuildProject_SourceType_noSourceInvalid(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBuildspec := `
version: 0.2
phases:
  build:
    commands:
      - rspec hello_world_spec.rb
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectConfig_sourceTypeNoSource(rName, "", ""),
				ExpectError: regexache.MustCompile("`buildspec` must be set when source's `type` is `NO_SOURCE`"),
			},
			{
				Config:      testAccProjectConfig_sourceTypeNoSource(rName, names.AttrLocation, rBuildspec),
				ExpectError: regexache.MustCompile("`location` must be empty when source's `type` is `NO_SOURCE`"),
			},
		},
	})
}

func TestAccCodeBuildProject_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_vpc2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", acctest.Ct2),
					resource.TestMatchResourceAttr(resourceName, "vpc_config.0.vpc_id", regexache.MustCompile(`^vpc-`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_vpc1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "vpc_config.0.vpc_id", regexache.MustCompile(`^vpc-`)),
				),
			},
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_windowsServer2019Container(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_windowsServer2019Container(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "environment.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "environment.0.compute_type", string(types.ComputeTypeBuildGeneral1Medium)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "environment.0.image", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "environment.0.privileged_mode", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "environment.0.image_pull_credentials_type", string(types.ImagePullCredentialsTypeCodebuild)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.type", string(types.EnvironmentTypeWindowsServer2019Container)),
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

func TestAccCodeBuildProject_armContainer(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_armContainer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
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

func TestAccCodeBuildProject_linuxLambdaContainer(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_linuxLambdaContainer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "environment.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "environment.0.compute_type", string(types.ComputeTypeBuildLambda1gb)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "environment.0.image", "aws/codebuild/amazonlinux-x86_64-lambda-standard:go1.21"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.privileged_mode", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "environment.0.image_pull_credentials_type", string(types.ImagePullCredentialsTypeCodebuild)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.type", string(types.EnvironmentTypeLinuxLambdaContainer)),
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

func TestAccCodeBuildProject_Artifacts_artifactIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	artifactIdentifier1 := "artifactIdentifier1"
	artifactIdentifier2 := "artifactIdentifier2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_artifactsArtifactIdentifier(rName, artifactIdentifier1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.artifact_identifier", artifactIdentifier1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_artifactsArtifactIdentifier(rName, artifactIdentifier2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.artifact_identifier", artifactIdentifier2),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_encryptionDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_artifactsEncryptionDisabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.encryption_disabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_artifactsEncryptionDisabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.encryption_disabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_location(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_artifactsLocation(rName1, rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.location", rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_artifactsLocation(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.location", rName2),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_name(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	name1 := "name1"
	name2 := "name2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_artifactsName(rName, name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.name", name1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_artifactsName(rName, name2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.name", name2),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_namespaceType(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_artifactsNamespaceType(rName, string(types.ArtifactNamespaceBuildId)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.namespace_type", string(types.ArtifactNamespaceBuildId)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_artifactsNamespaceType(rName, string(types.ArtifactNamespaceNone)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.namespace_type", string(types.ArtifactNamespaceNone)),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_overrideArtifactName(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_artifactsOverrideArtifactName(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.override_artifact_name", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_artifactsOverrideArtifactName(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.override_artifact_name", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_packaging(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_artifactsPackaging(rName, string(types.ArtifactPackagingZip)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.packaging", string(types.ArtifactPackagingZip)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_artifactsPackaging(rName, string(types.ArtifactPackagingNone)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.packaging", string(types.ArtifactPackagingNone)),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_path(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_artifactsPath(rName, "path1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.path", "path1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_artifactsPath(rName, "path2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.path", "path2"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_type(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	type1 := string(types.ArtifactsTypeS3)
	type2 := string(types.ArtifactsTypeCodepipeline)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_artifactsType(rName, type1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.type", type1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_artifactsType(rName, type2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.type", type2),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_bucketOwnerAccess(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_artifactsBucketOwnerAccess(rName, "FULL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.bucket_owner_access", "FULL"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_artifactsBucketOwnerAccess(rName, "READ_ONLY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.bucket_owner_access", "READ_ONLY"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_secondaryArtifacts(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_secondaryArtifacts(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct2),
				),
			},
			{
				Config: testAccProjectConfig_secondaryArtifactsNone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_artifactIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	artifactIdentifier1 := "artifactIdentifier1"
	artifactIdentifier2 := "artifactIdentifier2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_secondaryArtifactsArtifactIdentifier(rName, artifactIdentifier1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"artifact_identifier": artifactIdentifier1,
					}),
				),
			},
			{
				Config: testAccProjectConfig_secondaryArtifactsArtifactIdentifier(rName, artifactIdentifier2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"artifact_identifier": artifactIdentifier2,
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_overrideArtifactName(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_secondaryArtifactsOverrideArtifactName(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"override_artifact_name": acctest.CtTrue,
					}),
				),
			},
			{
				Config: testAccProjectConfig_secondaryArtifactsOverrideArtifactName(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"override_artifact_name": acctest.CtFalse,
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_encryptionDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_secondaryArtifactsEncryptionDisabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"encryption_disabled": acctest.CtTrue,
					}),
				),
			},
			{
				Config: testAccProjectConfig_secondaryArtifactsEncryptionDisabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"encryption_disabled": acctest.CtFalse,
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_location(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_secondaryArtifactsLocation(rName1, rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						names.AttrLocation: rName1,
					}),
				),
			},
			{
				Config: testAccProjectConfig_secondaryArtifactsLocation(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						names.AttrLocation: rName2,
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_name(t *testing.T) {
	acctest.Skip(t, "Currently no solution to allow updates on name attribute")

	ctx := acctest.Context(t)

	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	name1 := "name1"
	name2 := "name2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_secondaryArtifactsName(rName, name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						names.AttrName: name1,
					}),
				),
			},
			{
				Config: testAccProjectConfig_secondaryArtifactsName(rName, name2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						names.AttrName: name2,
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_namespaceType(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_secondaryArtifactsNamespaceType(rName, string(types.ArtifactNamespaceBuildId)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"namespace_type": string(types.ArtifactNamespaceBuildId),
					}),
				),
			},
			{
				Config: testAccProjectConfig_secondaryArtifactsNamespaceType(rName, string(types.ArtifactNamespaceNone)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"namespace_type": string(types.ArtifactNamespaceNone),
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_path(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	path1 := "path1"
	path2 := "path2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_secondaryArtifactsPath(rName, path1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						names.AttrPath: path1,
					}),
				),
			},
			{
				Config: testAccProjectConfig_secondaryArtifactsPath(rName, path2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						names.AttrPath: path2,
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_packaging(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_secondaryArtifactsPackaging(rName, string(types.ArtifactPackagingZip)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"packaging": string(types.ArtifactPackagingZip),
					}),
				),
			},
			{
				Config: testAccProjectConfig_secondaryArtifactsPackaging(rName, string(types.ArtifactPackagingNone)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"packaging": string(types.ArtifactPackagingNone),
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_type(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_secondaryArtifactsType(rName, string(types.ArtifactsTypeS3)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						names.AttrType: string(types.ArtifactsTypeS3),
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondarySources_codeCommit(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_secondarySourcesCodeCommit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.type", "CODECOMMIT"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource2",
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_concurrentBuildLimit(t *testing.T) {
	ctx := acctest.Context(t)
	var project types.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckSourceCredentialsForServerType(ctx, t, types.ServerTypeGithub)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_concurrentBuildLimit(rName, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "concurrent_build_limit", acctest.Ct4),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_concurrentBuildLimit(rName, 12),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "concurrent_build_limit", "12"),
				),
			},
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "concurrent_build_limit", acctest.Ct0),
				),
			},
		},
	})
}

func testAccCheckProjectExists(ctx context.Context, n string, v *types.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildClient(ctx)

		output, err := tfcodebuild.FindProjectByNameOrARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckProjectDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codebuild_project" {
				continue
			}

			_, err := tfcodebuild.FindProjectByNameOrARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CodeBuild Project %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProjectCertificate(project *types.Project, expectedCertificate string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(project.Environment.Certificate) != expectedCertificate {
			return fmt.Errorf("CodeBuild Project certificate (%s) did not match: %s", aws.ToString(project.Environment.Certificate), expectedCertificate)
		}
		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildClient(ctx)

	_, err := tfcodebuild.FindProjectByNameOrARN(ctx, conn, "tf-acc-test-precheck")

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil && !tfresource.NotFound(err) {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccProjectConfig_baseServiceRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "codebuild.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": "*",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ]
    },
    {
      "Effect": "Allow",
      "Resource": "*",
      "Action": [
        "s3:PutObject",
        "s3:GetObject",
        "s3:GetObjectVersion",
        "s3:GetBucketAcl",
        "s3:PutBucketAcl",
        "s3:GetBucketLocation"
      ]
    },
    {
      "Effect": "Allow",
      "Resource": "*",
      "Action": [
        "codebuild:CreateReportGroup",
        "codebuild:CreateReport",
        "codebuild:UpdateReport",
        "codebuild:BatchPutTestCases",
        "codebuild:BatchPutCodeCoverages"
      ]
    },
    {
      "Effect": "Allow",
      "Resource": "*",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:CreateNetworkInterfacePermission",
        "ec2:DescribeDhcpOptions",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface",
        "ec2:DescribeSubnets",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeVpcs"
      ]
    }
  ]
}
POLICY
}
`, rName)
}

func testAccProjectConfig_basic(rName string) string {
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
    location = %[2]q
    type     = "GITHUB"
  }
}
`, rName, testAccGitHubSourceLocationFromEnv()))
}

func testAccProjectConfig_tags1(rName, tagKey1, tagValue1 string) string {
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
    location = "https://github.com/hashicorp/packer.git"
    type     = "GITHUB"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccProjectConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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
    location = "https://github.com/hashicorp/packer.git"
    type     = "GITHUB"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccProjectConfig_visibility(rName, visibility string) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name               = %[1]q
  service_role       = aws_iam_role.test.arn
  project_visibility = %[3]q

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = %[2]q
    type     = "GITHUB"
  }
}
`, rName, testAccGitHubSourceLocationFromEnv(), visibility))
}

func testAccProjectConfig_visibilityResourceRole(rName, visibility string) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name                 = %[1]q
  service_role         = aws_iam_role.test.arn
  project_visibility   = %[3]q
  resource_access_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = %[2]q
    type     = "GITHUB"
  }
}
`, rName, testAccGitHubSourceLocationFromEnv(), visibility))
}

func testAccProjectConfig_badgeEnabled(rName string, badgeEnabled bool) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  badge_enabled = %[1]t
  name          = %[2]q
  service_role  = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, badgeEnabled, rName))
}

func testAccProjectConfig_buildTimeout(rName string, buildTimeout int) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  build_timeout = %[1]d
  name          = %[2]q
  service_role  = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, buildTimeout, rName))
}

func testAccProjectConfig_queuedTimeout(rName string, queuedTimeout int) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  queued_timeout = %[1]d
  name           = %[2]q
  service_role   = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, queuedTimeout, rName))
}

func testAccProjectConfig_cache(rName, cacheLocation, cacheType string) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test1" {
  bucket        = "%[1]s-1"
  force_destroy = true
}

resource "aws_s3_bucket" "test2" {
  bucket        = "%[1]s-2"
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  cache {
    location = %[2]q
    type     = %[3]q
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }

  depends_on = [aws_s3_bucket.test1, aws_s3_bucket.test2]
}
`, rName, cacheLocation, cacheType))
}

func testAccProjectConfig_localCache(rName, modeType string) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  cache {
    type  = "LOCAL"
    modes = [%[2]q]
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, modeType))
}

func testAccProjectConfig_s3ComputedLocation(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket_prefix = "cache"
}

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
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }

  cache {
    type     = "S3"
    location = aws_s3_bucket.test.bucket
  }
}
`, rName))
}

func testAccProjectConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  description  = %[1]q
  name         = %[2]q
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
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, description, rName))
}

func testAccProjectConfig_fileSystemLocations(rName, mountPoint string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "public" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "public" {
  route_table_id = aws_route_table.public.id
  subnet_id      = aws_subnet.public.id
}

resource "aws_route" "public" {
  route_table_id         = aws_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
}

resource "aws_subnet" "private" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.public.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_route.public]
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "private" {
  route_table_id = aws_route.private.route_table_id
  subnet_id      = aws_subnet.private.id
}

resource "aws_route" "private" {
  route_table_id         = aws_route_table.private.id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.test.id
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

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

    privileged_mode = true
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.private.id]
    vpc_id             = aws_vpc.test.id
  }

  file_system_locations {
    identifier    = "test"
    location      = "${aws_efs_file_system.test.dns_name}:/directory-path"
    type          = "EFS"
    mount_point   = %[2]q
    mount_options = "nfsvers=4.1,rsize=1048576,wsize=1048576,hard,timeo=450,retrans=3"
  }
}
`, rName, mountPoint))
}

func testAccProjectConfig_sourceVersion(rName, sourceVersion string) string {
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

  source_version = %[2]q

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, sourceVersion))
}

func testAccProjectConfig_encryptionKey(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codebuild_project" "test" {
  encryption_key = aws_kms_key.test.arn
  name           = %[1]q
  service_role   = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName))
}

func testAccProjectConfig_environmentVariableOne(rName, key1, value1 string) string {
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

    environment_variable {
      name  = %[2]q
      value = %[3]q
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, key1, value1))
}

func testAccProjectConfig_environmentVariableTwo(rName, key1, value1, key2, value2 string) string {
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

    environment_variable {
      name  = %[2]q
      value = %[3]q
    }

    environment_variable {
      name  = %[4]q
      value = %[5]q
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, key1, value1, key2, value2))
}

func testAccProjectConfig_environmentVariableZero(rName string) string {
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
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName))
}

func testAccProjectConfig_environmentVariableType(rName, environmentVariableType string) string {
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

    environment_variable {
      name  = "SOME_KEY"
      value = "SOME_VALUE"
    }

    environment_variable {
      name  = "SOME_KEY2"
      value = "SOME_VALUE2"
      type  = %[2]q
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, environmentVariableType))
}

func testAccProjectConfig_environmentCertificate(rName string, oName string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = %[1]q
  content = "test"
}

resource "aws_codebuild_project" "test" {
  name         = %[2]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
    certificate  = "${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, oName, rName))
}

func testAccProjectConfig_environmentRegistryCredential1(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type                = "BUILD_GENERAL1_SMALL"
    image                       = "2"
    type                        = "LINUX_CONTAINER"
    image_pull_credentials_type = "SERVICE_ROLE"

    registry_credential {
      credential          = aws_secretsmanager_secret_version.test.arn
      credential_provider = "SECRETS_MANAGER"
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}

resource "aws_secretsmanager_secret" "test" {
  name                    = "%[1]s-1"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ username : "user", password : "pass" })
}
`, rName))
}

func testAccProjectConfig_environmentRegistryCredential2(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type                = "BUILD_GENERAL1_SMALL"
    image                       = "2"
    type                        = "LINUX_CONTAINER"
    image_pull_credentials_type = "SERVICE_ROLE"

    registry_credential {
      credential          = aws_secretsmanager_secret_version.test.arn
      credential_provider = "SECRETS_MANAGER"
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}

resource "aws_secretsmanager_secret" "test" {
  name                    = "%[1]s-2"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ username : "user", password : "pass" })
}
`, rName))
}

func testAccProjectConfig_cloudWatchLogs(rName, status, gName, sName string) string {
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
    location = "https://github.com/hashicorp/packer.git"
    type     = "GITHUB"
  }

  logs_config {
    cloudwatch_logs {
      status      = %[2]q
      group_name  = %[3]q
      stream_name = %[4]q
    }
  }
}
`, rName, status, gName, sName))
}

func testAccProjectConfig_s3Logs(rName, status, location string, encryptionDisabled bool) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

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
    location = "https://github.com/hashicorp/packer.git"
    type     = "GITHUB"
  }

  logs_config {
    s3_logs {
      status              = %[2]q
      location            = %[3]q
      encryption_disabled = %[4]t
    }
  }
}
`, rName, status, location, encryptionDisabled))
}

func testAccProjectConfig_buildBatch(rName string, combineArtifacts bool, computeTypesAllowed string, maximumBuildsAllowed, timeoutInMins int) string {
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
    location = "https://github.com/hashicorp/packer.git"
    type     = "GITHUB"
  }

  build_batch_config {
    combine_artifacts = %[2]t

    restrictions {
      compute_types_allowed  = [%[3]q]
      maximum_builds_allowed = %[4]d
    }

    service_role    = aws_iam_role.test.arn
    timeout_in_mins = %[5]d
  }
}
`, rName, combineArtifacts, computeTypesAllowed, maximumBuildsAllowed, timeoutInMins))
}

func testAccProjectConfig_buildBatchConfigDelete(rName string, withBuildBatchConfig bool) string {
	template := `
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
    location = "https://github.com/hashicorp/packer.git"
    type     = "GITHUB"
  }

	%[2]s
}
	`

	buildBatchConfig := `
build_batch_config {
  combine_artifacts = true

  restrictions {
    compute_types_allowed  = []
    maximum_builds_allowed = 10
  }

  service_role    = aws_iam_role.test.arn
  timeout_in_mins = 2160
}
`

	if withBuildBatchConfig {
		return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(template, rName, buildBatchConfig))
	}
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(template, rName, ""))
}

func testAccProjectConfig_sourceGitCloneDepth(rName string, gitCloneDepth int) string {
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
    git_clone_depth = %[2]d
    location        = "https://github.com/hashicorp/packer.git"
    type            = "GITHUB"
  }
}
`, rName, gitCloneDepth))
}

func testAccProjectConfig_sourceGitSubmodulesCodeCommit(rName string, fetchSubmodules bool) string {
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
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/repo-name"
    type     = "CODECOMMIT"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }
}
`, rName, fetchSubmodules))
}

func testAccProjectConfig_sourceGitSubmodulesGitHub(rName string, fetchSubmodules bool) string {
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
    location = "https://github.com/hashicorp/packer.git"
    type     = "GITHUB"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }
}
`, rName, fetchSubmodules))
}

func testAccProjectConfig_sourceGitSubmodulesGitHubEnterprise(rName string, fetchSubmodules bool) string {
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
    location = "https://example.com/organization/repository.git"
    type     = "GITHUB_ENTERPRISE"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }
}
`, rName, fetchSubmodules))
}

func testAccProjectConfig_secondarySourcesGitSubmodulesCodeCommit(rName string, fetchSubmodules bool) string {
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
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/repo-name"
    type     = "CODECOMMIT"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }

  secondary_sources {
    location          = "https://git-codecommit.region-id.amazonaws.com/v1/repos/second-repo-name"
    type              = "CODECOMMIT"
    source_identifier = "secondarySource1"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }

  secondary_sources {
    location          = "https://git-codecommit.region-id.amazonaws.com/v1/repos/third-repo-name"
    type              = "CODECOMMIT"
    source_identifier = "secondarySource2"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }
}
`, rName, fetchSubmodules))
}

func testAccProjectConfig_secondarySourcesNone(rName string, fetchSubmodules bool) string {
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
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/repo-name"
    type     = "CODECOMMIT"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }
}
`, rName, fetchSubmodules))
}

func testAccProjectConfig_secondarySourcesGitSubmodulesGitHub(rName string, fetchSubmodules bool) string {
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
    location = "https://github.com/hashicorp/packer.git"
    type     = "GITHUB"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }

  secondary_sources {
    location          = "https://github.com/hashicorp/terraform.git"
    type              = "GITHUB"
    source_identifier = "secondarySource1"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }

  secondary_sources {
    location          = "https://github.com/hashicorp/vault.git"
    type              = "GITHUB"
    source_identifier = "secondarySource2"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }
}
`, rName, fetchSubmodules))
}

func testAccProjectConfig_secondarySourcesGitSubmodulesGitHubEnterprise(rName string, fetchSubmodules bool) string {
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
    location = "https://example.com/organization/repository-1.git"
    type     = "GITHUB_ENTERPRISE"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }

  secondary_sources {
    location          = "https://example.com/organization/repository-2.git"
    type              = "GITHUB_ENTERPRISE"
    source_identifier = "secondarySource1"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }

  secondary_sources {
    location          = "https://example.com/organization/repository-3.git"
    type              = "GITHUB_ENTERPRISE"
    source_identifier = "secondarySource2"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }
}
`, rName, fetchSubmodules))
}

func testAccProjectConfig_secondarySourceVersionsCodeCommit(rName string) string {
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
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/repo-name"
    type     = "CODECOMMIT"
  }

  secondary_sources {
    location          = "https://git-codecommit.region-id.amazonaws.com/v1/repos/second-repo-name"
    type              = "CODECOMMIT"
    source_identifier = "secondarySource1"
  }

  secondary_sources {
    location          = "https://git-codecommit.region-id.amazonaws.com/v1/repos/third-repo-name"
    type              = "CODECOMMIT"
    source_identifier = "secondarySource2"
  }

  secondary_source_version {
    source_version    = "master"
    source_identifier = "secondarySource1"
  }
}
`, rName))
}

func testAccProjectConfig_secondarySourceVersionsCodeCommitUpdated(rName string) string {
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
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/repo-name"
    type     = "CODECOMMIT"
  }

  secondary_sources {
    location          = "https://git-codecommit.region-id.amazonaws.com/v1/repos/second-repo-name"
    type              = "CODECOMMIT"
    source_identifier = "secondarySource1"
  }

  secondary_sources {
    location          = "https://git-codecommit.region-id.amazonaws.com/v1/repos/third-repo-name"
    type              = "CODECOMMIT"
    source_identifier = "secondarySource2"
  }

  secondary_source_version {
    source_version    = "master"
    source_identifier = "secondarySource1"
  }

  secondary_source_version {
    source_version    = "master"
    source_identifier = "secondarySource2"
  }
}
`, rName))
}

func testAccProjectConfig_sourceInsecureSSL(rName string, insecureSSL bool) string {
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
    insecure_ssl = %[2]t
    location     = "https://github.com/hashicorp/packer.git"
    type         = "GITHUB"
  }
}
`, rName, insecureSSL))
}

func testAccProjectConfig_sourceBuildStatusGitHubEnterprise(rName string) string {
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
    location = "https://example.com/organization/repository.git"
    type     = "GITHUB_ENTERPRISE"

    build_status_config {
      context    = "codebuild"
      target_url = "https://example.com/$${CODEBUILD_BUILD_ID}"
    }
  }
}
`, rName))
}

func testAccProjectConfig_sourceReportBuildStatusGitHubEnterprise(rName string, reportBuildStatus bool) string {
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
    location            = "https://example.com/organization/repository.git"
    report_build_status = %[2]t
    type                = "GITHUB_ENTERPRISE"
  }
}
`, rName, reportBuildStatus))
}

func testAccProjectConfig_sourceReportBuildStatusBitbucket(rName, sourceLocation string, reportBuildStatus bool) string {
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
    location            = %[2]q
    report_build_status = %[3]t
    type                = "BITBUCKET"
  }
}
`, rName, sourceLocation, reportBuildStatus))
}

func testAccProjectConfig_sourceReportBuildStatusGitHub(rName string, reportBuildStatus bool) string {
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
    location            = "https://github.com/hashicorp/packer.git"
    report_build_status = %[2]t
    type                = "GITHUB"
  }
}
`, rName, reportBuildStatus))
}

func testAccProjectConfig_sourceTypeBitbucket(rName, sourceLocation string) string {
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
    location = %[2]q
    type     = "BITBUCKET"
  }
}
`, rName, sourceLocation))
}

func testAccProjectConfig_sourceTypeCodeCommit(rName string) string {
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
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/repo-name"
    type     = "CODECOMMIT"
  }
}
`, rName))
}

func testAccProjectConfig_sourceTypeCodePipeline(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "CODEPIPELINE"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type = "CODEPIPELINE"
  }
}
`, rName))
}

func testAccProjectConfig_sourceTypeGitHubEnterprise(rName string) string {
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
    location = "https://example.com/organization/repository.git"
    type     = "GITHUB_ENTERPRISE"
  }
}
`, rName))
}

func testAccProjectConfig_sourceTypeS3(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  content = "test"
  key     = "test.txt"
}

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
    location = "${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    type     = "S3"
  }
}
`, rName))
}

func testAccProjectConfig_sourceTypeNoSource(rName string, rLocation string, rBuildspec string) string {
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
    type      = "NO_SOURCE"
    location  = %[2]q
    buildspec = %[3]q
  }
}
`, rName, rLocation, rBuildspec))
}

func testAccProjectConfig_baseVPC(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccProjectConfig_vpc1(rName string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		testAccProjectConfig_baseVPC(rName),
		fmt.Sprintf(`
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
    location = "https://github.com/hashicorp/packer.git"
    type     = "GITHUB"
  }

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.test[0].id]
    vpc_id             = aws_vpc.test.id
  }
}
`, rName))
}

func testAccProjectConfig_vpc2(rName string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		testAccProjectConfig_baseVPC(rName),
		fmt.Sprintf(`
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
    location = "https://github.com/hashicorp/packer.git"
    type     = "GITHUB"
  }

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnets            = aws_subnet.test[*].id
    vpc_id             = aws_vpc.test.id
  }
}
`, rName))
}

func testAccProjectConfig_windowsServer2019Container(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_MEDIUM"
    image        = "2"
    type         = "WINDOWS_SERVER_2019_CONTAINER"
  }

  source {
    location = %[2]q
    type     = "GITHUB"
  }
}
`, rName, testAccGitHubSourceLocationFromEnv()))
}

func testAccProjectConfig_armContainer(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_LARGE"
    image        = "2"
    type         = "ARM_CONTAINER"
  }

  source {
    location = %[2]q
    type     = "GITHUB"
  }
}
`, rName, testAccGitHubSourceLocationFromEnv()))
}

func testAccProjectConfig_linuxLambdaContainer(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_LAMBDA_1GB"
    image        = "aws/codebuild/amazonlinux-x86_64-lambda-standard:go1.21"
    type         = "LINUX_LAMBDA_CONTAINER"
  }

  source {
    location = %[2]q
    type     = "GITHUB"
  }
}
`, rName, testAccGitHubSourceLocationFromEnv()))
}

func testAccProjectConfig_artifactsArtifactIdentifier(rName string, artifactIdentifier string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    artifact_identifier = %[2]q
    location            = aws_s3_bucket.test.bucket
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, artifactIdentifier))
}

func testAccProjectConfig_artifactsEncryptionDisabled(rName string, encryptionDisabled bool) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    encryption_disabled = %[2]t
    location            = aws_s3_bucket.test.bucket
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, encryptionDisabled))
}

func testAccProjectConfig_artifactsLocation(rName, bucketName string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, bucketName))
}

func testAccProjectConfig_artifactsName(rName string, name string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    name     = %[2]q
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, name))
}

func testAccProjectConfig_artifactsNamespaceType(rName, namespaceType string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location       = aws_s3_bucket.test.bucket
    namespace_type = %[2]q
    type           = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, namespaceType))
}

func testAccProjectConfig_artifactsOverrideArtifactName(rName string, overrideArtifactName bool) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    override_artifact_name = %[2]t
    location               = aws_s3_bucket.test.bucket
    type                   = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, overrideArtifactName))
}

func testAccProjectConfig_artifactsPackaging(rName, packaging string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location  = aws_s3_bucket.test.bucket
    packaging = %[2]q
    type      = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, packaging))
}

func testAccProjectConfig_artifactsPath(rName, path string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    path     = %[2]q
    type     = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, path))
}

func testAccProjectConfig_artifactsType(rName string, artifactType string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type     = %[2]q
    location = aws_s3_bucket.test.bucket
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = %[2]q
    location = "${aws_s3_bucket.test.bucket}/"
  }
}
`, rName, artifactType))
}

func testAccProjectConfig_artifactsBucketOwnerAccess(rName string, typ string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type                = "S3"
    location            = aws_s3_bucket.test.bucket
    bucket_owner_access = %[2]q
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "S3"
    location = "${aws_s3_bucket.test.bucket}/"
  }
}
`, rName, typ))
}

func testAccProjectConfig_secondaryArtifacts(rName string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    location            = aws_s3_bucket.test.bucket
    type                = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact2"
    location            = aws_s3_bucket.test.bucket
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName))
}

func testAccProjectConfig_secondaryArtifactsNone(rName string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName))
}

func testAccProjectConfig_secondaryArtifactsArtifactIdentifier(rName string, artifactIdentifier string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = %[2]q
    location            = aws_s3_bucket.test.bucket
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, artifactIdentifier))
}

func testAccProjectConfig_secondaryArtifactsEncryptionDisabled(rName string, encryptionDisabled bool) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    encryption_disabled = %[2]t
    location            = aws_s3_bucket.test.bucket
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, encryptionDisabled))
}

func testAccProjectConfig_secondaryArtifactsLocation(rName, bucketName string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    location            = aws_s3_bucket.test.bucket
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, bucketName))
}

func testAccProjectConfig_secondaryArtifactsName(rName string, name string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    name                = %[2]q
    location            = aws_s3_bucket.test.bucket
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, name))
}

func testAccProjectConfig_secondaryArtifactsNamespaceType(rName, namespaceType string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    namespace_type      = %[2]q
    location            = aws_s3_bucket.test.bucket
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, namespaceType))
}

func testAccProjectConfig_secondaryArtifactsOverrideArtifactName(rName string, overrideArtifactName bool) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier    = "secondaryArtifact1"
    override_artifact_name = %[2]t
    location               = aws_s3_bucket.test.bucket
    type                   = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, overrideArtifactName))
}

func testAccProjectConfig_secondaryArtifactsPath(rName, path string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    path                = %[2]q
    location            = aws_s3_bucket.test.bucket
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, path))
}

func testAccProjectConfig_secondaryArtifactsPackaging(rName, packaging string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    packaging           = %[2]q
    location            = aws_s3_bucket.test.bucket
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, packaging))
}

func testAccProjectConfig_secondaryArtifactsType(rName string, artifactType string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_baseServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    type                = %[2]q
    location            = aws_s3_bucket.test.bucket
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/repo-name"
    type     = "CODECOMMIT"
  }
}
`, rName, artifactType))
}

func testAccProjectConfig_secondarySourcesCodeCommit(rName string) string {
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
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/repo-name"
    type     = "CODECOMMIT"
  }

  secondary_sources {
    location          = "https://git-codecommit.region-id.amazonaws.com/v1/repos/second-repo-name"
    type              = "CODECOMMIT"
    source_identifier = "secondarySource1"
  }

  secondary_sources {
    location          = "https://git-codecommit.region-id.amazonaws.com/v1/repos/third-repo-name"
    type              = "CODECOMMIT"
    source_identifier = "secondarySource2"
  }
}
`, rName))
}

func testAccProjectConfig_concurrentBuildLimit(rName string, concurrentBuildLimit int) string {
	return acctest.ConfigCompose(testAccProjectConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  concurrent_build_limit = %[1]d
  name                   = %[2]q
  service_role           = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, concurrentBuildLimit, rName))
}
