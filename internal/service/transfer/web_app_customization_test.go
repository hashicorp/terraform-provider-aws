// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTransferWebAppCustomization_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var webappcustomization awstypes.DescribedWebAppCustomization
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app_customization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppCustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppCustomizationConfig_basic(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppCustomizationExists(ctx, resourceName, &webappcustomization),
					resource.TestCheckResourceAttrPair(resourceName, "web_app_id", "aws_transfer_web_app.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "title", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWebAppCustomizationConfig_basic(rName, "test2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppCustomizationExists(ctx, resourceName, &webappcustomization),
					resource.TestCheckResourceAttrPair(resourceName, "web_app_id", "aws_transfer_web_app.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "title", "test2"),
				),
			},
		},
	})
}

func TestAccTransferWebAppCustomization_files(t *testing.T) {
	ctx := acctest.Context(t)

	var webappcustomization awstypes.DescribedWebAppCustomization
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app_customization.test"
	darkBytes, _ := os.ReadFile("test-fixtures/Terraform-LogoMark_onDark.png")
	lightBytes, _ := os.ReadFile("test-fixtures/Terraform-LogoMark_onLight.png")
	darkFileBase64Encoded := itypes.Base64Encode(darkBytes)
	lightFileBase64Encoded := itypes.Base64Encode(lightBytes)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppCustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppCustomizationConfig_files(rName, "test", "Dark", "Light"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppCustomizationExists(ctx, resourceName, &webappcustomization),
					resource.TestCheckResourceAttrPair(resourceName, "web_app_id", "aws_transfer_web_app.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "title", "test"),
					resource.TestCheckResourceAttr(resourceName, "logo_file", darkFileBase64Encoded),
					resource.TestCheckResourceAttr(resourceName, "favicon_file", lightFileBase64Encoded),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWebAppCustomizationConfig_files(rName, "test", "Light", "Dark"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppCustomizationExists(ctx, resourceName, &webappcustomization),
					resource.TestCheckResourceAttrPair(resourceName, "web_app_id", "aws_transfer_web_app.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "title", "test"),
					resource.TestCheckResourceAttr(resourceName, "logo_file", lightFileBase64Encoded),
					resource.TestCheckResourceAttr(resourceName, "favicon_file", darkFileBase64Encoded),
				),
			},
			{
				Config: testAccWebAppCustomizationConfig_basic(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppCustomizationExists(ctx, resourceName, &webappcustomization),
					resource.TestCheckResourceAttrPair(resourceName, "web_app_id", "aws_transfer_web_app.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "title", "test"),
					resource.TestCheckResourceAttr(resourceName, "logo_file", lightFileBase64Encoded),
					resource.TestCheckResourceAttr(resourceName, "favicon_file", darkFileBase64Encoded),
				),
			},
		},
	})
}

func TestAccTransferWebAppCustomization_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var webappcustomization awstypes.DescribedWebAppCustomization
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app_customization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppCustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppCustomizationConfig_basic(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppCustomizationExists(ctx, resourceName, &webappcustomization),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tftransfer.ResourceWebAppCustomization, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckWebAppCustomizationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_web_app_customization" {
				continue
			}

			_, err := tftransfer.FindWebAppCustomizationByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Transfer, create.ErrActionCheckingDestroyed, tftransfer.ResNameWebAppCustomization, rs.Primary.ID, err)
			}

			return create.Error(names.Transfer, create.ErrActionCheckingDestroyed, tftransfer.ResNameWebAppCustomization, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckWebAppCustomizationExists(ctx context.Context, name string, webappcustomization *awstypes.DescribedWebAppCustomization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Transfer, create.ErrActionCheckingExistence, tftransfer.ResNameWebAppCustomization, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Transfer, create.ErrActionCheckingExistence, tftransfer.ResNameWebAppCustomization, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferClient(ctx)

		resp, err := tftransfer.FindWebAppCustomizationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.Transfer, create.ErrActionCheckingExistence, tftransfer.ResNameWebAppCustomization, rs.Primary.ID, err)
		}

		*webappcustomization = *resp

		return nil
	}
}

func testAccWebAppCustomizationConfig_base(rName string) string {
	return acctest.ConfigCompose(
		testAccWebAppConfig_base(rName), `
resource "aws_transfer_web_app" "test" {
  identity_provider_details {
    identity_center_config {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      role         = aws_iam_role.test.arn
    }
  }
}
`)
}

func testAccWebAppCustomizationConfig_basic(rName, title string) string {
	return acctest.ConfigCompose(
		testAccWebAppCustomizationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_transfer_web_app_customization" "test" {
  web_app_id = aws_transfer_web_app.test.id
  title      = %[1]q
}
`, title))
}

func testAccWebAppCustomizationConfig_files(rName, title, logoFileSuffix, faviconFileSuffix string) string {
	return acctest.ConfigCompose(
		testAccWebAppCustomizationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_transfer_web_app_customization" "test" {
  web_app_id   = aws_transfer_web_app.test.id
  title        = %[1]q
  logo_file    = filebase64("test-fixtures/Terraform-LogoMark_on%[2]s.png")
  favicon_file = filebase64("test-fixtures/Terraform-LogoMark_on%[3]s.png")
}
`, title, logoFileSuffix, faviconFileSuffix))
}
