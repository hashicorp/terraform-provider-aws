// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecatalyst_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codecatalyst"
	"github.com/aws/aws-sdk-go-v2/service/codecatalyst/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfcodecatalyst "github.com/hashicorp/terraform-provider-aws/internal/service/codecatalyst"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeCatalystDevEnvironment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var DevEnvironment codecatalyst.GetDevEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecatalyst_dev_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeCatalyst)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCatalyst),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDevEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDevEnvironmentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDevEnvironmentExists(ctx, resourceName, &DevEnvironment),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, rName),
					resource.TestCheckResourceAttr(resourceName, "space_name", "terraform"),
					resource.TestCheckResourceAttr(resourceName, "project_name", "terraform"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "dev.standard1.small"),
					resource.TestCheckResourceAttr(resourceName, "persistent_storage.0.size", "16"),
					resource.TestCheckResourceAttr(resourceName, "ides.0.name", "VSCode"),
				),
			},
		},
	})
}

func TestAccCodeCatalystDevEnvironment_withRepositories(t *testing.T) {
	ctx := acctest.Context(t)
	var DevEnvironment codecatalyst.GetDevEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecatalyst_dev_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeCatalyst)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCatalyst),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDevEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDevEnvironmentConfig_withRepositories(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDevEnvironmentExists(ctx, resourceName, &DevEnvironment),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, rName),
					resource.TestCheckResourceAttr(resourceName, "space_name", "terraform"),
					resource.TestCheckResourceAttr(resourceName, "project_name", "terraform"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "dev.standard1.small"),
					resource.TestCheckResourceAttr(resourceName, "persistent_storage.0.size", "16"),
					resource.TestCheckResourceAttr(resourceName, "ides.0.name", "PyCharm"),
					resource.TestCheckResourceAttr(resourceName, "ides.0.runtime", "public.ecr.aws/jetbrains/py"),
					resource.TestCheckResourceAttr(resourceName, "inactivity_timeout_minutes", "40"),
					resource.TestCheckResourceAttr(resourceName, "repositories.0.repository_name", "terraform-provider-aws"),
					resource.TestCheckResourceAttr(resourceName, "repositories.0.branch_name", "main"),
				),
			},
		},
	})
}
func TestAccCodeCatalystDevEnvironment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var DevEnvironment codecatalyst.GetDevEnvironmentOutput
	resourceName := "aws_codecatalyst_dev_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeCatalyst)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCatalyst),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDevEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDevEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEnvironmentExists(ctx, resourceName, &DevEnvironment),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodecatalyst.ResourceDevEnvironment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDevEnvironmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCatalystClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codecatalyst_dev_environment" {
				continue
			}
			spaceName := rs.Primary.Attributes["space_name"]
			projectName := rs.Primary.Attributes["project_name"]

			_, err := conn.GetDevEnvironment(ctx, &codecatalyst.GetDevEnvironmentInput{
				Id:          aws.String(rs.Primary.ID),
				SpaceName:   aws.String(spaceName),
				ProjectName: aws.String(projectName),
			})
			if errs.IsA[*types.AccessDeniedException](err) {
				continue
			}
			if err != nil {
				return err
			}

			return create.Error(names.CodeCatalyst, create.ErrActionCheckingDestroyed, tfcodecatalyst.ResNameDevEnvironment, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDevEnvironmentExists(ctx context.Context, name string, DevEnvironment *codecatalyst.GetDevEnvironmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CodeCatalyst, create.ErrActionCheckingExistence, tfcodecatalyst.ResNameDevEnvironment, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CodeCatalyst, create.ErrActionCheckingExistence, tfcodecatalyst.ResNameDevEnvironment, name, errors.New("not set"))
		}
		spaceName := rs.Primary.Attributes["space_name"]
		projectName := rs.Primary.Attributes["project_name"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCatalystClient(ctx)
		resp, err := conn.GetDevEnvironment(ctx, &codecatalyst.GetDevEnvironmentInput{
			Id:          aws.String(rs.Primary.ID),
			SpaceName:   aws.String(spaceName),
			ProjectName: aws.String(projectName),
		})

		if err != nil {
			return create.Error(names.CodeCatalyst, create.ErrActionCheckingExistence, tfcodecatalyst.ResNameDevEnvironment, rs.Primary.ID, err)
		}

		*DevEnvironment = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCatalystClient(ctx)

	/*
		If no Amazon CodeCatalyst token is available then the Go SDK crashes:

		panic: runtime error: invalid memory address or nil pointer dereference [recovered]
			panic: runtime error: invalid memory address or nil pointer dereference
		[signal SIGSEGV: segmentation violation code=0x1 addr=0x18 pc=0x13b9c10]

		goroutine 51 [running]:
		testing.tRunner.func1.2({0xc732900, 0x15190320})
			/Users/ewbankkit/sdk/go1.20.5/src/testing/testing.go:1526 +0x24e
		testing.tRunner.func1()
			/Users/ewbankkit/sdk/go1.20.5/src/testing/testing.go:1529 +0x39f
		panic({0xc732900, 0x15190320})
			/Users/ewbankkit/sdk/go1.20.5/src/runtime/panic.go:884 +0x213
		github.com/aws/smithy-go/auth/bearer.(*AuthenticationMiddleware).HandleFinalize(0xc002690360, {0xf06b6e0, 0xc0023ab200}, {{0xdcae320?, 0xc0023aaa80?}}, {0xef7e740, 0xc000c4b170})
			/Users/ewbankkit/go/pkg/mod/github.com/aws/smithy-go@v1.14.0/auth/bearer/middleware.go:59 +0x70
	*/
	defer func() {
		if err := recover(); err != nil {
			t.Skipf("skipping acceptance testing: %s", err)
		}
	}()

	_, err := conn.VerifySession(ctx, &codecatalyst.VerifySessionInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDevEnvironmentConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecatalyst_dev_environment" "test" {
  alias         = %[1]q
  space_name    = "terraform"
  project_name  = "terraform"
  instance_type = "dev.standard1.small"
  persistent_storage {
    size = 16
  }
  ides {
    name = "VSCode"
  }


}
`, rName)
}

func testAccDevEnvironmentConfig_withRepositories(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecatalyst_dev_environment" "test" {
  alias         = %[1]q
  space_name    = "terraform"
  project_name  = "terraform"
  instance_type = "dev.standard1.small"

  persistent_storage {
    size = 16
  }

  ides {
    name    = "PyCharm"
    runtime = "public.ecr.aws/jetbrains/py"
  }

  inactivity_timeout_minutes = 40

  repositories {
    repository_name = "terraform-provider-aws"
    branch_name     = "main"
  }

}
`, rName)
}
