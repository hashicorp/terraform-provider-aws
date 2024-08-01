// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccessGrantsLocation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_access_grants_location.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantsLocationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantsLocationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessGrantsLocationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "access_grants_location_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "access_grants_location_id"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "location_scope", "s3://"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func testAccAccessGrantsLocation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_access_grants_location.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantsLocationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantsLocationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantsLocationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceAccessGrantsLocation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAccessGrantsLocation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_access_grants_location.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantsLocationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantsLocationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantsLocationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccessGrantsLocationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantsLocationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAccessGrantsLocationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantsLocationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccAccessGrantsLocation_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_access_grants_location.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantsLocationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantsLocationConfig_customLocation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantsLocationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrIAMRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "location_scope", fmt.Sprintf("s3://%s/prefixA*", rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccessGrantsLocationConfig_customLocationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantsLocationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrIAMRoleARN, "aws_iam_role.test2", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "location_scope", fmt.Sprintf("s3://%s/prefixA*", rName)),
				),
			},
		},
	})
}

func testAccCheckAccessGrantsLocationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_access_grants_location" {
				continue
			}

			_, err := tfs3control.FindAccessGrantsLocationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAccountID], rs.Primary.Attributes["access_grants_location_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Access Grants Location %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccessGrantsLocationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		_, err := tfs3control.FindAccessGrantsLocationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAccountID], rs.Primary.Attributes["access_grants_location_id"])

		return err
	}
}

func testAccAccessGrantsLocationConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = ["sts:AssumeRole", "sts:SetSourceIdentity"],
      Principal = {
        Service = "access-grants.s3.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "s3:*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_s3control_access_grants_instance" "test" {}
`, rName)
}

func testAccAccessGrantsLocationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAccessGrantsLocationConfig_base(rName), `
resource "aws_s3control_access_grants_location" "test" {
  depends_on = [aws_iam_role_policy.test, aws_s3control_access_grants_instance.test]

  iam_role_arn   = aws_iam_role.test.arn
  location_scope = "s3://"
}
`)
}

func testAccAccessGrantsLocationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAccessGrantsLocationConfig_base(rName), fmt.Sprintf(`
resource "aws_s3control_access_grants_location" "test" {
  depends_on = [aws_iam_role_policy.test, aws_s3control_access_grants_instance.test]

  iam_role_arn   = aws_iam_role.test.arn
  location_scope = "s3://"

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccAccessGrantsLocationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAccessGrantsLocationConfig_base(rName), fmt.Sprintf(`
resource "aws_s3control_access_grants_location" "test" {
  depends_on = [aws_iam_role_policy.test, aws_s3control_access_grants_instance.test]

  iam_role_arn   = aws_iam_role.test.arn
  location_scope = "s3://"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAccessGrantsLocationConfig_baseCustomLocation(rName string) string {
	return acctest.ConfigCompose(testAccAccessGrantsLocationConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "prefixA"
}
`, rName))
}

func testAccAccessGrantsLocationConfig_customLocation(rName string) string {
	return acctest.ConfigCompose(testAccAccessGrantsLocationConfig_baseCustomLocation(rName), `
resource "aws_s3control_access_grants_location" "test" {
  depends_on = [aws_iam_role_policy.test, aws_s3control_access_grants_instance.test]

  iam_role_arn   = aws_iam_role.test.arn
  location_scope = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}*"
}
`)
}

func testAccAccessGrantsLocationConfig_customLocationUpdated(rName string) string {
	return acctest.ConfigCompose(testAccAccessGrantsLocationConfig_baseCustomLocation(rName), fmt.Sprintf(`
resource "aws_iam_role" "test2" {
  name = "%[1]s-2"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = ["sts:AssumeRole", "sts:SetSourceIdentity"],
      Principal = {
        Service = "access-grants.s3.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
    }]
  })
}

resource "aws_iam_role_policy" "test2" {
  name = "%[1]s-2"
  role = aws_iam_role.test2.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "s3:*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_s3control_access_grants_location" "test" {
  depends_on = [aws_iam_role_policy.test2, aws_s3control_access_grants_instance.test]

  iam_role_arn   = aws_iam_role.test2.arn
  location_scope = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}*"
}
`, rName))
}
