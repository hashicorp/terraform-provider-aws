// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudWatchDashboard_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dashboard cloudwatch.GetDashboardOutput
	resourceName := "aws_cloudwatch_dashboard.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDashboardDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig_basic(rName, basicWidget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists(ctx, resourceName, &dashboard),
					resource.TestCheckResourceAttrSet(resourceName, "dashboard_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDashboardConfig_basic(rName, updatedWidget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists(ctx, resourceName, &dashboard),
					resource.TestCheckResourceAttrSet(resourceName, "dashboard_arn"),
				),
			},
		},
	})
}

func TestAccCloudWatchDashboard_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dashboard cloudwatch.GetDashboardOutput
	resourceName := "aws_cloudwatch_dashboard.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDashboardDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig_basic(rName, basicWidget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists(ctx, resourceName, &dashboard),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudwatch.ResourceDashboard(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDashboardExists(ctx context.Context, n string, v *cloudwatch.GetDashboardOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchClient(ctx)

		output, err := tfcloudwatch.FindDashboardByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDashboardDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_dashboard" {
				continue
			}

			_, err := tfcloudwatch.FindDashboardByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Dashboard %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

const (
	basicWidget = `{
  "widgets": [
    {
      "type": "text",
      "x": 0,
      "y": 0,
      "width": 6,
      "height": 6,
      "properties": {
        "markdown": "Hi there from Terraform: CloudWatch"
      }
    }
  ]
}`

	updatedWidget = `{
  "widgets": [
    {
      "type": "text",
      "x": 0,
      "y": 0,
      "width": 6,
      "height": 6,
      "properties": {
        "markdown": "Hi there from Terraform: CloudWatch - updated"
      }
    }
  ]
}`
)

func testAccDashboardConfig_basic(rName, body string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_dashboard" "test" {
  dashboard_name = %[1]q

  dashboard_body = <<EOF
  %[2]s
EOF
}
`, rName, body)
}
