// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
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

func TestAccCodeBuildSourceCredential_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var sourceCredentialsInfo types.SourceCredentialsInfo
	token := sdkacctest.RandomWithPrefix("token")
	resourceName := "aws_codebuild_source_credential.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSourceCredentialDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSourceCredentialConfig_basic("PERSONAL_ACCESS_TOKEN", "GITHUB", token),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceCredentialExists(ctx, resourceName, &sourceCredentialsInfo),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "codebuild", regexache.MustCompile(`token/github`)),
					resource.TestCheckResourceAttr(resourceName, "server_type", "GITHUB"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "PERSONAL_ACCESS_TOKEN"),
				),
			},
			{
				Config: testAccSourceCredentialConfig_basic("PERSONAL_ACCESS_TOKEN", "GITHUB_ENTERPRISE", token),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceCredentialExists(ctx, resourceName, &sourceCredentialsInfo),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "codebuild", regexache.MustCompile(`token/github_enterprise`)),
					resource.TestCheckResourceAttr(resourceName, "server_type", "GITHUB_ENTERPRISE"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "PERSONAL_ACCESS_TOKEN"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", names.AttrUserName},
			},
		},
	})
}

func TestAccCodeBuildSourceCredential_basicAuth(t *testing.T) {
	ctx := acctest.Context(t)
	var sourceCredentialsInfo types.SourceCredentialsInfo
	token := sdkacctest.RandomWithPrefix("token")
	resourceName := "aws_codebuild_source_credential.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSourceCredentialDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSourceCredentialConfig_basicAuth(token, "user1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceCredentialExists(ctx, resourceName, &sourceCredentialsInfo),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "codebuild", regexache.MustCompile(`token/bitbucket`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrUserName, "user1"),
					resource.TestCheckResourceAttr(resourceName, "server_type", "BITBUCKET"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "BASIC_AUTH"),
				),
			},
			{
				Config: testAccSourceCredentialConfig_basicAuth(token, "user2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceCredentialExists(ctx, resourceName, &sourceCredentialsInfo),
					resource.TestCheckResourceAttr(resourceName, names.AttrUserName, "user2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", names.AttrUserName},
			},
		},
	})
}

func TestAccCodeBuildSourceCredential_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var sourceCredentialsInfo types.SourceCredentialsInfo
	token := sdkacctest.RandomWithPrefix("token")
	resourceName := "aws_codebuild_source_credential.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSourceCredentialDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSourceCredentialConfig_basic("PERSONAL_ACCESS_TOKEN", "GITHUB_ENTERPRISE", token),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceCredentialExists(ctx, resourceName, &sourceCredentialsInfo),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodebuild.ResourceSourceCredential(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPreCheckSourceCredentialsForServerType(ctx context.Context, t *testing.T, serverType types.ServerType) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildClient(ctx)

	input := &codebuild.ListSourceCredentialsInput{}
	output, err := tfcodebuild.FindSourceCredentials(ctx, conn, input, func(v *types.SourceCredentialsInfo) bool {
		return v.ServerType == serverType
	})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if len(output) == 0 {
		t.Skipf("skipping acceptance testing: Source Credentials (%s) not found", serverType)
	}
}

func testAccCheckSourceCredentialDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codebuild_source_credential" {
				continue
			}

			_, err := tfcodebuild.FindSourceCredentialsByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CodeBuild Source Credential %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSourceCredentialExists(ctx context.Context, n string, v *types.SourceCredentialsInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildClient(ctx)

		output, err := tfcodebuild.FindSourceCredentialsByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSourceCredentialConfig_basic(authType, serverType, token string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_source_credential" "test" {
  auth_type   = "%s"
  server_type = "%s"
  token       = "%s"
}
`, authType, serverType, token)
}

func testAccSourceCredentialConfig_basicAuth(token, userName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_source_credential" "test" {
  auth_type   = "BASIC_AUTH"
  server_type = "BITBUCKET"
  token       = "%s"
  user_name   = "%s"
}
`, token, userName)
}
