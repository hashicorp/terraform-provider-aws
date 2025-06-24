// Copyright (c) HashiCorp, Inc.
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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfworkspacesweb "github.com/hashicorp/terraform-provider-aws/internal/service/workspacesweb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkSpacesWebTrustStore_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var trustStore awstypes.TrustStore
	resourceName := "aws_workspacesweb_trust_store.test"
	//caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, resourceName, &trustStore),
					resource.TestCheckResourceAttr(resourceName, "certificate_list.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "trust_store_arn", "workspaces-web", regexache.MustCompile(`trustStore/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"certificate_list"},
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "trust_store_arn"),
				ImportStateVerifyIdentifierAttribute: "trust_store_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebTrustStore_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var trustStore awstypes.TrustStore
	//caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	resourceName := "aws_workspacesweb_trust_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, resourceName, &trustStore),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfworkspacesweb.ResourceTrustStore, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWorkSpacesWebTrustStore_update(t *testing.T) {
	ctx := acctest.Context(t)
	var trustStore awstypes.TrustStore
	//caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	resourceName := "aws_workspacesweb_trust_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, resourceName, &trustStore),
					resource.TestCheckResourceAttr(resourceName, "certificate_list.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"certificate_list"},
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "trust_store_arn"),
				ImportStateVerifyIdentifierAttribute: "trust_store_arn",
			},
			{
				Config: testAccTrustStoreConfig_updatedAdd(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, resourceName, &trustStore),
					resource.TestCheckResourceAttr(resourceName, "certificate_list.#", "2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"certificate_list"},
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "trust_store_arn"),
				ImportStateVerifyIdentifierAttribute: "trust_store_arn",
			},
			{
				Config: testAccTrustStoreConfig_updatedRemove(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, resourceName, &trustStore),
					resource.TestCheckResourceAttr(resourceName, "certificate_list.#", "1"),
				),
			},
		},
	})
}

func testAccCheckTrustStoreDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_trust_store" {
				continue
			}

			_, err := tfworkspacesweb.FindTrustStoreByARN(ctx, conn, rs.Primary.Attributes["trust_store_arn"])

			if tfresource.NotFound(err) {
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

func testAccCheckTrustStoreExists(ctx context.Context, n string, v *awstypes.TrustStore) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesWebClient(ctx)

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
  signing_algorithm          = "SHA256WITHRSA"
  
  template_arn = "arn:aws:acm-pca:::template/RootCACertificate/V1"
  
  validity {
    type  = "YEARS"
    value = 1
  }
}

resource "aws_acmpca_certificate" "test2" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm          = "SHA256WITHRSA"
  
  template_arn = "arn:aws:acm-pca:::template/RootCACertificate/V1"
  
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
  certificate_list = [aws_acmpca_certificate.test1.certificate]
}
`)
}

func testAccTrustStoreConfig_updatedAdd() string {
	return acctest.ConfigCompose(
		testAccTrustStoreConfig_acmBase(),
		`
resource "aws_workspacesweb_trust_store" "test" {
  certificate_list = [aws_acmpca_certificate.test1.certificate, aws_acmpca_certificate.test2.certificate]
}
`)
}

func testAccTrustStoreConfig_updatedRemove() string {
	return acctest.ConfigCompose(
		testAccTrustStoreConfig_acmBase(),
		`
resource "aws_workspacesweb_trust_store" "test" {
  certificate_list = [aws_acmpca_certificate.test2.certificate]
}
`)
}
