package devicefarm_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdevicefarm "github.com/hashicorp/terraform-provider-aws/internal/service/devicefarm"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccDeviceFarmTestGridProject_basic(t *testing.T) {
	var proj devicefarm.TestGridProject
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-updated")
	resourceName := "aws_devicefarm_test_grid_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(devicefarm.EndpointsID, t)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:        acctest.ErrorCheck(t, devicefarm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeviceFarmProjectTestGridProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFarmProjectTestGridProjectConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmProjectTestGridProjectExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "devicefarm", regexp.MustCompile(`testgrid-project:.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeviceFarmProjectTestGridProjectConfig(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmProjectTestGridProjectExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "devicefarm", regexp.MustCompile(`testgrid-project:.+`)),
				),
			},
		},
	})
}

func TestAccDeviceFarmTestGridProject_vpc(t *testing.T) {
	var proj devicefarm.TestGridProject
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_test_grid_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(devicefarm.EndpointsID, t)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:        acctest.ErrorCheck(t, devicefarm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeviceFarmProjectTestGridProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFarmProjectTestGridProjectVpcConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmProjectTestGridProjectExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", "aws_vpc.test", "id"),
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
	var proj devicefarm.TestGridProject
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_test_grid_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(devicefarm.EndpointsID, t)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:        acctest.ErrorCheck(t, devicefarm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeviceFarmProjectTestGridProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFarmProjectTestGridProjectConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmProjectTestGridProjectExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeviceFarmProjectTestGridProjectConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmProjectTestGridProjectExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDeviceFarmProjectTestGridProjectConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmProjectTestGridProjectExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccDeviceFarmTestGridProject_disappears(t *testing.T) {
	var proj devicefarm.TestGridProject
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_test_grid_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(devicefarm.EndpointsID, t)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:        acctest.ErrorCheck(t, devicefarm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeviceFarmProjectTestGridProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFarmProjectTestGridProjectConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmProjectTestGridProjectExists(resourceName, &proj),
					acctest.CheckResourceDisappears(acctest.Provider, tfdevicefarm.ResourceTestGridProject(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfdevicefarm.ResourceTestGridProject(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDeviceFarmProjectTestGridProjectExists(n string, v *devicefarm.TestGridProject) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DeviceFarmConn
		resp, err := tfdevicefarm.FindTestGridProjectByArn(conn, rs.Primary.ID)
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

func testAccCheckDeviceFarmProjectTestGridProjectDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DeviceFarmConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_devicefarm_test_grid_project" {
			continue
		}

		// Try to find the resource
		_, err := tfdevicefarm.FindTestGridProjectByArn(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("DeviceFarm Test Grid Project %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccDeviceFarmProjectTestGridProjectConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_devicefarm_test_grid_project" "test" {
  name = %[1]q
}
`, rName)
}

func testAccDeviceFarmProjectTestGridProjectVpcConfig(rName string) string {
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
    subnet_ids         = aws_subnet.test.*.id
    security_group_ids = aws_security_group.test.*.id
  }
}
`, rName))
}

func testAccDeviceFarmProjectTestGridProjectConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_devicefarm_test_grid_project" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDeviceFarmProjectTestGridProjectConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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
