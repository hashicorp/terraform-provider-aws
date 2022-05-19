package elbv2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfelbv2 "github.com/hashicorp/terraform-provider-aws/internal/service/elbv2"
)

func TestAccELBV2HostedZoneIDDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSLbHostedZoneIdConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_lb_hosted_zone_id.main", "id", tfelbv2.HostedZoneIdPerRegionALBMap[acctest.Region()]),
				),
			},
			{
				Config: testAccCheckAWSLbHostedZoneIdExplicitRegionConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_lb_hosted_zone_id.regional", "id", "Z32O12XQLNTSW2"),
				),
			},
			{
				Config: testAccCheckAWSLbHostedZoneIdExplicitNetworkConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_lb_hosted_zone_id.network", "id", tfelbv2.HostedZoneIdPerRegionNLBMap[acctest.Region()]),
				),
			},
			{
				Config: testAccCheckAWSLbHostedZoneIdExplicitNetworkRegionConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_lb_hosted_zone_id.network-regional", "id", "Z2IFOLAFXWLO4F"),
				),
			},
		},
	})
}

const testAccCheckAWSLbHostedZoneIdConfig = `
data "aws_lb_hosted_zone_id" "main" {}
`

//lintignore:AWSAT003
const testAccCheckAWSLbHostedZoneIdExplicitRegionConfig = `
data "aws_lb_hosted_zone_id" "regional" {
  region = "eu-west-1"
}
`

const testAccCheckAWSLbHostedZoneIdExplicitNetworkConfig = `
data "aws_lb_hosted_zone_id" "network" {
  load_balancer_type = "network"
}
`

//lintignore:AWSAT003
const testAccCheckAWSLbHostedZoneIdExplicitNetworkRegionConfig = `
data "aws_lb_hosted_zone_id" "network-regional" {
  region             = "eu-west-1"
  load_balancer_type = "network"
}
`
