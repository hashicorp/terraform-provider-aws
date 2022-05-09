package emr_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/emr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfemr "github.com/hashicorp/terraform-provider-aws/internal/service/emr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEMRStudioSessionMapping_basic(t *testing.T) {
	var studio emr.SessionMappingDetail
	resourceName := "aws_emr_studio_session_mapping.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	uName := os.Getenv("AWS_IDENTITY_STORE_USER_ID")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckUserID(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStudioSessionMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStudioSessionMappingConfig_basic(rName, uName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioSessionMappingExists(resourceName, &studio),
					resource.TestCheckResourceAttr(resourceName, "identity_id", uName),
					resource.TestCheckResourceAttr(resourceName, "identity_type", "USER"),
					resource.TestCheckResourceAttrPair(resourceName, "studio_id", "aws_emr_studio.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "session_policy_arn", "aws_iam_policy.test", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStudioSessionMappingConfig_updated(rName, uName, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioSessionMappingExists(resourceName, &studio),
					resource.TestCheckResourceAttr(resourceName, "identity_id", uName),
					resource.TestCheckResourceAttr(resourceName, "identity_type", "USER"),
					resource.TestCheckResourceAttrPair(resourceName, "studio_id", "aws_emr_studio.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "session_policy_arn", "aws_iam_policy.test2", "arn"),
				),
			},
		},
	})
}

func TestAccEMRStudioSessionMapping_disappears(t *testing.T) {
	var studio emr.SessionMappingDetail
	resourceName := "aws_emr_studio_session_mapping.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	uName := os.Getenv("AWS_IDENTITY_STORE_USER_ID")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckUserID(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStudioSessionMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStudioSessionMappingConfig_basic(rName, uName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioSessionMappingExists(resourceName, &studio),
					acctest.CheckResourceDisappears(acctest.Provider, tfemr.ResourceStudioSessionMapping(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfemr.ResourceStudioSessionMapping(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStudioSessionMappingExists(resourceName string, studio *emr.SessionMappingDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn

		output, err := tfemr.FindStudioSessionMappingByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("EMR Studio (%s) not found", rs.Primary.ID)
		}

		*studio = *output

		return nil
	}
}

func testAccCheckStudioSessionMappingDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_emr_studio_session_mapping" {
			continue
		}

		_, err := tfemr.FindStudioSessionMappingByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EMR Studio %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccPreCheckUserID(t *testing.T) {
	if os.Getenv("AWS_IDENTITY_STORE_USER_ID") == "" {
		t.Skip("AWS_IDENTITY_STORE_USER_ID env var must be set for AWS Identity Store User acceptance test. " +
			"This is required until ListUsers API returns results without filtering by name.")
	}
}

func testAccStudioSessionMappingConfigBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy" "test" {
  name   = %[1]q
  role   = aws_iam_role.test.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}/*",
		"${aws_s3_bucket.test.arn}"
      ]
    }
  ]
}
EOF
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_emr_studio" "test" {
  auth_mode                   = "SSO"
  default_s3_location         = "s3://${aws_s3_bucket.test.bucket}/test"
  engine_security_group_id    = aws_security_group.test.id
  name                        = %[1]q
  service_role                = aws_iam_role.test.arn
  subnet_ids                  = [aws_subnet.test.id]
  user_role                   = aws_iam_role.test.arn
  vpc_id                      = aws_vpc.test.id
  workspace_security_group_id = aws_security_group.test.id
}

resource "aws_iam_policy" "test" {
  name   = %[1]q
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}/*",
		"${aws_s3_bucket.test.arn}"
      ]
    }
  ]
}
EOF
}
`, rName))
}

func testAccStudioSessionMappingConfig_basic(rName, uName string) string {
	return acctest.ConfigCompose(testAccStudioSessionMappingConfigBase(rName), fmt.Sprintf(`
resource "aws_emr_studio_session_mapping" "test" {
  studio_id          = aws_emr_studio.test.id
  identity_type      = "USER"
  identity_id        = %[1]q
  session_policy_arn = aws_iam_policy.test.arn
}
`, uName))
}

func testAccStudioSessionMappingConfig_updated(rName, uName, updatedName string) string {
	return acctest.ConfigCompose(testAccStudioSessionMappingConfigBase(rName), fmt.Sprintf(`
resource "aws_iam_policy" "test2" {
  name   = %[2]q
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}/*",
		"${aws_s3_bucket.test.arn}"
      ]
    }
  ]
}
EOF
}

resource "aws_emr_studio_session_mapping" "test" {
  studio_id          = aws_emr_studio.test.id
  identity_type      = "USER"
  identity_id        = %[1]q
  session_policy_arn = aws_iam_policy.test2.arn
}
`, uName, updatedName))
}
