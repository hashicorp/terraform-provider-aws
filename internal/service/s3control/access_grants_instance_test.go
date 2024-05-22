// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccessGrantsInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_access_grants_instance.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantsInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantsInstanceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessGrantsInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "access_grants_instance_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "access_grants_instance_id"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					resource.TestCheckNoResourceAttr(resourceName, "identity_center_application_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "identity_center_arn"),
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

func testAccAccessGrantsInstance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_access_grants_instance.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantsInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantsInstanceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantsInstanceExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceAccessGrantsInstance, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAccessGrantsInstance_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_access_grants_instance.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantsInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantsInstanceConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantsInstanceExists(ctx, resourceName),
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
				Config: testAccAccessGrantsInstanceConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantsInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAccessGrantsInstanceConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantsInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccAccessGrantsInstance_identityCenter(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_access_grants_instance.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantsInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantsInstanceConfig_identityCenter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessGrantsInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "identity_center_application_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "identity_center_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"identity_center_arn"},
			},
			{
				Config: testAccAccessGrantsInstanceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessGrantsInstanceExists(ctx, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "identity_center_application_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "identity_center_arn"),
				),
			},
			{
				Config: testAccAccessGrantsInstanceConfig_identityCenter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessGrantsInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "identity_center_application_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "identity_center_arn"),
				),
			},
		},
	})
}

func testAccCheckAccessGrantsInstanceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_access_grants_instance" {
				continue
			}

			_, err := tfs3control.FindAccessGrantsInstance(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Access Grants Instance %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccessGrantsInstanceExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		_, err := tfs3control.FindAccessGrantsInstance(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccAccessGrantsInstanceConfig_basic() string {
	return `
resource "aws_s3control_access_grants_instance" "test" {}
`
}

func testAccAccessGrantsInstanceConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_s3control_access_grants_instance" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccAccessGrantsInstanceConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_s3control_access_grants_instance" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAccessGrantsInstanceConfig_identityCenter() string {
	return `
data "aws_ssoadmin_instances" "test" {}

resource "aws_s3control_access_grants_instance" "test" {
  identity_center_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`
}
