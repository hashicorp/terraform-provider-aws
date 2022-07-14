package s3control_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3control"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccS3ControlAccessPointPolicy_basic(t *testing.T) {
	resourceName := "aws_s3control_access_point_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessPointPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", "true"),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`s3:GetObjectTagging`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAccessPointPolicyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3ControlAccessPointPolicy_disappears(t *testing.T) {
	resourceName := "aws_s3control_access_point_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessPointPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointPolicyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3control.ResourceAccessPointPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlAccessPointPolicy_disappears_AccessPoint(t *testing.T) {
	resourceName := "aws_s3control_access_point_policy.test"
	accessPointResourceName := "aws_s3_access_point.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessPointPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointPolicyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3control.ResourceAccessPoint(), accessPointResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlAccessPointPolicy_update(t *testing.T) {
	resourceName := "aws_s3control_access_point_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessPointPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", "true"),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`s3:GetObjectTagging`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAccessPointPolicyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAccessPointPolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", "true"),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`s3:GetObjectLegalHold`)),
				),
			},
		},
	})
}

func testAccAccessPointPolicyImportStateIdFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return rs.Primary.Attributes["access_point_arn"], nil
	}
}

func testAccCheckAccessPointPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3control_access_point_policy" {
			continue
		}

		accountID, name, err := tfs3control.AccessPointParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, _, err = tfs3control.FindAccessPointPolicyAndStatusByAccountIDAndName(conn, accountID, name)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("S3 Access Point Policy %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAccessPointPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Access Point Policy ID is set")
		}

		accountID, name, err := tfs3control.AccessPointParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn

		_, _, err = tfs3control.FindAccessPointPolicyAndStatusByAccountIDAndName(conn, accountID, name)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAccessPointPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.id
  name   = %[1]q

  public_access_block_configuration {
    block_public_acls       = true
    block_public_policy     = false
    ignore_public_acls      = true
    restrict_public_buckets = false
  }

  lifecycle {
    ignore_changes = [policy]
  }
}

resource "aws_s3control_access_point_policy" "test" {
  access_point_arn = aws_s3_access_point.test.arn

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Effect = "Allow"
      Action = "s3:GetObjectTagging"
      Principal = {
        AWS = "*"
      }
      Resource = "${aws_s3_access_point.test.arn}/object/*"
    }]
  })
}
`, rName)
}

func testAccAccessPointPolicyConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.id
  name   = %[1]q

  public_access_block_configuration {
    block_public_acls       = true
    block_public_policy     = false
    ignore_public_acls      = true
    restrict_public_buckets = false
  }

  lifecycle {
    ignore_changes = [policy]
  }
}

resource "aws_s3control_access_point_policy" "test" {
  access_point_arn = aws_s3_access_point.test.arn

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "s3:GetObjectLegalHold",
        "s3:GetObjectRetention",
      ]
      Principal = {
        AWS = "*"
      }
      Resource = "${aws_s3_access_point.test.arn}/object/prefix/*"
    }]
  })
}
`, rName)
}
