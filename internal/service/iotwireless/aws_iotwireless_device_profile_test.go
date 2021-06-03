package iotwireless_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotwireless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccAWSIotWirelessDeviceProfile_basic(t *testing.T) {
	resourceName := "aws_iotwireless_device_profile.test"
	rString := sdkacctest.RandString(8)
	rName := fmt.Sprintf("tf-acc-test-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iotwireless.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAWSIotWirelessDeviceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotWirelessDeviceProfileConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotWirelessDeviceProfileExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSIotWirelessDeviceProfile_Tags(t *testing.T) {
	resourceName := "aws_iotwireless_device_profile.test"
	rString := sdkacctest.RandString(8)
	rName := fmt.Sprintf("tf-acc-test-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iotwireless.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAWSIotWirelessDeviceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotWirelessDeviceProfileConfigTags(rName, "key1", "value1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotWirelessDeviceProfileExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
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

func testAccCheckAWSIotWirelessDeviceProfileDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTWirelessConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iotwireless_device_profile" {
			continue
		}

		output, err := conn.GetDeviceProfile(&iotwireless.GetDeviceProfileInput{
			Id: aws.String(rs.Primary.ID),
		})

		if tfawserr.ErrCodeEquals(err, iotwireless.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil && output.Arn != nil {
			return fmt.Errorf("IoT Wireless Device Profile (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSIotWirelessDeviceProfileExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no resource ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTWirelessConn

		output, err := conn.GetDeviceProfile(&iotwireless.GetDeviceProfileInput{
			Id: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if output == nil || output.Arn == nil {
			return fmt.Errorf("IoT Wireless Device Profile (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAWSIotWirelessDeviceProfileConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iotwireless_device_profile" "test" {
  name = %[1]q

  lorawan {
    factory_preset_freqs_list = [
      9168000,
      9170000,
      9172000,
      9174000,
      9176000,
      9178000,
      9180000,
      9182000,
      9175000,
    ]
    rx_freq_2           = 9233000
    mac_version         = "1.0.3"
    reg_params_revision = "Regional Parameters v1.0.3rA"
    max_eirp            = 15
    max_duty_cycle      = 10
    rf_region           = "AU915"
    supports_join       = true
    supports_32bit_fcnt = true
  }
}
`, rName)
}

func testAccAWSIotWirelessDeviceProfileConfigTags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iotwireless_device_profile" "test" {
  name = %[1]q

  lorawan {
    factory_preset_freqs_list = [
      9168000,
      9170000,
      9172000,
      9174000,
      9176000,
      9178000,
      9180000,
      9182000,
      9175000,
    ]
    rx_freq_2           = 9233000
    mac_version         = "1.0.3"
    reg_params_revision = "Regional Parameters v1.0.3rA"
    max_eirp            = 15
    max_duty_cycle      = 10
    rf_region           = "AU915"
    supports_join       = true
    supports_32bit_fcnt = true
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
