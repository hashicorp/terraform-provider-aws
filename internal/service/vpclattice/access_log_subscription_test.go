// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeAccessLogSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var accesslogsubscription vpclattice.GetAccessLogSubscriptionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_access_log_subscription.test"
	serviceNetworkResourceName := "aws_vpclattice_service_network.test"
	s3BucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessLogSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessLogSubscriptionConfig_basicS3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessLogSubscriptionExists(ctx, resourceName, &accesslogsubscription),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, names.VPCLatticeEndpointID, regexache.MustCompile(`accesslogsubscription/.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDestinationARN, s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, serviceNetworkResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "resource_identifier", serviceNetworkResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "service_network_log_type", "SERVICE"),
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

func TestAccVPCLatticeAccessLogSubscription_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var accesslogsubscription vpclattice.GetAccessLogSubscriptionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_access_log_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessLogSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessLogSubscriptionConfig_basicS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessLogSubscriptionExists(ctx, resourceName, &accesslogsubscription),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfvpclattice.ResourceAccessLogSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCLatticeAccessLogSubscription_arn(t *testing.T) {
	ctx := acctest.Context(t)
	var accesslogsubscription vpclattice.GetAccessLogSubscriptionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_access_log_subscription.test"
	serviceNetworkResourceName := "aws_vpclattice_service_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessLogSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessLogSubscriptionConfig_arn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessLogSubscriptionExists(ctx, resourceName, &accesslogsubscription),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, serviceNetworkResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "resource_identifier", serviceNetworkResourceName, names.AttrID),
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

func TestAccVPCLatticeAccessLogSubscription_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var accesslogsubscription1, accesslogsubscription2, accesslogsubscription3 vpclattice.GetAccessLogSubscriptionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_access_log_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessLogSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessLogSubscriptionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessLogSubscriptionExists(ctx, resourceName, &accesslogsubscription1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccessLogSubscriptionConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessLogSubscriptionExists(ctx, resourceName, &accesslogsubscription2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAccessLogSubscriptionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessLogSubscriptionExists(ctx, resourceName, &accesslogsubscription3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func TestAccVPCLatticeAccessLogSubscription_cloudwatchNoWildcard(t *testing.T) {
	ctx := acctest.Context(t)
	var accesslogsubscription vpclattice.GetAccessLogSubscriptionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_access_log_subscription.test"
	serviceResourceName := "aws_vpclattice_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessLogSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessLogSubscriptionConfig_cloudwatchNoWildcard(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessLogSubscriptionExists(ctx, resourceName, &accesslogsubscription),
					resource.TestCheckResourceAttrWith(resourceName, names.AttrDestinationARN, func(value string) error {
						if !strings.HasSuffix(value, ":*") {
							return fmt.Errorf("%s is not a wildcard ARN", value)
						}

						return nil
					}),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, serviceResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "resource_identifier", serviceResourceName, names.AttrID),
				),
			},
		},
	})
}

func TestAccVPCLatticeAccessLogSubscription_cloudwatchWildcard(t *testing.T) {
	ctx := acctest.Context(t)
	var accesslogsubscription vpclattice.GetAccessLogSubscriptionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_access_log_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessLogSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessLogSubscriptionConfig_cloudwatchWildcard(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessLogSubscriptionExists(ctx, resourceName, &accesslogsubscription),
					resource.TestCheckResourceAttrWith(resourceName, names.AttrDestinationARN, func(value string) error {
						if !strings.HasSuffix(value, ":*") {
							return fmt.Errorf("%s is not a wildcard ARN", value)
						}

						return nil
					}),
				),
			},
		},
	})
}

func TestAccVPCLatticeAccessLogSubscription_serviceNetworkLogType(t *testing.T) {
	ctx := acctest.Context(t)
	var accesslogsubscription vpclattice.GetAccessLogSubscriptionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_access_log_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessLogSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessLogSubscriptionConfig_serviceNetworkLogType(rName, "SERVICE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessLogSubscriptionExists(ctx, resourceName, &accesslogsubscription),
					resource.TestCheckResourceAttr(resourceName, "service_network_log_type", "SERVICE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccessLogSubscriptionConfig_serviceNetworkLogType(rName, "RESOURCE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessLogSubscriptionExists(ctx, resourceName, &accesslogsubscription),
					resource.TestCheckResourceAttr(resourceName, "service_network_log_type", "RESOURCE"),
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

func testAccCheckAccessLogSubscriptionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_access_log_subscription" {
				continue
			}

			_, err := tfvpclattice.FindAccessLogSubscriptionByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Lattice Access Log Subscription %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccessLogSubscriptionExists(ctx context.Context, n string, v *vpclattice.GetAccessLogSubscriptionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		output, err := tfvpclattice.FindAccessLogSubscriptionByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAccessLogSubscriptionConfig_baseS3(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}
`, rName)
}

