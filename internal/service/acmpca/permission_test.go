// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfacmpca "github.com/hashicorp/terraform-provider-aws/internal/service/acmpca"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccACMPCAPermission_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var permission types.Permission
	resourceName := "aws_acmpca_permission.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_basic(commonName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &permission),
					resource.TestCheckResourceAttr(resourceName, "actions.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttr(resourceName, "actions.*", "GetCertificate"),
					resource.TestCheckTypeSetElemAttr(resourceName, "actions.*", "IssueCertificate"),
					resource.TestCheckTypeSetElemAttr(resourceName, "actions.*", "ListPermissions"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, "acm.amazonaws.com"),
					acctest.CheckResourceAttrAccountID(resourceName, "source_account"),
				),
			},
		},
	})
}

func TestAccACMPCAPermission_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var permission types.Permission
	resourceName := "aws_acmpca_permission.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_basic(commonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &permission),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfacmpca.ResourcePermission(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccACMPCAPermission_sourceAccount(t *testing.T) {
	ctx := acctest.Context(t)
	var permission types.Permission
	resourceName := "aws_acmpca_permission.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_sourceAccount(commonName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &permission),
					acctest.CheckResourceAttrAccountID(resourceName, "source_account"),
				),
			},
		},
	})
}

func testAccCheckPermissionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMPCAClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_acmpca_permission" {
				continue
			}

			_, err := tfacmpca.FindPermissionByThreePartKey(ctx, conn, rs.Primary.Attributes["certificate_authority_arn"], rs.Primary.Attributes[names.AttrPrincipal], rs.Primary.Attributes["source_account"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ACM PCA Permission %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPermissionExists(ctx context.Context, n string, v *types.Permission) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMPCAClient(ctx)

		output, err := tfacmpca.FindPermissionByThreePartKey(ctx, conn, rs.Primary.Attributes["certificate_authority_arn"], rs.Primary.Attributes[names.AttrPrincipal], rs.Primary.Attributes["source_account"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPermissionConfig_basic(commonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

resource "aws_acmpca_permission" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
  principal                 = "acm.amazonaws.com"
  actions                   = ["IssueCertificate", "GetCertificate", "ListPermissions"]
}
`, commonName)
}

func testAccPermissionConfig_sourceAccount(commonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

data "aws_caller_identity" "current" {}

resource "aws_acmpca_permission" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
  principal                 = "acm.amazonaws.com"
  actions                   = ["IssueCertificate", "GetCertificate", "ListPermissions"]
  source_account            = data.aws_caller_identity.current.account_id
}
`, commonName)
}
