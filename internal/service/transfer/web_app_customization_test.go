// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTransferWebAppCustomization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedWebAppCustomization
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app_customization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppCustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppCustomizationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppCustomizationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "web_app_id",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "web_app_id"),
			},
		},
	})
}

func TestAccTransferWebAppCustomization_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedWebAppCustomization
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app_customization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppCustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppCustomizationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppCustomizationExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tftransfer.ResourceWebAppCustomization, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccTransferWebAppCustomization_Disappears_webApp(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedWebAppCustomization
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app_customization.test"
	webAppResourceName := "aws_transfer_web_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppCustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppCustomizationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppCustomizationExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tftransfer.ResourceWebApp, webAppResourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccTransferWebAppCustomization_title(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedWebAppCustomization
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app_customization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppCustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppCustomizationConfig_title(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppCustomizationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("title"), knownvalue.StringExact("test")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "web_app_id",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "web_app_id"),
			},
			{
				Config: testAccWebAppCustomizationConfig_title(rName, "test2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppCustomizationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("title"), knownvalue.StringExact("test2")),
				},
			},
			{
				Config: testAccWebAppCustomizationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppCustomizationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("title"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccTransferWebAppCustomization_files(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedWebAppCustomization
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app_customization.test"
	darkBytes, _ := os.ReadFile("test-fixtures/Terraform-LogoMark_onDark.png")
	lightBytes, _ := os.ReadFile("test-fixtures/Terraform-LogoMark_onLight.png")
	darkFileBase64Encoded := inttypes.Base64Encode(darkBytes)
	lightFileBase64Encoded := inttypes.Base64Encode(lightBytes)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppCustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppCustomizationConfig_files(rName, "Dark", "Light"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppCustomizationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("favicon_file"), knownvalue.StringExact(lightFileBase64Encoded)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("logo_file"), knownvalue.StringExact(darkFileBase64Encoded)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "web_app_id",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "web_app_id"),
			},
			{
				Config: testAccWebAppCustomizationConfig_files(rName, "Light", "Dark"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppCustomizationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("favicon_file"), knownvalue.StringExact(darkFileBase64Encoded)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("logo_file"), knownvalue.StringExact(lightFileBase64Encoded)),
				},
			},
			{
				Config: testAccWebAppCustomizationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppCustomizationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
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

			_, err := tftransfer.FindWebAppCustomizationByID(ctx, conn, rs.Primary.Attributes["web_app_id"])
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Transfer Web App Customization %s still exists", rs.Primary.Attributes["web_app_id"])
		}

		return nil
	}
}

func testAccCheckWebAppCustomizationExists(ctx context.Context, n string, v *awstypes.DescribedWebAppCustomization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferClient(ctx)

		resp, err := tftransfer.FindWebAppCustomizationByID(ctx, conn, rs.Primary.Attributes["web_app_id"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccWebAppCustomizationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccWebAppConfig_basic(rName), `
resource "aws_transfer_web_app_customization" "test" {
  web_app_id = aws_transfer_web_app.test.web_app_id
}
`)
}

func testAccWebAppCustomizationConfig_title(rName, title string) string {
	return acctest.ConfigCompose(testAccWebAppConfig_basic(rName),
		fmt.Sprintf(`
resource "aws_transfer_web_app_customization" "test" {
  web_app_id = aws_transfer_web_app.test.web_app_id
  title      = %[1]q
}
`, title))
}

func testAccWebAppCustomizationConfig_files(rName, logoFileSuffix, faviconFileSuffix string) string {
	return acctest.ConfigCompose(testAccWebAppConfig_basic(rName), fmt.Sprintf(`
resource "aws_transfer_web_app_customization" "test" {
  web_app_id   = aws_transfer_web_app.test.web_app_id
  logo_file    = filebase64("test-fixtures/Terraform-LogoMark_on%[1]s.png")
  favicon_file = filebase64("test-fixtures/Terraform-LogoMark_on%[2]s.png")
}
`, logoFileSuffix, faviconFileSuffix))
}
