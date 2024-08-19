// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rolesanywhere_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rolesanywhere"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrolesanywhere "github.com/hashicorp/terraform-provider-aws/internal/service/rolesanywhere"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRolesAnywhereTrustAnchor_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	caCommonName := acctest.RandomDomainName()
	resourceName := "aws_rolesanywhere_trust_anchor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RolesAnywhereServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustAnchorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustAnchorConfig_basic(rName, caCommonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustAnchorExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_data.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "source.0.source_data.0.acm_pca_arn", "aws_acmpca_certificate_authority.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_type", "AWS_ACM_PCA"),
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

func TestAccRolesAnywhereTrustAnchor_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	caCommonName := acctest.RandomDomainName()
	resourceName := "aws_rolesanywhere_trust_anchor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RolesAnywhereServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustAnchorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustAnchorConfig_tags1(rName, caCommonName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustAnchorExists(ctx, resourceName),
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
				Config: testAccTrustAnchorConfig_tags2(rName, caCommonName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustAnchorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTrustAnchorConfig_tags1(rName, caCommonName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustAnchorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRolesAnywhereTrustAnchor_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	caCommonName := acctest.RandomDomainName()
	resourceName := "aws_rolesanywhere_trust_anchor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RolesAnywhereServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustAnchorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustAnchorConfig_basic(rName, caCommonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustAnchorExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrolesanywhere.ResourceTrustAnchor(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRolesAnywhereTrustAnchor_certificateBundle(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rolesanywhere_trust_anchor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RolesAnywhereServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustAnchorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustAnchorConfig_certificateBundle(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustAnchorExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_data.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_type", "CERTIFICATE_BUNDLE"),
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

func TestAccRolesAnywhereTrustAnchor_enabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rolesanywhere_trust_anchor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RolesAnywhereServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustAnchorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustAnchorConfig_enabled(t, rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustAnchorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTrustAnchorConfig_enabled(t, rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustAnchorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				Config: testAccTrustAnchorConfig_enabled(t, rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustAnchorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckTrustAnchorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RolesAnywhereClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rolesanywhere_trust_anchor" {
				continue
			}

			_, err := tfrolesanywhere.FindTrustAnchorByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RolesAnywhere Trust Anchor %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTrustAnchorExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RolesAnywhere Trust Anchor ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RolesAnywhereClient(ctx)

		_, err := tfrolesanywhere.FindTrustAnchorByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccTrustAnchorConfig_acmBase(caCommonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"
    subject {
      common_name = %[1]q
    }
  }
}

data "aws_partition" "current" {}

resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 1
  }
}

resource "aws_acmpca_certificate_authority_certificate" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
  certificate               = aws_acmpca_certificate.test.certificate
  certificate_chain         = aws_acmpca_certificate.test.certificate_chain
}
`, caCommonName)
}

func testAccTrustAnchorConfig_basic(rName, caCommonName string) string {
	return acctest.ConfigCompose(
		testAccTrustAnchorConfig_acmBase(caCommonName),
		fmt.Sprintf(`
resource "aws_rolesanywhere_trust_anchor" "test" {
  name = %[1]q
  source {
    source_data {
      acm_pca_arn = aws_acmpca_certificate_authority.test.arn
    }
    source_type = "AWS_ACM_PCA"
  }
  depends_on = [aws_acmpca_certificate_authority_certificate.test]
}
`, rName))
}

func testAccTrustAnchorConfig_tags1(rName, caCommonName, tag, value string) string {
	return acctest.ConfigCompose(
		testAccTrustAnchorConfig_acmBase(caCommonName),
		fmt.Sprintf(`
resource "aws_rolesanywhere_trust_anchor" "test" {
  name = %[1]q
  source {
    source_data {
      acm_pca_arn = aws_acmpca_certificate_authority.test.arn
    }
    source_type = "AWS_ACM_PCA"
  }
  tags = {
    %[2]q = %[3]q
  }
  depends_on = [aws_acmpca_certificate_authority_certificate.test]
}
`, rName, tag, value))
}

func testAccTrustAnchorConfig_tags2(rName, caCommonName, tag1, value1, tag2, value2 string) string {
	return acctest.ConfigCompose(
		testAccTrustAnchorConfig_acmBase(caCommonName),
		fmt.Sprintf(`
resource "aws_rolesanywhere_trust_anchor" "test" {
  name = %[1]q
  source {
    source_data {
      acm_pca_arn = aws_acmpca_certificate_authority.test.arn
    }
    source_type = "AWS_ACM_PCA"
  }
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
  depends_on = [aws_acmpca_certificate_authority_certificate.test]
}
`, rName, tag1, value1, tag2, value2))
}

func testAccTrustAnchorConfig_certificateBundle(t *testing.T, rName string) string {
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificateForRolesAnywhereTrustAnchorPEM(t, caKey)

	return fmt.Sprintf(`
resource "aws_rolesanywhere_trust_anchor" "test" {
  name = %[1]q
  source {
    source_data {
      x509_certificate_data = "%[2]s"
    }
    source_type = "CERTIFICATE_BUNDLE"
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(caCertificate))
}

func testAccTrustAnchorConfig_enabled(t *testing.T, rName string, enabled bool) string {
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificateForRolesAnywhereTrustAnchorPEM(t, caKey)

	return fmt.Sprintf(`
resource "aws_rolesanywhere_trust_anchor" "test" {
  name = %[1]q
  source {
    source_data {
      x509_certificate_data = "%[2]s"
    }
    source_type = "CERTIFICATE_BUNDLE"
  }

  enabled = %[3]t
}
`, rName, acctest.TLSPEMEscapeNewlines(caCertificate), enabled)
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	acctest.PreCheckPartitionHasService(t, names.RolesAnywhereEndpointID)

	conn := acctest.Provider.Meta().(*conns.AWSClient).RolesAnywhereClient(ctx)

	input := &rolesanywhere.ListTrustAnchorsInput{}

	_, err := conn.ListTrustAnchors(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
