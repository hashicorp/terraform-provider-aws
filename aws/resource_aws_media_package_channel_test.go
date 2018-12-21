package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediapackage"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSMediaPackageChannel_basic(t *testing.T) {
	resourceName := "aws_media_package_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaPackageChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaPackageChannelConfig(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaPackageChannelExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.0.password", regexp.MustCompile("^[0-9a-f]*$")),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.0.url", regexp.MustCompile("^https://")),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.0.username", regexp.MustCompile("^[0-9a-f]*$")),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.1.password", regexp.MustCompile("^[0-9a-f]*$")),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.1.url", regexp.MustCompile("^https://")),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.1.username", regexp.MustCompile("^[0-9a-f]*$")),
				),
			},
		},
	})
}

func testAccCheckAwsMediaPackageChannelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).mediapackageconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_package_channel" {
			continue
		}

		input := &mediapackage.DescribeChannelInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeChannel(input)
		if err == nil {
			return fmt.Errorf("MediaPackage Channel (%s) not deleted", rs.Primary.ID)
		}

		if !isAWSErr(err, mediapackage.ErrCodeNotFoundException, "") {
			return err
		}
	}

	return nil
}

func testAccCheckAwsMediaPackageChannelExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).mediapackageconn

		input := &mediapackage.DescribeChannelInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeChannel(input)

		return err
	}
}

func testAccMediaPackageChannelConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_package_channel" "test" {
  channel_id = "tf_mediachannel_%s"
}`, rName)
}
