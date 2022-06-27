package macie_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccMacieS3BucketAssociation_basic(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, macie.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckS3BucketAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccS3BucketAssociationConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketAssociationExists("aws_macie_s3_bucket_association.test"),
					resource.TestCheckResourceAttr("aws_macie_s3_bucket_association.test", "classification_type.0.continuous", "FULL"),
					resource.TestCheckResourceAttr("aws_macie_s3_bucket_association.test", "classification_type.0.one_time", "NONE"),
				),
			},
			{
				Config: testAccS3BucketAssociationConfig_basicOneTime(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketAssociationExists("aws_macie_s3_bucket_association.test"),
					resource.TestCheckResourceAttr("aws_macie_s3_bucket_association.test", "classification_type.0.continuous", "FULL"),
					resource.TestCheckResourceAttr("aws_macie_s3_bucket_association.test", "classification_type.0.one_time", "FULL"),
				),
			},
		},
	})
}

func TestAccMacieS3BucketAssociation_accountIdAndPrefix(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, macie.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckS3BucketAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccS3BucketAssociationConfig_accountIdAndPrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketAssociationExists("aws_macie_s3_bucket_association.test"),
					resource.TestCheckResourceAttr("aws_macie_s3_bucket_association.test", "classification_type.0.continuous", "FULL"),
					resource.TestCheckResourceAttr("aws_macie_s3_bucket_association.test", "classification_type.0.one_time", "NONE"),
				),
			},
			{
				Config: testAccS3BucketAssociationConfig_accountIdAndPrefixOneTime(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketAssociationExists("aws_macie_s3_bucket_association.test"),
					resource.TestCheckResourceAttr("aws_macie_s3_bucket_association.test", "classification_type.0.continuous", "FULL"),
					resource.TestCheckResourceAttr("aws_macie_s3_bucket_association.test", "classification_type.0.one_time", "FULL"),
				),
			},
		},
	})
}

func testAccCheckS3BucketAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MacieConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie_s3_bucket_association" {
			continue
		}

		req := &macie.ListS3ResourcesInput{}
		acctId := rs.Primary.Attributes["member_account_id"]
		if acctId != "" {
			req.MemberAccountId = aws.String(acctId)
		}

		dissociated := true
		err := conn.ListS3ResourcesPages(req, func(page *macie.ListS3ResourcesOutput, lastPage bool) bool {
			for _, v := range page.S3Resources {
				if aws.StringValue(v.BucketName) == rs.Primary.Attributes["bucket_name"] && aws.StringValue(v.Prefix) == rs.Primary.Attributes["prefix"] {
					dissociated = false
					return false
				}
			}

			return true
		})
		if err != nil {
			return err
		}

		if !dissociated {
			return fmt.Errorf("S3 resource %s/%s is not dissociated from Macie", rs.Primary.Attributes["bucket_name"], rs.Primary.Attributes["prefix"])
		}
	}
	return nil
}

func testAccCheckS3BucketAssociationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MacieConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		req := &macie.ListS3ResourcesInput{}
		acctId := rs.Primary.Attributes["member_account_id"]
		if acctId != "" {
			req.MemberAccountId = aws.String(acctId)
		}

		exists := false
		err := conn.ListS3ResourcesPages(req, func(page *macie.ListS3ResourcesOutput, lastPage bool) bool {
			for _, v := range page.S3Resources {
				if aws.StringValue(v.BucketName) == rs.Primary.Attributes["bucket_name"] && aws.StringValue(v.Prefix) == rs.Primary.Attributes["prefix"] {
					exists = true
					return false
				}
			}

			return true
		})
		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("S3 resource %s/%s is not associated with Macie", rs.Primary.Attributes["bucket_name"], rs.Primary.Attributes["prefix"])
		}

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MacieConn

	input := &macie.ListS3ResourcesInput{}

	_, err := conn.ListS3Resources(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if tfawserr.ErrMessageContains(err, macie.ErrCodeInvalidInputException, "Macie is not enabled for this AWS account") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccS3BucketAssociationConfig_basic(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "tf-test-macie-bucket-%d"
}

resource "aws_macie_s3_bucket_association" "test" {
  bucket_name = aws_s3_bucket.test.id
}
`, randInt)
}

func testAccS3BucketAssociationConfig_basicOneTime(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "tf-test-macie-bucket-%d"
}

resource "aws_macie_s3_bucket_association" "test" {
  bucket_name = aws_s3_bucket.test.id

  classification_type {
    one_time = "FULL"
  }
}
`, randInt)
}

func testAccS3BucketAssociationConfig_accountIdAndPrefix(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "tf-test-macie-bucket-%d"
}

data "aws_caller_identity" "current" {}

resource "aws_macie_s3_bucket_association" "test" {
  bucket_name       = aws_s3_bucket.test.id
  member_account_id = data.aws_caller_identity.current.account_id
  prefix            = "data"
}
`, randInt)
}

func testAccS3BucketAssociationConfig_accountIdAndPrefixOneTime(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "tf-test-macie-bucket-%d"
}

data "aws_caller_identity" "current" {}

resource "aws_macie_s3_bucket_association" "test" {
  bucket_name       = aws_s3_bucket.test.id
  member_account_id = data.aws_caller_identity.current.account_id
  prefix            = "data"

  classification_type {
    one_time = "FULL"
  }
}
`, randInt)
}
