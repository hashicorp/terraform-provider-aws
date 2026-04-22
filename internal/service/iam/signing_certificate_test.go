// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMSigningCertificate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var cred awstypes.SigningCertificate

	resourceName := "aws_iam_signing_certificate.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningCertificateConfig_basic(rName, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningCertificateExists(ctx, t, resourceName, &cred),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, "aws_iam_user.test", names.AttrName),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_id"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_body"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Active"),
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

func TestAccIAMSigningCertificate_status(t *testing.T) {
	ctx := acctest.Context(t)
	var cred awstypes.SigningCertificate

	resourceName := "aws_iam_signing_certificate.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningCertificateConfig_status(rName, "Inactive", certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningCertificateExists(ctx, t, resourceName, &cred),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Inactive"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSigningCertificateConfig_status(rName, "Active", certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningCertificateExists(ctx, t, resourceName, &cred),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Active"),
				),
			},
			{
				Config: testAccSigningCertificateConfig_status(rName, "Inactive", certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningCertificateExists(ctx, t, resourceName, &cred),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Inactive"),
				),
			},
		},
	})
}

func TestAccIAMSigningCertificate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var cred awstypes.SigningCertificate
	resourceName := "aws_iam_signing_certificate.test"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningCertificateConfig_basic(rName, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningCertificateExists(ctx, t, resourceName, &cred),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceSigningCertificate(), resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceSigningCertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSigningCertificateExists(ctx context.Context, t *testing.T, n string, v *awstypes.SigningCertificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		output, err := tfiam.FindSigningCertificateByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserName], rs.Primary.Attributes["certificate_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSigningCertificateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_signing_certificate" {
				continue
			}

			output, err := tfiam.FindSigningCertificateByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserName], rs.Primary.Attributes["certificate_id"])

			if retry.NotFound(err) {
				continue
			}

			if output != nil {
				return fmt.Errorf("IAM Signing Certificate (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccSigningCertificateConfig_basic(rName, cert string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_signing_certificate" "test" {
  certificate_body = "%[2]s"
  user_name        = aws_iam_user.test.name
}
`, rName, acctest.TLSPEMEscapeNewlines(cert))
}

func testAccSigningCertificateConfig_status(rName, status, cert string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_signing_certificate" "test" {
  certificate_body = "%[3]s"
  user_name        = aws_iam_user.test.name
  status           = %[2]q
}
`, rName, status, acctest.TLSPEMEscapeNewlines(cert))
}
