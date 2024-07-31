// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/emr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfemr "github.com/hashicorp/terraform-provider-aws/internal/service/emr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEMRStudioSessionMapping_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var studio emr.SessionMappingDetail
	resourceName := "aws_emr_studio_session_mapping.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	uName := os.Getenv("AWS_IDENTITY_STORE_USER_ID")
	gName := os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckUserID(t)
			testAccPreCheckGroupName(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStudioSessionMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStudioSessionMappingConfig_basic(rName, uName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioSessionMappingExists(ctx, resourceName, &studio),
					resource.TestCheckResourceAttr(resourceName, "identity_id", uName),
					resource.TestCheckResourceAttr(resourceName, "identity_type", "USER"),
					resource.TestCheckResourceAttrPair(resourceName, "studio_id", "aws_emr_studio.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "session_policy_arn", "aws_iam_policy.test", names.AttrARN),
				),
			},
			{
				Config: testAccStudioSessionMappingConfig_basic2(rName, gName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioSessionMappingExists(ctx, resourceName, &studio),
					resource.TestCheckResourceAttr(resourceName, "identity_name", gName),
					resource.TestCheckResourceAttr(resourceName, "identity_type", "GROUP"),
					resource.TestCheckResourceAttrPair(resourceName, "studio_id", "aws_emr_studio.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "session_policy_arn", "aws_iam_policy.test", names.AttrARN),
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
					testAccCheckStudioSessionMappingExists(ctx, resourceName, &studio),
					resource.TestCheckResourceAttr(resourceName, "identity_id", uName),
					resource.TestCheckResourceAttr(resourceName, "identity_type", "USER"),
					resource.TestCheckResourceAttrPair(resourceName, "studio_id", "aws_emr_studio.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "session_policy_arn", "aws_iam_policy.test2", names.AttrARN),
				),
			},
			{
				Config: testAccStudioSessionMappingConfig_updated2(rName, gName, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioSessionMappingExists(ctx, resourceName, &studio),
					resource.TestCheckResourceAttr(resourceName, "identity_name", gName),
					resource.TestCheckResourceAttr(resourceName, "identity_type", "GROUP"),
					resource.TestCheckResourceAttrPair(resourceName, "studio_id", "aws_emr_studio.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "session_policy_arn", "aws_iam_policy.test2", names.AttrARN),
				),
			},
		},
	})
}

func TestAccEMRStudioSessionMapping_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var studio emr.SessionMappingDetail
	resourceName := "aws_emr_studio_session_mapping.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	uName := os.Getenv("AWS_IDENTITY_STORE_USER_ID")
	gName := os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckUserID(t)
			testAccPreCheckGroupName(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStudioSessionMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStudioSessionMappingConfig_basic(rName, uName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioSessionMappingExists(ctx, resourceName, &studio),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfemr.ResourceStudioSessionMapping(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfemr.ResourceStudioSessionMapping(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccStudioSessionMappingConfig_basic2(rName, gName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioSessionMappingExists(ctx, resourceName, &studio),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfemr.ResourceStudioSessionMapping(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfemr.ResourceStudioSessionMapping(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStudioSessionMappingExists(ctx context.Context, resourceName string, studio *emr.SessionMappingDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn(ctx)

		output, err := tfemr.FindStudioSessionMappingByIDOrName(ctx, conn, rs.Primary.ID)
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

func testAccCheckStudioSessionMappingDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_emr_studio_session_mapping" {
				continue
			}

			_, err := tfemr.FindStudioSessionMappingByIDOrName(ctx, conn, rs.Primary.ID)
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
}

func testAccPreCheckUserID(t *testing.T) {
	if os.Getenv("AWS_IDENTITY_STORE_USER_ID") == "" {
		t.Skip("AWS_IDENTITY_STORE_USER_ID env var must be set for AWS Identity Store User acceptance test. " +
			"This is required until ListUsers API returns results without filtering by name.")
	}
}

func testAccPreCheckGroupName(t *testing.T) {
	if os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME") == "" {
		t.Skip("AWS_IDENTITY_STORE_GROUP_NAME env var must be set for AWS Identity Store Group acceptance test. " +
			"This is required until ListGroups API returns results without filtering by name.")
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

func testAccStudioSessionMappingConfig_basic2(rName, gName string) string {
	return acctest.ConfigCompose(testAccStudioSessionMappingConfigBase(rName), fmt.Sprintf(`
resource "aws_emr_studio_session_mapping" "test" {
  studio_id          = aws_emr_studio.test.id
  identity_type      = "GROUP"
  identity_name      = %[1]q
  session_policy_arn = aws_iam_policy.test.arn
}
`, gName))
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

func testAccStudioSessionMappingConfig_updated2(rName, gName, updatedName string) string {
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
  identity_type      = "GROUP"
  identity_name      = %[1]q
  session_policy_arn = aws_iam_policy.test2.arn
}
`, gName, updatedName))
}
