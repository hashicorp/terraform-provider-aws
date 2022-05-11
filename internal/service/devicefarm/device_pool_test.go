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

func TestAccDeviceFarmDevicePool_basic(t *testing.T) {
	var pool devicefarm.DevicePool
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-updated")
	resourceName := "aws_devicefarm_device_pool.test"

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
		CheckDestroy:      testAccCheckDeviceFarmDevicePoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFarmDevicePoolConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmDevicePoolExists(resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "project_arn", "aws_devicefarm_project.test", "arn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "devicefarm", regexp.MustCompile(`devicepool:.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeviceFarmDevicePoolConfig(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmDevicePoolExists(resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "devicefarm", regexp.MustCompile(`devicepool:.+`)),
				),
			},
		},
	})
}

func TestAccDeviceFarmDevicePool_tags(t *testing.T) {
	var pool devicefarm.DevicePool
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_device_pool.test"

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
		CheckDestroy:      testAccCheckDeviceFarmDevicePoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFarmDevicePoolConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmDevicePoolExists(resourceName, &pool),
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
				Config: testAccDeviceFarmDevicePoolConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmDevicePoolExists(resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDeviceFarmDevicePoolConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmDevicePoolExists(resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccDeviceFarmDevicePool_disappears(t *testing.T) {
	var pool devicefarm.DevicePool
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_device_pool.test"

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
		CheckDestroy:      testAccCheckDeviceFarmDevicePoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFarmDevicePoolConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmDevicePoolExists(resourceName, &pool),
					acctest.CheckResourceDisappears(acctest.Provider, tfdevicefarm.ResourceDevicePool(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfdevicefarm.ResourceDevicePool(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDeviceFarmDevicePool_disappears_project(t *testing.T) {
	var pool devicefarm.DevicePool
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_device_pool.test"

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
		CheckDestroy:      testAccCheckDeviceFarmDevicePoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFarmDevicePoolConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmDevicePoolExists(resourceName, &pool),
					acctest.CheckResourceDisappears(acctest.Provider, tfdevicefarm.ResourceProject(), "aws_devicefarm_project.test"),
					acctest.CheckResourceDisappears(acctest.Provider, tfdevicefarm.ResourceDevicePool(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDeviceFarmDevicePoolExists(n string, v *devicefarm.DevicePool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DeviceFarmConn
		resp, err := tfdevicefarm.FindDevicepoolByArn(conn, rs.Primary.ID)
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

func testAccCheckDeviceFarmDevicePoolDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DeviceFarmConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_devicefarm_device_pool" {
			continue
		}

		// Try to find the resource
		_, err := tfdevicefarm.FindDevicepoolByArn(conn, rs.Primary.ID)
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

func testAccDeviceFarmDevicePoolConfig(rName string) string {
	return testAccDeviceFarmProjectConfig(rName) + fmt.Sprintf(`
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

func testAccDeviceFarmDevicePoolConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccDeviceFarmProjectConfig(rName) + fmt.Sprintf(`
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

func testAccDeviceFarmDevicePoolConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccDeviceFarmProjectConfig(rName) + fmt.Sprintf(`
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
