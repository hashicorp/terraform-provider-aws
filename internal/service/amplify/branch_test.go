// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amplify_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/amplify/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfamplify "github.com/hashicorp/terraform-provider-aws/internal/service/amplify"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccBranch_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var branch types.Branch
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_branch.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBranchDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBranchConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBranchExists(ctx, resourceName, &branch),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "amplify", regexache.MustCompile(`apps/.+/branches/.+`)),
					resource.TestCheckResourceAttr(resourceName, "associated_resources.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "backend_environment_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "branch_name", rName),
					resource.TestCheckResourceAttr(resourceName, "custom_domains.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "destination_branch", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_build", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_notification", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_performance_mode", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_pull_request_preview", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "framework", ""),
					resource.TestCheckResourceAttr(resourceName, "pull_request_environment_name", ""),
					resource.TestCheckResourceAttr(resourceName, "source_branch", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrStage, "NONE"),
					resource.TestCheckResourceAttr(resourceName, "ttl", "5"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func testAccBranch_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var branch types.Branch
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_branch.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBranchDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBranchConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBranchExists(ctx, resourceName, &branch),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfamplify.ResourceBranch(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccBranch_BasicAuthCredentials(t *testing.T) {
	ctx := acctest.Context(t)
	var branch types.Branch
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_branch.test"

	credentials1 := base64.StdEncoding.EncodeToString([]byte("username1:password1"))
	credentials2 := base64.StdEncoding.EncodeToString([]byte("username2:password2"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBranchDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBranchConfig_basicAuthCredentials(rName, credentials1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBranchExists(ctx, resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", credentials1),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBranchConfig_basicAuthCredentials(rName, credentials2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBranchExists(ctx, resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", credentials2),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", acctest.CtTrue),
				),
			},
			{
				Config: testAccBranchConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBranchExists(ctx, resourceName, &branch),
					// Clearing basic_auth_credentials not reflected in API.
					// resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccBranch_EnvironmentVariables(t *testing.T) {
	ctx := acctest.Context(t)
	var branch types.Branch
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_branch.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBranchDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBranchConfig_environmentVariables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBranchExists(ctx, resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENVVAR1", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBranchConfig_environmentVariablesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBranchExists(ctx, resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENVVAR1", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENVVAR2", acctest.Ct2),
				),
			},
			{
				Config: testAccBranchConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBranchExists(ctx, resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", acctest.Ct0),
				),
			},
		},
	})
}

func testAccBranch_OptionalArguments(t *testing.T) {
	ctx := acctest.Context(t)
	var branch types.Branch
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	environmentName := sdkacctest.RandStringFromCharSet(9, sdkacctest.CharSetAlpha)
	resourceName := "aws_amplify_branch.test"
	backendEnvironment1ResourceName := "aws_amplify_backend_environment.test1"
	backendEnvironment2ResourceName := "aws_amplify_backend_environment.test2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBranchDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBranchConfig_optionalArguments(rName, environmentName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBranchExists(ctx, resourceName, &branch),
					resource.TestCheckResourceAttrPair(resourceName, "backend_environment_arn", backendEnvironment1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "testdescription1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "testdisplayname1"),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_build", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_notification", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_performance_mode", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_pull_request_preview", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "framework", "React"),
					resource.TestCheckResourceAttr(resourceName, "pull_request_environment_name", "testpr1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStage, "DEVELOPMENT"),
					resource.TestCheckResourceAttr(resourceName, "ttl", acctest.Ct10),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBranchConfig_optionalArgumentsUpdated(rName, environmentName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBranchExists(ctx, resourceName, &branch),
					resource.TestCheckResourceAttrPair(resourceName, "backend_environment_arn", backendEnvironment2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "testdescription2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "testdisplayname2"),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_build", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_notification", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_performance_mode", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_pull_request_preview", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "framework", "Angular"),
					resource.TestCheckResourceAttr(resourceName, "pull_request_environment_name", "testpr2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStage, "EXPERIMENTAL"),
					resource.TestCheckResourceAttr(resourceName, "ttl", "15"),
				),
			},
		},
	})
}

func testAccCheckBranchExists(ctx context.Context, resourceName string, v *types.Branch) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyClient(ctx)

		output, err := tfamplify.FindBranchByTwoPartKey(ctx, conn, rs.Primary.Attributes["app_id"], rs.Primary.Attributes["branch_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckBranchDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_amplify_branch" {
				continue
			}

			_, err := tfamplify.FindBranchByTwoPartKey(ctx, conn, rs.Primary.Attributes["app_id"], rs.Primary.Attributes["branch_name"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Amplify Branch %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBranchConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q
}
`, rName)
}

func testAccBranchConfig_basicAuthCredentials(rName, basicAuthCredentials string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  basic_auth_credentials = %[2]q
  enable_basic_auth      = true
}
`, rName, basicAuthCredentials)
}

func testAccBranchConfig_environmentVariables(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  environment_variables = {
    ENVVAR1 = "1"
  }
}
`, rName)
}

func testAccBranchConfig_environmentVariablesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  environment_variables = {
    ENVVAR1 = "2",
    ENVVAR2 = "2"
  }
}
`, rName)
}

func testAccBranchConfig_optionalArguments(rName, environmentName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_backend_environment" "test1" {
  app_id           = aws_amplify_app.test.id
  environment_name = "%[2]sa"
}

resource "aws_amplify_backend_environment" "test2" {
  app_id           = aws_amplify_app.test.id
  environment_name = "%[2]sb"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  backend_environment_arn       = aws_amplify_backend_environment.test1.arn
  description                   = "testdescription1"
  display_name                  = "testdisplayname1"
  enable_auto_build             = false
  enable_notification           = true
  enable_performance_mode       = true
  enable_pull_request_preview   = false
  framework                     = "React"
  pull_request_environment_name = "testpr1"
  stage                         = "DEVELOPMENT"
  ttl                           = "10"
}
`, rName, environmentName)
}

func testAccBranchConfig_optionalArgumentsUpdated(rName, environmentName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_backend_environment" "test1" {
  app_id           = aws_amplify_app.test.id
  environment_name = "%[2]sa"
}

resource "aws_amplify_backend_environment" "test2" {
  app_id           = aws_amplify_app.test.id
  environment_name = "%[2]sb"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  backend_environment_arn       = aws_amplify_backend_environment.test2.arn
  description                   = "testdescription2"
  display_name                  = "testdisplayname2"
  enable_auto_build             = true
  enable_notification           = false
  enable_performance_mode       = false
  enable_pull_request_preview   = true
  framework                     = "Angular"
  pull_request_environment_name = "testpr2"
  stage                         = "EXPERIMENTAL"
  ttl                           = "15"
}
`, rName, environmentName)
}
