// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workspacesweb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfworkspacesweb "github.com/hashicorp/terraform-provider-aws/internal/service/workspacesweb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkSpacesWebTrustStore_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var trustStore awstypes.TrustStore
	resourceName := "aws_workspacesweb_trust_store.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, t, resourceName, &trustStore),
					resource.TestCheckResourceAttr(resourceName, "certificate.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "trust_store_arn", "workspaces-web", regexache.MustCompile(`trustStore/.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "certificate.0.body", "aws_acmpca_certificate.test1", names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.issuer"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.not_valid_after"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.not_valid_before"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.subject"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.thumbprint"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "trust_store_arn"),
				ImportStateVerifyIdentifierAttribute: "trust_store_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebTrustStore_multipleCerts(t *testing.T) {
	ctx := acctest.Context(t)
	var trustStore awstypes.TrustStore
	resourceName := "aws_workspacesweb_trust_store.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_multipleCerts(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, t, resourceName, &trustStore),
					resource.TestCheckResourceAttr(resourceName, "certificate.#", "2"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "trust_store_arn", "workspaces-web", regexache.MustCompile(`trustStore/.+$`)),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "certificate.*.body", "aws_acmpca_certificate.test1", names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.issuer"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.not_valid_after"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.not_valid_before"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.subject"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.thumbprint"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "certificate.*.body", "aws_acmpca_certificate.test2", names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.1.issuer"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.1.not_valid_after"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.1.not_valid_before"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.1.subject"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.1.thumbprint"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "trust_store_arn"),
				ImportStateVerifyIdentifierAttribute: "trust_store_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebTrustStore_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var trustStore awstypes.TrustStore
	resourceName := "aws_workspacesweb_trust_store.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, t, resourceName, &trustStore),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspacesweb.ResourceTrustStore, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWorkSpacesWebTrustStore_update(t *testing.T) {
	ctx := acctest.Context(t)
	var trustStore awstypes.TrustStore
	resourceName := "aws_workspacesweb_trust_store.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, t, resourceName, &trustStore),
					resource.TestCheckResourceAttr(resourceName, "certificate.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "trust_store_arn", "workspaces-web", regexache.MustCompile(`trustStore/.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "certificate.0.body", "aws_acmpca_certificate.test1", names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.issuer"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.not_valid_after"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.not_valid_before"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.subject"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.thumbprint"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "trust_store_arn"),
				ImportStateVerifyIdentifierAttribute: "trust_store_arn",
			},
			{
				Config: testAccTrustStoreConfig_updatedAdd(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, t, resourceName, &trustStore),
					resource.TestCheckResourceAttr(resourceName, "certificate.#", "2"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "trust_store_arn", "workspaces-web", regexache.MustCompile(`trustStore/.+$`)),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "certificate.*.body", "aws_acmpca_certificate.test1", names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.issuer"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.not_valid_after"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.not_valid_before"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.subject"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.thumbprint"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "certificate.*.body", "aws_acmpca_certificate.test2", names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.1.issuer"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.1.not_valid_after"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.1.not_valid_before"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.1.subject"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.1.thumbprint"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "trust_store_arn"),
				ImportStateVerifyIdentifierAttribute: "trust_store_arn",
			},
			{
				Config: testAccTrustStoreConfig_updatedRemove(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, t, resourceName, &trustStore),
					resource.TestCheckResourceAttr(resourceName, "certificate.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "trust_store_arn", "workspaces-web", regexache.MustCompile(`trustStore/.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "certificate.0.body", "aws_acmpca_certificate.test2", names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.issuer"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.not_valid_after"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.not_valid_before"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.subject"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate.0.thumbprint"),
				),
			},
		},
	})
}

func testAccCheckTrustStoreDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_trust_store" {
				continue
			}

			_, err := tfworkspacesweb.FindTrustStoreByARN(ctx, conn, rs.Primary.Attributes["trust_store_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WorkSpaces Web Trust Store %s still exists", rs.Primary.Attributes["trust_store_arn"])
		}

		return nil
	}
}

func testAccCheckTrustStoreExists(ctx context.Context, t *testing.T, n string, v *awstypes.TrustStore) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		output, err := tfworkspacesweb.FindTrustStoreByARN(ctx, conn, rs.Primary.Attributes["trust_store_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
func testAccTrustStoreConfig_acmBase() string {
	return (`

data "aws_partition" "current" {}

resource "aws_acmpca_certificate_authority" "test" {
  type = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_2048"
    signing_algorithm = "SHA256WITHRSA"

    subject {
      common_name = "example.com"
    }
  }
}

resource "aws_acmpca_certificate" "test1" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA256WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 1
  }
}

resource "aws_acmpca_certificate" "test2" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA256WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 1
  }
}
`)
}

func testAccTrustStoreConfig_basic() string {
	return acctest.ConfigCompose(
		testAccTrustStoreConfig_acmBase(),
		`
resource "aws_workspacesweb_trust_store" "test" {
  certificate {
    body = aws_acmpca_certificate.test1.certificate
  }
}
`)
}

func testAccTrustStoreConfig_multipleCerts() string {
	return acctest.ConfigCompose(
		testAccTrustStoreConfig_acmBase(),
		`
resource "aws_workspacesweb_trust_store" "test" {
  certificate {
    body = aws_acmpca_certificate.test1.certificate
  }
  certificate {
    body = aws_acmpca_certificate.test2.certificate
  }
}
`)
}

func testAccTrustStoreConfig_updatedAdd() string {
	return acctest.ConfigCompose(
		testAccTrustStoreConfig_acmBase(),
		`
resource "aws_workspacesweb_trust_store" "test" {
  certificate {
    body = aws_acmpca_certificate.test1.certificate
  }
  certificate {
    body = aws_acmpca_certificate.test2.certificate
  }
}
`)
}

func testAccTrustStoreConfig_updatedRemove() string {
	return acctest.ConfigCompose(
		testAccTrustStoreConfig_acmBase(),
		`
resource "aws_workspacesweb_trust_store" "test" {
  certificate {
    body = aws_acmpca_certificate.test2.certificate
  }
}
`)
}
