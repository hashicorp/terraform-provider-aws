// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCNetworkInsightsAnalysis_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_analysis.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsAnalysisDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsAnalysisConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`network-insights-analysis/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "filter_in_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "network_insights_path_id", "aws_ec2_network_insights_path.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "path_found", acctest.CtTrue),
					acctest.CheckResourceAttrRFC3339(resourceName, "start_date"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "succeeded"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "wait_for_completion", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_completion"},
			},
		},
	})
}

func TestAccVPCNetworkInsightsAnalysis_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_analysis.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsAnalysisDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsAnalysisConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceNetworkInsightsAnalysis(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCNetworkInsightsAnalysis_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_analysis.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsAnalysisDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsAnalysisConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_completion"},
			},
			{
				Config: testAccVPCNetworkInsightsAnalysisConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCNetworkInsightsAnalysisConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccVPCNetworkInsightsAnalysis_filterInARNs(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_analysis.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsAnalysisDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsAnalysisConfig_filterInARNs(rName, "vpc-peering-connection/pcx-fakearn1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "filter_in_arns.0", "ec2", regexache.MustCompile(`vpc-peering-connection/pcx-fakearn1$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_completion"},
			},
			{
				Config: testAccVPCNetworkInsightsAnalysisConfig_filterInARNs(rName, "vpc-peering-connection/pcx-fakearn2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "filter_in_arns.0", "ec2", regexache.MustCompile(`vpc-peering-connection/pcx-fakearn2$`)),
				),
			},
		},
	})
}

func TestAccVPCNetworkInsightsAnalysis_waitForCompletion(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_analysis.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsAnalysisDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsAnalysisConfig_waitForCompletion(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "wait_for_completion", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "running"),
				),
			},
			{
				Config: testAccVPCNetworkInsightsAnalysisConfig_waitForCompletion(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "wait_for_completion", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckNetworkInsightsAnalysisExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindNetworkInsightsAnalysisByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckNetworkInsightsAnalysisDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_network_insights_analysis" {
				continue
			}

			_, err := tfec2.FindNetworkInsightsAnalysisByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Network Insights Analysis %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVPCNetworkInsightsAnalysisConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  count = 2

  subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_path" "test" {
  source      = aws_network_interface.test[0].id
  destination = aws_network_interface.test[1].id
  protocol    = "tcp"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCNetworkInsightsAnalysisConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCNetworkInsightsAnalysisConfig_base(rName), `
resource "aws_ec2_network_insights_analysis" "test" {
  network_insights_path_id = aws_ec2_network_insights_path.test.id
}
`)
}

func testAccVPCNetworkInsightsAnalysisConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccVPCNetworkInsightsAnalysisConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_network_insights_analysis" "test" {
  network_insights_path_id = aws_ec2_network_insights_path.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccVPCNetworkInsightsAnalysisConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccVPCNetworkInsightsAnalysisConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_network_insights_analysis" "test" {
  network_insights_path_id = aws_ec2_network_insights_path.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccVPCNetworkInsightsAnalysisConfig_filterInARNs(rName, arnSuffix string) string {
	return acctest.ConfigCompose(testAccVPCNetworkInsightsAnalysisConfig_base(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_ec2_network_insights_analysis" "test" {
  network_insights_path_id = aws_ec2_network_insights_path.test.id
  filter_in_arns           = ["arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.id}:%[2]s"]

  tags = {
    Name = %[1]q
  }
}
`, rName, arnSuffix))
}

func testAccVPCNetworkInsightsAnalysisConfig_waitForCompletion(rName string, waitForCompletion bool) string {
	return acctest.ConfigCompose(testAccVPCNetworkInsightsAnalysisConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_network_insights_analysis" "test" {
  network_insights_path_id = aws_ec2_network_insights_path.test.id
  wait_for_completion      = %[2]t

  tags = {
    Name = %[1]q
  }
}
`, rName, waitForCompletion))
}
