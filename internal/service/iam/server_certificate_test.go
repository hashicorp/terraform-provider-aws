// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMServerCertificate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.ServerCertificate
	resourceName := "aws_iam_server_certificate.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServerCertificateConfig_basic(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerCertificateExists(ctx, t, resourceName, &v1),
					acctest.CheckResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "iam", fmt.Sprintf("server-certificate/%s", rName)),
					acctest.CheckResourceAttrRFC3339(resourceName, "expiration"),
					acctest.CheckResourceAttrRFC3339(resourceName, "upload_date"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrPath, "/"),
					resource.TestCheckResourceAttr(resourceName, "certificate_body", strings.TrimSpace(certificate)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{names.AttrPrivateKey},
			},
			{
				Config: testAccServerCertificateConfig_basic(rNameUpdated, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerCertificateExists(ctx, t, resourceName, &v2),
					testAccCheckServerCertficateNotRecreated(&v1, &v2),
					acctest.CheckResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "iam", fmt.Sprintf("server-certificate/%s", rNameUpdated)),
					acctest.CheckResourceAttrRFC3339(resourceName, "expiration"),
					acctest.CheckResourceAttrRFC3339(resourceName, "upload_date"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrPath, "/"),
					resource.TestCheckResourceAttr(resourceName, "certificate_body", strings.TrimSpace(certificate)),
				),
			},
		},
	})
}

func TestAccIAMServerCertificate_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var cert awstypes.ServerCertificate
	resourceName := "aws_iam_server_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServerCertificateConfig_nameGenerated(key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerCertificateExists(ctx, t, resourceName, &cert),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, sdkid.UniqueIdPrefix),
				),
			},
		},
	})
}

func TestAccIAMServerCertificate_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3, v4 awstypes.ServerCertificate
	resourceName := "aws_iam_server_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	namePrefix := "tf-acc-test-prefix-"
	namePrefixUpdated := "tf-acc-test-prefix-updated-"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServerCertificateConfig_namePrefix(namePrefix, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerCertificateExists(ctx, t, resourceName, &v1),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, namePrefix),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, namePrefix),
				),
			},
			{
				Config: testAccServerCertificateConfig_namePrefix(namePrefixUpdated, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerCertificateExists(ctx, t, resourceName, &v2),
					testAccCheckServerCertficateNotRecreated(&v1, &v2),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, namePrefixUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, namePrefixUpdated),
				),
			},
			// Change from name prefix to name
			{
				Config: testAccServerCertificateConfig_basic(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerCertificateExists(ctx, t, resourceName, &v3),
					testAccCheckServerCertficateNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
				),
			},
			// Change back from name to name prefix
			{
				Config: testAccServerCertificateConfig_namePrefix(namePrefix, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerCertificateExists(ctx, t, resourceName, &v4),
					testAccCheckServerCertficateNotRecreated(&v4, &v4),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, namePrefix),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, namePrefix),
				),
			},
		},
	})
}

func TestAccIAMServerCertificate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var cert awstypes.ServerCertificate
	resourceName := "aws_iam_server_certificate.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServerCertificateConfig_basic(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerCertificateExists(ctx, t, resourceName, &cert),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceServerCertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMServerCertificate_file(t *testing.T) {
	ctx := acctest.Context(t)
	var cert awstypes.ServerCertificate
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	unixFile := "test-fixtures/iam-ssl-unix-line-endings.pem"
	winFile := "test-fixtures/iam-ssl-windows-line-endings.pem.winfile"
	resourceName := "aws_iam_server_certificate.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServerCertificateConfig_file(rName, unixFile),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerCertificateExists(ctx, t, resourceName, &cert),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{names.AttrPrivateKey},
			},
			{
				Config: testAccServerCertificateConfig_file(rName, winFile),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerCertificateExists(ctx, t, resourceName, &cert),
				),
			},
		},
	})
}

