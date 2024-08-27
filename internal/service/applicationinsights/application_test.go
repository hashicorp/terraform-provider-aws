// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applicationinsights_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/applicationinsights/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapplicationinsights "github.com/hashicorp/terraform-provider-aws/internal/service/applicationinsights"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccApplicationInsightsApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_applicationinsights_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ApplicationInsightsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "resource_group_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "applicationinsights", fmt.Sprintf("application/resource-group/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "auto_config_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cwe_monitor_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ops_center_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
					testAccCheckApplicationExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "resource_group_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "applicationinsights", fmt.Sprintf("application/resource-group/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "auto_config_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cwe_monitor_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ops_center_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccApplicationInsightsApplication_autoConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_applicationinsights_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ApplicationInsightsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "resource_group_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "applicationinsights", fmt.Sprintf("application/resource-group/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "auto_config_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cwe_monitor_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ops_center_enabled", acctest.CtFalse),
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

func TestAccApplicationInsightsApplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_applicationinsights_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ApplicationInsightsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &app),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapplicationinsights.ResourceApplication(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapplicationinsights.ResourceApplication(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckApplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ApplicationInsightsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_applicationinsights_application" {
				continue
			}

			_, err := tfapplicationinsights.FindApplicationByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccCheckApplicationExists(ctx context.Context, n string, v *awstypes.ApplicationInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ApplicationInsightsClient(ctx)

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
