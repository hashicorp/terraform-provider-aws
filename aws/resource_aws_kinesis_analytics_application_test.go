package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSKinesisAnalyticsApplication_basic(t *testing.T) {
	var application kinesisanalytics.ApplicationDetail
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplication_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists("aws_kinesis_analytics_application.test", &application),
				),
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_update(t *testing.T) {
	var application kinesisanalytics.ApplicationDetail
	resName := "aws_kinesis_analytics_application.test"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplication_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &application),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplication_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "code", "testCode\n"),
				),
			},
		},
	})
}

func testAccCheckKinesisAnalyticsApplicationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kinesis_analytics_application" {
			continue
		}
		conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsconn
		describeOpts := &kinesisanalytics.DescribeApplicationInput{
			ApplicationName: aws.String(rs.Primary.Attributes["name"]),
		}
		resp, err := conn.DescribeApplication(describeOpts)
		if err == nil {
			if resp.ApplicationDetail != nil && *resp.ApplicationDetail.ApplicationStatus != "DELETING" {
				return fmt.Errorf("Error: Application still exists")
			}
		}
		return nil
	}
	return nil
}

func testAccCheckKinesisAnalyticsApplicationExists(n string, application *kinesisanalytics.ApplicationDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kinesis Analytics Application ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsconn
		describeOpts := &kinesisanalytics.DescribeApplicationInput{
			ApplicationName: aws.String(rs.Primary.Attributes["name"]),
		}
		resp, err := conn.DescribeApplication(describeOpts)
		if err != nil {
			return err
		}

		*application = *resp.ApplicationDetail

		return nil
	}
}

func testAccKinesisAnalyticsApplication_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = "testAcc-%d"
}
`, rInt)
}

func testAccKinesisAnalyticsApplication_update(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = "testAcc-%d"
  code = "testCode\n"
}`, rInt)
}
