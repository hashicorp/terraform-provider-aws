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

func TestAccDeviceFarmTestGridProject_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var proj awstypes.TestGridProject
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, "tf-acc-test-updated")
	resourceName := "aws_devicefarm_test_grid_project.test"

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
		CheckDestroy:             testAccCheckTestGridProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTestGridProjectConfig_project(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTestGridProjectExists(ctx, t, resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "devicefarm", regexache.MustCompile(`testgrid-project:.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTestGridProjectConfig_project(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTestGridProjectExists(ctx, t, resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "devicefarm", regexache.MustCompile(`testgrid-project:.+`)),
				),
			},
		},
	})
}

func TestAccDeviceFarmTestGridProject_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var proj awstypes.TestGridProject
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_test_grid_project.test"

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
		CheckDestroy:             testAccCheckTestGridProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTestGridProjectConfig_projectVPC(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTestGridProjectExists(ctx, t, resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", "aws_vpc.test", names.AttrID),
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

func TestAccDeviceFarmTestGridProject_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var proj awstypes.TestGridProject
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_test_grid_project.test"

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
		CheckDestroy:             testAccCheckTestGridProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTestGridProjectConfig_projectTags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTestGridProjectExists(ctx, t, resourceName, &proj),
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
				Config: testAccTestGridProjectConfig_projectTags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTestGridProjectExists(ctx, t, resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTestGridProjectConfig_projectTags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTestGridProjectExists(ctx, t, resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccDeviceFarmTestGridProject_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var proj awstypes.TestGridProject
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_test_grid_project.test"

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
		CheckDestroy:             testAccCheckTestGridProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTestGridProjectConfig_project(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTestGridProjectExists(ctx, t, resourceName, &proj),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdevicefarm.ResourceTestGridProject(), resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdevicefarm.ResourceTestGridProject(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTestGridProjectExists(ctx context.Context, t *testing.T, n string, v *awstypes.TestGridProject) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DeviceFarmClient(ctx)
		resp, err := tfdevicefarm.FindTestGridProjectByARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("DeviceFarm Test Grid Project not found")
		}

		*v = *resp

		return nil
	}
}

func testAccCheckTestGridProjectDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DeviceFarmClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_devicefarm_test_grid_project" {
				continue
			}

			// Try to find the resource
			_, err := tfdevicefarm.FindTestGridProjectByARN(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DeviceFarm Test Grid Project %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTestGridProjectConfig_project(rName string) string {
	return fmt.Sprintf(`
resource "aws_devicefarm_test_grid_project" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTestGridProjectConfig_projectVPC(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id
}

resource "aws_security_group" "test" {
  count = 2

  name        = "%[1]s-${count.index}"
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.test.id
  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }
}

resource "aws_devicefarm_test_grid_project" "test" {
  name = %[1]q

  vpc_config {
    vpc_id             = aws_vpc.test.id
    subnet_ids         = aws_subnet.test[*].id
    security_group_ids = aws_security_group.test[*].id
  }
}
`, rName))
}

func testAccTestGridProjectConfig_projectTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_devicefarm_test_grid_project" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccTestGridProjectConfig_projectTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_devicefarm_test_grid_project" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
