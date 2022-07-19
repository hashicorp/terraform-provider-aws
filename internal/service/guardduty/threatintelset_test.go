package guardduty_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfguardduty "github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
)

func testAccThreatintelset_basic(t *testing.T) {
	bucketName := fmt.Sprintf("tf-test-%s", sdkacctest.RandString(5))
	keyName1 := fmt.Sprintf("tf-%s", sdkacctest.RandString(5))
	keyName2 := fmt.Sprintf("tf-%s", sdkacctest.RandString(5))
	threatintelsetName1 := fmt.Sprintf("tf-%s", sdkacctest.RandString(5))
	threatintelsetName2 := fmt.Sprintf("tf-%s", sdkacctest.RandString(5))
	resourceName := "aws_guardduty_threatintelset.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckThreatintelsetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThreatintelsetConfig_basic(bucketName, keyName1, threatintelsetName1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThreatintelsetExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "guardduty", regexp.MustCompile("detector/.+/threatintelset/.+$")),
					resource.TestCheckResourceAttr(resourceName, "name", threatintelsetName1),
					resource.TestCheckResourceAttr(resourceName, "activate", "true"),
					resource.TestMatchResourceAttr(resourceName, "location", regexp.MustCompile(fmt.Sprintf("%s/%s$", bucketName, keyName1))),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccThreatintelsetConfig_basic(bucketName, keyName2, threatintelsetName2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThreatintelsetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", threatintelsetName2),
					resource.TestCheckResourceAttr(resourceName, "activate", "false"),
					resource.TestMatchResourceAttr(resourceName, "location", regexp.MustCompile(fmt.Sprintf("%s/%s$", bucketName, keyName2))),
				),
			},
		},
	})
}

func testAccThreatintelset_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_guardduty_threatintelset.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckThreatintelsetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThreatintelsetConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThreatintelsetExists(resourceName),
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
				Config: testAccThreatintelsetConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThreatintelsetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccThreatintelsetConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThreatintelsetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckThreatintelsetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_guardduty_threatintelset" {
			continue
		}

		threatIntelSetId, detectorId, err := tfguardduty.DecodeThreatintelsetID(rs.Primary.ID)
		if err != nil {
			return err
		}
		input := &guardduty.GetThreatIntelSetInput{
			ThreatIntelSetId: aws.String(threatIntelSetId),
			DetectorId:       aws.String(detectorId),
		}

		resp, err := conn.GetThreatIntelSet(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
				return nil
			}
			return err
		}

		if *resp.Status == guardduty.ThreatIntelSetStatusDeletePending || *resp.Status == guardduty.ThreatIntelSetStatusDeleted {
			return nil
		}

		return fmt.Errorf("Expected GuardDuty ThreatIntelSet to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckThreatintelsetExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		threatIntelSetId, detectorId, err := tfguardduty.DecodeThreatintelsetID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &guardduty.GetThreatIntelSetInput{
			DetectorId:       aws.String(detectorId),
			ThreatIntelSetId: aws.String(threatIntelSetId),
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn
		_, err = conn.GetThreatIntelSet(input)
		return err
	}
}

func testAccThreatintelsetConfig_basic(bucketName, keyName, threatintelsetName string, activate bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {}

resource "aws_s3_bucket" "test" {
  bucket        = "%s"
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_object" "test" {
  acl     = "public-read"
  content = "10.0.0.0/8\n"
  bucket  = aws_s3_bucket.test.id
  key     = "%s"
}

resource "aws_guardduty_threatintelset" "test" {
  name        = "%s"
  detector_id = aws_guardduty_detector.test.id
  format      = "TXT"
  location    = "https://s3.amazonaws.com/${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
  activate    = %t
}
`, bucketName, keyName, threatintelsetName, activate)
}

func testAccThreatintelsetConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_object" "test" {
  acl     = "public-read"
  content = "10.0.0.0/8\n"
  bucket  = aws_s3_bucket.test.id
  key     = %[1]q
}

resource "aws_guardduty_threatintelset" "test" {
  activate    = true
  detector_id = aws_guardduty_detector.test.id
  format      = "TXT"
  location    = "https://s3.amazonaws.com/${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
  name        = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccThreatintelsetConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_object" "test" {
  acl     = "public-read"
  content = "10.0.0.0/8\n"
  bucket  = aws_s3_bucket.test.id
  key     = %[1]q
}

resource "aws_guardduty_threatintelset" "test" {
  activate    = true
  detector_id = aws_guardduty_detector.test.id
  format      = "TXT"
  location    = "https://s3.amazonaws.com/${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
  name        = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
