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

func TestAccDeviceFarmDevicePool_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.DevicePool
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-updated")
	resourceName := "aws_devicefarm_device_pool.test"

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
		CheckDestroy:             testAccCheckDevicePoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDevicePoolConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevicePoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "project_arn", "aws_devicefarm_project.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "devicefarm", regexache.MustCompile(`devicepool:.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDevicePoolConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevicePoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "devicefarm", regexache.MustCompile(`devicepool:.+`)),
				),
			},
		},
	})
}

func TestAccDeviceFarmDevicePool_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.DevicePool
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_device_pool.test"

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
		CheckDestroy:             testAccCheckDevicePoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDevicePoolConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevicePoolExists(ctx, resourceName, &pool),
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
				Config: testAccDevicePoolConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevicePoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDevicePoolConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevicePoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccDeviceFarmDevicePool_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.DevicePool
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_device_pool.test"

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
		CheckDestroy:             testAccCheckDevicePoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDevicePoolConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevicePoolExists(ctx, resourceName, &pool),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdevicefarm.ResourceDevicePool(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdevicefarm.ResourceDevicePool(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDeviceFarmDevicePool_disappears_project(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.DevicePool
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_device_pool.test"

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
		CheckDestroy:             testAccCheckDevicePoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDevicePoolConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevicePoolExists(ctx, resourceName, &pool),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdevicefarm.ResourceProject(), "aws_devicefarm_project.test"),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdevicefarm.ResourceDevicePool(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDevicePoolExists(ctx context.Context, n string, v *awstypes.DevicePool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DeviceFarmClient(ctx)
		resp, err := tfdevicefarm.FindDevicePoolByARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("DeviceFarm Device Pool not found")
		}

		*v = *resp

		return nil
	}
}

func testAccCheckDevicePoolDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DeviceFarmClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_devicefarm_device_pool" {
				continue
			}

			// Try to find the resource
			_, err := tfdevicefarm.FindDevicePoolByARN(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DeviceFarm Device Pool %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDevicePoolConfig_basic(rName string) string {
	return testAccProjectConfig_basic(rName) + fmt.Sprintf(`
resource "aws_devicefarm_device_pool" "test" {
  name        = %[1]q
  project_arn = aws_devicefarm_project.test.arn
  rule {
    attribute = "OS_VERSION"
    operator  = "EQUALS"
    value     = "\"AVAILABLE\""
  }
}
`, rName)
}

func testAccDevicePoolConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return testAccProjectConfig_basic(rName) + fmt.Sprintf(`
resource "aws_devicefarm_device_pool" "test" {
  name        = %[1]q
  project_arn = aws_devicefarm_project.test.arn
  rule {
    attribute = "AVAILABILITY"
    operator  = "EQUALS"
    value     = "\"AVAILABLE\""
  }
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDevicePoolConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccProjectConfig_basic(rName) + fmt.Sprintf(`
resource "aws_devicefarm_device_pool" "test" {
  name        = %[1]q
  project_arn = aws_devicefarm_project.test.arn
  rule {
    attribute = "AVAILABILITY"
    operator  = "EQUALS"
    value     = "\"AVAILABLE\""
  }
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
