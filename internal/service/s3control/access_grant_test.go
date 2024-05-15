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

func testAccAccessGrant_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_access_grant.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "access_grant_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "access_grant_id"),
					resource.TestCheckResourceAttrSet(resourceName, "access_grants_location_id"),
					resource.TestCheckResourceAttr(resourceName, "access_grants_location_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttrSet(resourceName, "grant_scope"),
					resource.TestCheckResourceAttr(resourceName, "permission", "READ"),
					resource.TestCheckNoResourceAttr(resourceName, "s3_prefix_type"),
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

func testAccAccessGrant_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_access_grant.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceAccessGrant, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAccessGrant_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_access_grant.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, resourceName),
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
				Config: testAccAccessGrantConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantsLocationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAccessGrantConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantsLocationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccAccessGrant_locationConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_access_grant.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantConfig_locationConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_grants_location_configuration.#", acctest.Ct1),
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

func testAccCheckAccessGrantDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_access_grant" {
				continue
			}

			_, err := tfs3control.FindAccessGrantByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAccountID], rs.Primary.Attributes["access_grant_id"])

			if tfresource.NotFound(err) {
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

func testAccCheckAccessGrantExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

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
