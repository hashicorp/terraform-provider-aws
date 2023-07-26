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

func TestAccCodeCatalystDevenvironment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var devEnvironment codecatalyst.GetDevEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecatalyst_devenvironment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCatalyst),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDevenvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDevenvironmentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDevenvironmentExists(ctx, resourceName, &devEnvironment),
					resource.TestCheckResourceAttr(resourceName, "alias", rName),
					resource.TestCheckResourceAttr(resourceName, "space_name", "terraform"),
					resource.TestCheckResourceAttr(resourceName, "project_name", "terraform"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "dev.standard1.small"),
					resource.TestCheckResourceAttr(resourceName, "persistent_storage.0.size", "16"),
					resource.TestCheckResourceAttr(resourceName, "ides.0.name", "VSCode"),
				),
			},
		},
	})
}

func TestAccCodeCatalystDevenvironment_withRepositories(t *testing.T) {
	ctx := acctest.Context(t)
	var devEnvironment codecatalyst.GetDevEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecatalyst_devenvironment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCatalyst),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDevenvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDevenvironmentConfig_withRepositories(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDevenvironmentExists(ctx, resourceName, &devEnvironment),
					resource.TestCheckResourceAttr(resourceName, "alias", rName),
					resource.TestCheckResourceAttr(resourceName, "space_name", "terraform"),
					resource.TestCheckResourceAttr(resourceName, "project_name", "terraform"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "dev.standard1.small"),
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
func TestAccCodeCatalystDevenvironment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var devEnvironment codecatalyst.GetDevEnvironmentOutput
	resourceName := "aws_codecatalyst_devenvironment.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCatalyst),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDevenvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDevenvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevenvironmentExists(ctx, resourceName, &devEnvironment),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodecatalyst.ResourceDevenvironment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDevenvironmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCatalystClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codecatalyst_devenvironment" {
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

			return create.Error(names.CodeCatalyst, create.ErrActionCheckingDestroyed, tfcodecatalyst.ResNameDevenvironment, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDevenvironmentExists(ctx context.Context, name string, devenvironment *codecatalyst.GetDevEnvironmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CodeCatalyst, create.ErrActionCheckingExistence, tfcodecatalyst.ResNameDevenvironment, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CodeCatalyst, create.ErrActionCheckingExistence, tfcodecatalyst.ResNameDevenvironment, name, errors.New("not set"))
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
			return create.Error(names.CodeCatalyst, create.ErrActionCheckingExistence, tfcodecatalyst.ResNameDevenvironment, rs.Primary.ID, err)
		}

		*devenvironment = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCatalystClient(ctx)

	spaceName := "terraform"
	projectName := "terraform"

	input := &codecatalyst.ListDevEnvironmentsInput{
		SpaceName:   aws.String(spaceName),
		ProjectName: aws.String(projectName),
	}
	_, err := conn.ListDevEnvironments(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDevenvironmentConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecatalyst_devenvironment" "test" {
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

func testAccDevenvironmentConfig_withRepositories(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecatalyst_devenvironment" "test" {
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
