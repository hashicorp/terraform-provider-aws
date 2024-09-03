// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdmin_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"InstanceAccessControlAttributes": {
			acctest.CtBasic:      testAccInstanceAccessControlAttributes_basic,
			acctest.CtDisappears: testAccInstanceAccessControlAttributes_disappears,
			"multiple":           testAccInstanceAccessControlAttributes_multiple,
			"update":             testAccInstanceAccessControlAttributes_update,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccInstanceAccessControlAttributes_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_instance_access_control_attributes.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceAccessControlAttributesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAccessControlAttributesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceAccessControlAttributesExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "attribute.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ENABLED"),
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

func testAccInstanceAccessControlAttributes_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_instance_access_control_attributes.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionSetInlinePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAccessControlAttributesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceAccessControlAttributesExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssoadmin.ResourceAccessControlAttributes(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccInstanceAccessControlAttributes_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_instance_access_control_attributes.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceAccessControlAttributesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAccessControlAttributesConfig_multiple(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceAccessControlAttributesExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "attribute.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ENABLED"),
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

func testAccInstanceAccessControlAttributes_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_instance_access_control_attributes.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceAccessControlAttributesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAccessControlAttributesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceAccessControlAttributesExists(ctx, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceAccessControlAttributesConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceAccessControlAttributesExists(ctx, resourceName),
				),
			},
		},
	})
}

func testAccCheckInstanceAccessControlAttributesDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssoadmin_instance_access_control_attributes" {
				continue
			}

			_, err := tfssoadmin.FindInstanceAttributeControlAttributesByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSO Instance Access Control Attributes %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckInstanceAccessControlAttributesExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSO Instance Access Control Attributes ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		_, err := tfssoadmin.FindInstanceAttributeControlAttributesByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccInstanceAccessControlAttributesConfig_basic() string {
	return `
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_instance_access_control_attributes" "test" {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  attribute {
    key = "name"
    value {
      source = ["$${path:name.givenName}"]
    }
  }
}
`
}
func testAccInstanceAccessControlAttributesConfig_multiple() string {
	return `
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_instance_access_control_attributes" "test" {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  attribute {
    key = "name"
    value {
      source = ["$${path:name.givenName}"]
    }
  }
  attribute {
    key = "last"
    value {
      source = ["$${path:name.familyName}"]
    }
  }
}
`
}

func testAccInstanceAccessControlAttributesConfig_update() string {
	return `
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_instance_access_control_attributes" "test" {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  attribute {
    key = "name"
    value {
      source = ["$${path:name.familyName}"]
    }
  }
}
`
}
