// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workspacesweb_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfworkspacesweb "github.com/hashicorp/terraform-provider-aws/internal/service/workspacesweb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkSpacesWebTrustStoreAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var trustStore awstypes.TrustStore
	resourceName := "aws_workspacesweb_trust_store_association.test"
	trustStoreResourceName := "aws_workspacesweb_trust_store.test"
	portalResourceName := "aws_workspacesweb_portal.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreAssociationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustStoreAssociationExists(ctx, t, resourceName, &trustStore),
					resource.TestCheckResourceAttrPair(resourceName, "trust_store_arn", trustStoreResourceName, "trust_store_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "portal_arn", portalResourceName, "portal_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccTrustStoreAssociationImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "trust_store_arn",
			},
			{
				ResourceName: resourceName,
				RefreshState: true,
			},
			{
				Config: testAccTrustStoreAssociationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					//The following checks are for the TrustStore Resource and the PortalResource (and not for the association resource).
					resource.TestCheckResourceAttr(trustStoreResourceName, "associated_portal_arns.#", "1"),
					resource.TestCheckResourceAttrPair(trustStoreResourceName, "associated_portal_arns.0", portalResourceName, "portal_arn"),
					resource.TestCheckResourceAttrPair(portalResourceName, "trust_store_arn", trustStoreResourceName, "trust_store_arn"),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebTrustStoreAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var trustStore awstypes.TrustStore
	resourceName := "aws_workspacesweb_trust_store_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreAssociationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustStoreAssociationExists(ctx, t, resourceName, &trustStore),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspacesweb.ResourceTrustStoreAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTrustStoreAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_trust_store_association" {
				continue
			}

			trustStore, err := tfworkspacesweb.FindTrustStoreByARN(ctx, conn, rs.Primary.Attributes["trust_store_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// Check if the portal is still associated
			portalARN := rs.Primary.Attributes["portal_arn"]
			if slices.Contains(trustStore.AssociatedPortalArns, portalARN) {
				return fmt.Errorf("WorkSpaces Web Trust Store Association %s still exists", rs.Primary.Attributes["trust_store_arn"])
			}
		}

		return nil
	}
}

func testAccCheckTrustStoreAssociationExists(ctx context.Context, t *testing.T, n string, v *awstypes.TrustStore) resource.TestCheckFunc {
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

		// Check if the portal is associated
		portalARN := rs.Primary.Attributes["portal_arn"]
		if !slices.Contains(output.AssociatedPortalArns, portalARN) {
			return fmt.Errorf("Association not found")
		}

		*v = *output

		return nil
	}
}

func testAccTrustStoreAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["trust_store_arn"], rs.Primary.Attributes["portal_arn"]), nil
	}
}

func testAccTrustStoreAssociationConfig_acmBase() string {
	return `
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
  permanent_deletion_time_in_days = 7
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
`
}

func testAccTrustStoreAssociationConfig_basic() string {
	return acctest.ConfigCompose(
		testAccTrustStoreAssociationConfig_acmBase(),
		`
resource "aws_workspacesweb_portal" "test" {
  display_name = "test"
}

resource "aws_workspacesweb_trust_store" "test" {
  certificate {
    body = aws_acmpca_certificate.test1.certificate
  }
}

resource "aws_workspacesweb_trust_store_association" "test" {
  trust_store_arn = aws_workspacesweb_trust_store.test.trust_store_arn
  portal_arn      = aws_workspacesweb_portal.test.portal_arn
}
`)
}
