// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devicefarm_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devicefarm/types"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdevicefarm "github.com/hashicorp/terraform-provider-aws/internal/service/devicefarm"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDeviceFarmInstanceProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var profile awstypes.InstanceProfile
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-updated")
	resourceName := "aws_devicefarm_instance_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DeviceFarmEndpointID)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DeviceFarmServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "reboot_after_use", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "devicefarm", regexache.MustCompile(`instanceprofile:.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceProfileConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "reboot_after_use", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "devicefarm", regexache.MustCompile(`instanceprofile:.+`)),
				),
			},
		},
	})
}

func TestAccDeviceFarmInstanceProfile_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var profile awstypes.InstanceProfile
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_instance_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DeviceFarmEndpointID)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DeviceFarmServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceProfileConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccInstanceProfileConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccDeviceFarmInstanceProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var profile awstypes.InstanceProfile
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_instance_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DeviceFarmEndpointID)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DeviceFarmServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, resourceName, &profile),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdevicefarm.ResourceInstanceProfile(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdevicefarm.ResourceInstanceProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInstanceProfileExists(ctx context.Context, n string, v *awstypes.InstanceProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DeviceFarmClient(ctx)
		resp, err := tfdevicefarm.FindInstanceProfileByARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("DeviceFarm Instance Profile not found")
		}

		*v = *resp

		return nil
	}
}

func testAccCheckInstanceProfileDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DeviceFarmClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_devicefarm_instance_profile" {
				continue
			}

			// Try to find the resource
			_, err := tfdevicefarm.FindInstanceProfileByARN(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DeviceFarm Instance Profile %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccInstanceProfileConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_devicefarm_instance_profile" "test" {
  name = %[1]q
}
`, rName)
}

func testAccInstanceProfileConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_devicefarm_instance_profile" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccInstanceProfileConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_devicefarm_instance_profile" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
