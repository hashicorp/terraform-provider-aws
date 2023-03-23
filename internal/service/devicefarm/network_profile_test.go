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

func TestAccDeviceFarmNetworkProfile_basic(t *testing.T) {
	var pool devicefarm.NetworkProfile
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-updated")
	resourceName := "aws_devicefarm_network_profile.test"

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
		CheckDestroy:      testAccCheckNetworkProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkProfileExists(resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "downlink_bandwidth_bits", "104857600"),
					resource.TestCheckResourceAttr(resourceName, "uplink_bandwidth_bits", "104857600"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "project_arn", "aws_devicefarm_project.test", "arn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "devicefarm", regexp.MustCompile(`networkprofile:.+`)),
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
					testAccCheckNetworkProfileExists(resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "type", "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "downlink_bandwidth_bits", "104857600"),
					resource.TestCheckResourceAttr(resourceName, "uplink_bandwidth_bits", "104857600"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "project_arn", "aws_devicefarm_project.test", "arn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "devicefarm", regexp.MustCompile(`networkprofile:.+`)),
				),
			},
		},
	})
}

func TestAccDeviceFarmNetworkProfile_tags(t *testing.T) {
	var pool devicefarm.NetworkProfile
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_network_profile.test"

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
		CheckDestroy:      testAccCheckNetworkProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkProfileConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkProfileExists(resourceName, &pool),
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
				Config: testAccNetworkProfileConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkProfileExists(resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccNetworkProfileConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkProfileExists(resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccDeviceFarmNetworkProfile_disappears(t *testing.T) {
	var pool devicefarm.NetworkProfile
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_network_profile.test"

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
		CheckDestroy:      testAccCheckNetworkProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkProfileExists(resourceName, &pool),
					acctest.CheckResourceDisappears(acctest.Provider, tfdevicefarm.ResourceNetworkProfile(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfdevicefarm.ResourceNetworkProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDeviceFarmNetworkProfile_disappears_project(t *testing.T) {
	var pool devicefarm.NetworkProfile
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_network_profile.test"

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
		CheckDestroy:      testAccCheckNetworkProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkProfileExists(resourceName, &pool),
					acctest.CheckResourceDisappears(acctest.Provider, tfdevicefarm.ResourceProject(), "aws_devicefarm_project.test"),
					acctest.CheckResourceDisappears(acctest.Provider, tfdevicefarm.ResourceNetworkProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckNetworkProfileExists(n string, v *devicefarm.NetworkProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DeviceFarmConn
		resp, err := tfdevicefarm.FindNetworkProfileByArn(conn, rs.Primary.ID)
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

func testAccCheckNetworkProfileDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DeviceFarmConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_devicefarm_network_profile" {
			continue
		}

		// Try to find the resource
		_, err := tfdevicefarm.FindNetworkProfileByArn(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("DeviceFarm Network Profile %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccNetworkProfileConfig_basic(rName string) string {
	return testAccProjectConfig_basic(rName) + fmt.Sprintf(`
resource "aws_devicefarm_network_profile" "test" {
  name        = %[1]q
  project_arn = aws_devicefarm_project.test.arn
}
`, rName)
}

func testAccNetworkProfileConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return testAccProjectConfig_basic(rName) + fmt.Sprintf(`
resource "aws_devicefarm_network_profile" "test" {
  name        = %[1]q
  project_arn = aws_devicefarm_project.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccNetworkProfileConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccProjectConfig_basic(rName) + fmt.Sprintf(`
resource "aws_devicefarm_network_profile" "test" {
  name        = %[1]q
  project_arn = aws_devicefarm_project.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
