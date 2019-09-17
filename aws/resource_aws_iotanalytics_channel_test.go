package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIoTAnalyticsChannel_basic(t *testing.T) {
	rString := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsChannel_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsChannelExists_basic("aws_iotanalytics_channel.channel"),
					resource.TestCheckResourceAttr("aws_iotanalytics_channel.channel", "name", fmt.Sprintf("test_channel_%s", rString)),
				),
			},
		},
	})
}

func testAccCheckAWSIoTAnalyticsChannelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iotanalytics_channel" {
			continue
		}

		params := &iotanalytics.DescribeChannelInput{
			ChannelName: aws.String(rs.Primary.ID),
		}
		_, err := conn.DescribeChannel(params)

		if err != nil {
			if isAWSErr(err, iotanalytics.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}
		return fmt.Errorf("Expected IoTAnalytics Channel to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSIoTAnalyticsChannelExists_basic(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccAWSIoTAnalyticsChannel_basic(rString string) string {
	return fmt.Sprintf(`
resource "aws_iotanalytics_channel" "channel" {
  name = "test_channel_%[1]s"

  storage {
	  service_managed_s3 {}
  }
}
`, rString)
}
