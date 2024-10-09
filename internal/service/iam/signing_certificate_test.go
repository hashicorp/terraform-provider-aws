// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMSigningCertificate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var cred awstypes.SigningCertificate

	resourceName := "aws_iam_signing_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningCertificateConfig_basic(rName, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningCertificateExists(ctx, resourceName, &cred),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningCertificateConfig_status(rName, "Inactive", certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningCertificateExists(ctx, resourceName, &cred),
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
					testAccCheckSigningCertificateExists(ctx, resourceName, &cred),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Active"),
				),
			},
			{
				Config: testAccSigningCertificateConfig_status(rName, "Inactive", certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningCertificateExists(ctx, resourceName, &cred),
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

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningCertificateConfig_basic(rName, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningCertificateExists(ctx, resourceName, &cred),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceSigningCertificate(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceSigningCertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSigningCertificateExists(ctx context.Context, n string, cred *awstypes.SigningCertificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Server Cert ID is set")
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		certId, userName, err := tfiam.DecodeSigningCertificateId(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := tfiam.FindSigningCertificate(ctx, conn, userName, certId)
		if err != nil {
			return err
		}

		*cred = *output

		return nil
	}
}

func testAccCheckSigningCertificateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_signing_certificate" {
				continue
			}

			certId, userName, err := tfiam.DecodeSigningCertificateId(rs.Primary.ID)
			if err != nil {
				return err
			}

			output, err := tfiam.FindSigningCertificate(ctx, conn, userName, certId)

			if tfresource.NotFound(err) {
				continue
			}

			if output != nil {
				return fmt.Errorf("IAM Service Specific Credential (%s) still exists", rs.Primary.ID)
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
