// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package applicationinsights_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/applicationinsights/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfapplicationinsights "github.com/hashicorp/terraform-provider-aws/internal/service/applicationinsights"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccApplicationInsightsApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.ApplicationInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_applicationinsights_application.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ApplicationInsightsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "resource_group_name", rName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "applicationinsights", fmt.Sprintf("application/resource-group/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "auto_config_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cwe_monitor_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ops_center_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "resource_group_name", rName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "applicationinsights", fmt.Sprintf("application/resource-group/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "auto_config_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cwe_monitor_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ops_center_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccApplicationInsightsApplication_autoConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.ApplicationInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_applicationinsights_application.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ApplicationInsightsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "resource_group_name", rName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "applicationinsights", fmt.Sprintf("application/resource-group/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "auto_config_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cwe_monitor_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ops_center_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccApplicationInsightsApplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.ApplicationInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_applicationinsights_application.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ApplicationInsightsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &app),
					acctest.CheckSDKResourceDisappears(ctx, t, tfapplicationinsights.ResourceApplication(), resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfapplicationinsights.ResourceApplication(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckApplicationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ApplicationInsightsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_applicationinsights_application" {
				continue
			}

			_, err := tfapplicationinsights.FindApplicationByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ApplicationInsights Application %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckApplicationExists(ctx context.Context, t *testing.T, n string, v *awstypes.ApplicationInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ApplicationInsightsClient(ctx)

		output, err := tfapplicationinsights.FindApplicationByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccApplicationConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_resourcegroups_group" "test" {
  name = %[1]q

  resource_query {
    query = <<JSON
	{
		"ResourceTypeFilters": [
		  "AWS::EC2::Instance"
		],
		"TagFilters": [
		  {
			"Key": "Stage",
			"Values": [
			  "Test"
			]
		  }
		]
	  }
JSON
  }
}
`, rName)
}

func testAccApplicationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_base(rName), `
resource "aws_applicationinsights_application" "test" {
  resource_group_name = aws_resourcegroups_group.test.name
}
`)
}

func testAccApplicationConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_base(rName), `
resource "aws_applicationinsights_application" "test" {
  resource_group_name = aws_resourcegroups_group.test.name
  auto_config_enabled = true
}
`)
}
