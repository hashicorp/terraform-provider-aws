// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfcodecatalyst "github.com/hashicorp/terraform-provider-aws/internal/service/codecatalyst"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeCatalystSourceRepository_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var sourcerepository codecatalyst.GetSourceRepositoryOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codecatalyst_source_repository.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeCatalyst)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCatalyst),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSourceRepositoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSourceRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceRepositoryExists(ctx, t, resourceName, &sourcerepository),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "space_name", "tf-cc-aws-provider"),
					resource.TestCheckResourceAttr(resourceName, "project_name", "tf-cc"),
				),
			},
		},
	})
}

func TestAccCodeCatalystSourceRepository_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var sourcerepository codecatalyst.GetSourceRepositoryOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codecatalyst_source_repository.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeCatalyst)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCatalyst),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSourceRepositoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSourceRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceRepositoryExists(ctx, t, resourceName, &sourcerepository),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcodecatalyst.ResourceSourceRepository(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSourceRepositoryDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CodeCatalystClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codecatalyst_source_repository" {
				continue
			}
			spaceName := rs.Primary.Attributes["space_name"]
			projectName := rs.Primary.Attributes["project_name"]

			input := &codecatalyst.GetSourceRepositoryInput{
				Name:        aws.String(rs.Primary.ID),
				SpaceName:   aws.String(spaceName),
				ProjectName: aws.String(projectName),
			}
			_, err := conn.GetSourceRepository(ctx, input)

			if errs.IsA[*types.AccessDeniedException](err) {
				continue
			}
			if err != nil {
				return err
			}

			return create.Error(names.CodeCatalyst, create.ErrActionCheckingDestroyed, tfcodecatalyst.ResNameSourceRepository, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckSourceRepositoryExists(ctx context.Context, t *testing.T, name string, sourcerepository *codecatalyst.GetSourceRepositoryOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CodeCatalyst, create.ErrActionCheckingExistence, tfcodecatalyst.ResNameSourceRepository, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CodeCatalyst, create.ErrActionCheckingExistence, tfcodecatalyst.ResNameSourceRepository, name, errors.New("not set"))
		}
		spaceName := rs.Primary.Attributes["space_name"]
		projectName := rs.Primary.Attributes["project_name"]

		conn := acctest.ProviderMeta(ctx, t).CodeCatalystClient(ctx)
		input := codecatalyst.GetSourceRepositoryInput{
			Name:        aws.String(rs.Primary.ID),
			SpaceName:   aws.String(spaceName),
			ProjectName: aws.String(projectName),
		}
		resp, err := conn.GetSourceRepository(ctx, &input)

		if err != nil {
			return create.Error(names.CodeCatalyst, create.ErrActionCheckingExistence, tfcodecatalyst.ResNameSourceRepository, rs.Primary.ID, err)
		}

		*sourcerepository = *resp

		return nil
	}
}

func testAccSourceRepositoryConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecatalyst_source_repository" "test" {
  name         = %[1]q
  project_name = "tf-cc"
  space_name   = "tf-cc-aws-provider"
}
`, rName)
}
