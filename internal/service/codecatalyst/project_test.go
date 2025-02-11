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
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfcodecatalyst "github.com/hashicorp/terraform-provider-aws/internal/service/codecatalyst"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeCatalystProject_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var project codecatalyst.GetProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecatalyst_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeCatalyst)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCatalyst),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Sample CC project created by TF"),
					resource.TestCheckResourceAttr(resourceName, "space_name", "tf-cc-aws-provider"),
				),
			},
		},
	})
}

func TestAccCodeCatalystProject_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var project codecatalyst.GetProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecatalyst_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeCatalyst)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCatalyst),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodecatalyst.ResourceProject(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProjectDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCatalystClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codecatalyst_project" {
				continue
			}

			spaceName := rs.Primary.Attributes["space_name"]

			_, err := conn.GetProject(ctx, &codecatalyst.GetProjectInput{
				Name:      aws.String(rs.Primary.ID),
				SpaceName: aws.String(spaceName),
			})
			if errs.IsA[*types.AccessDeniedException](err) {
				continue
			}
			if err != nil {
				return err
			}

			return create.Error(names.CodeCatalyst, create.ErrActionCheckingDestroyed, tfcodecatalyst.ResNameProject, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckProjectExists(ctx context.Context, name string, project *codecatalyst.GetProjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CodeCatalyst, create.ErrActionCheckingExistence, tfcodecatalyst.ResNameProject, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CodeCatalyst, create.ErrActionCheckingExistence, tfcodecatalyst.ResNameProject, name, errors.New("not set"))
		}

		spaceName := rs.Primary.Attributes["space_name"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCatalystClient(ctx)
		resp, err := conn.GetProject(ctx, &codecatalyst.GetProjectInput{
			Name:      aws.String(rs.Primary.ID),
			SpaceName: aws.String(spaceName),
		})

		if err != nil {
			return create.Error(names.CodeCatalyst, create.ErrActionCheckingExistence, tfcodecatalyst.ResNameProject, rs.Primary.ID, err)
		}

		*project = *resp

		return nil
	}
}

func testAccProjectConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecatalyst_project" "test" {
  space_name   = "tf-cc-aws-provider"
  display_name = %[1]q
  description  = "Sample CC project created by TF"
}
`, rName)
}
