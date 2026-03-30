// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devicefarm_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devicefarm/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdevicefarm "github.com/hashicorp/terraform-provider-aws/internal/service/devicefarm"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDeviceFarmNetworkProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.NetworkProfile
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, "tf-acc-test-updated")
	resourceName := "aws_devicefarm_network_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DeviceFarmEndpointID)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DeviceFarmServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkProfileExists(ctx, t, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "downlink_bandwidth_bits", "104857600"),
					resource.TestCheckResourceAttr(resourceName, "uplink_bandwidth_bits", "104857600"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrPair(resourceName, "project_arn", "aws_devicefarm_project.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "devicefarm", regexache.MustCompile(`networkprofile:.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkProfileConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkProfileExists(ctx, t, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "downlink_bandwidth_bits", "104857600"),
					resource.TestCheckResourceAttr(resourceName, "uplink_bandwidth_bits", "104857600"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrPair(resourceName, "project_arn", "aws_devicefarm_project.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "devicefarm", regexache.MustCompile(`networkprofile:.+`)),
				),
			},
		},
	})
}

func TestAccDeviceFarmNetworkProfile_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.NetworkProfile
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_network_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DeviceFarmEndpointID)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DeviceFarmServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkProfileConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkProfileExists(ctx, t, resourceName, &pool),
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
				Config: testAccNetworkProfileConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkProfileExists(ctx, t, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccNetworkProfileConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkProfileExists(ctx, t, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccDeviceFarmNetworkProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.NetworkProfile
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_network_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DeviceFarmEndpointID)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DeviceFarmServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkProfileExists(ctx, t, resourceName, &pool),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdevicefarm.ResourceNetworkProfile(), resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdevicefarm.ResourceNetworkProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDeviceFarmNetworkProfile_disappears_project(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.NetworkProfile
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_network_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DeviceFarmEndpointID)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DeviceFarmServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkProfileExists(ctx, t, resourceName, &pool),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdevicefarm.ResourceProject(), "aws_devicefarm_project.test"),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdevicefarm.ResourceNetworkProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckNetworkProfileExists(ctx context.Context, t *testing.T, n string, v *awstypes.NetworkProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DeviceFarmClient(ctx)
		resp, err := tfdevicefarm.FindNetworkProfileByARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("DeviceFarm Network Profile not found")
		}

		*v = *resp

		return nil
	}
}

func testAccCheckNetworkProfileDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DeviceFarmClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_devicefarm_network_profile" {
				continue
			}

			// Try to find the resource
			_, err := tfdevicefarm.FindNetworkProfileByARN(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DeviceFarm Network Profile %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccNetworkProfileConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basic(rName), fmt.Sprintf(`
resource "aws_devicefarm_network_profile" "test" {
  name        = %[1]q
  project_arn = aws_devicefarm_project.test.arn
}
`, rName))
}

func testAccNetworkProfileConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basic(rName), fmt.Sprintf(`
resource "aws_devicefarm_network_profile" "test" {
  name        = %[1]q
  project_arn = aws_devicefarm_project.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccNetworkProfileConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basic(rName), fmt.Sprintf(`
resource "aws_devicefarm_network_profile" "test" {
  name        = %[1]q
  project_arn = aws_devicefarm_project.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
