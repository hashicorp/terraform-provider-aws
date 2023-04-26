package macie_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/macie"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmacie "github.com/hashicorp/terraform-provider-aws/internal/service/macie"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccMacieS3BucketAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_macie_s3_bucket_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, macie.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckS3BucketAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccS3BucketAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "classification_type.0.continuous", "FULL"),
					resource.TestCheckResourceAttr(resourceName, "classification_type.0.one_time", "NONE"),
				),
			},
			{
				Config: testAccS3BucketAssociationConfig_basicOneTime(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "classification_type.0.continuous", "FULL"),
					resource.TestCheckResourceAttr(resourceName, "classification_type.0.one_time", "FULL"),
				),
			},
		},
	})
}

func TestAccMacieS3BucketAssociation_accountIDAndPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_macie_s3_bucket_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, macie.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckS3BucketAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccS3BucketAssociationConfig_accountIDAndPrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "classification_type.0.continuous", "FULL"),
					resource.TestCheckResourceAttr(resourceName, "classification_type.0.one_time", "NONE"),
				),
			},
			{
				Config: testAccS3BucketAssociationConfig_accountIDAndPrefixOneTime(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "classification_type.0.continuous", "FULL"),
					resource.TestCheckResourceAttr(resourceName, "classification_type.0.one_time", "FULL"),
				),
			},
		},
	})
}

func testAccCheckS3BucketAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MacieConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_macie_s3_bucket_association" {
				continue
			}

			_, err := tfmacie.FindS3ResourceClassificationByThreePartKey(ctx, conn, rs.Primary.Attributes["member_account_id"], rs.Primary.Attributes["bucket_name"], rs.Primary.Attributes["prefix"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf(" Macie Classic S3 Bucket Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckS3BucketAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Macie Classic S3 Bucket Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MacieConn()

		_, err := tfmacie.FindS3ResourceClassificationByThreePartKey(ctx, conn, rs.Primary.Attributes["member_account_id"], rs.Primary.Attributes["bucket_name"], rs.Primary.Attributes["prefix"])

		return err
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MacieConn()

	input := &macie.ListS3ResourcesInput{}

	_, err := conn.ListS3ResourcesWithContext(ctx, input)

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

func testAccS3BucketAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_macie_s3_bucket_association" "test" {
  bucket_name = aws_s3_bucket.test.id
}
`, rName)
}

func testAccS3BucketAssociationConfig_basicOneTime(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_macie_s3_bucket_association" "test" {
  bucket_name = aws_s3_bucket.test.id

  classification_type {
    one_time = "FULL"
  }
}
`, rName)
}

func testAccS3BucketAssociationConfig_accountIDAndPrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_caller_identity" "current" {}

resource "aws_macie_s3_bucket_association" "test" {
  bucket_name       = aws_s3_bucket.test.id
  member_account_id = data.aws_caller_identity.current.account_id
  prefix            = "data"
}
`, rName)
}

func testAccS3BucketAssociationConfig_accountIDAndPrefixOneTime(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
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
`, rName)
}
