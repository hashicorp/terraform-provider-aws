package emr_test

import (
	"fmt"
	"regexp"
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

func TestAccEMRStudio_sso(t *testing.T) {
	var studio emr.Studio
	resourceName := "aws_emr_studio.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStudioDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStudioConfig_sso(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioExists(resourceName, &studio),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "elasticmapreduce", regexp.MustCompile(`studio/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "auth_mode", "SSO"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_security_group_id", "aws_security_group.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "engine_security_group_id", "aws_security_group.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "user_role", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
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

func TestAccEMRStudio_iam(t *testing.T) {
	var studio emr.Studio
	resourceName := "aws_emr_studio.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStudioDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStudioConfig_iam(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioExists(resourceName, &studio),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "auth_mode", "IAM"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_security_group_id", "aws_security_group.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "engine_security_group_id", "aws_security_group.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
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

func TestAccEMRStudio_disappears(t *testing.T) {
	var studio emr.Studio
	resourceName := "aws_emr_studio.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStudioDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStudioConfig_sso(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioExists(resourceName, &studio),
					acctest.CheckResourceDisappears(acctest.Provider, tfemr.ResourceStudio(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfemr.ResourceStudio(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEMRStudio_tags(t *testing.T) {
	var studio emr.Studio
	resourceName := "aws_emr_studio.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStudioDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStudioConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioExists(resourceName, &studio),
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
				Config: testAccStudioConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioExists(resourceName, &studio),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccStudioConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioExists(resourceName, &studio),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckStudioExists(resourceName string, studio *emr.Studio) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn

		output, err := tfemr.FindStudioByID(conn, rs.Primary.ID)
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

func testAccCheckStudioDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_emr_studio" {
			continue
		}

		_, err := tfemr.FindStudioByID(conn, rs.Primary.ID)
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

func testAccStudioConfigBase(rName string) string {
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
`, rName))
}

func testAccStudioConfig_sso(rName string) string {
	return acctest.ConfigCompose(testAccStudioConfigBase(rName), fmt.Sprintf(`
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
`, rName))
}

func testAccStudioConfig_iam(rName string) string {
	return acctest.ConfigCompose(testAccStudioConfigBase(rName), fmt.Sprintf(`
resource "aws_emr_studio" "test" {
  auth_mode                   = "IAM"
  default_s3_location         = "s3://${aws_s3_bucket.test.bucket}/test"
  engine_security_group_id    = aws_security_group.test.id
  name                        = %[1]q
  service_role                = aws_iam_role.test.arn
  subnet_ids                  = [aws_subnet.test.id]
  vpc_id                      = aws_vpc.test.id
  workspace_security_group_id = aws_security_group.test.id
}
`, rName))
}

func testAccStudioConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccStudioConfigBase(rName), fmt.Sprintf(`
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

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccStudioConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccStudioConfigBase(rName), fmt.Sprintf(`
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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
