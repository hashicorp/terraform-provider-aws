// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCNetworkInsightsAnalysisDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_analysis.test"
	datasourceName := "data.aws_ec2_network_insights_analysis.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsAnalysisDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "alternate_path_hints.#", resourceName, "alternate_path_hints.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "explanations.#", resourceName, "explanations.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "filter_in_arns.#", resourceName, "filter_in_arns.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "forward_path_components.#", resourceName, "forward_path_components.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "network_insights_analysis_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, "network_insights_path_id", resourceName, "network_insights_path_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "path_found", resourceName, "path_found"),
					resource.TestCheckResourceAttrPair(datasourceName, "return_path_components.#", resourceName, "return_path_components.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "start_date", resourceName, "start_date"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrStatusMessage, resourceName, names.AttrStatusMessage),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "warning_message", resourceName, "warning_message"),
				),
			},
		},
	})
}

func testAccVPCNetworkInsightsAnalysisDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCNetworkInsightsAnalysisConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_network_insights_analysis" "test" {
  network_insights_path_id = aws_ec2_network_insights_path.test.id
  wait_for_completion      = true

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_network_insights_analysis" "test" {
  network_insights_analysis_id = aws_ec2_network_insights_analysis.test.id
}
`, rName))
}