func testAccAccessLogSubscriptionConfig_baseCloudWatch(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_group" "test" {
  name = "/aws/vpclattice/%[1]s"
}
`, rName)
}

func testAccAccessLogSubscriptionConfig_basicS3(rName string) string {
	return acctest.ConfigCompose(testAccAccessLogSubscriptionConfig_baseS3(rName), `
resource "aws_vpclattice_access_log_subscription" "test" {
  resource_identifier = aws_vpclattice_service_network.test.id
  destination_arn     = aws_s3_bucket.test.arn
}
`)
}

func testAccAccessLogSubscriptionConfig_arn(rName string) string {
	return acctest.ConfigCompose(testAccAccessLogSubscriptionConfig_baseS3(rName), `
resource "aws_vpclattice_access_log_subscription" "test" {
  resource_identifier = aws_vpclattice_service_network.test.arn
  destination_arn     = aws_s3_bucket.test.arn
}
`)
}

func testAccAccessLogSubscriptionConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAccessLogSubscriptionConfig_baseS3(rName), fmt.Sprintf(`
resource "aws_vpclattice_access_log_subscription" "test" {
  resource_identifier = aws_vpclattice_service_network.test.id
  destination_arn     = aws_s3_bucket.test.arn

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccAccessLogSubscriptionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAccessLogSubscriptionConfig_baseS3(rName), fmt.Sprintf(`
resource "aws_vpclattice_access_log_subscription" "test" {
  resource_identifier = aws_vpclattice_service_network.test.id
  destination_arn     = aws_s3_bucket.test.arn

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAccessLogSubscriptionConfig_cloudwatchNoWildcard(rName string) string {
	return acctest.ConfigCompose(testAccAccessLogSubscriptionConfig_baseCloudWatch(rName), `
resource "aws_vpclattice_access_log_subscription" "test" {
  resource_identifier = aws_vpclattice_service.test.id
  destination_arn     = aws_cloudwatch_log_group.test.arn
}
`)
}

func testAccAccessLogSubscriptionConfig_cloudwatchWildcard(rName string) string {
	return acctest.ConfigCompose(testAccAccessLogSubscriptionConfig_baseCloudWatch(rName), `
resource "aws_vpclattice_access_log_subscription" "test" {
  resource_identifier = aws_vpclattice_service.test.id
  destination_arn     = "${aws_cloudwatch_log_group.test.arn}:*"
}
`)
}

func testAccAccessLogSubscriptionConfig_serviceNetworkLogType(rName, serviceNetworkLogType string) string {
	return acctest.ConfigCompose(testAccAccessLogSubscriptionConfig_baseS3(rName), fmt.Sprintf(`
resource "aws_vpclattice_access_log_subscription" "test" {
  resource_identifier      = aws_vpclattice_service_network.test.arn
  destination_arn          = aws_s3_bucket.test.arn
  service_network_log_type = %[1]q
}
`, serviceNetworkLogType))
}
