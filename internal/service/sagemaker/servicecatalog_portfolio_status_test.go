// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccServicecatalogPortfolioStatus_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var config sagemaker.GetSagemakerServicecatalogPortfolioStatusOutput
	resourceName := "aws_sagemaker_servicecatalog_portfolio_status.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccServicecatalogPortfolioStatusConfigConfig_basic("Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServicecatalogPortfolioStatusExists(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Enabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServicecatalogPortfolioStatusConfigConfig_basic("Disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServicecatalogPortfolioStatusExists(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Disabled"),
				),
			},
			{
				Config: testAccServicecatalogPortfolioStatusConfigConfig_basic("Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServicecatalogPortfolioStatusExists(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Enabled"),
				),
			},
		},
	})
}

func testAccCheckServicecatalogPortfolioStatusExists(ctx context.Context, n string, config *sagemaker.GetSagemakerServicecatalogPortfolioStatusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SageMaker AI Studio Lifecycle Config ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerClient(ctx)

		output, err := tfsagemaker.FindServicecatalogPortfolioStatus(ctx, conn)

		if err != nil {
			return err
		}

		*config = *output

		return nil
	}
}

func testAccServicecatalogPortfolioStatusConfigConfig_basic(status string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_servicecatalog_portfolio_status" "test" {
  status = %[1]q
}
`, status)
}
