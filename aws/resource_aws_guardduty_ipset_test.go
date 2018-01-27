package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsGuardDutyIpset_basic(t *testing.T) {
	rName := acctest.RandString(5)
	modName := acctest.RandString(5)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyIpsetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyIpsetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyIpsetExists("aws_guardduty_ipset.test"),
					resource.TestCheckResourceAttr("aws_guardduty_ipset.test", "name", fmt.Sprintf("tf-%s", rName)),
					resource.TestCheckResourceAttr("aws_guardduty_ipset.test", "activate", "true"),
					resource.TestMatchResourceAttr(
						"aws_guardduty_ipset.test", "location", regexp.MustCompile(fmt.Sprintf("tf-test-%s/tf-%s$", rName, rName)),
					),
				),
			},
			{
				Config: testAccGuardDutyIpsetConfig_update(rName, modName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyIpsetExists("aws_guardduty_ipset.test"),
					resource.TestCheckResourceAttr("aws_guardduty_ipset.test", "name", fmt.Sprintf("tf-%s", modName)),
					resource.TestCheckResourceAttr("aws_guardduty_ipset.test", "activate", "false"),
					resource.TestMatchResourceAttr(
						"aws_guardduty_ipset.test", "location", regexp.MustCompile(fmt.Sprintf("tf-test-%s/tf-update-%s$", rName, rName)),
					),
				),
			},
		},
	})
}

func testAccCheckAwsGuardDutyIpsetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).guarddutyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_guardduty_ipset" {
			continue
		}

		input := &guardduty.GetIPSetInput{
			IpSetId:    aws.String(rs.Primary.ID),
			DetectorId: aws.String(rs.Primary.Attributes["detector_id"]),
		}

		resp, err := conn.GetIPSet(input)
		if err != nil {
			if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
				return nil
			}
			return err
		}

		if *resp.Status == guardduty.IpSetStatusDeletePending || *resp.Status == guardduty.IpSetStatusDeleted {
			return nil
		}

		return fmt.Errorf("Expected GuardDuty Ipset to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsGuardDutyIpsetExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccGuardDutyIpsetConfig_basic(rName string) string {
	return fmt.Sprintf(`
%s

resource "aws_s3_bucket" "test" {
  acl = "private"
  bucket = "tf-test-%s"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  acl     = "public-read"
  content = "10.0.0.0/8\n"
  bucket  = "${aws_s3_bucket.test.id}"
  key     = "tf-%s"
}

resource "aws_guardduty_ipset" "test" {
  name = "tf-%s"
  activate = true
  detector_id = "${aws_guardduty_detector.test.id}"
  format = "TXT"
  location = "https://s3.amazonaws.com/${aws_s3_bucket_object.test.bucket}/${aws_s3_bucket_object.test.key}"
}
`, testAccGuardDutyDetectorConfig_basic1, rName, rName, rName)
}

func testAccGuardDutyIpsetConfig_update(rName, modName string) string {
	return fmt.Sprintf(`
%s

resource "aws_s3_bucket" "test" {
  acl = "private"
  bucket = "tf-test-%s"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test_update" {
  acl     = "public-read"
  content = "10.0.0.0/8\n"
  bucket  = "${aws_s3_bucket.test.id}"
  key     = "tf-update-%s"
}

resource "aws_guardduty_ipset" "test" {
  name = "tf-%s"
  activate = false
  detector_id = "${aws_guardduty_detector.test.id}"
  format = "TXT"
  location = "https://s3.amazonaws.com/${aws_s3_bucket_object.test_update.bucket}/${aws_s3_bucket_object.test_update.key}"
}
`, testAccGuardDutyDetectorConfig_basic1, rName, rName, modName)
}
