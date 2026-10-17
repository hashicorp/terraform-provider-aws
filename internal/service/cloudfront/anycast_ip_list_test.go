// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontAnycastIPList_serial(t *testing.T) {
	t.Parallel()

	acctest.Skip(t, "Amazon CloudFront Anycast static IP lists cost $3000 per list per month")

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccAnycastIPList_basic,
		acctest.CtDisappears: testAccAnycastIPList_disappears,
		"ipamCidrConfig":     testAccAnycastIPList_ipamCidrConfig,
		"tags":               testAccCloudFrontAnycastIPList_tagsSerial,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccAnycastIPList_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_anycast_ip_list.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnycastIPListDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnycastIPListConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnycastIPListExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("anycast_ips"), knownvalue.ListSizeExact(3)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.GlobalARNRegexp("cloudfront", regexache.MustCompile(`anycast-ip-list/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("etag"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ip_count"), knownvalue.Int32Exact(3)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAnycastIPList_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_anycast_ip_list.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnycastIPListDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnycastIPListConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnycastIPListExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudfront.ResourceAnycastIPList, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccAnycastIPList_ipamCidrConfig(t *testing.T) {
	ctx := acctest.Context(t)

	message := os.Getenv("IPAM_BYOIP_IPV4_MESSAGE")
	signature := os.Getenv("IPAM_BYOIP_IPV4_SIGNATURE")
	cidr := os.Getenv("IPAM_BYOIP_IPV4_PROVISIONED_CIDR")
	if message == "" || signature == "" || cidr == "" {
		t.Skip("Environment variable IPAM_BYOIP_IPV4_MESSAGE, IPAM_BYOIP_IPV4_SIGNATURE, or IPAM_BYOIP_IPV4_PROVISIONED_CIDR is not set")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_anycast_ip_list.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnycastIPListDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnycastIPListConfig_ipamCidrConfig(rName, cidr, message, signature),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnycastIPListExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ip_count"), knownvalue.Int32Exact(3)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ipam_cidr_config"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ipam_cidr_config").AtSliceIndex(0).AtMapKey("cidr"), knownvalue.StringExact(cidr)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ipam_config"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ipam_config").AtSliceIndex(0).AtMapKey("ipam_cidr_configs").AtSliceIndex(0).AtMapKey(names.AttrStatus), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ipam_cidr_config"},
			},
		},
	})
}

func testAccCheckAnycastIPListExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		_, err := tfcloudfront.FindAnycastIPListByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckAnycastIPListDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_anycast_ip_list" {
				continue
			}

			_, err := tfcloudfront.FindAnycastIPListByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Anycast IP List %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccAnycastIPListConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_anycast_ip_list" "test" {
  name     = %[1]q
  ip_count = 3
}
`, rName)
}

func testAccAnycastIPListConfig_ipamCidrConfig(rName, cidr, message, signature string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.region
  }
}

resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.public_default_scope_id
  locale         = data.aws_region.current.region
  aws_service    = "global-services"
}

resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = %[2]q

  cidr_authorization_context {
    message   = %[3]q
    signature = %[4]q
  }
}

resource "aws_cloudfront_anycast_ip_list" "test" {
  name     = %[1]q
  ip_count = 3

  ipam_cidr_config {
    cidr          = %[2]q
    ipam_pool_arn = aws_vpc_ipam_pool.test.arn
  }

  depends_on = [aws_vpc_ipam_pool_cidr.test]
}
`, rName, cidr, message, signature)
}