func TestAccIAMServerCertificate_path(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 awstypes.ServerCertificate
	resourceName := "aws_iam_server_certificate.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	path := "/test/"
	pathUpdated := "/test/updated/"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServerCertificateConfig_path(rName, path, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerCertificateExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrPath, path),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{names.AttrPrivateKey},
			},
			{
				Config: testAccServerCertificateConfig_path(rName, pathUpdated, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerCertificateExists(ctx, t, resourceName, &v2),
					testAccCheckServerCertficateNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrPath, pathUpdated),
				),
			},
			// Change both name and path
			{
				Config: testAccServerCertificateConfig_path(rNameUpdated, path, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerCertificateExists(ctx, t, resourceName, &v3),
					testAccCheckServerCertficateNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrPath, path),
				),
			},
		},
	})
}

func testAccCheckServerCertificateExists(ctx context.Context, t *testing.T, n string, v *awstypes.ServerCertificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IAM Server Certificate ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		output, err := tfiam.FindServerCertificateByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckServerCertificateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_server_certificate" {
				continue
			}

			_, err := tfiam.FindServerCertificateByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM Server Certificate %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckServerCertficateNotRecreated(v1, v2 *awstypes.ServerCertificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(v1.ServerCertificateMetadata.ServerCertificateId) != aws.ToString(v2.ServerCertificateMetadata.ServerCertificateId) {
			return fmt.Errorf("IAM Server Certificate recreated")
		}
		return nil
	}
}

func testAccServerCertificateConfig_basic(rName, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccServerCertificateConfig_nameGenerated(key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_iam_server_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccServerCertificateConfig_namePrefix(namePrefix, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_iam_server_certificate" "test" {
  name_prefix      = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, namePrefix, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccServerCertificateConfig_path(rName, path, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  path             = %[2]q
  certificate_body = "%[3]s"
  private_key      = "%[4]s"
}
`, rName, path, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

// iam-ssl-unix-line-endings
func testAccServerCertificateConfig_file(rName, fName string) string {
	return fmt.Sprintf(`
resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = file(%[2]q)

  private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDKdH6BU9Q0xBVPfeX5NjCC/B2Pm3WsFGnTtRw4abkD+r4to9wD
eYUgjH2yPCyonNOA8mNiCQgDTtaLfbA8LjBYoodt7rgaTO7C0ugRtmTNK96DmYxm
f8Gs5ZS6eC3yeaFv58d1w2mow7tv0+DRk8uXwzVfaaMxoalsCtlLznmZHwIDAQAB
AoGABZj69nBu6ZaSUERW23EYHkcCOjo+Iqfd1TCouxaROv7vyytApgfyGlhIEWmA
gpjzcBlDji5Zvl2rqOesu707MOuJavZvluo+JHy/VIuU+yGUrWuO/QVCu6Jn3yns
vS7g48ConuZ962cTzRPcpPDspONBVOAhVCF33Y8PsnxV0wECQQD5RqeoqxEUupsy
QhrDui0KkYXLdT0uhrEQ69n9rvAiQoHPsiX0MswfEKnj/g9N3VwGLdgWytT0TvcI
8fDPRB4/AkEAz+qF3taX77gB69XRPQwCGWqE1fHIFMwX7QeYdEsk3iRZ0EKVcdp6
vIPCB2Cq4a4eXcaFa/bXen4yeYgyTbeNIQJBAO92dWctdoowPRiJskZmGhC1/Q6X
gH+qenyj5VSy8hInS6anH5i4F6icDGhtzmvhgx6YeaZjkTFkjiG0sb2aVWcCQQDD
WL7UwtzX/xPXB/ril5C1Xo5WESgC2ks0ielkgmGuUYsNEDInWbXtvwGjOuDyz0x6
oRYkfTSxQzabVyqkOGvhAkBtbjUxOD8wgBIjb4T6mAMokQo6PeEAZGUTyPifjJNo
detWVr2WRvgNgQvcRnNPECwfq1RtMJJpavaI3kgeaSxg
-----END RSA PRIVATE KEY-----
EOF
}
`, rName, fName)
}
