// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccessGrant_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3control_access_grant.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "access_grant_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "access_grant_id"),
					resource.TestCheckResourceAttrSet(resourceName, "access_grants_location_id"),
					resource.TestCheckResourceAttr(resourceName, "access_grants_location_configuration.#", "0"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttrSet(resourceName, "grant_scope"),
					resource.TestCheckResourceAttr(resourceName, "permission", "READ"),
					resource.TestCheckNoResourceAttr(resourceName, "s3_prefix_type"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func testAccAccessGrant_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3control_access_grant.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfs3control.ResourceAccessGrant, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAccessGrant_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3control_access_grant.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccessGrantConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantsLocationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAccessGrantConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantsLocationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccAccessGrant_locationConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3control_access_grant.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantConfig_locationConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_grants_location_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_grants_location_configuration.0.s3_sub_prefix", "prefix1/prefix2/data.txt"),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix_type", "Object"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"s3_prefix_type"},
			},
		},
	})
}

func testAccCheckAccessGrantDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_access_grant" {
				continue
			}

			_, err := tfs3control.FindAccessGrantByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAccountID], rs.Primary.Attributes["access_grant_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Access Grant %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccessGrantExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).S3ControlClient(ctx)

		_, err := tfs3control.FindAccessGrantByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAccountID], rs.Primary.Attributes["access_grant_id"])

		return err
	}
}

func testAccAccessGrantConfig_baseCustomLocation(rName string) string {
	return acctest.ConfigCompose(testAccAccessGrantsLocationConfig_baseCustomLocation(rName), fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_s3control_access_grants_location" "test" {
  depends_on = [aws_iam_role_policy.test, aws_s3control_access_grants_instance.test]

  iam_role_arn   = aws_iam_role.test.arn
  location_scope = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}*"
}
`, rName))
}

func testAccAccessGrantConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAccessGrantConfig_baseCustomLocation(rName), `
resource "aws_s3control_access_grant" "test" {
  access_grants_location_id = aws_s3control_access_grants_location.test.access_grants_location_id
  permission                = "READ"

  grantee {
    grantee_type       = "IAM"
    grantee_identifier = aws_iam_user.test.arn
  }
}
`)
}

func testAccAccessGrantConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAccessGrantConfig_baseCustomLocation(rName), fmt.Sprintf(`
resource "aws_s3control_access_grant" "test" {
  access_grants_location_id = aws_s3control_access_grants_location.test.access_grants_location_id
  permission                = "READWRITE"

  grantee {
    grantee_type       = "IAM"
    grantee_identifier = aws_iam_user.test.arn
  }

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccAccessGrantConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAccessGrantConfig_baseCustomLocation(rName), fmt.Sprintf(`
resource "aws_s3control_access_grant" "test" {
  access_grants_location_id = aws_s3control_access_grants_location.test.access_grants_location_id
  permission                = "READWRITE"

  grantee {
    grantee_type       = "IAM"
    grantee_identifier = aws_iam_user.test.arn
  }

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAccessGrantConfig_locationConfiguration(rName string) string {
	return acctest.ConfigCompose(testAccAccessGrantConfig_baseCustomLocation(rName), `
resource "aws_s3control_access_grant" "test" {
  access_grants_location_id = aws_s3control_access_grants_location.test.access_grants_location_id
  permission                = "WRITE"

  grantee {
    grantee_type       = "IAM"
    grantee_identifier = aws_iam_user.test.arn
  }

  access_grants_location_configuration {
    s3_sub_prefix = "prefix1/prefix2/data.txt"
  }

  s3_prefix_type = "Object"
}
`)
}
