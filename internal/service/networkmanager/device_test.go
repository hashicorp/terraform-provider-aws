package networkmanager

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func init() {
	resource.AddTestSweepers("aws_networkmanager_device", &resource.Sweeper{
		Name: "aws_networkmanager_device",
		F:    testSweepDevice,
	})
}

func testSweepDevice(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).networkmanagerconn
	var sweeperErrs *multierror.Error

	err = conn.GetDevicesPages(&networkmanager.GetDevicesInput{},
		func(page *networkmanager.GetDevicesOutput, lastPage bool) bool {
			for _, device := range page.Devices {
				input := &networkmanager.DeleteDeviceInput{
					GlobalNetworkId: device.GlobalNetworkId,
					DeviceId:        device.DeviceId,
				}
				id := aws.StringValue(device.DeviceId)
				globalNetworkID := aws.StringValue(device.GlobalNetworkId)

				log.Printf("[INFO] Deleting Network Manager Device: %s", id)
				_, err := conn.DeleteDevice(input)

				if tfawserr.ErrCodeEquals(err, "InvalidDeviceID.NotFound", "") {
					continue
				}

				if err != nil {
					sweeperErr := fmt.Errorf("failed to delete Network Manager Device %s: %s", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}

				if err := waitForDeviceDeletion(conn, globalNetworkID, id); err != nil {
					sweeperErr := fmt.Errorf("error waiting for Network Manager Device (%s) deletion: %s", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}
			}
			return !lastPage
		})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Network Manager Device sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving Network Manager Devices: %s", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccDevice_basic(t *testing.T) {
	resourceName := "aws_networkmanager_device.test"
	siteResourceName := "aws_networkmanager_site.test"
	site2ResourceName := "aws_networkmanager_site.test2"
	gloablNetworkResourceName := "aws_networkmanager_global_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceConfig("test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDeviceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "global_network_id", gloablNetworkResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "site_id", siteResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "model", "abc"),
					resource.TestCheckResourceAttr(resourceName, "serial_number", "123"),
					resource.TestCheckResourceAttr(resourceName, "type", "office device"),
					resource.TestCheckResourceAttr(resourceName, "vendor", "company"),
					resource.TestCheckResourceAttr(resourceName, "location.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeviceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccDeviceConfig_Update("test updated", "def", "456", "home device", "othercompany", "18.0029784", "-76.7897987"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDeviceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "global_network_id", gloablNetworkResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "site_id", site2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "description", "test updated"),
					resource.TestCheckResourceAttr(resourceName, "model", "def"),
					resource.TestCheckResourceAttr(resourceName, "serial_number", "456"),
					resource.TestCheckResourceAttr(resourceName, "type", "home device"),
					resource.TestCheckResourceAttr(resourceName, "vendor", "othercompany"),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.address", ""),
					resource.TestCheckResourceAttr(resourceName, "location.0.latitude", "18.0029784"),
					resource.TestCheckResourceAttr(resourceName, "location.0.longitude", "-76.7897987"),
				),
			},
		},
	})
}

func TestAccDevice_tags(t *testing.T) {
	resourceName := "aws_networkmanager_device.test"
	description := "test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceConfigTags1(description, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeviceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccDeviceConfigTags2(description, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDeviceConfigTags1(description, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsDeviceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).networkmanagerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkmanager_device" {
			continue
		}

		device, err := networkmanagerDescribeDevice(conn, rs.Primary.Attributes["global_network_id"], rs.Primary.ID)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeValidationException, "") {
				return nil
			}
			return err
		}

		if device == nil {
			continue
		}

		return fmt.Errorf("Expected Device to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsDeviceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).networkmanagerconn

		device, err := networkmanagerDescribeDevice(conn, rs.Primary.Attributes["global_network_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		if device == nil {
			return fmt.Errorf("Network Manager Device not found")
		}

		if aws.StringValue(device.State) != networkmanager.DeviceStateAvailable && aws.StringValue(device.State) != networkmanager.DeviceStatePending {
			return fmt.Errorf("Network Manager Device (%s) exists in (%s) state", rs.Primary.ID, aws.StringValue(device.State))
		}

		return err
	}
}

func testAccDeviceConfig(description string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
 description = "test"
}

resource "aws_networkmanager_site" "test" {
 description       = "test"
 global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_device" "test" {
 description       = %q
 global_network_id = aws_networkmanager_global_network.test.id
 site_id           = aws_networkmanager_site.test.id
 model             = "abc"
 serial_number     = "123"
 type              = "office device"
 vendor            = "company"
}
`, description)
}

func testAccDeviceConfigTags1(description, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
 description = "test"
}

resource "aws_networkmanager_site" "test" {
 description       = "test"
 global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_device" "test" {
 description       = %q
 global_network_id = aws_networkmanager_global_network.test.id
 site_id           = aws_networkmanager_site.test.id

  tags = {
    %q = %q
  }
}
`, description, tagKey1, tagValue1)
}

func testAccDeviceConfigTags2(description, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
 description = "test"
}

resource "aws_networkmanager_site" "test" {
 description       = "test"
 global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_device" "test" {
 description       = %q
 global_network_id = aws_networkmanager_global_network.test.id

  tags = {
   %q = %q
   %q = %q
  }
}
`, description, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDeviceConfig_Update(description, model, serial_number, device_type, vendor, latitude, longitude string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
 description = "test"
}

resource "aws_networkmanager_site" "test" {
 description       = "test"
 global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_site" "test2" {
 description       = "test2"
 global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_device" "test" {
 description       = %q
 global_network_id = aws_networkmanager_global_network.test.id
 site_id           = aws_networkmanager_site.test2.id
 model             = %q
 serial_number     = %q
 type              = %q
 vendor            = %q

 location {
  latitude  = %q	
  longitude = %q
 }
}
`, description, model, serial_number, device_type, vendor, latitude, longitude)
}

func testAccDeviceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["arn"], nil
	}
}
